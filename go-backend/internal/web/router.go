package web

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	r.Get("/api/report/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		report, err := client.GetReport(code)
		if err != nil {
			http.Error(w, `{"detail":"Report not found"}`, 404)
			return
		}

		// Filter fights: only encounters
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

		// Filter druids
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

	return r
}
