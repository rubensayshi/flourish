package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)



// --- Grove Guardians (Keeper version with divisor) ---

func TestGroveGuardiansBasicNourish(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, nourish, 10000, withSource(petSource), withTarget(target)),
	}
	gg := talents.NewGroveGuardiansAttributor()
	gg.SetPlayerPetIDs(map[int]bool{petSource: true})
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{gg}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Grove Guardians"], 1.0)
}

func TestGroveGuardiansDivisorWithWildSynthesis(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withTalentTree([]map[string]any{
			{"nodeID": 82043, "id": 1},
			{"nodeID": 94535, "id": 2}, // Wild Synthesis
		})),
		makeHeal(100, nourish, 13000, withSource(petSource), withTarget(target)),
	}
	gg := talents.NewGroveGuardiansAttributor()
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{gg}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Grove Guardians"], 1.0)
}

// --- Wild Synthesis ---

func TestWildSynthesisNourish(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, nourish, 13000, withSource(petSource), withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	expected := 13000.0 - 13000.0/1.3
	require.InDelta(t, expected, results.TalentHealing["Wild Synthesis"], 1.0)
}

func TestWildSynthesisEfflorescence(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, efflorescence, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 5000.0 - 5000.0/1.3
	require.InDelta(t, expected, results.TalentHealing["Wild Synthesis"], 1.0)
}

// --- Dream Surge ---

func TestDreamSurgeDreamBloom(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, dreamBloom, 8000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewDreamSurgeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 8000.0, results.TalentHealing["Dream Surge"], 1.0)
}

// --- Spirit of the Thicket ---

func TestSpiritOfTheThicketDryadBeam(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, dryadBeam, 6000, withSource(petSource), withTarget(target)),
	}
	a := talents.NewSpiritOfTheThicketAttributor()
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	require.InDelta(t, 6000.0, results.TalentHealing["Spirit of the Thicket"], 1.0)
}

// --- Grove's Inspiration ---

func TestGrovesInspirationRegrowth(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, rejuv, 10000, withTarget(target)), // not boosted
		makeHeal(200, 8936, 10000, withTarget(target)),  // Regrowth - boosted
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewGrovesInspirationAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.09
	require.InDelta(t, expected, results.TalentHealing["Grove's Inspiration"], 1.0)
}

// --- Cenarius' Might ---

func TestCenariusMightSwiftmend(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, swiftmend, 20000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewCenariusMightAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 20000.0 - 20000.0/1.2
	require.InDelta(t, expected, results.TalentHealing["Cenarius' Might"], 1.0)
}

// --- Bounteous Bloom ---

func TestBountifulBloomTreantHeal(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, treantHeal, 13000, withSource(petSource), withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBountifulBloomAttributor()}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	expected := 13000.0 - 13000.0/1.3
	require.InDelta(t, expected, results.TalentHealing["Bounteous Bloom"], 1.0)
}

// --- Patient Custodian ---

func TestPatientCustodianRejuv(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPatientCustodianAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.06
	require.InDelta(t, expected, results.TalentHealing["Patient Custodian"], 1.0)
}

// --- Harmony of the Grove ---

func TestHarmonyOfTheGroveNoGuardians(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewHarmonyOfTheGroveAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Harmony of the Grove"])
}

func TestHarmonyOfTheGroveOneGuardian(t *testing.T) {
	events := []map[string]any{
		makeSummon(100, guardianAbil),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewHarmonyOfTheGroveAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.05
	require.InDelta(t, expected, results.TalentHealing["Harmony of the Grove"], 1.0)
}

func TestHarmonyOfTheGroveTwoGuardians(t *testing.T) {
	events := []map[string]any{
		makeSummon(100, guardianAbil),
		makeSummon(200, guardianAbil),
		makeHeal(300, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewHarmonyOfTheGroveAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.10
	require.InDelta(t, expected, results.TalentHealing["Harmony of the Grove"], 1.0)
}

func TestHarmonyOfTheGroveGuardianExpired(t *testing.T) {
	events := []map[string]any{
		makeSummon(100, guardianAbil),
		makeHeal(9000, rejuv, 10000, withTarget(target)), // after 8s duration
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewHarmonyOfTheGroveAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Harmony of the Grove"])
}

// --- Power of Nature ---

func TestPowerOfNatureRejuvWithGuardian(t *testing.T) {
	events := []map[string]any{
		makeSummon(100, guardianAbil),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPowerOfNatureAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 - 10000.0/1.10
	require.InDelta(t, expected, results.TalentHealing["Power of Nature"], 1.0)
}

func TestPowerOfNatureNonMatchingSpell(t *testing.T) {
	events := []map[string]any{
		makeSummon(100, guardianAbil),
		makeHeal(200, swiftmend, 10000, withTarget(target)), // not in spell list
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPowerOfNatureAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Power of Nature"])
}

func TestPowerOfNatureNoGuardians(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPowerOfNatureAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Power of Nature"])
}

// --- Sylvan Beckoning ---

func TestSylvanBeckoningDryadTranq(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, dryadTranq, 12000, withSource(petSource), withTarget(target)),
	}
	a := talents.NewSylvanBeckoningAttributor()
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	require.InDelta(t, 12000.0, results.TalentHealing["Sylvan Beckoning"], 1.0)
}

func TestSylvanBeckoningNonPetNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, dryadTranq, 12000, withTarget(target)), // sourceID=1, not pet
	}
	a := talents.NewSylvanBeckoningAttributor()
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{a}, nil, map[int]bool{petSource: true})
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Sylvan Beckoning"])
}

// --- Bursting Growth ---

func TestBurstingGrowth(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, 440121, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBurstingGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Bursting Growth"], 1.0)
}

// --- Thriving Growth ---

func TestThrivingGrowth(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, 474760, 7000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 7000.0, results.TalentHealing["Thriving Growth"], 1.0)
}
