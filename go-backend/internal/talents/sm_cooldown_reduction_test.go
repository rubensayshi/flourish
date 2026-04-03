package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

const (
	dryadTranq = 1264659
	ggNourish  = 422090
)

func smCombatantInfo(ts int, nodeIDs []int, talentIDs []int) map[string]any {
	tree := []any{}
	added := map[int]bool{}
	for _, nid := range nodeIDs {
		tree = append(tree, map[string]any{"nodeID": nid, "id": nid})
		added[nid] = true
	}
	for _, tid := range talentIDs {
		if !added[tid] {
			tree = append(tree, map[string]any{"nodeID": 0, "id": tid})
		}
	}
	return map[string]any{
		"timestamp": ts, "type": "combatantinfo", "sourceID": 1,
		"talentTree": tree,
		"critSpell": 0, "hasteSpell": 0, "mastery": 0, "specID": 105,
	}
}

func TestSmTracksCasts(t *testing.T) {
	attr := talents.NewSmCooldownReductionAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, nil)
	events := []map[string]any{
		makeCast(0, swiftmend),
		makeCast(12000, swiftmend),
		makeCast(24000, swiftmend),
	}
	pipeline.Run(events)
	require.Equal(t, []int{0, 12000, 24000}, attr.SmCastTimestamps())
}

func TestSmTracksDryadWindowsFromPetHeals(t *testing.T) {
	attr := talents.NewSmCooldownReductionAttributor(nil)
	petIDs := map[int]bool{99: true}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, petIDs)
	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.DryadsDanceNodeID, talents.EarlySpringNodeID},
			[]int{talents.EarlySpringTalentID}),
		makeHeal(1000, dryadTranq, 500, withSource(99)),
		makeHeal(1500, dryadTranq, 500, withSource(99)),
		makeHeal(2000, dryadTranq, 500, withSource(99)),
		makeHeal(10000, dryadTranq, 500, withSource(99)),
		makeHeal(10500, dryadTranq, 500, withSource(99)),
	}
	pipeline.Run(events)
	windows := attr.DryadWindows()
	require.Len(t, windows, 2)
	require.Equal(t, [2]int{1000, 2000}, windows[0])
	require.Equal(t, [2]int{10000, 10500}, windows[1])
}

func TestSmDryadWindowClosesInFinalize(t *testing.T) {
	attr := talents.NewSmCooldownReductionAttributor(nil)
	petIDs := map[int]bool{99: true}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, petIDs)
	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.DryadsDanceNodeID, talents.EarlySpringNodeID},
			[]int{talents.EarlySpringTalentID}),
		makeHeal(1000, dryadTranq, 500, withSource(99)),
		makeHeal(1500, dryadTranq, 500, withSource(99)),
	}
	pipeline.Run(events)
	windows := attr.DryadWindows()
	require.Len(t, windows, 1)
	require.Equal(t, [2]int{1000, 1500}, windows[0])
}

func TestSmIgnoresPlayerSourceHeals(t *testing.T) {
	attr := talents.NewSmCooldownReductionAttributor(nil)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, nil)
	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.DryadsDanceNodeID, talents.EarlySpringNodeID},
			[]int{talents.EarlySpringTalentID}),
		makeHeal(1000, dryadTranq, 500, withSource(1)),
	}
	pipeline.Run(events)
	require.Len(t, attr.DryadWindows(), 0)
}

// --- compute_effective_cd tests ---

func TestEffectiveCdBaselineWithRenewingSurge(t *testing.T) {
	cd := talents.ComputeEffectiveCd(true, false, 0)
	require.InDelta(t, 12075.0, cd, 1.0)
}

func TestEffectiveCdWithEarlySpring(t *testing.T) {
	cd := talents.ComputeEffectiveCd(true, true, 0)
	require.InDelta(t, 11075.0, cd, 1.0)
}

