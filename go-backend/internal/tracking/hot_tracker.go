package tracking

import "github.com/rdruid-talent-analyzer/go-backend/internal/models"

type HotInstance struct {
	SpellID     int
	TargetID    int
	AppliedAt   int
	LastRefresh int
	Tags        map[string]bool
}

type HotTracker struct {
	hots map[[2]int]*HotInstance // key: {targetID, spellID}
}

func NewHotTracker() *HotTracker {
	return &HotTracker{hots: make(map[[2]int]*HotInstance)}
}

func (t *HotTracker) Process(event models.Event) {
	switch e := event.(type) {
	case *models.ApplyBuffEvent:
		key := [2]int{e.TargetID, e.AbilityID}
		t.hots[key] = &HotInstance{
			SpellID:   e.AbilityID,
			TargetID:  e.TargetID,
			AppliedAt: e.Timestamp,
			Tags:      make(map[string]bool),
		}
	case *models.RefreshBuffEvent:
		key := [2]int{e.TargetID, e.AbilityID}
		if existing, ok := t.hots[key]; ok {
			existing.LastRefresh = e.Timestamp
			existing.Tags = make(map[string]bool)
		} else {
			t.hots[key] = &HotInstance{
				SpellID:     e.AbilityID,
				TargetID:    e.TargetID,
				AppliedAt:   e.Timestamp,
				LastRefresh: e.Timestamp,
				Tags:        make(map[string]bool),
			}
		}
	case *models.RemoveBuffEvent:
		key := [2]int{e.TargetID, e.AbilityID}
		delete(t.hots, key)
	}
}

func (t *HotTracker) Get(targetID, spellID int) *HotInstance {
	return t.hots[[2]int{targetID, spellID}]
}

func (t *HotTracker) GetAll(targetID int) []*HotInstance {
	var result []*HotInstance
	for k, h := range t.hots {
		if k[0] == targetID {
			result = append(result, h)
		}
	}
	return result
}

// CountByTarget returns the number of active HoTs on a target whose spell IDs
// are in the provided set. Pass nil to count all.
func (t *HotTracker) CountByTarget(targetID int, spellIDs map[int]bool) int {
	count := 0
	for k := range t.hots {
		if k[0] == targetID && (spellIDs == nil || spellIDs[k[1]]) {
			count++
		}
	}
	return count
}

func (t *HotTracker) GetAllBySpell(spellID int) []*HotInstance {
	var result []*HotInstance
	for k, h := range t.hots {
		if k[1] == spellID {
			result = append(result, h)
		}
	}
	return result
}
