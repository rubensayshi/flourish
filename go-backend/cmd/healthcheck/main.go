// healthcheck validates our health reconstruction against WCL death events.
//
// Usage: go run ./cmd/healthcheck/ <report_code> <fight_id> <player_id>
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run ./cmd/healthcheck/ <report_code> <fight_id> <player_id>")
		os.Exit(1)
	}

	code := os.Args[1]
	fightID, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.Atoi(os.Args[3])

	client := wcl.NewClient(os.Getenv("WCL_CLIENT_ID"), os.Getenv("WCL_CLIENT_SECRET"))
	cachedClient := wcl.NewCachedClient(client, "data/cache")

	report, err := cachedClient.GetReport(code)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	rawFights, _ := report["fights"].([]any)
	var startTime, endTime float64
	var fightName string
	for _, f := range rawFights {
		fm := f.(map[string]any)
		if int(fm["id"].(float64)) == fightID {
			startTime = fm["startTime"].(float64)
			endTime = fm["endTime"].(float64)
			fightName, _ = fm["name"].(string)
			break
		}
	}

	masterData, _ := report["masterData"].(map[string]any)
	rawActors, _ := masterData["actors"].([]any)
	var playerName string
	for _, a := range rawActors {
		am := a.(map[string]any)
		if int(am["id"].(float64)) == playerID {
			playerName, _ = am["name"].(string)
			break
		}
	}
	fmt.Printf("Fight: %s (%.0fs) — Player: %s (ID %d)\n\n", fightName, (endTime-startTime)/1000, playerName, playerID)

	// Fetch WCL death events for this player
	fmt.Println("Fetching death events from WCL...")
	deathQuery := `query($code: String!, $startTime: Float!, $endTime: Float!, $targetID: Int!, $fightIDs: [Int!]) {
		reportData {
			report(code: $code) {
				events(
					startTime: $startTime,
					endTime: $endTime,
					targetID: $targetID,
					fightIDs: $fightIDs,
					dataType: Deaths,
					limit: 100
				) {
					data
					nextPageTimestamp
				}
			}
		}
	}`
	data, err := client.Query(deathQuery, map[string]any{
		"code":      code,
		"startTime": startTime,
		"endTime":   endTime,
		"targetID":  playerID,
		"fightIDs":  []int{fightID},
	})
	if err != nil {
		fmt.Printf("Error fetching deaths: %v\n", err)
	}

	var deathTimestamps []int
	if data != nil {
		eventsData := data["reportData"].(map[string]any)["report"].(map[string]any)["events"].(map[string]any)
		eventsList, _ := eventsData["data"].([]any)
		fmt.Printf("WCL reports %d death events\n", len(eventsList))
		for _, e := range eventsList {
			em := e.(map[string]any)
			targetIDVal := int(em["targetID"].(float64))
			if targetIDVal != playerID {
				continue // WCL returns all deaths in the fight, filter to our player
			}
			ts := int(em["timestamp"].(float64))
			deathTimestamps = append(deathTimestamps, ts)
			killingAbility := ""
			if v, ok := em["killingAbilityGameID"]; ok {
				killingAbility = fmt.Sprintf(" (ability %v)", v)
			}
			fmt.Printf("  Death at t=%ds%s\n", (ts-int(startTime))/1000, killingAbility)
		}
	}

	// Build our reconstruction
	fmt.Println("\nBuilding health reconstruction...")
	fightEvents, err := cachedClient.GetFightEvents(code, fightID, startTime, endTime)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Processing %d fight events\n", len(fightEvents))
	ht := tracking.NewHealthTracker(fightEvents)

	// Check: at each WCL death timestamp, is our reconstruction near 0%?
	if len(deathTimestamps) > 0 {
		fmt.Println("\n=== Death timestamp validation ===")
		fmt.Printf("%-10s  %-10s  %-10s\n", "Time(s)", "Our HP%", "Match?")
		for _, ts := range deathTimestamps {
			pct := ht.GetHealthPct(playerID, ts)
			match := "YES"
			if pct > 0.05 {
				match = "NO (expected ~0%)"
			}
			fmt.Printf("%-10d  %8.1f%%   %s\n", (ts-int(startTime))/1000, pct*100, match)
		}
	}

	// Show health timeline around interesting moments
	fmt.Println("\n=== Health timeline (every 5s, first 200s) ===")
	for t := int(startTime); t < int(startTime)+200000 && t < int(endTime); t += 5000 {
		pct := ht.GetHealthPct(playerID, t)
		sec := (t - int(startTime)) / 1000
		bars := int(pct * 30)
		bar := ""
		for i := 0; i < bars; i++ {
			bar += "█"
		}
		// Mark deaths
		marker := ""
		for _, d := range deathTimestamps {
			if d >= t && d < t+5000 {
				marker = " ☠"
			}
		}
		fmt.Printf("  %4ds  %5.1f%%  %s%s\n", sec, pct*100, bar, marker)
	}
}
