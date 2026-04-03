package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	iwgWildGrowth  = 48438
	iwgTolBuff     = 33891
	iwgBaseTargets = 5
	iwgExtraTargets = 2
)

type ImprovedWildGrowthAttributor struct {
	BaseAttributor
}

func NewImprovedWildGrowthAttributor() *ImprovedWildGrowthAttributor {
	return &ImprovedWildGrowthAttributor{
		BaseAttributor: NewBaseAttributor("Improved Wild Growth", intPtr(82045), nil),
	}
}

func (a *ImprovedWildGrowthAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != iwgWildGrowth {
		return 0.0
	}
	if buff.IsActive(iwgTolBuff) {
		return 0.0
	}
	totalTargets := iwgBaseTargets + iwgExtraTargets
	return float64(event.Amount) * float64(iwgExtraTargets) / float64(totalTargets)
}
