package talents

import (
	"math"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	photoWindowMS    = 200
	photoEverbloomMS = 1200
)

type PhotosynthesisAttributor struct {
	BaseAttributor
	bloomEvents      []bloomEvent
	lbTransitions    [][2]int // {timestamp, targetID}
	lbCasts          [][2]int
	sotfConsumptions []int
}

type bloomEvent struct {
	timestamp int
	targetID  int
	amount    float64
}

func NewPhotosynthesisAttributor() *PhotosynthesisAttributor {
	return &PhotosynthesisAttributor{
		BaseAttributor: NewBaseAttributor("Photosynthesis", intPtr(82073), nil),
	}
}

func (a *PhotosynthesisAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	switch e := event.(type) {
	case *models.RemoveBuffEvent:
		if e.AbilityID == Lifebloom {
			a.lbTransitions = append(a.lbTransitions, [2]int{e.Timestamp, e.TargetID})
		} else if e.AbilityID == SoulOfTheForestBuff {
			a.sotfConsumptions = append(a.sotfConsumptions, e.Timestamp)
		}
	case *models.RefreshBuffEvent:
		if e.AbilityID == Lifebloom {
			a.lbTransitions = append(a.lbTransitions, [2]int{e.Timestamp, e.TargetID})
		}
	case *models.CastEvent:
		if e.AbilityID == Lifebloom {
			a.lbCasts = append(a.lbCasts, [2]int{e.Timestamp, e.TargetID})
		}
	}
}

func (a *PhotosynthesisAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == LifebloomBloom {
		a.bloomEvents = append(a.bloomEvents, bloomEvent{event.Timestamp, event.TargetID, float64(event.Amount)})
	}
	return 0.0
}

func (a *PhotosynthesisAttributor) Finalize() float64 {
	total := 0.0
	for _, b := range a.bloomEvents {
		explained := false

		for _, t := range a.lbTransitions {
			if t[1] == b.targetID && math.Abs(float64(t[0]-b.timestamp)) < photoWindowMS {
				explained = true
				break
			}
		}

		if !explained {
			for _, c := range a.lbCasts {
				if c[1] == b.targetID && math.Abs(float64(c[0]-b.timestamp)) < photoWindowMS {
					explained = true
					break
				}
			}
		}

		if !explained {
			for _, sts := range a.sotfConsumptions {
				diff := b.timestamp - sts
				if diff >= 0 && diff <= photoEverbloomMS {
					explained = true
					break
				}
			}
		}

		if !explained {
			total += b.amount
		}
	}
	return total
}
