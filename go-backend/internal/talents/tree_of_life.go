package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	iwgNodeID    = 82045
	tickWindowMS = 200
)

type TreeOfLifeAttributor struct {
	BaseAttributor
	tolActive          bool
	wgBuffer           []*models.HealEvent
	bufferStart        int
	deferredWGHealing  float64
}

func NewTreeOfLifeAttributor() *TreeOfLifeAttributor {
	return &TreeOfLifeAttributor{
		BaseAttributor: NewBaseAttributor("Incarnation: Tree of Life", intPtr(82064), intPtr(103120)),
	}
}

func (a *TreeOfLifeAttributor) baseWGTargets() int {
	if a.HasTalent(iwgNodeID) {
		return 7
	}
	return 5
}

func (a *TreeOfLifeAttributor) flushWGBuffer() float64 {
	if len(a.wgBuffer) == 0 {
		return 0.0
	}
	targets := make(map[int]bool)
	totalHealing := 0
	for _, e := range a.wgBuffer {
		targets[e.TargetID] = true
		totalHealing += e.Amount
	}
	baseTargets := a.baseWGTargets()
	baseBuff := float64(totalHealing) - float64(totalHealing)/TolOtherDivisor
	extraShare := 0.0
	targetCount := len(targets)
	if targetCount > baseTargets {
		extraShare = float64(totalHealing) * float64(targetCount-baseTargets) / float64(targetCount)
	}
	a.wgBuffer = nil
	return baseBuff + extraShare
}

func (a *TreeOfLifeAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if ab, ok := event.(*models.ApplyBuffEvent); ok && ab.AbilityID == TreeOfLifeBuff {
		a.tolActive = true
	}
	if rb, ok := event.(*models.RemoveBuffEvent); ok && rb.AbilityID == TreeOfLifeBuff {
		a.tolActive = false
		a.deferredWGHealing += a.flushWGBuffer()
	}
}

func (a *TreeOfLifeAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if !a.tolActive {
		return 0.0
	}

	if event.AbilityID == Rejuvenation || event.AbilityID == GerminationRejuv {
		return float64(event.Amount) - float64(event.Amount)/TolRejuvDivisor
	}

	if event.AbilityID == WildGrowth {
		flushed := 0.0
		if len(a.wgBuffer) > 0 && event.Timestamp-a.bufferStart > tickWindowMS {
			flushed = a.flushWGBuffer()
		}
		if len(a.wgBuffer) == 0 {
			a.bufferStart = event.Timestamp
		}
		a.wgBuffer = append(a.wgBuffer, event)
		return flushed
	}

	return float64(event.Amount) - float64(event.Amount)/TolOtherDivisor
}

func (a *TreeOfLifeAttributor) Finalize() float64 {
	return a.deferredWGHealing + a.flushWGBuffer()
}
