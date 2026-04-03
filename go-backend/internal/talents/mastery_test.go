package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)



// --- Harmonious Blooming ---

func TestHarmoniousBloomingWithLBOnTarget(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Greater(t, results.TalentHealing["Harmonious Blooming"], 0.0)
}

func TestHarmoniousBloomingNoLBNoBonus(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Harmonious Blooming"])
}

func TestHarmoniousBloomingLBOwnHealingExcluded(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, lifebloom, 5000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Harmonious Blooming"])
}

func TestHarmoniousBloomingWithMasteryFromCombatant(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)), // 40% mastery
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	// With 40% mastery, base_stacks=2: table[2]=2.3, table[4]=3.2
	// mult_base = 1 + 0.4*2.3 = 1.92
	// mult_with = 1 + 0.4*3.2 = 2.28
	// fraction = 1 - 1.92/2.28 ≈ 0.1579
	expected := 10000.0 * (1.0 - (1.0 + 0.4*2.3)/(1.0 + 0.4*3.2))
	require.InDelta(t, expected, results.TalentHealing["Harmonious Blooming"], 1.0)
}

// --- Symbiotic Bloom Mastery ---

func TestSymbioticBloomMasteryWithBloomOnTarget(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Greater(t, results.TalentHealing["Symbiotic Bloom Mastery"], 0.0)
}

func TestSymbioticBloomMasteryNoBloomNoBonus(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Symbiotic Bloom Mastery"])
}

func TestSymbioticBloomMasteryOwnHealExcluded(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, symBloom, 5000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Symbiotic Bloom Mastery"])
}

func TestSymbioticBloomMasteryFraction(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(2, nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	// base_stacks=2: table[1]=1.7, table[2]=2.3
	// mult_base = 1 + 0.4*1.7 = 1.68
	// mult_with = 1 + 0.4*2.3 = 1.92
	// fraction = 1 - 1.68/1.92 = 0.125
	expected := 10000.0 * (1.0 - (1.0 + 0.4*1.7)/(1.0 + 0.4*2.3))
	require.InDelta(t, expected, results.TalentHealing["Symbiotic Bloom Mastery"], 1.0)
}
