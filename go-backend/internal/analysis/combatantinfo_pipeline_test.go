package analysis_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

// TestAttr tracks whether it received combatant info.
type TestAttr struct {
	talents.BaseAttributor
	SawInfo bool
}

func NewTestAttr() *TestAttr {
	return &TestAttr{
		BaseAttributor: talents.NewBaseAttributor("Test", nil, nil),
	}
}

func (a *TestAttr) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	a.SawInfo = true
}

// Unused but required to keep interface satisfied if BaseAttributor methods are pointer-receiver.
func (a *TestAttr) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	return 0.0
}

func TestCombatantInfoPassedToAttributors(t *testing.T) {
	attr := NewTestAttr()
	rawEvents := []map[string]any{
		{
			"timestamp":  1000,
			"type":       "combatantinfo",
			"sourceID":   3,
			"talentTree": []any{map[string]any{"id": 1, "rank": 1, "nodeID": 82047}},
			"critSpell":  256, "hasteSpell": 564, "mastery": 893, "specID": 105,
		},
		makeHealRaw(2000, 774, 1000, 0),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, nil)
	pipeline.Run(rawEvents)
	require.True(t, attr.SawInfo)
	require.True(t, attr.HasTalent(82047))
	require.False(t, attr.HasTalent(99999))
}

func TestTalentFilteringRemovesUnselected(t *testing.T) {
	nodeID82047 := 82047
	nodeID99999 := 99999

	selected := talents.NewBaseAttributorPtr("Selected", &nodeID82047, nil)
	unselected := talents.NewBaseAttributorPtr("Unselected", &nodeID99999, nil)
	noNode := talents.NewBaseAttributorPtr("NoNode", nil, nil)

	rawEvents := []map[string]any{
		{
			"timestamp":  1000,
			"type":       "combatantinfo",
			"sourceID":   3,
			"talentTree": []any{map[string]any{"id": 103098, "rank": 1, "nodeID": 82047}},
			"critSpell":  256, "hasteSpell": 564, "mastery": 893, "specID": 105,
		},
	}
	pipeline := analysis.NewPipeline(
		[]talents.TalentAttributor{selected, unselected, noNode}, nil, nil,
	)
	results := pipeline.Run(rawEvents)
	_, hasSelected := results.TalentHealing["Selected"]
	_, hasUnselected := results.TalentHealing["Unselected"]
	_, hasNoNode := results.TalentHealing["NoNode"]
	require.True(t, hasSelected)
	require.False(t, hasUnselected)
	require.True(t, hasNoNode)
}

func TestChoiceNodeFiltering(t *testing.T) {
	nodeID := 82064
	talentIDa := 108125
	talentIDb := 108124

	choiceA := talents.NewBaseAttributorPtr("Choice A", &nodeID, &talentIDa)
	choiceB := talents.NewBaseAttributorPtr("Choice B", &nodeID, &talentIDb)

	rawEvents := []map[string]any{
		{
			"timestamp":  1000,
			"type":       "combatantinfo",
			"sourceID":   3,
			"talentTree": []any{map[string]any{"id": 108125, "rank": 1, "nodeID": 82064}},
			"critSpell":  256, "hasteSpell": 564, "mastery": 893, "specID": 105,
		},
	}
	pipeline := analysis.NewPipeline(
		[]talents.TalentAttributor{choiceA, choiceB}, nil, nil,
	)
	results := pipeline.Run(rawEvents)
	_, hasA := results.TalentHealing["Choice A"]
	_, hasB := results.TalentHealing["Choice B"]
	require.True(t, hasA)
	require.False(t, hasB)
}
