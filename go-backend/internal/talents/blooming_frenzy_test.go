package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestBloomingFrenzySotFTriggersLBBlooms(t *testing.T) {
	events := []map[string]any{
		makeRemove(100, sotfBuffID),
		makeHeal(200, lifebloomBloom, 5000, withTarget(target)),
		makeHeal(300, lifebloomBloom, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBloomingFrenzyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Everbloom: Blooming Frenzy"], 1.0)
}

func TestBloomingFrenzyMaxFiveBlooms(t *testing.T) {
	events := []map[string]any{
		makeRemove(100, sotfBuffID),
	}
	for i := 0; i < 7; i++ {
		events = append(events, makeHeal(200+i*50, lifebloomBloom, 1000, withTarget(target)))
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBloomingFrenzyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Everbloom: Blooming Frenzy"], 1.0)
}

func TestBloomingFrenzyExpiredWindow(t *testing.T) {
	events := []map[string]any{
		makeRemove(100, sotfBuffID),
		makeHeal(2000, lifebloomBloom, 5000, withTarget(target)), // >1500ms after SotF removal
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBloomingFrenzyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Blooming Frenzy"])
}

func TestBloomingFrenzyNoSotFNoAttribution(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, lifebloomBloom, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBloomingFrenzyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Blooming Frenzy"])
}

func TestBloomingFrenzyNonLBNotCounted(t *testing.T) {
	events := []map[string]any{
		makeRemove(100, sotfBuffID),
		makeHeal(200, rejuv, 5000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewBloomingFrenzyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Blooming Frenzy"])
}
