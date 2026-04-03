package tracking

import "github.com/rdruid-talent-analyzer/go-backend/internal/models"

type BuffTracker struct {
	active map[int]int // buffID -> timestamp
}

func NewBuffTracker() *BuffTracker {
	return &BuffTracker{active: make(map[int]int)}
}

func (t *BuffTracker) Process(event models.Event) {
	switch e := event.(type) {
	case *models.ApplyBuffEvent:
		t.active[e.AbilityID] = e.Timestamp
	case *models.RefreshBuffEvent:
		t.active[e.AbilityID] = e.Timestamp
	case *models.RemoveBuffEvent:
		delete(t.active, e.AbilityID)
	}
}

func (t *BuffTracker) IsActive(buffID int) bool {
	_, ok := t.active[buffID]
	return ok
}

func (t *BuffTracker) GetAppliedAt(buffID int) *int {
	ts, ok := t.active[buffID]
	if !ok {
		return nil
	}
	return &ts
}