func TestEffectiveCdWithDryadFullOverlap(t *testing.T) {
	cd := talents.ComputeEffectiveCd(true, true, 11075)
	require.InDelta(t, 8860.0, cd, 1.0)
}

func TestEffectiveCdWithDryadPartialOverlap(t *testing.T) {
	cd := talents.ComputeEffectiveCd(true, true, 5000)
	require.InDelta(t, 10075.0, cd, 1.0)
}

func TestEffectiveCdNoRenewingSurge(t *testing.T) {
	cd := talents.ComputeEffectiveCd(false, true, 0)
	require.InDelta(t, 14000.0, cd, 1.0)
}

// --- finalize tests ---

func TestSmFullAttributionOnCooldown(t *testing.T) {
	sotf := talents.NewSoulOfTheForestAttributor()
	gg := talents.NewGroveGuardiansAttributor()
	smCd := talents.NewSmCooldownReductionAttributor([]talents.TalentAttributor{sotf, gg})

	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.EarlySpringNodeID, talents.RenewingSurgeNodeID, 82055, 82043},
			[]int{talents.EarlySpringTalentID}),
		// SM 1
		makeCast(1000, swiftmend),
		makeApply(1001, sotfBuffID, withTarget(1)),
		makeCast(1002, rejuv, withTarget(3)),
		makeApply(1003, rejuv, withTarget(3)),
		makeRemove(1004, sotfBuffID, withTarget(1)),
		makeHeal(1100, rejuv, 10000, withTarget(3)),
		makeHeal(1200, ggNourish, 5000, withSource(99)),
		// SM 2
		makeCast(12000, swiftmend),
		makeApply(12001, sotfBuffID, withTarget(1)),
		makeCast(12002, rejuv, withTarget(4)),
		makeApply(12003, rejuv, withTarget(4)),
		makeRemove(12004, sotfBuffID, withTarget(1)),
		makeHeal(12100, rejuv, 10000, withTarget(4)),
		makeHeal(12200, ggNourish, 5000, withSource(99)),
		// SM 3
		makeCast(23000, swiftmend),
		makeApply(23001, sotfBuffID, withTarget(1)),
		makeCast(23002, rejuv, withTarget(5)),
		makeApply(23003, rejuv, withTarget(5)),
		makeRemove(23004, sotfBuffID, withTarget(1)),
		makeHeal(23100, rejuv, 10000, withTarget(5)),
		makeHeal(23200, ggNourish, 5000, withSource(99)),
	}
	petIDs := map[int]bool{99: true}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{sotf, gg, smCd}, nil, petIDs)
	results := pipeline.Run(events)

	ratio := 1.0 - 11075.0/12075.0
	fraction := (ratio * 2) / 3
	expected := fraction * 26250
	require.InDelta(t, expected, results.TalentHealing["Early Spring + Dryad's Dance"], expected*0.02)
}

func TestSmNoAttributionWhenNotOnCooldown(t *testing.T) {
	sotf := talents.NewSoulOfTheForestAttributor()
	smCd := talents.NewSmCooldownReductionAttributor([]talents.TalentAttributor{sotf})

	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.EarlySpringNodeID, talents.RenewingSurgeNodeID, 82055},
			[]int{talents.EarlySpringTalentID}),
		makeCast(1000, swiftmend),
		makeApply(1001, sotfBuffID, withTarget(1)),
		makeCast(1002, rejuv, withTarget(3)),
		makeApply(1003, rejuv, withTarget(3)),
		makeRemove(1004, sotfBuffID, withTarget(1)),
		makeHeal(1100, rejuv, 10000, withTarget(3)),
		// SM 2 — gap far too large
		makeCast(31000, swiftmend),
		makeApply(31001, sotfBuffID, withTarget(1)),
		makeCast(31002, rejuv, withTarget(4)),
		makeApply(31003, rejuv, withTarget(4)),
		makeRemove(31004, sotfBuffID, withTarget(1)),
		makeHeal(31100, rejuv, 10000, withTarget(4)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{sotf, smCd}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Early Spring + Dryad's Dance"])
}

