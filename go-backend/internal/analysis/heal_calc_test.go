package analysis_test

import (
	"math"
	"sort"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func testCoeffs() *models.SpellCoefficients {
	return &models.SpellCoefficients{
		Spells: map[int]models.SpellConfig{
			774: {
				Name:       "Rejuvenation",
				DurationMS: 12000,
				Effects: []models.SpellEffectConfig{
					{Type: "periodic", Coefficient: 0.803, PeriodMS: 3000},
				},
			},
			8936: {
				Name:       "Regrowth",
				DurationMS: 6000,
				Effects: []models.SpellEffectConfig{
					{Type: "direct", Coefficient: 5.36},
					{Type: "periodic", Coefficient: 0.225, PeriodMS: 2000},
				},
			},
		},
	}
}

func TestGetCoefficient(t *testing.T) {
	hc := &analysis.HealCalculator{Coefficients: testCoeffs()}

	require.InDelta(t, 0.803, hc.GetCoefficient(774, true), 0.001)
	require.InDelta(t, 0.0, hc.GetCoefficient(774, false), 0.001) // no direct effect
	require.InDelta(t, 5.36, hc.GetCoefficient(8936, false), 0.01)
	require.InDelta(t, 0.225, hc.GetCoefficient(8936, true), 0.001)
	require.Equal(t, 0.0, hc.GetCoefficient(99999, true)) // unknown spell
}

func TestMasteryMult(t *testing.T) {
	hc := &analysis.HealCalculator{
		MasteryBase: 0.1069, // 10.69% mastery
		DRTable:     []float64{0, 1.0, 1.7, 2.3, 2.8, 3.2, 3.6, 4.0, 4.4, 4.8},
	}

	require.InDelta(t, 1.0, hc.MasteryMult(0), 0.001)
	require.InDelta(t, 1.1069, hc.MasteryMult(1), 0.001)
	require.InDelta(t, 1.0+0.1069*1.7, hc.MasteryMult(2), 0.001)
	require.InDelta(t, 1.0+0.1069*2.3, hc.MasteryMult(3), 0.001)
}

func TestExpectedHeal(t *testing.T) {
	hc := &analysis.HealCalculator{
		EffectiveSP: 5000,
		VersPercent: 0.02,   // 2% vers
		MasteryBase: 0.1069,
		DRTable:     []float64{0, 1.0, 1.7, 2.3},
	}

	// Rejuv tick: 5000 * 0.803 * 1.02 * 1.1069 = 4535
	expected := hc.ExpectedHeal(0.803, 1, false)
	require.InDelta(t, 5000*0.803*1.02*1.1069, expected, 1.0)

	// Same but crit: x2
	expectedCrit := hc.ExpectedHeal(0.803, 1, true)
	require.InDelta(t, expected*2, expectedCrit, 1.0)
}

func TestCalibrateFromTicks(t *testing.T) {
	hc := analysis.NewHealCalculator(
		&models.CombatantInfoEvent{
			Intellect:   2847,
			Versatility: 384,
			Mastery:     1069,
		},
		testCoeffs(),
		[]float64{0, 1.0, 1.7, 2.3},
	)

	// Simulate clean ticks: SP=5000, coeff=0.803, vers=1.019, mastery(1)=1.1069
	baseTickF := 5000.0 * 0.803 * (1.0 + 384.0/20500.0) * (1.0 + 0.1069*1.0)
	baseTick := int(baseTickF)
	ticks := []analysis.CleanTick{}
	for i := 0; i < 20; i++ {
		ticks = append(ticks, analysis.CleanTick{RawAmount: baseTick + i*10})
	}
	sort.Slice(ticks, func(i, j int) bool { return ticks[i].RawAmount < ticks[j].RawAmount })

	hc.CalibrateFromTicks(774, ticks)
	require.InDelta(t, 5000, hc.EffectiveSP, 50) // should recover ~5000 SP
}

func TestDecomposeHeal(t *testing.T) {
	// Heal of 10000 with vers=1.05, mastery=1.15, talent=1.6
	mults := map[string]float64{
		"Versatility":     1.05,
		"Mastery: Harmony": 1.15,
		"SotF":            1.60,
	}
	result := analysis.DecomposeHeal(10000, mults)

	// Should sum to 10000
	total := 0.0
	for _, v := range result {
		total += v
	}
	require.InDelta(t, 10000, total, 0.01)

	// Spell Power (base) should be 10000 / (1.05 * 1.15 * 1.6) = 5176
	require.InDelta(t, 10000/(1.05*1.15*1.6), result["Spell Power"], 1.0)

	// SotF should get the largest share of the bonus (it's the biggest multiplier)
	require.Greater(t, result["SotF"], result["Versatility"])
	require.Greater(t, result["SotF"], result["Mastery: Harmony"])
}

func TestDecomposeHealNoMultipliers(t *testing.T) {
	result := analysis.DecomposeHeal(5000, map[string]float64{})
	require.InDelta(t, 5000, result["Spell Power"], 0.01)
}

func TestDecomposeHealProportions(t *testing.T) {
	// With two equal multipliers, they should get equal shares
	mults := map[string]float64{
		"A": 1.10,
		"B": 1.10,
	}
	result := analysis.DecomposeHeal(10000, mults)

	require.InDelta(t, result["A"], result["B"], 0.01)

	// With a larger multiplier, it should get proportionally more
	mults2 := map[string]float64{
		"Small": 1.05,
		"Big":   1.50,
	}
	result2 := analysis.DecomposeHeal(10000, mults2)
	// ln(1.5)/ln(1.05) ≈ 8.31, so Big should get ~8x more than Small
	ratio := result2["Big"] / result2["Small"]
	expectedRatio := math.Log(1.5) / math.Log(1.05)
	require.InDelta(t, expectedRatio, ratio, 0.01)
}
