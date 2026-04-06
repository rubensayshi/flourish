package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	refTolBaseDurationMS     = 10000
	refTolExtendedDurationMS = 16000
	PotentEnchantmentsNode   = 94595
	PotentEnchantmentsTalent = 117188
)

type ReforestationAttributor struct {
	BaseAttributor
	smCount          int
	reforestationStart int
	reforestationEnd   int
	realTolActive    bool
	potentEnch       *PotentEnchantmentsAttributor
}

func NewReforestationAttributor(potentEnch *PotentEnchantmentsAttributor) *ReforestationAttributor {
	return &ReforestationAttributor{
		BaseAttributor: NewBaseAttributor("Reforestation", intPtr(82069), nil),
		potentEnch:     potentEnch,
	}
}

func (a *ReforestationAttributor) durationMS() int {
	if a.HasTalent(PotentEnchantmentsNode) {
		return refTolExtendedDurationMS
	}
	return refTolBaseDurationMS
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
			a.reforestationStart = ce.Timestamp
			a.reforestationEnd = ce.Timestamp + a.durationMS()
		}
	}
}

func (a *ReforestationAttributor) GetMultiplier(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.realTolActive {
		return 1.0
	}
	if event.Timestamp < a.reforestationStart || event.Timestamp > a.reforestationEnd {
		return 1.0
	}
	if event.AbilityID == Rejuvenation || event.AbilityID == GerminationRejuv {
		return TolRejuvDivisor
	}
	return TolOtherDivisor
}

func (a *ReforestationAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.realTolActive {
		return 0.0
	}
	if event.Timestamp > a.reforestationEnd {
		return 0.0
	}
	if event.Timestamp < a.reforestationStart {
		return 0.0
	}

	var bonus float64
	if event.AbilityID == Rejuvenation || event.AbilityID == GerminationRejuv {
		bonus = float64(event.Amount) - float64(event.Amount)/TolRejuvDivisor
	} else {
		bonus = float64(event.Amount) - float64(event.Amount)/TolOtherDivisor
	}

	// Healing in the extended window (10-16s) goes to Potent Enchantments
	if a.potentEnch != nil && event.Timestamp >= a.reforestationStart+refTolBaseDurationMS {
		a.potentEnch.AddHealing(bonus)
		return 0.0
	}

	return bonus
}
