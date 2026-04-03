package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)



func TestSotfAttributesBonusFromRejuv(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(150, rejuv, withTarget(target)),
		makeApply(150, rejuv, withTarget(target)),
		makeRemove(150, sotfBuffID, withTarget(player)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestSotfDoesNotAttributeUnbuffedRejuv(t *testing.T) {
	events := []map[string]any{
		makeApply(100, rejuv, withTarget(target)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["SotF + PotA"])
}

func TestSotfOnlyAppliesToConsumingCast(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(150, rejuv, withTarget(10)),
		makeApply(150, rejuv, withTarget(10)),
		makeRemove(150, sotfBuffID, withTarget(player)),
		makeApply(1000, rejuv, withTarget(20)),
		makeHeal(1100, rejuv, 10000, withTarget(10)),
		makeHeal(1110, rejuv, 10000, withTarget(20)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestSotfRegrowthDirectHealAttributed(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(150, regrowth, withTarget(2)),
		makeHeal(150, regrowth, 60000),
		makeApply(150, regrowth),
		makeRemove(150, sotfBuffID, withTarget(player)),
		makeHeal(200, regrowth, 5000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 22500.0+1875.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestSotfGerminationRejuv(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(150, rejuv, withTarget(2)),
		makeApply(150, germinationRejuv),
		makeRemove(150, sotfBuffID, withTarget(player)),
		makeHeal(200, germinationRejuv, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestPotaSpreadsGetFullAttribution(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(200, rejuv, withTarget(target)),
		makeApply(200, rejuv, withTarget(target)),
		makeRemove(200, sotfBuffID, withTarget(player)),
		makeApply(210, rejuv, withTarget(spread1)),
		makeApply(220, rejuv, withTarget(spread2)),
		makeHeal(500, rejuv, 10000, withTarget(target)),
		makeHeal(510, rejuv, 10000, withTarget(spread1)),
		makeHeal(520, rejuv, 10000, withTarget(spread2)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 23750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestPotaSpreadOutsideWindowNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(200, rejuv, withTarget(target)),
		makeApply(200, rejuv, withTarget(target)),
		makeRemove(200, sotfBuffID, withTarget(player)),
		makeApply(800, rejuv, withTarget(spread1)),
		makeHeal(1000, rejuv, 10000, withTarget(target)),
		makeHeal(1010, rejuv, 10000, withTarget(spread1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestPotaSpreadWrongSpellNotAttributed(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(200, rejuv, withTarget(target)),
		makeApply(200, rejuv, withTarget(target)),
		makeRemove(200, sotfBuffID, withTarget(player)),
		makeApply(210, regrowth, withTarget(spread1)),
		makeHeal(500, rejuv, 10000, withTarget(target)),
		makeHeal(510, regrowth, 10000, withTarget(spread1)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestSotfTagClearedOnRefresh(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeCast(150, rejuv, withTarget(2)),
		makeApply(150, rejuv),
		makeRemove(150, sotfBuffID, withTarget(player)),
		makeHeal(200, rejuv, 10000),
		makeRefresh(300, rejuv),
		makeHeal(400, rejuv, 10000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}

func TestInterferingApplyBuffDoesNotStealSotf(t *testing.T) {
	events := []map[string]any{
		makeCast(100, swiftmend),
		makeApply(100, sotfBuffID, withTarget(player)),
		makeApply(120, rejuv, withTarget(99)),
		makeCast(150, rejuv, withTarget(target)),
		makeApply(150, rejuv, withTarget(target)),
		makeRemove(150, sotfBuffID, withTarget(player)),
		makeHeal(200, rejuv, 10000, withTarget(target)),
		makeHeal(210, rejuv, 10000, withTarget(99)),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewSoulOfTheForestAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3750.0, results.TalentHealing["SotF + PotA"], 1.0)
}
