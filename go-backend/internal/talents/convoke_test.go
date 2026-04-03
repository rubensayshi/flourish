package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)



func TestConvokeAttributesDuringChannel(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convoke),
		makeHeal(1500, 774, 10000),
		makeHeal(2000, 8936, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 10500.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeLegacySpellID(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convokeLegacy),
		makeHeal(1500, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 7000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeNoAttributionOutsideChannel(t *testing.T) {
	events := []map[string]any{
		makeHeal(500, 774, 10000),
		makeCast(1000, convoke),
		makeHeal(6000, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0)}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Convoke the Spirits"])
}

func TestConvokeBoundaryAtWindowEnd(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convoke),
		makeHeal(5000, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 7000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeCustomRatio(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convoke),
		makeHeal(1500, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0.5)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokeTagsHotAppliedDuringChannel(t *testing.T) {
	events := []map[string]any{
		makeCast(1000, convoke),
		makeApply(1200, 774, withTarget(5)),
		makeHeal(1500, 774, 3000, withTarget(5)),
		makeHeal(6000, 774, 3000, withTarget(5)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 4200.0, results.TalentHealing["Convoke the Spirits"], 1.0)
}

func TestConvokePreexistingHotNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, 774, withTarget(5)),
		makeCast(1000, convoke),
		makeHeal(1500, 774, 10000, withTarget(5)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewConvokeAttributor(0)}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Convoke the Spirits"])
}
