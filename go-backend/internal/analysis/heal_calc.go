package analysis

import (
	"math"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
)

// VersRatingPerPercent is the versatility rating needed for 1% healing bonus in Midnight.
const VersRatingPerPercent = 20500.0 / 100.0 // 205 rating per 1%

// HealCalculator computes expected heal amounts from spell coefficients and player stats.
type HealCalculator struct {
	Coefficients *models.SpellCoefficients
	EffectiveSP  float64 // calibrated from clean ticks
	VersPercent  float64 // versatility healing bonus as decimal (e.g., 0.0187)
	MasteryBase  float64 // mastery percentage as decimal (e.g., 0.1069)
	DRTable      []float64
}

// NewHealCalculator creates a calculator from combatant info and config.
func NewHealCalculator(ci *models.CombatantInfoEvent, coeffs *models.SpellCoefficients, drTable []float64) *HealCalculator {
	hc := &HealCalculator{
		Coefficients: coeffs,
		DRTable:      drTable,
	}
	if ci != nil {
		hc.VersPercent = ci.Versatility / VersRatingPerPercent / 100.0
		hc.MasteryBase = ci.Mastery / 100.0 / 100.0
	}
	return hc
}

// CalibrateFromTicks derives effective spell power from observed clean heal ticks.
// It finds the minimum non-crit ticks for a given periodic spell and works backwards.
func (hc *HealCalculator) CalibrateFromTicks(spellID int, ticks []CleanTick) {
	if len(ticks) == 0 || hc.Coefficients == nil {
		return
	}

	spell, ok := hc.Coefficients.Spells[spellID]
	if !ok {
		return
	}

	var coeff float64
	for _, eff := range spell.Effects {
		if eff.Type == "periodic" {
			coeff = eff.Coefficient
			break
		}
	}
	if coeff == 0 {
		return
	}

	// Use the lowest ticks (least talent multipliers active)
	// Assume minimum mastery = 1 HoT (the spell itself)
	versMult := 1.0 + hc.VersPercent
	masteryMult := hc.MasteryMult(1) // 1 HoT minimum

	// Take 10th percentile to avoid outlier lows (partial ticks)
	idx := len(ticks) / 10
	if idx < 0 {
		idx = 0
	}
	if idx >= len(ticks) {
		idx = len(ticks) - 1
	}
	tick := ticks[idx]

	hc.EffectiveSP = float64(tick.RawAmount) / (coeff * versMult * masteryMult)
}

// CleanTick is a non-crit heal tick used for SP calibration.
type CleanTick struct {
	RawAmount int // amount + overheal + absorb
}

// MasteryMult returns the mastery multiplier for a given number of HoTs on target.
func (hc *HealCalculator) MasteryMult(hotCount int) float64 {
	if hotCount <= 0 || hc.MasteryBase <= 0 || len(hc.DRTable) == 0 {
		return 1.0
	}
	idx := hotCount
	if idx >= len(hc.DRTable) {
		idx = len(hc.DRTable) - 1
	}
	return 1.0 + hc.MasteryBase*hc.DRTable[idx]
}

// ExpectedHeal computes the expected heal amount for a spell effect.
func (hc *HealCalculator) ExpectedHeal(coeff float64, hotCount int, isCrit bool) float64 {
	if hc.EffectiveSP == 0 || coeff == 0 {
		return 0
	}
	amount := hc.EffectiveSP * coeff
	amount *= 1.0 + hc.VersPercent
	amount *= hc.MasteryMult(hotCount)
	if isCrit {
		amount *= 2.0
	}
	return amount
}

// GetCoefficient returns the SP coefficient for a spell's heal type.
// Returns 0 if not found.
func (hc *HealCalculator) GetCoefficient(spellID int, isTick bool) float64 {
	if hc.Coefficients == nil {
		return 0
	}
	spell, ok := hc.Coefficients.Spells[spellID]
	if !ok {
		return 0
	}
	wantType := "direct"
	if isTick {
		wantType = "periodic"
	}
	for _, eff := range spell.Effects {
		if eff.Type == wantType {
			return eff.Coefficient
		}
	}
	return 0
}

// DecomposeHeal splits a heal amount into proportional contributions from each multiplier.
// Uses log-proportional method: each factor gets credit proportional to ln(Mi).
// Returns map of factor name -> healing amount. Sum equals healAmount.
func DecomposeHeal(healAmount float64, multipliers map[string]float64) map[string]float64 {
	result := make(map[string]float64)
	if healAmount <= 0 {
		return result
	}

	// Compute total product of all multipliers
	totalProduct := 1.0
	for _, m := range multipliers {
		if m > 0 {
			totalProduct *= m
		}
	}

	if totalProduct <= 1.0 {
		result["Spell Power"] = healAmount
		return result
	}

	base := healAmount / totalProduct
	bonus := healAmount - base
	result["Spell Power"] = base

	// Log-proportional split of the bonus
	logSum := 0.0
	for _, m := range multipliers {
		if m > 1.0 {
			logSum += math.Log(m)
		}
	}

	if logSum <= 0 {
		return result
	}

	for name, m := range multipliers {
		if m > 1.0 {
			result[name] = bonus * math.Log(m) / logSum
		}
	}

	return result
}
