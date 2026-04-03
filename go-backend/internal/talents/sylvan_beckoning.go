package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	dryadTranq    = 1264659
	dryadRegrowth = 1264664
)

var dryadSpells = map[int]bool{dryadTranq: true, dryadRegrowth: true}

// SylvanBeckoningAttributor attributes healing from Dryad pet casts.
type SylvanBeckoningAttributor struct {
	BaseAttributor
}

func NewSylvanBeckoningAttributor() *SylvanBeckoningAttributor {
	return &SylvanBeckoningAttributor{
		BaseAttributor: NewBaseAttributor("Sylvan Beckoning", intPtr(109714), nil),
	}
}

func (a *SylvanBeckoningAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if !dryadSpells[event.AbilityID] {
		return 0.0
	}
	if a.IsPlayerPet(event.GetBase().SourceID) {
		return float64(event.Amount)
	}
	return 0.0
}
