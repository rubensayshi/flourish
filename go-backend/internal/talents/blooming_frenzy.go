package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	frenzyWindowMs  = 1500
	frenzyMaxBlooms = 5
)

// BloomingFrenzyAttributor: Everbloom rank 4 — LB blooms 5 times rapidly when SotF is consumed.
type BloomingFrenzyAttributor struct {
	BaseAttributor
	frenzyStart *int
	frenzyCount int
}

func NewBloomingFrenzyAttributor() *BloomingFrenzyAttributor {
	return &BloomingFrenzyAttributor{
		BaseAttributor: NewBaseAttributor("Everbloom: Blooming Frenzy", intPtr(110424), intPtr(137038)),
	}
}

func (a *BloomingFrenzyAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if e, ok := event.(*models.RemoveBuffEvent); ok && e.AbilityID == SoulOfTheForestBuff {
		ts := e.Timestamp
		a.frenzyStart = &ts
		a.frenzyCount = 0
	}
	// Expire window
	if a.frenzyStart != nil {
		ts := event.GetBase().Timestamp
		if ts-*a.frenzyStart > frenzyWindowMs {
			a.frenzyStart = nil
			a.frenzyCount = 0
		}
	}
}

func (a *BloomingFrenzyAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != LifebloomBloom {
		return 0.0
	}
	if a.frenzyStart == nil {
		return 0.0
	}
	if event.Timestamp-*a.frenzyStart > frenzyWindowMs {
		return 0.0
	}
	if a.frenzyCount >= frenzyMaxBlooms {
		return 0.0
	}
	a.frenzyCount++
	return float64(event.Amount)
}
