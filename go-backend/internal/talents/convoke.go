package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	convokeSpellID       = 391528
	convokeLegacyID      = 323764
	convokeDurationMS    = 4000
	convokeDurationCGMS  = 3000
	cenariusGuidanceNode = 82063
	defaultHealingRatio  = 0.7
	convokeTag           = "convoke"
)

type ConvokeAttributor struct {
	BaseAttributor
	channelEnd   int
	healingRatio float64
}

func NewConvokeAttributor(healingRatio float64) *ConvokeAttributor {
	if healingRatio == 0 {
		healingRatio = defaultHealingRatio
	}
	return &ConvokeAttributor{
		BaseAttributor: NewBaseAttributor("Convoke the Spirits", intPtr(82064), intPtr(103119)),
		healingRatio:   healingRatio,
	}
}

func (a *ConvokeAttributor) channelDuration() int {
	if a.HasTalent(cenariusGuidanceNode) {
		return convokeDurationCGMS
	}
	return convokeDurationMS
}

func (a *ConvokeAttributor) isChanneling(ts int) bool {
	return a.channelEnd > 0 && ts <= a.channelEnd
}

func (a *ConvokeAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if ce, ok := event.(*models.CastEvent); ok {
		if ce.AbilityID == convokeSpellID || ce.AbilityID == convokeLegacyID {
			a.channelEnd = ce.Timestamp + a.channelDuration()
		}
	}

	if ab, ok := event.(*models.ApplyBuffEvent); ok {
		if a.isChanneling(ab.Timestamp) {
			h := hot.Get(ab.TargetID, ab.AbilityID)
			if h != nil {
				h.Tags[convokeTag] = true
			}
		}
	}
}

func (a *ConvokeAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	h := hot.Get(event.TargetID, event.AbilityID)

	if h != nil && h.Tags[convokeTag] {
		return float64(event.Amount) * a.healingRatio
	}

	if a.isChanneling(event.Timestamp) && h == nil {
		return float64(event.Amount) * a.healingRatio
	}

	return 0.0
}
