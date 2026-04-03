package wcl_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateAndQuery(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]any{"access_token": "fake_token"})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"reportData": map[string]any{
						"report": map[string]any{"title": "Test"},
					},
				},
			})
		}
	}))
	defer server.Close()

	client := wcl.NewClient("test_id", "test_secret", wcl.WithBaseURL(server.URL), wcl.WithOAuthURL(server.URL))
	result, err := client.GetReport("abc123")
	require.NoError(t, err)
	require.Equal(t, "Test", result["title"])
	require.Equal(t, 2, callCount)
}

func TestGetEventsPaginates(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp map[string]any
		if callCount == 1 {
			resp = map[string]any{
				"data": map[string]any{
					"reportData": map[string]any{
						"report": map[string]any{
							"events": map[string]any{
								"data":              []any{map[string]any{"type": "heal", "timestamp": 1}},
								"nextPageTimestamp": 5000,
							},
						},
					},
				},
			}
		} else {
			resp = map[string]any{
				"data": map[string]any{
					"reportData": map[string]any{
						"report": map[string]any{
							"events": map[string]any{
								"data":              []any{map[string]any{"type": "heal", "timestamp": 5001}},
								"nextPageTimestamp": nil,
							},
						},
					},
				},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := wcl.NewClient("test_id", "test_secret",
		wcl.WithBaseURL(server.URL),
		wcl.WithOAuthURL(server.URL),
		wcl.WithToken("fake_token"),
	)
	events, err := client.GetEvents("abc", 1, 1, 0, 10000)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, 2, callCount)
}
