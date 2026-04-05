package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestConvokeDirectHealDuringChannel(t *testing.T) {
	// Swiftmend (non-skippable) during channel = full attribution
	events := []map[string]any{
		makeCast(1000, convoke),
		makeHeal(1500, swiftmend, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeLegacySpellID(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convokeLegacy),
		makeHeal(1500, swiftmend, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeNoAttributionOutsideChannel(t *testing.T) {
	events := []map[string]any{
		makeHeal(500, swiftmend, 10000),
		makeCast(1000, convoke),
		makeHeal(6000, swiftmend, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Convoke the Spirits"])
}

func TestConvokeBoundaryAtWindowEnd(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convoke),
		makeHeal(5000, swiftmend, 10000), // exactly at 4s boundary
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeTagsHotAppliedDuringChannel(t *testing.T) {
	// WG apply (not skippable) during channel, ticks attributed at 1.0
	events := []map[string]any{
		makeCast(1000, convoke),
		makeApply(1200, wildGrowth, withTarget(5)),
		makeHeal(1500, wildGrowth, 3000, withTarget(5)),
		makeHeal(6000, wildGrowth, 3000, withTarget(5)), // after channel, still tagged
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 6000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokePreexistingHotNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, rejuv, withTarget(5)),
		makeCast(1000, convoke),
		makeHeal(1500, rejuv, 10000, withTarget(5)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Convoke the Spirits"])
}

func TestConvokeSkipsFirst3RejuvRegrowth(t *testing.T) {
	// First 3 rejuv/regrowth applies are skipped (opportunity cost)
	events := []map[string]any{
		makeCast(1000, convoke),
		makeApply(1100, rejuv, withTarget(5)),  // skip 1
		makeApply(1200, rejuv, withTarget(6)),  // skip 2
		makeApply(1300, regrowth, withTarget(7)), // skip 3
		makeApply(1400, rejuv, withTarget(8)),  // 4th — attributed
		makeHeal(1500, rejuv, 1000, withTarget(5)), // skipped HoT
		makeHeal(1600, rejuv, 1000, withTarget(8)), // attributed HoT
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 1000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeSkipCountResetsPerChannel(t *testing.T) {
	events := []map[string]any{
		// First convoke
		makeCast(1000, convoke),
		makeApply(1100, rejuv, withTarget(5)),  // skip 1
		makeApply(1200, rejuv, withTarget(6)),  // skip 2
		makeApply(1300, rejuv, withTarget(7)),  // skip 3
		makeApply(1400, rejuv, withTarget(8)),  // attributed
		// Second convoke
		makeCast(200000, convoke),
		makeApply(200100, rejuv, withTarget(9)),  // skip 1 (reset)
		makeApply(200200, rejuv, withTarget(10)), // skip 2
		makeApply(200300, rejuv, withTarget(11)), // skip 3
		makeApply(200400, rejuv, withTarget(12)), // attributed
		makeHeal(200500, rejuv, 2000, withTarget(12)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 2000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeNonSkippableSpellsAlwaysAttributed(t *testing.T) {
	// Wild Growth and Swiftmend are never skipped
	events := []map[string]any{
		makeCast(1000, convoke),
		makeApply(1100, wildGrowth, withTarget(5)),
		makeHeal(1200, swiftmend, 5000),
		makeHeal(1300, wildGrowth, 3000, withTarget(5)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 8000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}
