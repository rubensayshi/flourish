package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

// StaticBuffAttributor attributes bonus from flat percentage buffs.
// bonus = amount - amount / (1 + multiplier)
type StaticBuffAttributor struct {
	BaseAttributor
	SpellIDs   map[int]bool
	Multiplier float64
}

func newStaticBuff(name string, nodeID *int, talentID *int, spellIDs []int, multiplier float64) *StaticBuffAttributor {
	m := make(map[int]bool)
	for _, id := range spellIDs {
		m[id] = true
	}
	return &StaticBuffAttributor{
		BaseAttributor: NewBaseAttributor(name, nodeID, talentID),
		SpellIDs:       m,
		Multiplier:     multiplier,
	}
}

func (a *StaticBuffAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.SpellIDs[event.AbilityID] && a.Multiplier > 0 {
		return float64(event.Amount) - float64(event.Amount)/(1+a.Multiplier)
	}
	return 0.0
}

func NewWildSynthesisAttributor() *StaticBuffAttributor {
	return newStaticBuff("Wild Synthesis", intPtr(94535), nil, []int{422090, 142421, 81269, 434141}, 0.3)
}

func NewWildstalkersPowerAttributor() *StaticBuffAttributor {
	return newStaticBuff("Wildstalker's Power", intPtr(94621), nil, []int{774, 155777, 81269, 33763, 33778}, 0.1)
}

func NewLifetreadingAttributor() *StaticBuffAttributor {
	return newStaticBuff("Lifetreading", intPtr(103874), nil, []int{81269}, 0.25)
}

func NewGrovesInspirationAttributor() *StaticBuffAttributor {
	return newStaticBuff("Grove's Inspiration", intPtr(94595), intPtr(117189), []int{8936, 1264664, 48438, 18562}, 0.09)
}

func NewCenariusMightAttributor() *StaticBuffAttributor {
	return newStaticBuff("Cenarius' Might", intPtr(94604), nil, []int{18562}, 0.2)
}

func NewBountifulBloomAttributor() *StaticBuffAttributor {
	return newStaticBuff("Bounteous Bloom", intPtr(94591), intPtr(117184), []int{422090, 142421}, 0.3)
}

func NewPatientCustodianAttributor() *StaticBuffAttributor {
	return newStaticBuff("Patient Custodian", intPtr(94630), nil, []int{774, 155777, 8936, 48438, 33763, 33778, 1244341, 1264664}, 0.06)
}

// IntensityAttributor: Regrowth crits at 260% instead of 200%.
// On Regrowth crits, bonus = amount - amount/1.3
type IntensityAttributor struct {
	BaseAttributor
}

var intensitySpells = map[int]bool{8936: true, 1264664: true} // Regrowth + Rampant Growth Regrowth

func NewIntensityAttributor() *IntensityAttributor {
	return &IntensityAttributor{
		BaseAttributor: NewBaseAttributor("Intensity", intPtr(82052), nil),
	}
}

func (a *IntensityAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if intensitySpells[event.AbilityID] && event.HitType == 2 {
		return float64(event.Amount) - float64(event.Amount)/1.3
	}
	return 0.0
}
