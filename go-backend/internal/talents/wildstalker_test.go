package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)


// --- Vigorous Creepers ---

func TestVigorousCreepersBuffOnTargetBoostsHeal(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, rejuv, 12000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewVigorousCreepersAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 2000.0, results.TalentHealing["Vigorous Creepers"], 1.0)
}

func TestVigorousCreepersNoBloomNoBonus(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 12000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewVigorousCreepersAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Vigorous Creepers"])
}

func TestVigorousCreepersBloomOwnHealingNotCounted(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, symBloom, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewVigorousCreepersAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Vigorous Creepers"])
}

func TestVigorousCreepersBloomRemovedNoBonus(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeRemove(150, symBloom, withTarget(target)),
		makeHeal(200, rejuv, 12000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewVigorousCreepersAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Vigorous Creepers"])
}

// --- Implant ---

func TestImplantSmTriggersBloom(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(200, symBloom, withTarget(target)),
		makeHeal(300, symBloom, 8000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImplantAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 8000.0, results.TalentHealing["Implant"], 1.0)
}

func TestImplantWgTriggersBloom(t *testing.T) {
	events := []map[string]any{
		makeCast(100, wildGrowth, withTarget(target)),
		makeApply(200, symBloom, withTarget(20)),
		makeHeal(300, symBloom, 5000, withTarget(20)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImplantAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Implant"], 1.0)
}

func TestImplantNaturalBloomNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, symBloom, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImplantAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Implant"])
}

func TestImplantBloomOutsideWindowNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(700, symBloom, withTarget(target)),
		makeHeal(800, symBloom, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImplantAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Implant"])
}

func TestImplantNonBloomHealingNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(200, symBloom, withTarget(target)),
		makeHeal(300, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewImplantAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Implant"])
}

// --- Root Network ---

func TestRootNetworkSingleBloomGives2Pct(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(20)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRootNetworkAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.02
	require.InDelta(t, expected, results.TalentHealing["Root Network"], 1.0)
}

func TestRootNetworkMultipleBlooms(t *testing.T) {
	events := []map[string]any{
		makeApply(100, symBloom, withTarget(target)),
		makeApply(110, symBloom, withTarget(20)),
		makeHeal(200, rejuv, 10000, withTarget(30)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRootNetworkAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.04
	require.InDelta(t, expected, results.TalentHealing["Root Network"], 1.0)
}

func TestRootNetworkNoBlooms(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRootNetworkAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Root Network"])
}

// --- Strategic Infusion ---

const critRatingPerPercent = 700.0

func TestStrategicInfusionPeriodicCritAttributed(t *testing.T) {
	critRating := 0.21 * critRatingPerPercent
	events := []map[string]any{
		makeCombatantInfo(0, withCritSpell(critRating), withTalentTree([]map[string]any{
			{"nodeID": 94623, "id": 117223},
		})),
		makeHeal(100, rejuv, 10000, withHitType(2), withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewStrategicInfusionAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 5000.0 * (0.04 / 0.25)
	require.InDelta(t, expected, results.TalentHealing["Strategic Infusion"], expected*0.02)
}

func TestStrategicInfusionPeriodicNonCritNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withCritSpell(0.21*critRatingPerPercent), withTalentTree([]map[string]any{
			{"nodeID": 94623, "id": 117223},
		})),
		makeHeal(100, rejuv, 10000, withHitType(1), withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewStrategicInfusionAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Strategic Infusion"])
}

func TestStrategicInfusionDirectCritNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withCritSpell(0.21*critRatingPerPercent), withTalentTree([]map[string]any{
			{"nodeID": 94623, "id": 117223},
		})),
		makeHeal(100, swiftmend, 20000, withHitType(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewStrategicInfusionAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Strategic Infusion"])
}

func TestStrategicInfusionNoCombatantInfoUsesMinimumCrit(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, rejuv, 10000, withHitType(2), withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewStrategicInfusionAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 5000.0 * (0.04 / 0.09)
	require.InDelta(t, expected, results.TalentHealing["Strategic Infusion"], expected*0.02)
}
