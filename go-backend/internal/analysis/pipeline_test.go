package analysis_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

// FakeAttributor claims 50% of Rejuv (774) healing.
type FakeAttributor struct {
	talents.BaseAttributor
}

func NewFakeAttributor() *FakeAttributor {
	return &FakeAttributor{
		BaseAttributor: talents.NewBaseAttributor("Fake Talent", nil, nil),
	}
}

func (f *FakeAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == 774 {
		return float64(event.Amount) * 0.5
	}
	return 0.0
}

func makeHealRaw(ts, ability, amount, overheal int) map[string]any {
	return map[string]any{
		"timestamp":     ts,
		"type":          "heal",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
		"amount":        amount,
		"overheal":      overheal,
		"hitType":       1,
	}
}

func TestPipelineAttributesHealing(t *testing.T) {
	rawEvents := []map[string]any{makeHealRaw(100, 774, 10000, 0)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{NewFakeAttributor()}, nil, nil)
	results := pipeline.Run(rawEvents)
	require.Equal(t, 5000.0, results.TalentHealing["Fake Talent"])
}

func TestPipelineSkipsWastedHeals(t *testing.T) {
	rawEvents := []map[string]any{makeHealRaw(100, 774, 2000, 3000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{NewFakeAttributor()}, nil, nil)
	results := pipeline.Run(rawEvents)
	require.Equal(t, 0.0, results.TalentHealing["Fake Talent"])
	require.Greater(t, results.Wasted, 0)
}

func TestPipelineTracksTotalHealing(t *testing.T) {
	rawEvents := []map[string]any{
		makeHealRaw(100, 774, 10000, 0),
		makeHealRaw(200, 48438, 5000, 0),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{NewFakeAttributor()}, nil, nil)
	results := pipeline.Run(rawEvents)
	require.Equal(t, 15000, results.TotalHealing)
	require.Equal(t, 5000.0, results.TalentHealing["Fake Talent"])
}
