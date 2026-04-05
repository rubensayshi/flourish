package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	ndBaseRejuvMS    = 12000
	lingeringHealMS  = 3000
	germinationDurMS = 2000
	lingeringHealNode = 82240
	germinationNode   = 82071
)

type NurturingDormancyAttributor struct {
	BaseAttributor
	baseDurationMS int
	pandemicCapMS  int
	// expectedExpiry tracks the non-ND expiry per {targetID, spellID}.
	expectedExpiry map[[2]int]int
}

func NewNurturingDormancyAttributor() *NurturingDormancyAttributor {
	dur := ndBaseRejuvMS + lingeringHealMS + germinationDurMS // default 17s
	return &NurturingDormancyAttributor{
		BaseAttributor: NewBaseAttributor("Nurturing Dormancy", intPtr(82076), nil),
		baseDurationMS: dur,
		pandemicCapMS:  dur * 3 / 10,
		expectedExpiry: make(map[[2]int]int),
	}
}

func (a *NurturingDormancyAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	dur := ndBaseRejuvMS
	if a.HasTalent(lingeringHealNode) {
		dur += lingeringHealMS
	}
	if a.HasTalent(germinationNode) {
		dur += germinationDurMS
	}
	a.baseDurationMS = dur
	a.pandemicCapMS = dur * 3 / 10 // 30% pandemic cap
}

func (a *NurturingDormancyAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	switch e := event.(type) {
	case *models.ApplyBuffEvent:
		if IsRejuv(e.AbilityID) {
			key := [2]int{e.TargetID, e.AbilityID}
			a.expectedExpiry[key] = e.Timestamp + a.baseDurationMS
		}
	case *models.RefreshBuffEvent:
		if IsRejuv(e.AbilityID) {
			key := [2]int{e.TargetID, e.AbilityID}
			remaining := 0
			if exp, ok := a.expectedExpiry[key]; ok && exp > e.Timestamp {
				remaining = exp - e.Timestamp
			}
			pandemicBonus := remaining
			if pandemicBonus > a.pandemicCapMS {
				pandemicBonus = a.pandemicCapMS
			}
			a.expectedExpiry[key] = e.Timestamp + a.baseDurationMS + pandemicBonus
		}
	case *models.RemoveBuffEvent:
		if IsRejuv(e.AbilityID) {
			delete(a.expectedExpiry, [2]int{e.TargetID, e.AbilityID})
		}
	}
}

func (a *NurturingDormancyAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if !IsRejuv(event.AbilityID) {
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