func TestSmNoAttributionSingleSmCast(t *testing.T) {
	sotf := talents.NewSoulOfTheForestAttributor()
	smCd := talents.NewSmCooldownReductionAttributor([]talents.TalentAttributor{sotf})

	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.EarlySpringNodeID, talents.RenewingSurgeNodeID, 82055},
			[]int{talents.EarlySpringTalentID}),
		makeCast(1000, swiftmend),
		makeApply(1001, sotfBuffID, withTarget(1)),
		makeCast(1002, rejuv, withTarget(3)),
		makeApply(1003, rejuv, withTarget(3)),
		makeRemove(1004, sotfBuffID, withTarget(1)),
		makeHeal(1100, rejuv, 10000, withTarget(3)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{sotf, smCd}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Early Spring + Dryad's Dance"])
}

// --- WG CD reduction tests ---

func TestEffectiveWgCdWith4pcAndEarlySpring(t *testing.T) {
	cd := talents.ComputeEffectiveWgCd(true, true)
	require.InDelta(t, 7000.0, cd, 1.0)
}

func TestEffectiveWgCdWith4pcNoEarlySpring(t *testing.T) {
	cd := talents.ComputeEffectiveWgCd(false, true)
	require.InDelta(t, 8000.0, cd, 1.0)
}

func TestEffectiveWgCdNo4pcWithEarlySpring(t *testing.T) {
	cd := talents.ComputeEffectiveWgCd(true, false)
	require.InDelta(t, 9000.0, cd, 1.0)
}

func TestWgTracksCasts(t *testing.T) {
	attr := talents.NewWgCooldownReductionAttributor(nil, false)
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, nil)
	events := []map[string]any{
		makeBegincast(0, wildGrowth),
		makeBegincast(8000, wildGrowth),
		makeBegincast(16000, wildGrowth),
	}
	pipeline.Run(events)
	require.Equal(t, []int{0, 8000, 16000}, attr.WgCastTimestamps())
}

func TestWgAttributionOnCooldown(t *testing.T) {
	gg := talents.NewGroveGuardiansAttributor()
	wgCd := talents.NewWgCooldownReductionAttributor([]talents.TalentAttributor{gg}, true)

	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.EarlySpringNodeID, 82043},
			[]int{talents.EarlySpringTalentID}),
		makeBegincast(1000, wildGrowth),
		makeHeal(1100, ggNourish, 5000, withSource(99)),
		makeBegincast(8500, wildGrowth),
		makeHeal(8600, ggNourish, 5000, withSource(99)),
		makeBegincast(16000, wildGrowth),
		makeHeal(16100, ggNourish, 5000, withSource(99)),
	}
	petIDs := map[int]bool{99: true}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{gg, wgCd}, nil, petIDs)
	results := pipeline.Run(events)

	ratio := 1.0 - 7000.0/8000.0
	fraction := (ratio * 2) / 3
	expected := fraction * 15000
	require.InDelta(t, expected, results.TalentHealing["Early Spring (WG)"], expected*0.02)
}

func TestWgNoAttributionWhenNotOnCooldown(t *testing.T) {
	gg := talents.NewGroveGuardiansAttributor()
	wgCd := talents.NewWgCooldownReductionAttributor([]talents.TalentAttributor{gg}, true)

	events := []map[string]any{
		smCombatantInfo(0,
			[]int{talents.EarlySpringNodeID, 82043},
			[]int{talents.EarlySpringTalentID}),
		makeBegincast(1000, wildGrowth),
		makeHeal(1100, ggNourish, 5000, withSource(99)),
		makeBegincast(31000, wildGrowth),
		makeHeal(31100, ggNourish, 5000, withSource(99)),
	}
	petIDs := map[int]bool{99: true}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{gg, wgCd}, nil, petIDs)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Early Spring (WG)"])
}
