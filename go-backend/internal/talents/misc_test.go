package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)



// --- Abundance ---

func TestAbundanceRegrowthCritWithRejuvs(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withCritSpell(350)),
		makeApply(100, rejuv, withTarget(target)),
		makeApply(110, rejuv, withTarget(20)),
		makeHeal(200, regrowth, 20000, withHitType(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewAbundanceAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (0.16 / 0.66)
	require.InDelta(t, expected, results.TalentHealing["Abundance"], expected*0.02)
}

func TestAbundanceNonCritRegrowthNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, rejuv),
		makeHeal(200, regrowth, 10000, withHitType(1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewAbundanceAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Abundance"])
}

func TestAbundanceNoRejuvsNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeHeal(200, regrowth, 10000, withHitType(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewAbundanceAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Abundance"])
}

func TestAbundanceNonRegrowthNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, rejuv),
		makeHeal(200, rejuv, 10000, withHitType(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewAbundanceAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Abundance"])
}

func TestAbundanceGerminationRejuvCounted(t *testing.T) {
	events := []map[string]any{
		makeCombatantInfo(0, withCritSpell(350)),
		makeApply(100, germinationRejuv, withTarget(target)),
		makeHeal(200, regrowth, 20000, withHitType(2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewAbundanceAttributor()}, nil, nil)
	results := pipeline.Run(events)
	expected := 10000.0 * (0.08 / 0.58)
	require.InDelta(t, expected, results.TalentHealing["Abundance"], expected*0.02)
}

// --- Photosynthesis ---

func TestPhotosynthesisUnexplainedBloomAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeHeal(500, lifebloomBloom, 15000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 15000.0, results.TalentHealing["Photosynthesis"], 1.0)
}

func TestPhotosynthesisBloomFromExpiryNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeRemove(499, lifebloom),
		makeHeal(500, lifebloomBloom, 15000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Photosynthesis"])
}

func TestPhotosynthesisBloomFromExpiryRemoveAfter(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeHeal(500, lifebloomBloom, 15000),
		makeRemove(501, lifebloom),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Photosynthesis"])
}

func TestPhotosynthesisBloomFromRefreshNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeRefresh(493, lifebloom),
		makeHeal(500, lifebloomBloom, 15000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Photosynthesis"])
}

func TestPhotosynthesisMixedBlooms(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeHeal(500, lifebloomBloom, 10000),
		makeRemove(999, lifebloom),
		makeHeal(1000, lifebloomBloom, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 10000.0, results.TalentHealing["Photosynthesis"], 1.0)
}

func TestPhotosynthesisLbRecastBloomNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeCast(498, lifebloom),
		makeHeal(500, lifebloomBloom, 15000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Photosynthesis"])
}

func TestPhotosynthesisEverbloomBloomsAfterSotfNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeApply(200, sotfBuffID),
		makeRemove(300, sotfBuffID),
		makeHeal(330, lifebloomBloom, 10000),
		makeHeal(580, lifebloomBloom, 10000),
		makeHeal(830, lifebloomBloom, 10000),
		makeHeal(1080, lifebloomBloom, 10000),
		makeHeal(1330, lifebloomBloom, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Photosynthesis"])
}

func TestPhotosynthesisPhotoProcOutsideSotfWindow(t *testing.T) {
	events := []map[string]any{
		makeApply(100, lifebloom),
		makeApply(200, sotfBuffID),
		makeRemove(300, sotfBuffID),
		makeHeal(2000, lifebloomBloom, 15000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewPhotosynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 15000.0, results.TalentHealing["Photosynthesis"], 1.0)
}

// --- Nurturing Dormancy ---

func TestNurturingDormancyTickPastBaseDuration(t *testing.T) {
	events := []map[string]any{
		makeApply(0, rejuv),
		makeHeal(18000, rejuv, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Nurturing Dormancy"], 1.0)
}

func TestNurturingDormancyTickWithinBaseDuration(t *testing.T) {
	events := []map[string]any{
		makeApply(0, rejuv),
		makeHeal(16000, rejuv, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Nurturing Dormancy"])
}

func TestNurturingDormancyNoHotTracked(t *testing.T) {
	events := []map[string]any{
		makeHeal(13000, rejuv, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Nurturing Dormancy"])
}

func TestNurturingDormancyNonRejuvNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(0, regrowth),
		makeHeal(13000, regrowth, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Nurturing Dormancy"])
}

func TestNurturingDormancyRefreshResetsBaseDuration(t *testing.T) {
	events := []map[string]any{
		makeApply(0, rejuv),
		makeRefresh(10000, rejuv),
		makeHeal(26000, rejuv, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Nurturing Dormancy"])
}

func TestNurturingDormancyRefreshThenTickWithinPandemic(t *testing.T) {
	// Refresh at 5s → 12s remaining → pandemic cap 5.1s → expiry = 5000+17000+5100 = 27100
	// Tick at 23s is within pandemic window, NOT ND.
	events := []map[string]any{
		makeApply(0, rejuv),
		makeRefresh(5000, rejuv),
		makeHeal(23000, rejuv, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Nurturing Dormancy"])
}

func TestNurturingDormancyRefreshThenTickPastPandemic(t *testing.T) {
	// Refresh at 5s → pandemic expiry = 27100. Tick at 28s > 27100 → ND.
	events := []map[string]any{
		makeApply(0, rejuv),
		makeRefresh(5000, rejuv),
		makeHeal(28000, rejuv, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewNurturingDormancyAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Nurturing Dormancy"], 1.0)
}
