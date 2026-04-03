package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/output"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: flourish <report_code> [--fight ID] [--player NAME]")
		os.Exit(1)
	}

	reportCode := os.Args[1]
	var fightID int
	var playerName string

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--fight":
			if i+1 < len(os.Args) {
				fightID, _ = strconv.Atoi(os.Args[i+1])
				i++
			}
		case "--player":
			if i+1 < len(os.Args) {
				playerName = os.Args[i+1]
				i++
			}
		}
	}

	clientID := os.Getenv("WCL_CLIENT_ID")
	clientSecret := os.Getenv("WCL_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Println("Error: WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set")
		os.Exit(1)
	}

	// Create cached client
	inner := wcl.NewClient(clientID, clientSecret)
	cacheDir := "data/cache"
	client := wcl.NewCachedClient(inner, cacheDir)

	// Fetch report
	report, err := client.GetReport(reportCode)
	if err != nil {
		fmt.Printf("Error fetching report: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Report: %s\n", report["title"])

	// Get fights and actors
	rawFights, _ := report["fights"].([]any)
	masterData, _ := report["masterData"].(map[string]any)
	rawActors, _ := masterData["actors"].([]any)

	// Filter boss fights
	type fightInfo struct {
		id        int
		name      string
		kill      bool
		startTime float64
		endTime   float64
	}
	var fights []fightInfo
	for _, f := range rawFights {
		fm := f.(map[string]any)
		eid := toInt(fm["encounterID"])
		if eid > 0 {
			fights = append(fights, fightInfo{
				id:        toInt(fm["id"]),
				name:      fmt.Sprint(fm["name"]),
				kill:      toBool(fm["kill"]),
				startTime: toFloat(fm["startTime"]),
				endTime:   toFloat(fm["endTime"]),
			})
		}
	}

	// Select fight
	if fightID == 0 {
		fmt.Println("\nFights:")
		for _, f := range fights {
			status := "Kill"
			if !f.kill {
				status = "Wipe"
			}
			dur := (f.endTime - f.startTime) / 1000
			fmt.Printf("  %3d: %s (%s, %.0fs)\n", f.id, f.name, status, dur)
		}
		fmt.Print("Select fight ID: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		fightID, _ = strconv.Atoi(strings.TrimSpace(line))
	}

	var selectedFight *fightInfo
	for i := range fights {
		if fights[i].id == fightID {
			selectedFight = &fights[i]
			break
		}
	}
	if selectedFight == nil {
		fmt.Println("Fight not found")
		os.Exit(1)
	}

	// Find druids
	type actorInfo struct {
		id       int
		name     string
		subType  string
		server   string
		petOwner int
	}
	var allActors []actorInfo
	for _, a := range rawActors {
		am := a.(map[string]any)
		allActors = append(allActors, actorInfo{
			id:       toInt(am["id"]),
			name:     fmt.Sprint(am["name"]),
			subType:  fmt.Sprint(am["subType"]),
			server:   fmt.Sprint(am["server"]),
			petOwner: toInt(am["petOwner"]),
		})
	}

	var druids []actorInfo
	for _, a := range allActors {
		if a.subType == "Druid" {
			druids = append(druids, a)
		}
	}

	// Select player
	var selectedPlayer *actorInfo
	if playerName == "" {
		if len(druids) == 1 {
			selectedPlayer = &druids[0]
			fmt.Printf("Auto-selected: %s\n", selectedPlayer.name)
		} else {
			fmt.Println("\nResto Druids:")
			for _, d := range druids {
				fmt.Printf("  %3d: %s (%s)\n", d.id, d.name, d.server)
			}
			fmt.Print("Select player ID: ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			pid, _ := strconv.Atoi(strings.TrimSpace(line))
			for i := range druids {
				if druids[i].id == pid {
					selectedPlayer = &druids[i]
					break
				}
			}
		}
	} else {
		for i := range druids {
			if strings.EqualFold(druids[i].name, playerName) {
				selectedPlayer = &druids[i]
				break
			}
		}
	}
	if selectedPlayer == nil {
		fmt.Println("Player not found")
		os.Exit(1)
	}

	// Fetch events
	fmt.Printf("\nFetching events for %s in %s...\n", selectedPlayer.name, selectedFight.name)
	events, err := client.GetEvents(reportCode, selectedFight.id, selectedPlayer.id, selectedFight.startTime, selectedFight.endTime)
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched %d events\n", len(events))

	// Fetch damage taken with regrowth
	regrowthFilter := `IN RANGE FROM (type = "applybuff" OR type = "refreshbuff") AND ability.id = 8936 TO type = "removebuff" AND ability.id = 8936 GROUP BY target ON target END`
	damageTaken, _ := client.GetDamageTaken(reportCode, selectedFight.id, selectedPlayer.id, selectedFight.startTime, selectedFight.endTime, regrowthFilter)

	// Load config
	configPath := "config/talents.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try from go-backend dir
		configPath = "../config/talents.yaml"
	}
	config, err := models.LoadConfig(configPath)
	if err != nil {
		config = &models.Config{
			Mastery: models.MasteryConfig{BaseStacks: 2, DRTable: []float64{1.0, 1.7, 2.3, 2.8, 3.2}},
			Talents: map[string]models.TalentConfig{},
		}
	}

	// Build attributors
	attributors := buildAttributors(config, damageTaken)

	// Build pet ID sets
	petIDs := make(map[int]bool)
	playerPetIDs := make(map[int]bool)
	for _, a := range allActors {
		if a.petOwner > 0 {
			petIDs[a.id] = true
			if a.petOwner == selectedPlayer.id {
				playerPetIDs[a.id] = true
			}
		}
	}

	// Run pipeline
	pipeline := analysis.NewPipeline(attributors, petIDs, playerPetIDs)
	results := pipeline.Run(events)

	// Output
	fmt.Println()
	fmt.Print(output.RenderResults(results, selectedFight.name, selectedPlayer.name))
}

func buildAttributors(config *models.Config, damageTaken int) []talents.TalentAttributor {
	convokeCfg := config.Talents["convoke_the_spirits"]
	convokeRatio := 0.7
	if convokeCfg.Multiplier != nil {
		convokeRatio = *convokeCfg.Multiplier
	}

	sotf := talents.NewSoulOfTheForestAttributor()
	gg := talents.NewGroveGuardiansAttributor()
	smCd := talents.NewSmCooldownReductionAttributor([]talents.TalentAttributor{sotf, gg})
	wgCd := talents.NewWgCooldownReductionAttributor([]talents.TalentAttributor{gg}, false)

	baseStacks := config.Mastery.BaseStacks
	drTable := config.Mastery.DRTable

	all := []talents.TalentAttributor{
		sotf,
		talents.NewEverbloomSplashAttributor(),
		talents.NewBloomingFrenzyAttributor(),
		gg,
		talents.NewDreamSurgeAttributor(),
		talents.NewEfflorescenceAttributor(),
		talents.NewVerdancyAttributor(),
		talents.NewRampantGrowthAttributor(),
		talents.NewFlourishAttributor(),
		talents.NewCultivationAttributor(),
		talents.NewYserasGiftAttributor(),
		talents.NewEmbraceOfTheDreamAttributor(),
		talents.NewImprovedSwiftmendAttributor(),
		talents.NewUnstoppableGrowthAttributor(),
		talents.NewLivelinessAttributor(),
		talents.NewRegenesisAttributor(),
		talents.NewThrivingVegetationAttributor(),
		talents.NewWildSynthesisAttributor(),
		talents.NewGrovesInspirationAttributor(),
		talents.NewCenariusMightAttributor(),
		talents.NewBountifulBloomAttributor(),
		talents.NewHarmonyOfTheGroveAttributor(),
		talents.NewPowerOfNatureAttributor(),
		talents.NewSpiritOfTheThicketAttributor(),
		talents.NewSylvanBeckoningAttributor(),
		talents.NewWildstalkersPowerAttributor(),
		talents.NewPatientCustodianAttributor(),
		talents.NewLifetreadingAttributor(),
		talents.NewTreeOfLifeAttributor(),
		talents.NewConvokeAttributor(convokeRatio),
		talents.NewImprovedWildGrowthAttributor(),
		talents.NewReforestationAttributor(),
		talents.NewVigorousCreepersAttributor(),
		talents.NewImplantAttributor(),
		talents.NewRootNetworkAttributor(),
		talents.NewStrategicInfusionAttributor(),
		talents.NewBurstingGrowthAttributor(),
		talents.NewThrivingGrowthAttributor(),
		talents.NewHarmoniousBloomingAttributor(baseStacks, drTable),
		talents.NewSymbioticBloomMasteryAttributor(baseStacks, drTable),
		talents.NewIntensityAttributor(),
		talents.NewNaturesBountyAttributor(),
		talents.NewRegenerativeHeartwoodAttributor(),
		talents.NewAbundanceAttributor(),
		talents.NewPhotosynthesisAttributor(),
		talents.NewNurturingDormancyAttributor(),
		talents.NewProtectiveGrowthAttributor(damageTaken),
		smCd,
		wgCd,
	}

	// Filter by config skip
	var active []talents.TalentAttributor
	for _, a := range all {
		key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(a.Name(), " ", "_"), "'", ""))
		if cfg, ok := config.Talents[key]; ok && cfg.Skip {
			continue
		}
		active = append(active, a)
	}
	return active
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

func toBool(v any) bool {
	b, _ := v.(bool)
	return b
}
