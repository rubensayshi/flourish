package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

// DirectSpellAttributor attributes all effective healing from specific spell IDs.
type DirectSpellAttributor struct {
	BaseAttributor
	SpellIDs       map[int]bool
	AllowPetSource bool
}

func newDirectSpell(name string, nodeID *int, talentID *int, spellIDs []int) *DirectSpellAttributor {
	m := make(map[int]bool)
	for _, id := range spellIDs {
		m[id] = true
	}
	return &DirectSpellAttributor{
		BaseAttributor: NewBaseAttributor(name, nodeID, talentID),
		SpellIDs:       m,
	}
}

func (a *DirectSpellAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if a.SpellIDs[event.AbilityID] {
		if !a.AllowPetSource && a.IsPlayerPet(event.GetBase().SourceID) {
			return 0.0
		}
		return float64(event.Amount)
	}
	return 0.0
}

func intPtr(v int) *int { return &v }

func NewEverbloomSplashAttributor() *DirectSpellAttributor {
	return newDirectSpell("Everbloom: Splash", intPtr(110424), intPtr(137039), []int{1244341})
}

func NewEfflorescenceAttributor() *DirectSpellAttributor {
	return newDirectSpell("Efflorescence", intPtr(82057), nil, []int{81269})
}

func NewVerdancyAttributor() *DirectSpellAttributor {
	return newDirectSpell("Verdancy", intPtr(82059), nil, []int{392329})
}

func NewGroveGuardiansAttributor() *DirectSpellAttributor {
	a := newDirectSpell("Grove Guardians", intPtr(82043), nil, []int{422090})
	a.AllowPetSource = true
	return a
}

func NewDreamSurgeAttributor() *DirectSpellAttributor {
	return newDirectSpell("Dream Surge", nil, nil, []int{434141})
}

func NewCultivationAttributor() *DirectSpellAttributor {
	return newDirectSpell("Cultivation", intPtr(82056), nil, []int{200390})
}

func NewFlourishAttributor() *DirectSpellAttributor {
	return newDirectSpell("Flourish", intPtr(82053), intPtr(103106), []int{1264659})
}

// RampantGrowthAttributor credits bonus portion of Regrowth HoT ticks (+100%).
type RampantGrowthAttributor struct {
	BaseAttributor
}

func NewRampantGrowthAttributor() *RampantGrowthAttributor {
	return &RampantGrowthAttributor{
		BaseAttributor: NewBaseAttributor("Rampant Growth", intPtr(82058), nil),
	}
}

func (a *RampantGrowthAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == 8936 && event.Tick {
		return float64(event.Amount) - float64(event.Amount)/2.0
	}
	return 0.0
}
