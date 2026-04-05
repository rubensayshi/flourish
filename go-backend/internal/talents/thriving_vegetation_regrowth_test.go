package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

const tvRegrowth = "Thriving Vegetation: Regrowth"

func TestTVRegrowthTickPastBaseDuration(t *testing.T) {
	events := []map[string]any{
		makeApply(0, regrowth),
		makeHeal(13000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing[tvRegrowth], 1.0)
}

func TestTVRegrowthTickWithinBaseDuration(t *testing.T) {
	events := []map[string]any{
		makeApply(0, regrowth),
		makeHeal(11000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthDirectHealNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(0, regrowth),
		makeHeal(13000, regrowth, 5000), // not a tick
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthNonRegrowthNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeApply(0, rejuv),
		makeHeal(13000, rejuv, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthNoHotTracked(t *testing.T) {
	events := []map[string]any{
		makeHeal(13000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthRefreshResetsBaseDuration(t *testing.T) {
	// Apply at 0, refresh at 10s. Non-TV expiry: 10000 + 12000 + min(2000, 3600) = 24000.
	// Tick at 23s is within non-TV window.
	events := []map[string]any{
		makeApply(0, regrowth),
		makeRefresh(10000, regrowth),
		makeHeal(23000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthRefreshThenTickPastPandemic(t *testing.T) {
	// Apply at 0, refresh at 10s. Non-TV expiry: 10000 + 12000 + 2000 = 24000.
	// Tick at 25s > 24000 → attributed to TV.
	events := []map[string]any{
		makeApply(0, regrowth),
		makeRefresh(10000, regrowth),
		makeHeal(25000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing[tvRegrowth], 1.0)
}

func TestTVRegrowthRefreshAfterNonTVExpiry(t *testing.T) {
	// Apply at 0, non-TV expiry = 12000. Refresh at 14s (past non-TV expiry).
	// Treated as fresh apply: new non-TV expiry = 14000 + 12000 = 26000.
	// Tick at 25s is within non-TV window → not attributed.
	events := []map[string]any{
		makeApply(0, regrowth),
		makeRefresh(14000, regrowth),
		makeHeal(25000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthRefreshAfterNonTVExpiryThenTickPastNewExpiry(t *testing.T) {
	// Apply at 0, refresh at 14s (past 12s non-TV expiry) → fresh: expiry = 26000.
	// Tick at 27s > 26000 → attributed.
	events := []map[string]any{
		makeApply(0, regrowth),
		makeRefresh(14000, regrowth),
		makeHeal(27000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing[tvRegrowth], 1.0)
}

func TestTVRegrowthRemoveStopsTracking(t *testing.T) {
	events := []map[string]any{
		makeApply(0, regrowth),
		makeRemove(5000, regrowth),
		makeHeal(13000, regrowth, 5000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing[tvRegrowth])
}

func TestTVRegrowthPandemicCapped(t *testing.T) {
	// Apply at 0, refresh at 1s. Remaining = 11000, capped at 3600.
	// Non-TV expiry = 1000 + 12000 + 3600 = 16600.
	// Tick at 16000 is within → not attributed.
	// Tick at 17000 is past → attributed.
	events := []map[string]any{
		makeApply(0, regrowth),
		makeRefresh(1000, regrowth),
		makeHeal(16000, regrowth, 5000, withTick()),
		makeHeal(17000, regrowth, 3000, withTick()),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewThrivingVegetationRegrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3000.0, results.TalentHealing[tvRegrowth], 1.0)
}
