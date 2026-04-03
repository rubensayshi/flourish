package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestWildSynthesisAttributesBonusPortion(t *testing.T) {
	events := []map[string]any{makeHeal(100, 422090, 13000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3000.0, results.TalentHealing["Wild Synthesis"], 50.0)
}

func TestWildstalkersPowerOnRejuv(t *testing.T) {
	events := []map[string]any{makeHeal(100, 774, 11000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildstalkersPowerAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 1000.0, results.TalentHealing["Wildstalker's Power"], 50.0)
}

func TestStaticBuffIgnoresUnrelatedSpells(t *testing.T) {
	events := []map[string]any{makeHeal(100, 999, 10000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Wild Synthesis"])
}

func TestLifetreadingOnEfflorescence(t *testing.T) {
	events := []map[string]any{makeHeal(100, 81269, 12500)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewLifetreadingAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 2500.0, results.TalentHealing["Lifetreading"], 50.0)
}

func TestStaticBuffSkipsWasted(t *testing.T) {
	events := []map[string]any{makeHeal(100, 422090, 2000, withOverheal(3000))}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Wild Synthesis"])
}
