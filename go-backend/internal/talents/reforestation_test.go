package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestReforestationTriggersAfter4thSwiftmend(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeCast(400, swiftmend),
		makeHeal(500, 8936, 11000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 1000.0, results.TalentHealing["Reforestation"], 1.0)
}

func TestReforestationRejuvGets50Pct(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeCast(400, swiftmend),
		makeHeal(500, 774, 15000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Reforestation"], 1.0)
}

func TestReforestationNoTriggerBefore4th(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeHeal(400, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Reforestation"])
}

func TestReforestationExpiresAfter10Sec(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeCast(400, swiftmend),
		makeHeal(10500, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Reforestation"])
}

func TestReforestationTriggersAgainAt8th(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeCast(400, swiftmend),
		makeCast(20000, swiftmend),
		makeCast(20100, swiftmend),
		makeCast(20200, swiftmend),
		makeCast(20300, swiftmend),
		makeHeal(20400, 8936, 11000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 1000.0, results.TalentHealing["Reforestation"], 1.0)
}

func TestReforestationNoTriggerDuringRealTol(t *testing.T) {
	events := []map[string]any{
		makeApply(50, tolBuff, withTarget(1)),
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeCast(400, swiftmend),
		makeRemove(450, tolBuff, withTarget(1)),
		makeHeal(500, 8936, 11000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Reforestation"])
}

func TestReforestationNoAttributionDuringRealTol(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeCast(200, swiftmend),
		makeCast(300, swiftmend),
		makeCast(400, swiftmend),
		makeApply(500, tolBuff, withTarget(1)),
		makeHeal(600, 8936, 11000),
		makeRemove(700, tolBuff, withTarget(1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewReforestationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Reforestation"])
}
