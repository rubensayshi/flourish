package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

var defaultDRTable = []float64{1.0, 1.7, 2.3, 2.8, 3.2}

const lifebloom = 33763

// HarmoniousBloomingAttributor attributes mastery bonus from Lifebloom counting as 3 stacks instead of 1.
type HarmoniousBloomingAttributor struct {
	BaseAttributor
	mastery    float64
	baseStacks int
	drTable    []float64
	fraction   float64
}

func NewHarmoniousBloomingAttributor(baseStacks int, drTable []float64) *HarmoniousBloomingAttributor {
	if drTable == nil {
		drTable = defaultDRTable
	}
	a := &HarmoniousBloomingAttributor{
		BaseAttributor: NewBaseAttributor("Harmonious Blooming", intPtr(82077), nil),
		mastery:        0.25,
		baseStacks:     baseStacks,
		drTable:        drTable,
	}
	a.fraction = a.computeFraction()
	return a
}

func (a *HarmoniousBloomingAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	if info != nil && info.Mastery > 0 {
		a.mastery = info.Mastery / 100.0
		a.fraction = a.computeFraction()
	}
}

func (a *HarmoniousBloomingAttributor) computeFraction() float64 {
	n := a.baseStacks
	table := a.drTable
	nWith := n + 2
	if nWith > len(table)-1 {
		nWith = len(table) - 1
	}
	if n < 1 || n >= len(table) {
		return 0.0
	}
	multBase := 1.0 + a.mastery*table[n]
	multWith := 1.0 + a.mastery*table[nWith]
	return 1.0 - multBase/multWith
}

func (a *HarmoniousBloomingAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == lifebloom {
		return 0.0
	}
	if hot.Get(event.TargetID, lifebloom) != nil {
		return float64(event.Amount) * a.fraction
	}
	return 0.0
}

// SymbioticBloomMasteryAttributor attributes mastery bonus from Symbiotic Bloom adding an extra HoT stack.
type SymbioticBloomMasteryAttributor struct {
	BaseAttributor
	mastery    float64
	baseStacks int
	drTable    []float64
	fraction   float64
}

const symbioticBloomSpell = 439530

func NewSymbioticBloomMasteryAttributor(baseStacks int, drTable []float64) *SymbioticBloomMasteryAttributor {
	if drTable == nil {
		drTable = defaultDRTable
	}
	a := &SymbioticBloomMasteryAttributor{
		BaseAttributor: NewBaseAttributor("Symbiotic Bloom Mastery", intPtr(94626), nil),
		mastery:        0.25,
		baseStacks:     baseStacks,
		drTable:        drTable,
	}
	a.fraction = a.computeFraction()
	return a
}

func (a *SymbioticBloomMasteryAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	if info != nil && info.Mastery > 0 {
		a.mastery = info.Mastery / 100.0
		a.fraction = a.computeFraction()
	}
}

func (a *SymbioticBloomMasteryAttributor) computeFraction() float64 {
	n := a.baseStacks
	table := a.drTable
	if n < 1 || n >= len(table) {
		return 0.0
	}
	multBase := 1.0 + a.mastery*table[n-1]
	multWith := 1.0 + a.mastery*table[n]
	return 1.0 - multBase/multWith
}

func (a *SymbioticBloomMasteryAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == symbioticBloomSpell {
		return 0.0
	}
	if hot.Get(event.TargetID, symbioticBloomSpell) != nil {
		return float64(event.Amount) * a.fraction
	}
	return 0.0
}
