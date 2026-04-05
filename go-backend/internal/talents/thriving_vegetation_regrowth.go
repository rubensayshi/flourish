package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	baseRegrowthHotMS = 12000
	regrowthPandemicCapMS = baseRegrowthHotMS * 3 / 10 // 3600ms
)

// ThrivingVegetationRegrowthAttributor attributes Regrowth HoT ticks that
// occur beyond the base duration, i.e. only in the window added by
// Thriving Vegetation's "+3 sec per rank" extension.
//
// On refresh: if the non-TV expiry has already passed, the refresh is treated
// as a fresh apply (no pandemic carry-over) because without TV the HoT would
// have fallen off.
type ThrivingVegetationRegrowthAttributor struct {
	BaseAttributor
	// expectedExpiry tracks the non-TV expiry per {targetID, spellID}.
	expectedExpiry map[[2]int]int
}

func NewThrivingVegetationRegrowthAttributor() *ThrivingVegetationRegrowthAttributor {
	return &ThrivingVegetationRegrowthAttributor{
		BaseAttributor: NewBaseAttributor("Thriving Vegetation: Regrowth", intPtr(82068), nil),
		expectedExpiry: make(map[[2]int]int),
	}
}

func (a *ThrivingVegetationRegrowthAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	switch e := event.(type) {
	case *models.ApplyBuffEvent:
		if e.AbilityID == Regrowth {
			key := [2]int{e.TargetID, e.AbilityID}
			a.expectedExpiry[key] = e.Timestamp + baseRegrowthHotMS
		}
	case *models.RefreshBuffEvent:
		if e.AbilityID == Regrowth {
			key := [2]int{e.TargetID, e.AbilityID}
			exp, ok := a.expectedExpiry[key]
			if !ok || e.Timestamp >= exp {
				// Without TV the HoT would have expired; treat as fresh apply.
				a.expectedExpiry[key] = e.Timestamp + baseRegrowthHotMS
			} else {
				remaining := exp - e.Timestamp
				pandemicBonus := remaining
				if pandemicBonus > regrowthPandemicCapMS {
					pandemicBonus = regrowthPandemicCapMS
				}
				a.expectedExpiry[key] = e.Timestamp + baseRegrowthHotMS + pandemicBonus
			}
		}
	case *models.RemoveBuffEvent:
		if e.AbilityID == Regrowth {
			delete(a.expectedExpiry, [2]int{e.TargetID, e.AbilityID})
		}
	}
}

func (a *ThrivingVegetationRegrowthAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != Regrowth || !event.Tick {
		return 0.0
	}
	key := [2]int{event.TargetID, event.AbilityID}
	exp, ok := a.expectedExpiry[key]
	if !ok {
		return 0.0
	}
	if event.Timestamp > exp {
		return float64(event.Amount)
	}
	return 0.0
}
