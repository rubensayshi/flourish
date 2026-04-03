package wcl_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
	"github.com/stretchr/testify/require"
)

// MockWCLClient implements wcl.Querier for testing.
type MockWCLClient struct {
	GetReportResult map[string]any
	GetEventsResult []map[string]any
	ReportCalled    bool
	EventsCalled    bool
}

func (m *MockWCLClient) GetReport(code string) (map[string]any, error) {
	m.ReportCalled = true
	return m.GetReportResult, nil
}

func (m *MockWCLClient) GetEvents(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
	m.EventsCalled = true
	return m.GetEventsResult, nil
}

func (m *MockWCLClient) GetFightEvents(code string, fightID int, startTime, endTime float64) ([]map[string]any, error) {
	return m.GetEventsResult, nil
}

func (m *MockWCLClient) GetResources(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
	return nil, nil
}

func (m *MockWCLClient) GetDamageTaken(code string, fightID, sourceID int, startTime, endTime float64, filter string) (int, error) {
	return 0, nil
}

func TestGetReportCachesToDisk(t *testing.T) {
	inner := &MockWCLClient{
		GetReportResult: map[string]any{"title": "Test Report", "fights": []any{}},
	}
	client := wcl.NewCachedClient(inner, t.TempDir())

	result, err := client.GetReport("ABC123")
	require.NoError(t, err)
	require.Equal(t, "Test Report", result["title"])
}

func TestGetReportReadsFromCache(t *testing.T) {
	inner := &MockWCLClient{
		GetReportResult: map[string]any{"title": "Test Report"},
	}
	dir := t.TempDir()
	client := wcl.NewCachedClient(inner, dir)

	_, err := client.GetReport("ABC123")
	require.NoError(t, err)

	inner2 := &MockWCLClient{}
	client2 := wcl.NewCachedClient(inner2, dir)
	result, err := client2.GetReport("ABC123")
	require.NoError(t, err)
	require.Equal(t, "Test Report", result["title"])
	require.False(t, inner2.ReportCalled)
}

func TestGetEventsCachesToDisk(t *testing.T) {
	inner := &MockWCLClient{
		GetEventsResult: []map[string]any{{"type": "heal", "timestamp": 1}},
	}
	client := wcl.NewCachedClient(inner, t.TempDir())

	result, err := client.GetEvents("ABC123", 1, 5, 0, 10000)
	require.NoError(t, err)
	require.Len(t, result, 1)
}

func TestGetEventsReadsFromCache(t *testing.T) {
	inner := &MockWCLClient{
		GetEventsResult: []map[string]any{{"type": "heal", "timestamp": 1}},
	}
	dir := t.TempDir()
	client := wcl.NewCachedClient(inner, dir)

	_, err := client.GetEvents("ABC123", 1, 5, 0, 10000)
	require.NoError(t, err)

	inner2 := &MockWCLClient{}
	client2 := wcl.NewCachedClient(inner2, dir)
	result, err := client2.GetEvents("ABC123", 1, 5, 0, 10000)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.False(t, inner2.EventsCalled)
}
