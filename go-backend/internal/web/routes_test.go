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

func (m *MockReportClient) GetFightEvents(code string, fightID int, startTime, endTime float64) ([]map[string]any, error) {
	return nil, nil
}

func (m *MockReportClient) GetResources(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
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

func TestAnalyzeEndpointReturnsResults(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var data map[string]any
	err := json.NewDecoder(w.Body).Decode(&data)
	require.NoError(t, err)
	require.Equal(t, "Boss", data["fight_name"])
	require.Equal(t, "Saikó", data["player_name"])
	require.NotNil(t, data["talents"])
	require.NotNil(t, data["hero_trees"])
}

func TestAnalyzeEndpointCachesResults(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())

	// First request
	req1 := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	require.Equal(t, 200, w1.Code)

	// Second request should be cached
	req2 := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	require.Equal(t, 200, w2.Code)

	var data1, data2 map[string]any
	json.NewDecoder(w1.Body).Decode(&data1)
	json.NewDecoder(w2.Body).Decode(&data2)
	require.Equal(t, data1["total_healing"], data2["total_healing"])
}

func TestAnalyzeEndpoint404OnBadFight(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/analyze/ABC123/999/Saik%C3%B3", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 404, w.Code)
}

func TestAnalyzeEndpoint404OnBadPlayer(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Nobody", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 404, w.Code)
}
