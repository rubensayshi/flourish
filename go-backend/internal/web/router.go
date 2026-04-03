package web

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
)

func getIntFromAny(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func getFloatFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

func NewRouter(client wcl.Querier, cacheDir string) http.Handler {
	return NewRouterWithAuth(client, cacheDir, NewAuthState())
}

func NewRouterWithAuth(client wcl.Querier, cacheDir string, authState *AuthState) http.Handler {
	r := chi.NewRouter()
	resultCache := NewResultCache(cacheDir)
	reportLimiter := NewRateLimiter(15, time.Minute)
	analyzeLimiter := NewRateLimiter(10, time.Minute)

	MountAuthRoutes(r, authState)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	r.Get("/api/report/{code}", func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if !reportLimiter.Allow(ip) {
			http.Error(w, `{"detail":"Rate limit exceeded"}`, 429)
			return
		}

		code := chi.URLParam(r, "code")
		reqClient := wcl.Querier(client)
		if token := GetUserToken(r); token != "" {
			reqClient = wcl.NewUserClient(token)
		}
		report, err := reqClient.GetReport(code)
		if err != nil {
			http.Error(w, `{"detail":"Report not found"}`, 404)
			return
		}

		rawFights, _ := report["fights"].([]any)
		var fights []map[string]any
		for _, f := range rawFights {
			fight := f.(map[string]any)
			encounterID := getIntFromAny(fight["encounterID"])
			if encounterID > 0 {
				startTime := getFloatFromAny(fight["startTime"])
				endTime := getFloatFromAny(fight["endTime"])
				id := getIntFromAny(fight["id"])
				fights = append(fights, map[string]any{
					"id":       id,
					"name":     fight["name"],
					"kill":     fight["kill"],
					"duration": int((endTime - startTime) / 1000),
				})
			}
		}

		masterData, _ := report["masterData"].(map[string]any)
		rawActors, _ := masterData["actors"].([]any)
		var druids []map[string]any
		for _, a := range rawActors {
			actor := a.(map[string]any)
			if actor["subType"] == "Druid" {
				id := getIntFromAny(actor["id"])
				druids = append(druids, map[string]any{
					"id":     id,
					"name":   actor["name"],
					"server": actor["server"],
				})
			}
		}

		if fights == nil {
			fights = []map[string]any{}
		}
		if druids == nil {
			druids = []map[string]any{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"title":  report["title"],
			"fights": fights,
			"druids": druids,
		})
	})

	r.Get("/api/analyze/{code}/{fightID}/{playerName}", func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		fightIDStr := chi.URLParam(r, "fightID")
		playerName := chi.URLParam(r, "playerName")
		fightID, _ := strconv.Atoi(fightIDStr)

		baseStacksStr := r.URL.Query().Get("base_stacks")

		// Check cache first (doesn't count against rate limit)
		if baseStacksStr == "" {
			if cached := resultCache.Get(code, fightID, playerName); cached != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(cached)
				return
			}
		}

		ip := GetClientIP(r)
		if !analyzeLimiter.Allow(ip) {
			http.Error(w, `{"detail":"Rate limit exceeded"}`, 429)
			return
		}

		// Check anonymous limit
		userToken := GetUserToken(r)
		if userToken == "" && !authState.CheckAnonLimit(ip) {
			w.WriteHeader(403)
			json.NewEncoder(w).Encode(map[string]any{
				"detail": "You've used your 2 free analyses. Log in with WarcraftLogs to continue. " +
					"This helps us stay within API rate limits — we only use your login to " +
					"analyze logs on your behalf, nothing else.",
			})
			return
		}

		// Use user-authenticated client if token provided
		reqClient := wcl.Querier(client)
		if userToken != "" {
			reqClient = wcl.NewUserClient(userToken)
		}

		report, err := reqClient.GetReport(code)
		if err != nil {
			http.Error(w, `{"detail":"Report not found"}`, 404)
			return
		}

		// Find fight
		rawFights, _ := report["fights"].([]any)
		var selectedFight map[string]any
		for _, f := range rawFights {
			fight := f.(map[string]any)
			if getIntFromAny(fight["id"]) == fightID {
				selectedFight = fight
				break
			}
		}
		if selectedFight == nil {
			http.Error(w, `{"detail":"Fight not found"}`, 404)
			return
		}

		// Find player
		masterData, _ := report["masterData"].(map[string]any)
		rawActors, _ := masterData["actors"].([]any)
		var selectedPlayer map[string]any
		for _, a := range rawActors {
			actor := a.(map[string]any)
			name, _ := actor["name"].(string)
			subType, _ := actor["subType"].(string)
			if subType == "Druid" && strings.EqualFold(name, playerName) {
				selectedPlayer = actor
				break
			}
		}
		if selectedPlayer == nil {
			http.Error(w, `{"detail":"Player not found"}`, 404)
			return
		}

		playerID := getIntFromAny(selectedPlayer["id"])
		startTime := getFloatFromAny(selectedFight["startTime"])
		endTime := getFloatFromAny(selectedFight["endTime"])

		// Record anonymous usage
		if userToken == "" {
			authState.RecordAnonUsage(ip)
		}

		events, err := reqClient.GetEvents(code, fightID, playerID, startTime, endTime)
		if err != nil {
			http.Error(w, `{"detail":"Error fetching events"}`, 500)
			return
		}

		regrowthFilter := talents.RegrowthDamageTakenFilter
		damageTaken, _ := reqClient.GetDamageTaken(code, fightID, playerID, startTime, endTime, regrowthFilter)

		// Load config
		configPath := "config/talents.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = "../config/talents.yaml"
		}
		config, err := models.LoadConfig(configPath)
		if err != nil {
			config = models.DefaultConfig()
		}

		if baseStacksStr != "" {
			if bs, err := strconv.Atoi(baseStacksStr); err == nil {
				if bs < 1 {
					bs = 1
				}
				if bs > 5 {
					bs = 5
				}
				config.Mastery.BaseStacks = bs
			}
		}

		// Build pet ID sets
		petIDs := make(map[int]bool)
		playerPetIDs := make(map[int]bool)
		for _, a := range rawActors {
			actor := a.(map[string]any)
			petOwner := getIntFromAny(actor["petOwner"])
			if petOwner > 0 {
				actorID := getIntFromAny(actor["id"])
				petIDs[actorID] = true
				if petOwner == playerID {
					playerPetIDs[actorID] = true
				}
			}
		}

		attributors := BuildAttributors(config, damageTaken)
		pipeline := analysis.NewPipeline(attributors, petIDs, playerPetIDs)

		healthThreshold := 1.0
			if ht := r.URL.Query().Get("healthThreshold"); ht != "" {
				if v, err := strconv.ParseFloat(ht, 64); err == nil && v >= 0 && v <= 1 {
					healthThreshold = v
				}
			} else if r.URL.Query().Get("trackHealth") == "true" {
				healthThreshold = 0.8
			}
			if healthThreshold < 1.0 {
				fightEvents, err := reqClient.GetFightEvents(code, fightID, startTime, endTime)
				if err == nil {
					pipeline.HealthTracker = tracking.NewHealthTracker(fightEvents)
					pipeline.HighHealthThreshold = healthThreshold
				}
			}

		results := pipeline.Run(events)

		// Format response
		durationSec := math.Max(float64(results.FightDurationMs)/1000.0, 1.0)
		total := float64(results.TotalHealing)

		talentEntry := func(name string, amount float64) map[string]any {
			entry := map[string]any{
				"name":       name,
				"attributed": int(math.Round(amount)),
				"pct":        0.0,
				"hps":        int(math.Round(amount / durationSec)),
			}
			if total > 0 {
				entry["pct"] = math.Round(amount/total*1000) / 10
			}
			if rank, ok := results.TalentRanks[name]; ok {
				entry["rank"] = rank
			}
			return entry
		}

		type nameAmount struct {
			name   string
			amount float64
		}

		var nonHero []nameAmount
		heroGroups := map[string][]nameAmount{}

		for name, amount := range results.TalentHealing {
			if amount <= 0 {
				continue
			}
			tree := heroTreeFor(name)
			if tree != "" {
				heroGroups[tree] = append(heroGroups[tree], nameAmount{name, amount})
			} else {
				nonHero = append(nonHero, nameAmount{name, amount})
			}
		}

		sort.Slice(nonHero, func(i, j int) bool { return nonHero[i].amount > nonHero[j].amount })

		talentsList := make([]map[string]any, 0, len(nonHero))
		for _, t := range nonHero {
			talentsList = append(talentsList, talentEntry(t.name, t.amount))
		}

		type treeEntry struct {
			name    string
			total   float64
			entries []nameAmount
		}
		var trees []treeEntry
		for treeName, entries := range heroGroups {
			treeTotal := 0.0
			for _, e := range entries {
				treeTotal += e.amount
			}
			sort.Slice(entries, func(i, j int) bool { return entries[i].amount > entries[j].amount })
			trees = append(trees, treeEntry{treeName, treeTotal, entries})
		}
		sort.Slice(trees, func(i, j int) bool { return trees[i].total > trees[j].total })

		heroTreesList := make([]map[string]any, 0, len(trees))
		for _, tree := range trees {
			treeTalents := make([]map[string]any, 0, len(tree.entries))
			for _, e := range tree.entries {
				treeTalents = append(treeTalents, talentEntry(e.name, e.amount))
			}
			entry := map[string]any{
				"name":       tree.name,
				"attributed": int(math.Round(tree.total)),
				"pct":        0.0,
				"hps":        int(math.Round(tree.total / durationSec)),
				"talents":    treeTalents,
			}
			if total > 0 {
				entry["pct"] = math.Round(tree.total/total*1000) / 10
			}
			heroTreesList = append(heroTreesList, entry)
		}

		allAttributed := 0.0
		for _, t := range nonHero {
			allAttributed += t.amount
		}
		for _, g := range heroGroups {
			for _, e := range g {
				allAttributed += e.amount
			}
		}
		unattributed := total - math.Round(allAttributed) - float64(results.Wasted) - float64(results.HighHealthHealing)
		if unattributed < 0 {
			unattributed = 0
		}

		response := map[string]any{
			"fight_name":    selectedFight["name"],
			"player_name":   selectedPlayer["name"],
			"total_healing": results.TotalHealing,
			"duration_sec":  int(math.Round(durationSec)),
			"talents":       talentsList,
			"hero_trees":    heroTreesList,
			"wasted":        results.Wasted,
			"high_health":   results.HighHealthHealing,
			"unattributed":  int(unattributed),
		}

		resultCache.Set(code, fightID, playerName, response)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return r
}
