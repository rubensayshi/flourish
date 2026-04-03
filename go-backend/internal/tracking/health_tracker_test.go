package tracking_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

func makeEvent(typ string, ts int, fields map[string]any) map[string]any {
	e := map[string]any{"type": typ, "timestamp": float64(ts)}
	for k, v := range fields {
		e[k] = v
	}
	return e
}

func combatantInfo(ts, sourceID, stamina int) map[string]any {
	return makeEvent("combatantinfo", ts, map[string]any{
		"sourceID": float64(sourceID),
		"stamina":  float64(stamina),
	})
}

func damage(ts, targetID, amount int) map[string]any {
	return makeEvent("damage", ts, map[string]any{
		"targetID":      float64(targetID),
		"abilityGameID": float64(1),
		"amount":        float64(amount),
	})
}

func heal(ts, targetID, amount, overheal int) map[string]any {
	return makeEvent("heal", ts, map[string]any{
		"targetID":      float64(targetID),
		"abilityGameID": float64(1),
		"sourceID":      float64(99),
		"amount":        float64(amount),
		"overheal":      float64(overheal),
	})
}

func TestHealthTracker_StartsAtFullHP(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500), // maxHP = 500*20 = 10000
	}
	ht := tracking.NewHealthTracker(events)

	// Before any events, should be full
	require.InDelta(t, 1.0, ht.GetHealthPct(1, 100), 0.001)
}

func TestHealthTracker_DamageReducesHealth(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500), // maxHP = 10000
		damage(100, 1, 3000),     // 3000 damage → 7000/10000 = 70%
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 0.7, ht.GetHealthPct(1, 100), 0.001)
	require.InDelta(t, 0.7, ht.GetHealthPct(1, 200), 0.001) // stays at 70% after
}

func TestHealthTracker_HealRestoresHealth(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500), // maxHP = 10000
		damage(100, 1, 5000),     // 50%
		heal(200, 1, 2000, 0),    // +2000 → 70%
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 0.5, ht.GetHealthPct(1, 100), 0.001)
	require.InDelta(t, 0.7, ht.GetHealthPct(1, 200), 0.001)
}

func TestHealthTracker_HealCapsAtMax(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500),  // maxHP = 10000
		damage(100, 1, 2000),      // 80%
		heal(200, 1, 2000, 1000),  // effective 2000, overheal 1000 → capped at 100%
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 0.8, ht.GetHealthPct(1, 100), 0.001)
	require.InDelta(t, 1.0, ht.GetHealthPct(1, 200), 0.001)
}

func TestHealthTracker_MultiplePlayers(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500),  // player 1: maxHP = 10000
		combatantInfo(0, 2, 1000), // player 2: maxHP = 20000
		damage(100, 1, 4000),      // player 1 → 60%
		damage(100, 2, 10000),     // player 2 → 50%
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 0.6, ht.GetHealthPct(1, 100), 0.001)
	require.InDelta(t, 0.5, ht.GetHealthPct(2, 100), 0.001)
}

func TestHealthTracker_UnknownTargetReturnsFull(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500),
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 1.0, ht.GetHealthPct(999, 100), 0.001)
}

func TestHealthTracker_BeforeAnyEventsReturnsFull(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500),
		damage(1000, 1, 5000),
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 1.0, ht.GetHealthPct(1, 500), 0.001) // before damage at t=1000
}

func TestHealthTracker_OverkillNotDoubleSubtracted(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500),   // maxHP = 10000
		damage(100, 1, 9000),       // 10%
		makeEvent("damage", 200, map[string]any{
			"targetID":      float64(1),
			"abilityGameID": float64(1),
			"amount":        float64(3000),
			"overkill":      float64(2000), // 3000 total but 2000 overkill → only 1000 effective
		}),
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 0.1, ht.GetHealthPct(1, 100), 0.001)
	require.InDelta(t, 0.0, ht.GetHealthPct(1, 200), 0.001)
}

func TestHealthTracker_DamageSequenceThenHeal(t *testing.T) {
	events := []map[string]any{
		combatantInfo(0, 1, 500), // maxHP = 10000
		damage(100, 1, 2000),     // 80%
		damage(200, 1, 3000),     // 50%
		damage(300, 1, 1000),     // 40%
		heal(400, 1, 4000, 0),    // 80%
		heal(500, 1, 3000, 1000), // 100% (overheal)
	}
	ht := tracking.NewHealthTracker(events)

	require.InDelta(t, 0.8, ht.GetHealthPct(1, 100), 0.001)
	require.InDelta(t, 0.5, ht.GetHealthPct(1, 200), 0.001)
	require.InDelta(t, 0.4, ht.GetHealthPct(1, 300), 0.001)
	require.InDelta(t, 0.8, ht.GetHealthPct(1, 400), 0.001)
	require.InDelta(t, 1.0, ht.GetHealthPct(1, 500), 0.001)
}
