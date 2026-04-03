package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestEverbloomAttributesAllHealing(t *testing.T) {
	events := []map[string]any{makeHeal(100, 1244341, 5000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewEverbloomSplashAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 5000.0, results.TalentHealing["Everbloom: Splash"])
}

func TestGroveGuardiansAttributesNourish(t *testing.T) {
	events := []map[string]any{makeHeal(100, 422090, 3000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewGroveGuardiansAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 3000.0, results.TalentHealing["Grove Guardians"])
}

func TestDreamSurgeAttributesDreamBloom(t *testing.T) {
	events := []map[string]any{makeHeal(100, 434141, 2000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewDreamSurgeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 2000.0, results.TalentHealing["Dream Surge"])
}

func TestDirectSpellIgnoresUnrelatedSpells(t *testing.T) {
	events := []map[string]any{makeHeal(100, 774, 10000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewEverbloomSplashAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Splash"])
}

func TestDirectSpellSkipsWastedHeals(t *testing.T) {
	events := []map[string]any{makeHeal(100, 1244341, 2000, withOverheal(3000))}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewEverbloomSplashAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Splash"])
}

func TestMultipleDirectAttributors(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, 1244341, 5000),
		makeHeal(200, 422090, 3000),
		makeHeal(300, 434141, 2000),
		makeHeal(400, 81269, 1000),
		makeHeal(500, 774, 8000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{
		talents.NewEverbloomSplashAttributor(),
		talents.NewGroveGuardiansAttributor(),
		talents.NewDreamSurgeAttributor(),
		talents.NewEfflorescenceAttributor(),
	}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 5000.0, results.TalentHealing["Everbloom: Splash"])
	require.Equal(t, 3000.0, results.TalentHealing["Grove Guardians"])
	require.Equal(t, 2000.0, results.TalentHealing["Dream Surge"])
	require.Equal(t, 1000.0, results.TalentHealing["Efflorescence"])
	require.Equal(t, 19000, results.TotalHealing)
}

func TestRampantGrowthAttributesBonusOnRegrowthTicks(t *testing.T) {
	events := []map[string]any{makeHeal(100, 8936, 10000, withTick())}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Rampant Growth"], 1.0)
}

func TestRampantGrowthIgnoresDirectHeals(t *testing.T) {
	events := []map[string]any{makeHeal(100, 8936, 10000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Rampant Growth"])
}

func TestRampantGrowthIgnoresOtherSpells(t *testing.T) {
	events := []map[string]any{makeHeal(100, 774, 10000, withTick())}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Rampant Growth"])
}

func TestRampantGrowthSkipsWastedHeals(t *testing.T) {
	events := []map[string]any{makeHeal(100, 8936, 2000, withOverheal(3000), withTick())}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Rampant Growth"])
}
