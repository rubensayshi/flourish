package tracking_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

const (
	rejuvID = 774
	targetA = 10
)

func makeApplyBuff(ts, target, ability int) *models.ApplyBuffEvent {
	return &models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: ts, SourceID: 1, Type: "applybuff"},
		TargetID:  target, AbilityID: ability,
	}
}

func makeRemoveBuff(ts, target, ability int) *models.RemoveBuffEvent {
	return &models.RemoveBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: ts, SourceID: 1, Type: "removebuff"},
		TargetID:  target, AbilityID: ability,
	}
}

func makeRefreshBuff(ts, target, ability int) *models.RefreshBuffEvent {
	return &models.RefreshBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: ts, SourceID: 1, Type: "refreshbuff"},
		TargetID:  target, AbilityID: ability,
	}
}

func TestApplyCreatesHot(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	hot := tracker.Get(targetA, rejuvID)
	require.NotNil(t, hot)
	require.Equal(t, rejuvID, hot.SpellID)
	require.Equal(t, 100, hot.AppliedAt)
}

func TestRemoveClearsHot(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	tracker.Process(makeRemoveBuff(200, targetA, rejuvID))
	require.Nil(t, tracker.Get(targetA, rejuvID))
}

func TestRefreshUpdatesTimestamp(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	tracker.Process(makeRefreshBuff(200, targetA, rejuvID))
	hot := tracker.Get(targetA, rejuvID)
	require.Equal(t, 100, hot.AppliedAt)
	require.Equal(t, 200, hot.LastRefresh)
}

func TestTagsClearedOnRefresh(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	hot := tracker.Get(targetA, rejuvID)
	hot.Tags["sotf"] = true
	tracker.Process(makeRefreshBuff(200, targetA, rejuvID))
	require.False(t, tracker.Get(targetA, rejuvID).Tags["sotf"])
}

func TestGetAllHotsOnTarget(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, 774))
	tracker.Process(makeApplyBuff(100, targetA, 33763))
	hots := tracker.GetAll(targetA)
	require.Len(t, hots, 2)
}
