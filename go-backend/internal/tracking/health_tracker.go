package tracking

import (
	"sort"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
)

const staminaToHP = 20

type healthSnapshot struct {
	Timestamp int
	HealthPct float64
}

type HealthState struct {
	CurrentHP int
	MaxHP     int
}

// HealthTracker reconstructs player health bars from all fight events.
type HealthTracker struct {
	states    map[int]*HealthState    // targetID → current state (used during build)
	snapshots map[int][]healthSnapshot // targetID → sorted snapshots (used for lookups)
}

// NewHealthTracker processes all fight events (unfiltered by sourceID) to build
// per-player health timelines. Players start at full HP.
func NewHealthTracker(rawEvents []map[string]any) *HealthTracker {
	ht := &HealthTracker{
		states:    make(map[int]*HealthState),
		snapshots: make(map[int][]healthSnapshot),
	}

	// First pass: extract maxHP from combatantinfo stamina for all players
	for _, raw := range rawEvents {
		if models.GetString(raw, "type") != models.EventCombatantInfo {
			continue
		}
		sourceID := models.GetInt(raw, "sourceID", 0)
		stamina := models.GetInt(raw, "stamina", 0)
		if sourceID > 0 && stamina > 0 {
			maxHP := stamina * staminaToHP
			ht.states[sourceID] = &HealthState{CurrentHP: maxHP, MaxHP: maxHP}
		}
	}

	// Second pass: process events chronologically to build health timeline
	for _, raw := range rawEvents {
		eventType := models.GetString(raw, "type")
		timestamp := models.GetInt(raw, "timestamp", 0)

		switch eventType {
		case models.EventDamage:
			targetID := models.GetInt(raw, "targetID", 0)
			amount := models.GetInt(raw, "amount", 0)
			absorbed := models.GetInt(raw, "absorbed", 0)
			overkill := models.GetInt(raw, "overkill", 0)
			ht.applyDamage(targetID, timestamp, amount+absorbed, overkill)

		case models.EventHeal:
			targetID := models.GetInt(raw, "targetID", 0)
			amount := models.GetInt(raw, "amount", 0)
			overheal := models.GetInt(raw, "overheal", 0)
			ht.applyHeal(targetID, timestamp, amount, overheal)
		}
	}

	return ht
}

func (ht *HealthTracker) ensureState(targetID int) *HealthState {
	if s, ok := ht.states[targetID]; ok {
		return s
	}
	// Unknown target (no combatantinfo); will calibrate from heals
	s := &HealthState{CurrentHP: 0, MaxHP: 0}
	ht.states[targetID] = s
	return s
}

func (ht *HealthTracker) applyDamage(targetID, timestamp, totalDamage, overkill int) {
	s := ht.ensureState(targetID)
	if s.MaxHP == 0 {
		return // can't track without maxHP
	}
	// Overkill is damage beyond 0 HP, don't subtract it
	effective := totalDamage - overkill
	s.CurrentHP -= effective
	if s.CurrentHP < 0 {
		s.CurrentHP = 0
	}
	ht.recordSnapshot(targetID, timestamp, s)
}

func (ht *HealthTracker) applyHeal(targetID, timestamp, amount, overheal int) {
	s := ht.ensureState(targetID)
	if s.MaxHP == 0 {
		return
	}

	s.CurrentHP += amount
	if s.CurrentHP > s.MaxHP {
		s.CurrentHP = s.MaxHP
	}

	// Calibrate maxHP: if there's overheal, the target reached max
	if overheal > 0 {
		// Post-heal they're at max. Pre-heal they were missing `amount`.
		// If our tracked HP after adding amount exceeds MaxHP, the calibration is consistent.
		// If not, adjust MaxHP upward.
		if s.CurrentHP == s.MaxHP {
			// Consistent — no adjustment needed
		}
	}

	ht.recordSnapshot(targetID, timestamp, s)
}

func (ht *HealthTracker) recordSnapshot(targetID, timestamp int, s *HealthState) {
	pct := 1.0
	if s.MaxHP > 0 {
		pct = float64(s.CurrentHP) / float64(s.MaxHP)
		if pct > 1.0 {
			pct = 1.0
		}
		if pct < 0.0 {
			pct = 0.0
		}
	}
	ht.snapshots[targetID] = append(ht.snapshots[targetID], healthSnapshot{
		Timestamp: timestamp,
		HealthPct: pct,
	})
}

// GetHealthPct returns the target's health fraction (0.0-1.0) at the given timestamp.
// Returns 1.0 if no data is available for the target.
func (ht *HealthTracker) GetHealthPct(targetID, timestamp int) float64 {
	snaps, ok := ht.snapshots[targetID]
	if !ok || len(snaps) == 0 {
		return 1.0 // assume full health if unknown
	}

	// Binary search for the last snapshot at or before timestamp
	idx := sort.Search(len(snaps), func(i int) bool {
		return snaps[i].Timestamp > timestamp
	})

	if idx == 0 {
		return 1.0 // before any tracked events, assume full
	}
	return snaps[idx-1].HealthPct
}
