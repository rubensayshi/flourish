package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

const wildGrowth = 48438

func TestIwgAttributesExtraTargetShare(t *testing.T) {
	events := []map[string]any{makeHeal(100, wildGrowth, 7000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImprovedWildGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 2000.0, results.TalentHealing["Improved Wild Growth"], 1.0)
}

func TestIwgIgnoresNonWg(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, 774, 10000),
		makeHeal(200, 8936, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImprovedWildGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Improved Wild Growth"])
}

func TestIwgMultipleTicks(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, wildGrowth, 7000, withTarget(2)),
		makeHeal(100, wildGrowth, 7000, withTarget(3)),
		makeHeal(100, wildGrowth, 7000, withTarget(4)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImprovedWildGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 6000.0, results.TalentHealing["Improved Wild Growth"], 1.0)
}

func TestIwgSkipsDuringTol(t *testing.T) {
	events := []map[string]any{
		makeApply(50, tolBuff, withTarget(1)),
		makeHeal(100, wildGrowth, 7000),
		makeHeal(200, wildGrowth, 7000),
		makeRemove(300, tolBuff, withTarget(1)),
		makeHeal(400, wildGrowth, 7000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImprovedWildGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 2000.0, results.TalentHealing["Improved Wild Growth"], 1.0)
}
