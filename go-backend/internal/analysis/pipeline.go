package analysis

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

type AnalysisResults struct {
	TotalHealing      int
	Wasted            int
	HighHealthHealing int
	TalentHealing     map[string]float64
	TalentRanks       map[string]int
	FightDurationMs   int
	CombatantInfo     *models.CombatantInfoEvent
}

const HighHealthThreshold = 0.8

type Pipeline struct {
	Attributors    []talents.TalentAttributor
	HotTracker     *tracking.HotTracker
	BuffTracker    *tracking.BuffTracker
	HealthTracker  *tracking.HealthTracker
	PetIDs         map[int]bool
	PlayerPetIDs   map[int]bool
}

func NewPipeline(attributors []talents.TalentAttributor, petIDs, playerPetIDs map[int]bool) *Pipeline {
	if petIDs == nil {
		petIDs = make(map[int]bool)
	}
	if playerPetIDs == nil {
		playerPetIDs = make(map[int]bool)
	}
	for _, attr := range attributors {
		attr.SetPlayerPetIDs(playerPetIDs)
	}
	return &Pipeline{
		Attributors:  attributors,
		HotTracker:   tracking.NewHotTracker(),
		BuffTracker:  tracking.NewBuffTracker(),
		PetIDs:       petIDs,
		PlayerPetIDs: playerPetIDs,
	}
}

func (p *Pipeline) Run(rawEvents []map[string]any) *AnalysisResults {
	results := &AnalysisResults{
		TalentHealing: make(map[string]float64),
		TalentRanks:   make(map[string]int),
	}
	for _, attr := range p.Attributors {
		results.TalentHealing[attr.Name()] = 0.0
	}

	// Parse events
	var events []models.Event
	for _, raw := range rawEvents {
		if e := models.ParseEvent(raw); e != nil {
			events = append(events, e)
		}
	}

	if len(events) > 0 {
		results.FightDurationMs = events[len(events)-1].GetBase().Timestamp - events[0].GetBase().Timestamp
	}

	for _, event := range events {
		// Handle combatant info
		if ci, ok := event.(*models.CombatantInfoEvent); ok {
			if results.CombatantInfo == nil {
				results.CombatantInfo = ci
				for _, attr := range p.Attributors {
					attr.SetCombatantInfo(ci)
				}
				if len(ci.TalentNodes) > 0 {
					var filtered []talents.TalentAttributor
					for _, a := range p.Attributors {
						if a.IsSelected() {
							filtered = append(filtered, a)
						}
					}
					p.Attributors = filtered
					results.TalentHealing = make(map[string]float64)
					for _, a := range p.Attributors {
						results.TalentHealing[a.Name()] = 0.0
						if rank := a.GetTalentRank(); rank != nil {
							results.TalentRanks[a.Name()] = *rank
						}
					}
				}
			}
			continue
		}

		// Update trackers
		switch event.(type) {
		case *models.ApplyBuffEvent, *models.RefreshBuffEvent, *models.RemoveBuffEvent:
			p.HotTracker.Process(event)
			p.BuffTracker.Process(event)
		}

		// Let attributors see every event
		for _, attr := range p.Attributors {
			attr.ProcessEvent(event, p.HotTracker, p.BuffTracker)
		}

		// Process heals
		if he, ok := event.(*models.HealEvent); ok {
			if p.PetIDs[he.TargetID] {
				continue
			}
			results.TotalHealing += he.Amount

			if he.IsWasted() {
				results.Wasted += he.Amount
				continue
			}

			// High-health filter: skip talent attribution for heals on targets >80% HP
			if p.HealthTracker != nil {
				healthPct := p.HealthTracker.GetHealthPct(he.TargetID, he.Timestamp)
				if healthPct > HighHealthThreshold {
					results.HighHealthHealing += he.Amount
					continue
				}
			}

			for _, attr := range p.Attributors {
				attributed := attr.ProcessHeal(he, p.HotTracker, p.BuffTracker)
				results.TalentHealing[attr.Name()] += attributed
				attr.AddTotalAttributed(attributed)
			}
		}
	}

	// Finalize
	for _, attr := range p.Attributors {
		finalized := attr.Finalize()
		results.TalentHealing[attr.Name()] += finalized
		attr.AddTotalAttributed(finalized)
	}

	return results
}
