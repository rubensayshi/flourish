package models_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestParseCombatantInfo(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000,
		"type":      "combatantinfo",
		"sourceID":  3,
		"talentTree": []any{
			map[string]any{"id": 103098, "rank": 1, "nodeID": 82047},
			map[string]any{"id": 103100, "rank": 1, "nodeID": 82049},
		},
		"critSpell":  256,
		"hasteSpell": 564,
		"mastery":    893,
		"specID":     105,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	ci, ok := event.(*models.CombatantInfoEvent)
	require.True(t, ok)
	require.True(t, ci.TalentNodes[82047])
	require.True(t, ci.TalentNodes[82049])
	require.Equal(t, 256.0, ci.CritSpell)
	require.Equal(t, 105, ci.SpecID)
}
