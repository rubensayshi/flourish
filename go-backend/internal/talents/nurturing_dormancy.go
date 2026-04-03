package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	ndRejuv           = 774
	ndGermRejuv       = 155777
	ndBaseDurationMS  = 17000
)

type NurturingDormancyAttributor struct {
	BaseAttributor
}

func NewNurturingDormancyAttributor() *NurturingDormancyAttributor {
	return &NurturingDormancyAttributor{
		BaseAttributor: NewBaseAttributor("Nurturing Dormancy", intPtr(82076), nil),
	}
}

func (a *NurturingDormancyAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != ndRejuv && event.AbilityID != ndGermRejuv {
		return 0.0
	}
	h := hot.Get(event.TargetID, event.AbilityID)
	if h == nil {
		return 0.0
	}
	baseTime := h.AppliedAt
	if h.LastRefresh > 0 {
		baseTime = h.LastRefresh
	}
	elapsed := event.Timestamp - baseTime
	if elapsed > ndBaseDurationMS {
		return float64(event.Amount)
	}
	return 0.0
}
