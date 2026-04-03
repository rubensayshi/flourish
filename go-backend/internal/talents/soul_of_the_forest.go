package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	sotfSwiftmend    = 18562
	sotfRejuv        = 774
	sotfGermRejuv    = 155777
	sotfRegrowth     = 8936
	sotfBuff         = 114108
	sotfMultiplier   = 0.6
	sotfTag          = "sotf"
	potaTag          = "pota"
	potaWindowMS     = 500
)

type SoulOfTheForestAttributor struct {
	BaseAttributor
	primaryTarget    int
	primarySpell     int
	consumeTimestamp int
	pendingCast      *[3]int // {ts, target, spell}
}

func NewSoulOfTheForestAttributor() *SoulOfTheForestAttributor {
	return &SoulOfTheForestAttributor{
		BaseAttributor: NewBaseAttributor("SotF + PotA", intPtr(82055), nil),
	}
}

func isSotfSpell(id int) bool {
	return id == sotfRejuv || id == sotfGermRejuv || id == sotfRegrowth
}

func isRejuvID(id int) bool {
	return id == sotfRejuv || id == sotfGermRejuv
}

func (a *SoulOfTheForestAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	base := event.GetBase()

	// Track Rejuv/Regrowth casts as potential SotF consumers
	if ce, ok := event.(*models.CastEvent); ok {
		if ce.AbilityID == sotfRejuv || ce.AbilityID == sotfRegrowth {
			a.pendingCast = &[3]int{ce.Timestamp, ce.TargetID, ce.AbilityID}
		}
	}

	// SotF buff removal = consumed
	if re, ok := event.(*models.RemoveBuffEvent); ok && re.AbilityID == sotfBuff {
		if a.pendingCast != nil {
			targetID := a.pendingCast[1]
			spellID := a.pendingCast[2]
			a.primaryTarget = targetID
			a.primarySpell = spellID
			a.consumeTimestamp = re.Timestamp
			a.pendingCast = nil

			// Tag primary HoT
			if isRejuvID(spellID) {
				for _, sid := range []int{sotfRejuv, sotfGermRejuv} {
					h := hot.Get(targetID, sid)
					if h != nil && !h.Tags[sotfTag] {
						h.Tags[sotfTag] = true
						break
					}
				}
			} else {
				h := hot.Get(targetID, spellID)
				if h != nil {
					h.Tags[sotfTag] = true
				}
			}
		}
	}

	// Tag PotA spread HoTs
	switch e := event.(type) {
	case *models.ApplyBuffEvent, *models.RefreshBuffEvent:
		var abilityID, targetID, timestamp int
		if ab, ok := e.(*models.ApplyBuffEvent); ok {
			abilityID, targetID, timestamp = ab.AbilityID, ab.TargetID, ab.Timestamp
		} else {
			rb := e.(*models.RefreshBuffEvent)
			abilityID, targetID, timestamp = rb.AbilityID, rb.TargetID, rb.Timestamp
		}
		if isSotfSpell(abilityID) && a.consumeTimestamp > 0 &&
			targetID != a.primaryTarget &&
			timestamp-a.consumeTimestamp <= potaWindowMS {
			// Check spell matches primary
			if (isRejuvID(a.primarySpell) && isRejuvID(abilityID)) || abilityID == a.primarySpell {
				h := hot.Get(targetID, abilityID)
				if h != nil {
					h.Tags[sotfTag] = true
					h.Tags[potaTag] = true
				}
			}
		}
	}

	// Expire PotA window
	if a.consumeTimestamp > 0 && base.Timestamp-a.consumeTimestamp > potaWindowMS {
		a.primaryTarget = 0
		a.primarySpell = 0
		a.consumeTimestamp = 0
	}
}

func (a *SoulOfTheForestAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if !isSotfSpell(event.AbilityID) {
		return 0.0
	}

	// Regrowth direct heal before HoT is tagged
	if event.AbilityID == sotfRegrowth && buff.IsActive(sotfBuff) &&
		a.pendingCast != nil && event.TargetID == a.pendingCast[1] && a.pendingCast[2] == sotfRegrowth {
		return float64(event.Amount) - float64(event.Amount)/(1+sotfMultiplier)
	}

	h := hot.Get(event.TargetID, event.AbilityID)
	if h == nil || !h.Tags[sotfTag] {
		return 0.0
	}

	// PotA spread: 100%
	if h.Tags[potaTag] {
		return float64(event.Amount)
	}

	// Primary SotF: bonus portion
	return float64(event.Amount) - float64(event.Amount)/(1+sotfMultiplier)
}
