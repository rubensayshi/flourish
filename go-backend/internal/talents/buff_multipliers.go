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
	return newStaticBuff("Wild Synthesis", nil, nil, []int{422090}, 0.3)
}

func NewWildstalkersPowerAttributor() *StaticBuffAttributor {
	return newStaticBuff("Wildstalker's Power", nil, nil, []int{774, 155777}, 0.1)
}

func NewLifetreadingAttributor() *StaticBuffAttributor {
	return newStaticBuff("Lifetreading", intPtr(103874), nil, []int{81269}, 0.25)
}
