package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

// Mastery: Harmony DR table — cumulative mastery coefficient per HoT stack count.
// Increments: 1.0, 0.7, 0.6, 0.5, 0.4, 0.4, 0.4, 0.4, ...  (plateaus at 0.4)
var defaultDRTable = []float64{0, 1.0, 1.7, 2.3, 2.8, 3.2, 3.6, 4.0, 4.4, 4.8}

// MasteryHoTs is the set of HoTs that count as mastery stacks for Mastery: Harmony.
var MasteryHoTs = map[int]bool{
	Rejuvenation:     true,
	GerminationRejuv: true,
	Regrowth:         true,
	WildGrowth:       true,
	Lifebloom:        true,
	CenarionWard:     true,
	SymbioticBloomSpell: true,
	CultivationSpell: true,
}

// masteryFraction computes the fraction of healing attributable to going from
// `without` mastery stacks to `with` mastery stacks, capped to the DR table bounds.
func masteryFraction(mastery float64, drTable []float64, without, with int) float64 {
	maxIdx := len(drTable) - 1
	if without < 0 {
		without = 0
	}
	if without > maxIdx {
		without = maxIdx
	}
	if with > maxIdx {
		with = maxIdx
	}
	if without >= with {
		return 0.0
	}
	multBase := 1.0 + mastery*drTable[without]
	multWith := 1.0 + mastery*drTable[with]
	return 1.0 - multBase/multWith
}

// HarmoniousBloomingAttributor attributes mastery bonus from Lifebloom counting as 3 stacks instead of 1.
type HarmoniousBloomingAttributor struct {
	BaseAttributor
	mastery float64
	drTable []float64
}

func NewHarmoniousBloomingAttributor(drTable []float64) *HarmoniousBloomingAttributor {
	if drTable == nil {
		drTable = defaultDRTable
	}
	return &HarmoniousBloomingAttributor{
		BaseAttributor: NewBaseAttributor("Harmonious Blooming", intPtr(82077), nil),
		mastery:        0.25,
		drTable:        drTable,
	}
}

func (a *HarmoniousBloomingAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	if info != nil && info.Mastery > 0 {
		a.mastery = info.Mastery / 100.0
	}
}

func (a *HarmoniousBloomingAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == Lifebloom {
		return 0.0
	}
	if hot.Get(event.TargetID, Lifebloom) == nil {
		return 0.0
	}
	// Count only mastery-eligible HoTs on this target (LB counted as 1 in tracker)
	stacks := hot.CountByTarget(event.TargetID, MasteryHoTs)
	// HB makes LB count as 3 instead of 1 → without HB: stacks, with HB: stacks+2
	fraction := masteryFraction(a.mastery, a.drTable, stacks, stacks+2)
	return float64(event.Amount) * fraction
}

// SymbioticBloomMasteryAttributor attributes mastery bonus from Symbiotic Bloom adding an extra HoT stack.
type SymbioticBloomMasteryAttributor struct {
	BaseAttributor
	mastery  float64
	drTable  []float64
	hbActive bool
}

func NewSymbioticBloomMasteryAttributor(drTable []float64) *SymbioticBloomMasteryAttributor {
	if drTable == nil {
		drTable = defaultDRTable
	}
	return &SymbioticBloomMasteryAttributor{
		BaseAttributor: NewBaseAttributor("Symbiotic Bloom Mastery", intPtr(94626), nil),
		mastery:        0.25,
		drTable:        drTable,
	}
}

// SetHBActive marks whether Harmonious Blooming is active in this build,
// so SB can account for LB's +2 virtual mastery stacks.
func (a *SymbioticBloomMasteryAttributor) SetHBActive(active bool) {
	a.hbActive = active
}

func (a *SymbioticBloomMasteryAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	if info != nil && info.Mastery > 0 {
		a.mastery = info.Mastery / 100.0
	}
}

func (a *SymbioticBloomMasteryAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == SymbioticBloomSpell {
		return 0.0
	}
	if hot.Get(event.TargetID, SymbioticBloomSpell) == nil {
		return 0.0
	}
	// Count only mastery-eligible HoTs on target (includes SB itself)
	stacks := hot.CountByTarget(event.TargetID, MasteryHoTs)
	// If HB is active and LB is on this target, LB counts as 3 instead of 1 (+2 virtual)
	if a.hbActive && hot.Get(event.TargetID, Lifebloom) != nil {
		stacks += 2
	}
	// Without SB: stacks-1, with SB: stacks
	fraction := masteryFraction(a.mastery, a.drTable, stacks-1, stacks)
	return float64(event.Amount) * fraction
}
