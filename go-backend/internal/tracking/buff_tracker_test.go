package tracking_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

const sotfBuff = 114108

func TestBuffApplied(t *testing.T) {
	tracker := tracking.NewBuffTracker()
	tracker.Process(&models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "applybuff"},
		TargetID:  1, AbilityID: sotfBuff,
	})
	require.True(t, tracker.IsActive(sotfBuff))
}

func TestBuffRemoved(t *testing.T) {
	tracker := tracking.NewBuffTracker()
	tracker.Process(&models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "applybuff"},
		TargetID:  1, AbilityID: sotfBuff,
	})
	tracker.Process(&models.RemoveBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 200, SourceID: 1, Type: "removebuff"},
		TargetID:  1, AbilityID: sotfBuff,
	})
	require.False(t, tracker.IsActive(sotfBuff))
}

func TestBuffNotActiveByDefault(t *testing.T) {
	tracker := tracking.NewBuffTracker()
	require.False(t, tracker.IsActive(sotfBuff))
}
