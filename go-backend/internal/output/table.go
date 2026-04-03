package output

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
)

func RenderResults(results *analysis.AnalysisResults, fightName, playerName string) string {
	var sb strings.Builder

	if fightName != "" || playerName != "" {
		sb.WriteString(fmt.Sprintf("%s — %s\n", fightName, playerName))
	}

	durationSec := math.Max(float64(results.FightDurationMs)/1000.0, 1.0)
	total := float64(results.TotalHealing)

	// Sort talents by healing descending
	type entry struct {
		name   string
		amount float64
	}
	var entries []entry
	allAttributed := 0.0
	for name, amount := range results.TalentHealing {
		if amount > 0 {
			entries = append(entries, entry{name, amount})
			allAttributed += amount
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].amount > entries[j].amount })

	sb.WriteString(fmt.Sprintf("%-30s %10s %6s %8s\n", "Talent", "Healing", "%", "HPS"))
	sb.WriteString(strings.Repeat("-", 56) + "\n")

	for _, e := range entries {
		pct := 0.0
		if total > 0 {
			pct = e.amount / total * 100
		}
		hps := e.amount / durationSec
		sb.WriteString(fmt.Sprintf("%-30s %10.0f %5.1f%% %8.0f\n", e.name, e.amount, pct, hps))
	}

	sb.WriteString(strings.Repeat("-", 56) + "\n")

	// Wasted
	if results.Wasted > 0 {
		sb.WriteString(fmt.Sprintf("%-30s %10d\n", "Wasted (>50% OH)", results.Wasted))
	}

	// High health
	if results.HighHealthHealing > 0 {
		sb.WriteString(fmt.Sprintf("%-30s %10d\n", "High Health (>80% HP)", results.HighHealthHealing))
	}

	// Unattributed
	unattributed := total - allAttributed - float64(results.Wasted) - float64(results.HighHealthHealing)
	if unattributed < 0 {
		sb.WriteString(fmt.Sprintf("%-30s %10s\n", "Unattributed", "—"))
	} else {
		sb.WriteString(fmt.Sprintf("%-30s %10.0f\n", "Unattributed", unattributed))
	}

	// Overlap disclaimer
	if allAttributed > total {
		sb.WriteString("\nTalents can overlap — total may exceed 100%.\n")
	}

	return sb.String()
}
