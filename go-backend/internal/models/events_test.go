package models_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestParseHealEvent(t *testing.T) {
	raw := map[string]any{
		"timestamp":     1000,
		"type":          "heal",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": 774,
		"amount":        5000,
		"overheal":      1000,
		"hitType":       1,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	he, ok := event.(*models.HealEvent)
	require.True(t, ok)
	require.Equal(t, 5000, he.Amount)
	require.Equal(t, 1000, he.Overheal)
	require.Equal(t, 6000, he.RawHeal())
	require.InDelta(t, 1000.0/6000.0, he.OverhealPct(), 0.0001)
}

func TestParseCastEvent(t *testing.T) {
	raw := map[string]any{
		"timestamp":     1000,
		"type":          "cast",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": 18562,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	ce, ok := event.(*models.CastEvent)
	require.True(t, ok)
	require.Equal(t, 18562, ce.AbilityID)
}

func TestParseApplyBuffEvent(t *testing.T) {
	raw := map[string]any{
		"timestamp":     1000,
		"type":          "applybuff",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": 774,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	_, ok := event.(*models.ApplyBuffEvent)
	require.True(t, ok)
}

func TestParseUnknownEventReturnsNil(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000,
		"type":      "totally_unknown",
		"sourceID":  1,
	}
	event := models.ParseEvent(raw)
	require.Nil(t, event)
}

func TestHealEventIsWasted(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 2000, "overheal": 3000, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.True(t, he.IsWasted()) // 3000/5000 = 60% > 50%
}

func TestHealEventNotWasted(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 4000, "overheal": 1000, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.False(t, he.IsWasted()) // 1000/5000 = 20% < 50%
}

func TestHealEventAbsorbIncludedInRawHeal(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 1000, "absorb": 500, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.Equal(t, 500, he.Absorb)
	require.Equal(t, 6500, he.RawHeal())
}

func TestHealEventAbsorbDefaultsToZero(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 1000, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.Equal(t, 0, he.Absorb)
	require.Equal(t, 6000, he.RawHeal())
}

func TestHealEventTickParsed(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 0, "hitType": 1,
		"tick": true,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.True(t, he.Tick)
}

func TestHealEventTickDefaultsToFalse(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 0, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.False(t, he.Tick)
}
