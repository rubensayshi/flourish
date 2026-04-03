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
	return newStaticBuff("Wild Synthesis", intPtr(94535), nil, []int{GroveGuardianNourish, GroveGuardianHeal, Efflorescence, DreamSurge}, 0.3)
}

func NewWildstalkersPowerAttributor() *StaticBuffAttributor {
	return newStaticBuff("Wildstalker's Power", intPtr(94621), nil, []int{Rejuvenation, GerminationRejuv, Efflorescence, Lifebloom, LifebloomBloom}, 0.1)
}

func NewLifetreadingAttributor() *StaticBuffAttributor {
	return newStaticBuff("Lifetreading", intPtr(103874), nil, []int{Efflorescence}, 0.25)
}

func NewGrovesInspirationAttributor() *StaticBuffAttributor {
	return newStaticBuff("Grove's Inspiration", intPtr(94595), intPtr(117189), []int{Regrowth, DryadRegrowthSpell, WildGrowth, Swiftmend}, 0.09)
}

func NewCenariusMightAttributor() *StaticBuffAttributor {
	return newStaticBuff("Cenarius' Might", intPtr(94604), nil, []int{Swiftmend}, 0.2)
}

func NewBountifulBloomAttributor() *StaticBuffAttributor {
	return newStaticBuff("Bounteous Bloom", intPtr(94591), intPtr(117184), []int{GroveGuardianNourish, GroveGuardianHeal}, 0.3)
}

func NewPatientCustodianAttributor() *StaticBuffAttributor {
	return newStaticBuff("Patient Custodian", intPtr(94630), nil, []int{Rejuvenation, GerminationRejuv, Regrowth, WildGrowth, Lifebloom, LifebloomBloom, EverbloomSplash, DryadRegrowthSpell}, 0.06)
}

func NewImprovedSwiftmendAttributor() *StaticBuffAttributor {
	return newStaticBuff("Improved Swiftmend", intPtr(82063), nil, []int{Swiftmend}, 0.3)
}

func NewUnstoppableGrowthAttributor() *StaticBuffAttributor {
	// Reduced falloff ≈ 27.7% average bonus
	return newStaticBuff("Unstoppable Growth", intPtr(82061), nil, []int{WildGrowth}, 0.277)
}

func NewLivelinessAttributor() *StaticBuffAttributor {
	return newStaticBuff("Liveliness", intPtr(82064), intPtr(103130),
		[]int{Rejuvenation, GerminationRejuv, Regrowth, DryadRegrowthSpell, WildGrowth, Lifebloom, LifebloomBloom, EverbloomSplash}, 0.05)
}

func NewRegenesisAttributor() *StaticBuffAttributor {
	// Approximation of conditional low-health bonus as flat 15%
	return newStaticBuff("Regenesis", intPtr(82062), nil, []int{Rejuvenation, GerminationRejuv, CenarionWard, DryadTranquility}, 0.15)
}

// IntensityAttributor: Regrowth crits at 260% instead of 200%.
// On Regrowth crits, bonus = amount - amount/1.3
type IntensityAttributor struct {
	BaseAttributor
}

var intensitySpells = map[int]bool{Regrowth: true, DryadRegrowthSpell: true}

func NewIntensityAttributor() *IntensityAttributor {
	return &IntensityAttributor{
		BaseAttributor: NewBaseAttributor("Intensity", intPtr(82052), nil),
	}
}

// intensityCritDivisor: crits are 260% instead of 200%, so bonus = amount - amount/1.3
const intensityCritDivisor = 1.3

func (a *IntensityAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if intensitySpells[event.AbilityID] && event.HitType == 2 {
		return float64(event.Amount) - float64(event.Amount)/intensityCritDivisor
	}
	return 0.0
}
