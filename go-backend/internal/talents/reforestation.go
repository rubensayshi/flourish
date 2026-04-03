package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	refSwiftmend       = 18562
	refTolBuff         = 33891
	refTolDurationMS   = 10000
)

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
	if ab, ok := event.(*models.ApplyBuffEvent); ok && ab.AbilityID == refTolBuff {
		a.realTolActive = true
	}
	if rb, ok := event.(*models.RemoveBuffEvent); ok && rb.AbilityID == refTolBuff {
		a.realTolActive = false
	}
	if ce, ok := event.(*models.CastEvent); ok && ce.AbilityID == refSwiftmend {
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
	if event.AbilityID == sotfRejuv || event.AbilityID == sotfGermRejuv {
		return float64(event.Amount) - float64(event.Amount)/1.5
	}
	return float64(event.Amount) - float64(event.Amount)/1.1
}
