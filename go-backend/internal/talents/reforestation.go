package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const refTolDurationMS = 10000

type ReforestationAttributor struct {
	BaseAttributor
	smCount           int
	reforestationEnd  int
	realTolActive     bool
}

func NewReforestationAttributor() *ReforestationAttributor {
	return &ReforestationAttributor{
		BaseAttributor: NewBaseAttributor("Reforestation", intPtr(82069), nil),
	}
}

func (a *ReforestationAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if ab, ok := event.(*models.ApplyBuffEvent); ok && ab.AbilityID == TreeOfLifeBuff {
		a.realTolActive = true
	}
	if rb, ok := event.(*models.RemoveBuffEvent); ok && rb.AbilityID == TreeOfLifeBuff {
		a.realTolActive = false
	}
	if ce, ok := event.(*models.CastEvent); ok && ce.AbilityID == Swiftmend {
		a.smCount++
		if a.smCount%4 == 0 && !a.realTolActive {
			a.reforestationEnd = ce.Timestamp + refTolDurationMS
		}
	}
}

func (a *ReforestationAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.realTolActive {
		return 0.0
	}
	if event.Timestamp > a.reforestationEnd {
		return 0.0
	}
	if event.AbilityID == Rejuvenation || event.AbilityID == GerminationRejuv {
		return float64(event.Amount) - float64(event.Amount)/TolRejuvDivisor
	}
	return float64(event.Amount) - float64(event.Amount)/TolOtherDivisor
}
