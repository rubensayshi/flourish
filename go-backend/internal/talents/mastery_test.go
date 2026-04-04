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
	a := talents.NewHarmoniousBloomingAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Greater(t, results.TalentHealing["Harmonious Blooming"], 0.0)
}

func TestHarmoniousBloomingNoLBNoBonus(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Harmonious Blooming"])
}

func TestHarmoniousBloomingLBOwnHealingExcluded(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, lifebloom, 5000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Harmonious Blooming"])
}

func TestHarmoniousBloomingDynamicStacks(t *testing.T) {
	// Target has LB + Rejuv + WG = 3 mastery HoTs
	// HB: without = 3, with = 5 → table[3]=2.3 to table[5]=3.2
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, lifebloom, withTarget(target)),
		makeApply(100, rejuv, withTarget(target)),
		makeApply(100, wildGrowth, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (1.0 - (1.0+0.4*2.3)/(1.0+0.4*3.2))
	require.InDelta(t, expected, results.TalentHealing["Harmonious Blooming"], 1.0)
}

func TestHarmoniousBloomingOnlyLB(t *testing.T) {
	// Target has only LB = 1 mastery HoT
	// HB: without = 1, with = 3 → table[1]=1.0 to table[3]=2.3
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewHarmoniousBloomingAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (1.0 - (1.0+0.4*1.0)/(1.0+0.4*2.3))
	require.InDelta(t, expected, results.TalentHealing["Harmonious Blooming"], 1.0)
}

// --- Symbiotic Bloom Mastery ---

func TestSymbioticBloomMasteryWithBloomOnTarget(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Greater(t, results.TalentHealing["Symbiotic Bloom Mastery"], 0.0)
}

func TestSymbioticBloomMasteryNoBloomNoBonus(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Symbiotic Bloom Mastery"])
}

func TestSymbioticBloomMasteryOwnHealExcluded(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, symBloom, 5000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Symbiotic Bloom Mastery"])
}

func TestSymbioticBloomMasteryDynamicStacks(t *testing.T) {
	// Target has SB + Rejuv = 2 mastery HoTs, no LB
	// Without SB: 1, with SB: 2 → table[1]=1.0 to table[2]=1.7
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, symBloom, withTarget(target)),
		makeApply(100, rejuv, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (1.0 - (1.0+0.4*1.0)/(1.0+0.4*1.7))
	require.InDelta(t, expected, results.TalentHealing["Symbiotic Bloom Mastery"], 1.0)
}

func TestSymbioticBloomMasteryWithHBAndLB(t *testing.T) {
	// Target has SB + LB + Rejuv = 3 mastery HoTs
	// HB active → +2 virtual = 5 effective
	// Without SB: 4, with SB: 5 → table[4]=2.8 to table[5]=3.2
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, symBloom, withTarget(target)),
		makeApply(100, lifebloom, withTarget(target)),
		makeApply(100, rejuv, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	a.SetHBActive(true)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (1.0 - (1.0+0.4*2.8)/(1.0+0.4*3.2))
	require.InDelta(t, expected, results.TalentHealing["Symbiotic Bloom Mastery"], 1.0)
}

func TestSymbioticBloomMasteryWithHBAndLBFewHots(t *testing.T) {
	// Target has only SB + LB = 2 mastery HoTs
	// HB active → +2 virtual = 4 effective stacks
	// Without SB: 3, with SB: 4 → table[3]=2.3 to table[4]=2.8
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, symBloom, withTarget(target)),
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	a.SetHBActive(true)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (1.0 - (1.0+0.4*2.3)/(1.0+0.4*2.8))
	require.InDelta(t, expected, results.TalentHealing["Symbiotic Bloom Mastery"], 1.0)
}

func TestSymbioticBloomMasteryWithoutHBIgnoresLB(t *testing.T) {
	// Same scenario but HB not active → LB doesn't add virtual stacks
	// Target has SB + LB = 2 mastery HoTs
	// Without SB: 1, with SB: 2 → table[1]=1.0 to table[2]=1.7
	events := []map[string]any{
		makeCombatantInfo(0, withMastery(40.0)),
		makeApply(100, symBloom, withTarget(target)),
		makeApply(100, lifebloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	a := talents.NewSymbioticBloomMasteryAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (1.0 - (1.0+0.4*1.0)/(1.0+0.4*1.7))
	require.InDelta(t, expected, results.TalentHealing["Symbiotic Bloom Mastery"], 1.0)
}
