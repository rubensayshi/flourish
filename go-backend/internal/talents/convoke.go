package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	convokeDurationMS    = 4000
	convokeDurationCGMS  = 3000
	cenariusGuidanceNode = 82063
	convokeTag           = "convoke"
	convokeSkipCount     = 3 // first N rejuv/regrowth casts are opportunity cost
)

type ConvokeAttributor struct {
	BaseAttributor
	channelEnd     int
	skippableCasts int // rejuv/regrowth casts seen during current channel
}

func NewConvokeAttributor() *ConvokeAttributor {
	return &ConvokeAttributor{
		BaseAttributor: NewBaseAttributor("Convoke the Spirits", intPtr(82064), intPtr(103119)),
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

func isSkippableConvokeSpell(abilityID int) bool {
	return abilityID == Rejuvenation || abilityID == GerminationRejuv || abilityID == Regrowth
}

func (a *ConvokeAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if ce, ok := event.(*models.CastEvent); ok {
		if ce.AbilityID == ConvokeTheSpirits || ce.AbilityID == ConvokeLegacy {
			a.channelEnd = ce.Timestamp + a.channelDuration()
			a.skippableCasts = 0
		}
	}

	if ab, ok := event.(*models.ApplyBuffEvent); ok {
		if a.isChanneling(ab.Timestamp) {
			// Skip first N rejuv/regrowth applications (opportunity cost)
			if isSkippableConvokeSpell(ab.AbilityID) {
				a.skippableCasts++
				if a.skippableCasts <= convokeSkipCount {
					return
				}
			}
			h := hot.Get(ab.TargetID, ab.AbilityID)
			if h != nil {
				h.Tags[convokeTag] = true
			}
		}
	}
}

func (a *ConvokeAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	h := hot.Get(event.TargetID, event.AbilityID)

	// Tagged HoT tick — this HoT was applied by Convoke (and not skipped)
	if h != nil && h.Tags[convokeTag] {
		return float64(event.Amount)
	}

	// Direct heal during channel (no HoT, e.g. Swiftmend, Wild Growth initial)
	if a.isChanneling(event.Timestamp) && h == nil {
		// Skip skippable direct heals that correspond to skipped casts
		if isSkippableConvokeSpell(event.AbilityID) {
			return 0
		}
		return float64(event.Amount)
	}

	return 0
}
