package analysis

import (
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

type AnalysisResults struct {
	TotalHealing        int
	Wasted              int
	HighHealthHealing   int
	HighHealthThreshold float64 // the threshold used (0 = not tracked)
	TalentHealing       map[string]float64
	TalentRanks         map[string]int
	StatHealing         map[string]float64 // Versatility, Mastery: Harmony, Critical Strike
	BaseSpellHealing    map[string]float64 // per-spell unattributed (base) healing
	FightDurationMs     int
	CombatantInfo       *models.CombatantInfoEvent
}

type Pipeline struct {
	Attributors         []talents.TalentAttributor
	HotTracker          *tracking.HotTracker
	BuffTracker         *tracking.BuffTracker
	HealthTracker       *tracking.HealthTracker
	HighHealthThreshold float64 // heals on targets above this % are excluded (1.0 = disabled)
	PetIDs              map[int]bool
	PlayerPetIDs        map[int]bool
	DRTable             []float64

	versPercent float64
	masteryBase float64
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
		TalentHealing:       make(map[string]float64),
		TalentRanks:         make(map[string]int),
		StatHealing:         make(map[string]float64),
		BaseSpellHealing:    make(map[string]float64),
		HighHealthThreshold: p.HighHealthThreshold,
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
				p.versPercent = ci.Versatility / VersRatingPerPercent / 100.0
				p.masteryBase = ci.Mastery / 100.0 / 100.0

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

			// High-health filter
			if p.HealthTracker != nil && p.HighHealthThreshold < 1.0 {
				healthPct := p.HealthTracker.GetHealthPct(he.TargetID, he.Timestamp)
				if healthPct > p.HighHealthThreshold {
					results.HighHealthHealing += he.Amount
					continue
				}
			}

			amount := float64(he.Amount)

			// Talent attribution (counterfactual: "what would I lose")
			for _, attr := range p.Attributors {
				attributed := attr.ProcessHeal(he, p.HotTracker, p.BuffTracker)
				results.TalentHealing[attr.Name()] += attributed
				attr.AddTotalAttributed(attributed)
			}

			// Stat attribution (counterfactual on full heal amount)
			if p.versPercent > 0 {
				results.StatHealing["Versatility"] += amount - amount/(1.0+p.versPercent)
			}

			hotCount := p.HotTracker.CountByTarget(he.TargetID, talents.MasteryHoTs)
			if hotCount > 0 && p.masteryBase > 0 && len(p.DRTable) > 0 {
				idx := hotCount
				if idx >= len(p.DRTable) {
					idx = len(p.DRTable) - 1
				}
				masteryMult := 1.0 + p.masteryBase*p.DRTable[idx]
				results.StatHealing["Mastery: Harmony"] += amount - amount/masteryMult
			}

			if he.HitType == 2 {
				results.StatHealing["Critical Strike"] += amount / 2.0
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
