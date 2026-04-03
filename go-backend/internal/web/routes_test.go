package web_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
	"github.com/stretchr/testify/require"
)

type MockReportClient struct {
	FailOnCode string
}

func (m *MockReportClient) GetReport(code string) (map[string]any, error) {
	if code == m.FailOnCode {
		return nil, errors.New("not found")
	}
	return map[string]any{
		"title": "Test Raid",
		"fights": []any{
			map[string]any{"id": 1, "name": "Boss", "kill": true, "startTime": 0,
				"endTime": 60000, "encounterID": 123, "difficulty": 4},
			map[string]any{"id": 2, "name": "Trash", "kill": true, "startTime": 0,
				"endTime": 30000, "encounterID": 0, "difficulty": 0},
		},
		"masterData": map[string]any{
			"actors": []any{
				map[string]any{"id": 10, "name": "Saikó", "type": "Player",
					"subType": "Druid", "server": "Draenor"},
				map[string]any{"id": 11, "name": "Warrior", "type": "Player",
					"subType": "Warrior", "server": "Draenor"},
			},
		},
	}, nil
}

func (m *MockReportClient) GetEvents(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
	return nil, nil
}

func (m *MockReportClient) GetDamageTaken(code string, fightID, sourceID int, startTime, endTime float64, filter string) (int, error) {
	return 0, nil
}

func TestReportEndpointReturnsFightsAndDruids(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/report/ABC123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var data map[string]any
	err := json.NewDecoder(w.Body).Decode(&data)
	require.NoError(t, err)
	require.Equal(t, "Test Raid", data["title"])
	fights := data["fights"].([]any)
	require.Len(t, fights, 1) // Trash filtered out (encounterID=0)
	druids := data["druids"].([]any)
	require.Len(t, druids, 1) // Only Druid, not Warrior
}

func TestReportEndpoint404OnInvalidCode(t *testing.T) {
	router := web.NewRouter(&MockReportClient{FailOnCode: "INVALID"}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/report/INVALID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 404, w.Code)
}

func TestHealthEndpoint(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
}
