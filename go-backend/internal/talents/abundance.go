package talents

import (
	"math"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	abundRejuv          = 774
	abundGermRejuv      = 155777
	abundRegrowth       = 8936
	abundCritPerRejuv   = 0.08
	abundCritRatingPPct = 700.0
)

type AbundanceAttributor struct {
	BaseAttributor
}

func NewAbundanceAttributor() *AbundanceAttributor {
	return &AbundanceAttributor{
		BaseAttributor: NewBaseAttributor("Abundance", intPtr(103876), nil),
	}
}

func (a *AbundanceAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != abundRegrowth || event.HitType != 2 {
		return 0.0
	}

	rejuvCount := len(hot.GetAllBySpell(abundRejuv)) + len(hot.GetAllBySpell(abundGermRejuv))
	if rejuvCount <= 0 {
		return 0.0
	}

	abundanceCrit := math.Min(float64(rejuvCount)*abundCritPerRejuv, 0.96)

	baseCrit := 0.0
	if a.CombatantInfo != nil {
		baseCrit = a.CombatantInfo.CritSpell / abundCritRatingPPct
	}
	baseCrit = math.Max(baseCrit, 0.05)

	totalCrit := math.Min(baseCrit+abundanceCrit, 1.0)
	abundanceShare := abundanceCrit / totalCrit

	critBonus := float64(event.Amount) / 2.0
	return critBonus * abundanceShare
}
