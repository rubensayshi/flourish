package talents

import (
	"math"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	wsImplantTag           = "implant"
	wsImplantWindowMS      = 500
	wsTwinSproutsTag       = "twin_sprouts"
	wsTwinSproutsWindowMS  = 50
	wsSICritBonus          = 0.04
	vigorousCreepers       = 1.2
	rootNetworkPerBloom    = 0.02
)

// VigorousCreepersAttributor: +20% healing to targets with Symbiotic Bloom.
type VigorousCreepersAttributor struct {
	BaseAttributor
}

func NewVigorousCreepersAttributor() *VigorousCreepersAttributor {
	return &VigorousCreepersAttributor{
		BaseAttributor: NewBaseAttributor("Vigorous Creepers", intPtr(94627), nil),
	}
}

func (a *VigorousCreepersAttributor) GetMultiplier(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == SymbioticBloomSpell {
		return 1.0
	}
	if hot.Get(event.TargetID, SymbioticBloomSpell) != nil {
		return vigorousCreepers
	}
	return 1.0
}

func (a *VigorousCreepersAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == SymbioticBloomSpell {
		return 0.0
	}
	if hot.Get(event.TargetID, SymbioticBloomSpell) != nil {
		return float64(event.Amount) - float64(event.Amount)/vigorousCreepers
	}
	return 0.0
}

// ImplantAttributor: SM/WG spawns Symbiotic Bloom.
type ImplantAttributor struct {
	BaseAttributor
	recentCasts []int
}

func NewImplantAttributor() *ImplantAttributor {
	return &ImplantAttributor{
		BaseAttributor: NewBaseAttributor("Implant", intPtr(94628), intPtr(117229)),
	}
}

func (a *ImplantAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if ce, ok := event.(*models.CastEvent); ok {
		if ce.AbilityID == Swiftmend || ce.AbilityID == WildGrowth {
			a.recentCasts = append(a.recentCasts, ce.Timestamp)
			// Clean old entries
			var cleaned []int
			for _, t := range a.recentCasts {
				if ce.Timestamp-t < wsImplantWindowMS*2 {
					cleaned = append(cleaned, t)
				}
			}
			a.recentCasts = cleaned
		}
	}

	if ab, ok := event.(*models.ApplyBuffEvent); ok && ab.AbilityID == SymbioticBloomSpell {
		for _, ts := range a.recentCasts {
			if ab.Timestamp-ts < wsImplantWindowMS {
				h := hot.Get(ab.TargetID, SymbioticBloomSpell)
				if h != nil {
					h.Tags[wsImplantTag] = true
				}
				break
			}
		}
	}
}

func (a *ImplantAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != SymbioticBloomSpell {
		return 0.0
	}
	h := hot.Get(event.TargetID, SymbioticBloomSpell)
	if h != nil && h.Tags[wsImplantTag] {
		return float64(event.Amount)
	}
	return 0.0
}

// TwinSproutsAttributor: 30% chance for bloom growth to spawn another on a nearby target.
// Detected by finding bloom applications within 50ms of another bloom on a different target.
// WCL has no distinct event for the proc, so we use timestamp proximity as a heuristic.
type TwinSproutsAttributor struct {
	BaseAttributor
	lastBloomTime   int
	lastBloomTarget int
}

func NewTwinSproutsAttributor() *TwinSproutsAttributor {
	return &TwinSproutsAttributor{
		BaseAttributor: NewBaseAttributor("Twin Sprouts", intPtr(94628), intPtr(117230)),
	}
}

func (a *TwinSproutsAttributor) processBloomEvent(timestamp, targetID int, hot *tracking.HotTracker) {
	if a.lastBloomTime > 0 && targetID != a.lastBloomTarget {
		if timestamp-a.lastBloomTime <= wsTwinSproutsWindowMS {
			h := hot.Get(targetID, SymbioticBloomSpell)
			if h != nil {
				h.Tags[wsTwinSproutsTag] = true
			}
		}
	}
	a.lastBloomTime = timestamp
	a.lastBloomTarget = targetID
}

func (a *TwinSproutsAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	switch e := event.(type) {
	case *models.ApplyBuffEvent:
		if e.AbilityID == SymbioticBloomSpell {
			a.processBloomEvent(e.Timestamp, e.TargetID, hot)
		}
	case *models.RefreshBuffEvent:
		if e.AbilityID == SymbioticBloomSpell {
			a.processBloomEvent(e.Timestamp, e.TargetID, hot)
		}
	}
}

func (a *TwinSproutsAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID != SymbioticBloomSpell {
		return 0.0
	}
	h := hot.Get(event.TargetID, SymbioticBloomSpell)
	if h != nil && h.Tags[wsTwinSproutsTag] {
		return float64(event.Amount)
	}
	return 0.0
}

// RootNetworkAttributor: +2% healing per active Symbiotic Bloom.
type RootNetworkAttributor struct {
	BaseAttributor
}

func NewRootNetworkAttributor() *RootNetworkAttributor {
	return &RootNetworkAttributor{
		BaseAttributor: NewBaseAttributor("Root Network", intPtr(94631), intPtr(117233)),
	}
}

func (a *RootNetworkAttributor) GetMultiplier(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	bloomCount := len(hot.GetAllBySpell(SymbioticBloomSpell))
	if bloomCount <= 0 {
		return 1.0
	}
	return 1.0 + rootNetworkPerBloom*float64(bloomCount)
}

func (a *RootNetworkAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	bloomCount := len(hot.GetAllBySpell(SymbioticBloomSpell))
	if bloomCount <= 0 {
		return 0.0
	}
	multiplier := rootNetworkPerBloom * float64(bloomCount)
	return float64(event.Amount) - float64(event.Amount)/(1+multiplier)
}

// StrategicInfusionAttributor: +4% crit on periodic heals.
type StrategicInfusionAttributor struct {
	BaseAttributor
}

func NewStrategicInfusionAttributor() *StrategicInfusionAttributor {
	return &StrategicInfusionAttributor{
		BaseAttributor: NewBaseAttributor("Strategic Infusion", intPtr(94623), nil),
	}
}

func (a *StrategicInfusionAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if !event.Tick || event.HitType != 2 {
		return 0.0
	}

	baseCrit := 0.0
	if a.CombatantInfo != nil {
		baseCrit = a.CombatantInfo.CritSpell / CritRatingPerPercent
	}
	baseCrit = math.Max(baseCrit, 0.05)

	totalCrit := baseCrit + wsSICritBonus
	infusionShare := wsSICritBonus / totalCrit

	critBonus := float64(event.Amount) / 2.0
	return critBonus * infusionShare
}
