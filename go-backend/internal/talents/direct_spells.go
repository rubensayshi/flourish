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

// GroveGuardiansAttributor attributes treant pet healing, dividing out
// Wild Synthesis and Bounteous Bloom multipliers to avoid double-counting.
type GroveGuardiansAttributor struct {
	DirectSpellAttributor
	divisor float64
}

func NewGroveGuardiansAttributor() *GroveGuardiansAttributor {
	base := newDirectSpell("Grove Guardians", intPtr(82043), nil, []int{422090, 142421})
	base.AllowPetSource = true
	return &GroveGuardiansAttributor{DirectSpellAttributor: *base, divisor: 1.0}
}

const (
	wildSynthesisNode    = 94535
	bounteousBloomNode   = 94591
	bounteousBloomEntry  = 117184
)

func (a *GroveGuardiansAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	a.divisor = 1.0
	if a.HasTalent(wildSynthesisNode) {
		a.divisor *= 1.3
	}
	if info != nil && info.TalentNodes[bounteousBloomNode] && info.TalentIDs[bounteousBloomEntry] {
		a.divisor *= 1.3
	}
}

func (a *GroveGuardiansAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	base := a.DirectSpellAttributor.ProcessHeal(event, hot, buff)
	if base > 0 && a.divisor > 1.0 {
		return base / a.divisor
	}
	return base
}

func NewDreamSurgeAttributor() *DirectSpellAttributor {
	return newDirectSpell("Dream Surge", intPtr(94600), nil, []int{434141})
}

func NewSpiritOfTheThicketAttributor() *DirectSpellAttributor {
	a := newDirectSpell("Spirit of the Thicket", intPtr(109712), nil, []int{1264905})
	a.AllowPetSource = true
	return a
}

func NewBurstingGrowthAttributor() *DirectSpellAttributor {
	return newDirectSpell("Bursting Growth", intPtr(109716), nil, []int{440121})
}

func NewThrivingGrowthAttributor() *DirectSpellAttributor {
	return newDirectSpell("Thriving Growth", intPtr(94626), nil, []int{474760})
}

func NewNaturesBountyAttributor() *DirectSpellAttributor {
	return newDirectSpell("Nature's Bounty", intPtr(82072), nil, []int{1264376})
}

func NewRegenerativeHeartwoodAttributor() *DirectSpellAttributor {
	return newDirectSpell("Regenerative Heartwood", intPtr(82075), nil, []int{392117})
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
