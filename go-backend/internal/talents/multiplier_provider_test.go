package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

// TestMultiplierProviderInterface verifies that all multiplicative attributors
// implement MultiplierProvider and that GetMultiplier is consistent with ProcessHeal.
func TestMultiplierProviderInterface(t *testing.T) {
	// Verify these types implement MultiplierProvider at compile time.
	var _ talents.MultiplierProvider = (*talents.StaticBuffAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.IntensityAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.HarmonyOfTheGroveAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.PowerOfNatureAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.TreeOfLifeAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.ReforestationAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.VigorousCreepersAttributor)(nil)
	var _ talents.MultiplierProvider = (*talents.RootNetworkAttributor)(nil)
}

// TestStaticBuffMultiplierConsistency checks that GetMultiplier and ProcessHeal agree.
func TestStaticBuffMultiplierConsistency(t *testing.T) {
	a := talents.NewWildSynthesisAttributor()
	hot := tracking.NewHotTracker()
	buff := tracking.NewBuffTracker()

	he := &models.HealEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "heal"},
		AbilityID: 422090, // Grove Guardian Nourish
		Amount:    10000,
		HitType:   1,
	}

	mult := a.GetMultiplier(he, hot, buff)
	bonus := a.ProcessHeal(he, hot, buff)

	require.InDelta(t, 1.3, mult, 0.001)
	expected := float64(he.Amount) - float64(he.Amount)/mult
	require.InDelta(t, expected, bonus, 0.01)
}

// TestStaticBuffMultiplierNoMatch returns 1.0 for non-matching spells.
func TestStaticBuffMultiplierNoMatch(t *testing.T) {
	a := talents.NewWildSynthesisAttributor()
	hot := tracking.NewHotTracker()
	buff := tracking.NewBuffTracker()

	he := &models.HealEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "heal"},
		AbilityID: 774, // Rejuvenation - not in Wild Synthesis spell list
		Amount:    10000,
		HitType:   1,
	}

	require.Equal(t, 1.0, a.GetMultiplier(he, hot, buff))
}

// TestVigorousCreepersMultiplier verifies 1.2x when SB is active on target.
func TestVigorousCreepersMultiplier(t *testing.T) {
	a := talents.NewVigorousCreepersAttributor()
	hot := tracking.NewHotTracker()
	buff := tracking.NewBuffTracker()

	// Apply Symbiotic Bloom on target
	hot.Process(&models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 50, SourceID: 1, Type: "applybuff"},
		TargetID:  2, AbilityID: 439530,
	})

	he := &models.HealEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "heal"},
		TargetID:  2,
		AbilityID: 774, // Rejuv
		Amount:    10000,
		HitType:   1,
	}

	require.InDelta(t, 1.2, a.GetMultiplier(he, hot, buff), 0.001)

	// SB heal itself should return 1.0
	sbHeal := &models.HealEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "heal"},
		TargetID:  2,
		AbilityID: 439530,
		Amount:    5000,
		HitType:   1,
	}
	require.Equal(t, 1.0, a.GetMultiplier(sbHeal, hot, buff))
}

// TestTreeOfLifeMultiplier verifies ToL returns correct multipliers.
func TestTreeOfLifeMultiplier(t *testing.T) {
	a := talents.NewTreeOfLifeAttributor()
	hot := tracking.NewHotTracker()
	buff := tracking.NewBuffTracker()

	rejuvHeal := &models.HealEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "heal"},
		AbilityID: 774, Amount: 10000, HitType: 1,
	}
	otherHeal := &models.HealEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "heal"},
		AbilityID: 8936, Amount: 10000, HitType: 1,
	}

	// Not in ToL -> 1.0
	require.Equal(t, 1.0, a.GetMultiplier(rejuvHeal, hot, buff))

	// Activate ToL
	a.ProcessEvent(&models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 50, SourceID: 1, Type: "applybuff"},
		TargetID:  1, AbilityID: 33891,
	}, hot, buff)

	require.InDelta(t, 1.5, a.GetMultiplier(rejuvHeal, hot, buff), 0.001)
	require.InDelta(t, 1.1, a.GetMultiplier(otherHeal, hot, buff), 0.001)
}
