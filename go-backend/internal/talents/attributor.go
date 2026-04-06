package talents

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

// MultiplierProvider is implemented by attributors that apply a multiplicative
// bonus to healing. GetMultiplier returns the multiplier (e.g. 1.2 for +20%).
// Returns 1.0 when the talent does not affect this heal event.
// Used by the decomposition engine for log-proportional allocation.
type MultiplierProvider interface {
	GetMultiplier(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64
}

// TalentAttributor is the interface all talent attributors implement.
type TalentAttributor interface {
	Name() string
	TalentNodeID() *int
	TalentID() *int
	SetCombatantInfo(info *models.CombatantInfoEvent)
	SetPlayerPetIDs(ids map[int]bool)
	IsSelected() bool
	GetTalentRank() *int
	HasTalent(nodeID int) bool
	IsPlayerPet(sourceID int) bool
	ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker)
	ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64
	Finalize() float64
	GetTotalAttributed() float64
	AddTotalAttributed(amount float64)
}

// BaseAttributor provides default implementations for TalentAttributor.
type BaseAttributor struct {
	name            string
	talentNodeID    *int
	talentID        *int
	CombatantInfo   *models.CombatantInfoEvent
	TotalAttributedVal float64
	PlayerPetIDs    map[int]bool
}

func NewBaseAttributor(name string, nodeID *int, talentID *int) BaseAttributor {
	return BaseAttributor{
		name:         name,
		talentNodeID: nodeID,
		talentID:     talentID,
		PlayerPetIDs: make(map[int]bool),
	}
}

// NewBaseAttributorPtr returns a *BaseAttributor that satisfies TalentAttributor.
func NewBaseAttributorPtr(name string, nodeID *int, talentID *int) *BaseAttributor {
	b := NewBaseAttributor(name, nodeID, talentID)
	return &b
}

func (a *BaseAttributor) Name() string       { return a.name }
func (a *BaseAttributor) TalentNodeID() *int { return a.talentNodeID }
func (a *BaseAttributor) TalentID() *int     { return a.talentID }

func (a *BaseAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.CombatantInfo = info
}

func (a *BaseAttributor) SetPlayerPetIDs(ids map[int]bool) {
	a.PlayerPetIDs = ids
}

func (a *BaseAttributor) IsSelected() bool {
	if a.CombatantInfo == nil {
		return true
	}
	if a.talentNodeID == nil {
		return true
	}
	if !a.CombatantInfo.TalentNodes[*a.talentNodeID] {
		return false
	}
	if a.talentID != nil {
		return a.CombatantInfo.TalentIDs[*a.talentID]
	}
	return true
}

func (a *BaseAttributor) GetTalentRank() *int {
	if a.CombatantInfo != nil && a.talentID != nil {
		if rank, ok := a.CombatantInfo.TalentRanks[*a.talentID]; ok {
			return &rank
		}
	}
	return nil
}

func (a *BaseAttributor) HasTalent(nodeID int) bool {
	if a.CombatantInfo != nil {
		return a.CombatantInfo.TalentNodes[nodeID]
	}
	return false
}

func (a *BaseAttributor) IsPlayerPet(sourceID int) bool {
	return a.PlayerPetIDs[sourceID]
}

func (a *BaseAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
}

func (a *BaseAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	return 0.0
}

func (a *BaseAttributor) Finalize() float64 {
	return 0.0
}

func (a *BaseAttributor) GetTotalAttributed() float64 {
	return a.TotalAttributedVal
}

func (a *BaseAttributor) AddTotalAttributed(amount float64) {
	a.TotalAttributedVal += amount
}
