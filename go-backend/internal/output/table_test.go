package output_test

import (
	"strings"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/output"
	"github.com/stretchr/testify/require"
)

func TestRenderResultsReturnsString(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		Wasted:          10000,
		TalentHealing:   map[string]float64{"Soul of the Forest": 15000.0, "Cultivation": 8000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "Mythic Boss", "TestDruid")
	require.Contains(t, out, "Soul of the Forest")
	require.Contains(t, out, "Cultivation")
	require.Contains(t, out, "Wasted")
}

func TestRenderOverlapDisclaimerWhenExceedingTotal(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		TalentHealing:   map[string]float64{"Talent A": 80000.0, "Talent B": 50000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "", "")
	require.Contains(t, out, "Talents can overlap")
}

func TestRenderNoDisclaimerWhenWithinTotal(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		TalentHealing:   map[string]float64{"Talent A": 30000.0, "Talent B": 20000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "", "")
	require.NotContains(t, out, "Talents can overlap")
}

func TestRenderUnattributedDashWhenNegative(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		TalentHealing:   map[string]float64{"Talent A": 80000.0, "Talent B": 50000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "", "")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Unattributed") {
			require.Contains(t, line, "—")
		}
	}
}
