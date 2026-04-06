package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

const tolBuff = 33891

func TestTolRejuvBuff(t *testing.T) {
	events := []map[string]any{
		makeApply(100, tolBuff, withTarget(1)),
		makeHeal(200, 774, 15000),
		makeRemove(500, tolBuff, withTarget(1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Incarnation: Tree of Life"], 1.0)
}

func TestTolGerminationRejuvBuff(t *testing.T) {
	events := []map[string]any{
		makeApply(100, tolBuff, withTarget(1)),
		makeHeal(200, 155777, 15000),
		makeRemove(500, tolBuff, withTarget(1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Incarnation: Tree of Life"], 1.0)
}

func TestTolOtherSpellBuff(t *testing.T) {
	events := []map[string]any{
		makeApply(100, tolBuff, withTarget(1)),
		makeHeal(200, 8936, 11000),
		makeRemove(500, tolBuff, withTarget(1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 1000.0, results.TalentHealing["Incarnation: Tree of Life"], 1.0)
}

func TestTolNoAttributionOutside(t *testing.T) {
	events := []map[string]any{
		makeHeal(50, 774, 10000),
		makeApply(100, tolBuff, withTarget(1)),
		makeRemove(200, tolBuff, withTarget(1)),
		makeHeal(300, 774, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Incarnation: Tree of Life"])
}

func TestTolWgBaseBuff(t *testing.T) {
	events := []map[string]any{
		makeApply(100, tolBuff, withTarget(1)),
		makeHeal(200, 48438, 10000, withTarget(2)),
		makeRemove(500, tolBuff, withTarget(1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 909.09, results.TalentHealing["Incarnation: Tree of Life"], 10.0)
}

func TestTolNoAttributionOnUnrelatedEventAfterDeactivation(t *testing.T) {
	events := []map[string]any{
		makeApply(100, tolBuff, withTarget(1)),
		makeHeal(200, 48438, 10000, withTarget(2)),
		makeRemove(500, tolBuff, withTarget(1)),
		makeHeal(600, 8936, 5000, withTarget(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 909.09, results.TalentHealing["Incarnation: Tree of Life"], 10.0)
}

func TestTolWgBufferFlushedAtFightEnd(t *testing.T) {
	events := []map[string]any{
		makeApply(100, tolBuff, withTarget(1)),
		makeHeal(200, 48438, 10000, withTarget(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewTreeOfLifeAttributor(nil)}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 909.09, results.TalentHealing["Incarnation: Tree of Life"], 10.0)
}
