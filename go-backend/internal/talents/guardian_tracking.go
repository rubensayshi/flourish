package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	groveGuardianBaseDurationMs = 8000
	guardianDurabilityBonus     = 0.2
	durabilityOfNatureNode      = 94605
	durabilityOfNatureTalentID  = 117200
	harmonyBonusPerGuardian     = 0.05
	powerOfNatureBonus          = 0.10
)

// guardianTracker tracks active Grove Guardian count via summon events.
type guardianTracker struct {
	guardianCount      int
	despawnTimes       []int
	guardianDurationMs int
}

func (g *guardianTracker) initGuardianDuration(info *models.CombatantInfoEvent) {
	g.guardianDurationMs = groveGuardianBaseDurationMs
	if info != nil && info.TalentNodes[durabilityOfNatureNode] && info.TalentIDs[durabilityOfNatureTalentID] {
		g.guardianDurationMs = int(float64(groveGuardianBaseDurationMs) * (1 + guardianDurabilityBonus))
	}
}

func (g *guardianTracker) updateGuardians(event models.Event) {
	ts := event.GetBase().Timestamp
	// Remove expired guardians
	alive := g.despawnTimes[:0]
	for _, t := range g.despawnTimes {
		if t > ts {
			alive = append(alive, t)
		}
	}
	g.despawnTimes = alive
	g.guardianCount = len(g.despawnTimes)

	// Track new summons
	if e, ok := event.(*models.SummonEvent); ok && e.AbilityID == GroveGuardianSummon {
		g.despawnTimes = append(g.despawnTimes, ts+g.guardianDurationMs)
		// Recount after adding
		count := 0
		for _, t := range g.despawnTimes {
			if t > ts {
				count++
			}
		}
		g.guardianCount = count
	}
}

// HarmonyOfTheGroveAttributor: each Grove Guardian increases healing done by 5%.
type HarmonyOfTheGroveAttributor struct {
	BaseAttributor
	guardianTracker
}

func NewHarmonyOfTheGroveAttributor() *HarmonyOfTheGroveAttributor {
	return &HarmonyOfTheGroveAttributor{
		BaseAttributor:  NewBaseAttributor("Harmony of the Grove", intPtr(94606), nil),
		guardianTracker: guardianTracker{guardianDurationMs: groveGuardianBaseDurationMs},
	}
}

func (a *HarmonyOfTheGroveAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	a.guardianTracker.initGuardianDuration(info)
}

func (a *HarmonyOfTheGroveAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	a.guardianTracker.updateGuardians(event)
}

func (a *HarmonyOfTheGroveAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.guardianCount <= 0 {
		return 0.0
	}
	multiplier := harmonyBonusPerGuardian * float64(a.guardianCount)
	return float64(event.Amount) - float64(event.Amount)/(1+multiplier)
}

// PowerOfNatureAttributor: Grove Guardians increase Rejuv, Efflorescence, and Lifebloom healing by 10%.
type PowerOfNatureAttributor struct {
	BaseAttributor
	guardianTracker
}

var powerOfNatureSpells = map[int]bool{Rejuvenation: true, GerminationRejuv: true, Efflorescence: true, Lifebloom: true, LifebloomBloom: true}

func NewPowerOfNatureAttributor() *PowerOfNatureAttributor {
	return &PowerOfNatureAttributor{
		BaseAttributor:  NewBaseAttributor("Power of Nature", intPtr(94605), intPtr(117201)),
		guardianTracker: guardianTracker{guardianDurationMs: groveGuardianBaseDurationMs},
	}
}

func (a *PowerOfNatureAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	a.guardianTracker.initGuardianDuration(info)
}

func (a *PowerOfNatureAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	a.guardianTracker.updateGuardians(event)
}

func (a *PowerOfNatureAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.guardianCount <= 0 || !powerOfNatureSpells[event.AbilityID] {
		return 0.0
	}
	return float64(event.Amount) - float64(event.Amount)/(1+powerOfNatureBonus)
}
