package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
	"github.com/stretchr/testify/require"
)

func TestLoginRedirectsToWCL(t *testing.T) {
	t.Setenv("WCL_CLIENT_ID", "test-client-id")
	router := web.NewRouterWithAuth(&MockReportClient{}, t.TempDir(), web.NewAuthState())
	req := httptest.NewRequest("GET", "/api/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, 302, w.Code)
	loc := w.Header().Get("Location")
	require.Contains(t, loc, "warcraftlogs.com/oauth/authorize")
	require.Contains(t, loc, "test-client-id")
}

func TestCallbackMissingCode(t *testing.T) {
	router := web.NewRouterWithAuth(&MockReportClient{}, t.TempDir(), web.NewAuthState())
	req := httptest.NewRequest("GET", "/api/auth/callback", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 400, w.Code)
}

func TestCallbackInvalidState(t *testing.T) {
	router := web.NewRouterWithAuth(&MockReportClient{}, t.TempDir(), web.NewAuthState())
	req := httptest.NewRequest("GET", "/api/auth/callback?code=abc&state=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 400, w.Code)
}

func TestAnonLimitEnforced(t *testing.T) {
	authState := web.NewAuthState()
	router := web.NewRouterWithAuth(&MockReportClient{}, t.TempDir(), authState)

	// Use base_stacks param to bypass cache
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3?base_stacks=2", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code, "request %d should succeed", i+1)
	}

	// Third request should be blocked
	req := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3?base_stacks=2", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 403, w.Code)
}

func TestAuthenticatedBypassesAnonLimit(t *testing.T) {
	authState := web.NewAuthState()
	router := web.NewRouterWithAuth(&MockReportClient{}, t.TempDir(), authState)

	// Exhaust anonymous limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
	}

	// Authenticated request should still work
	req := httptest.NewRequest("GET", "/api/analyze/ABC123/1/Saik%C3%B3", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	req.Header.Set("Authorization", "Bearer some-user-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Will fail at WCL API level (mock doesn't handle user tokens), but won't be 403
	// The mock returns data for any code, so it should be 200
	require.NotEqual(t, 403, w.Code)
}

func TestAuthStateCSRF(t *testing.T) {
	state := web.NewAuthState()
	state.AddState("abc123")
	require.True(t, state.ConsumeState("abc123"))
	require.False(t, state.ConsumeState("abc123")) // can't reuse
	require.False(t, state.ConsumeState("unknown"))
}

func TestAuthStateAnonTracking(t *testing.T) {
	state := web.NewAuthState()
	require.True(t, state.CheckAnonLimit("1.2.3.4"))
	state.RecordAnonUsage("1.2.3.4")
	require.True(t, state.CheckAnonLimit("1.2.3.4"))
	state.RecordAnonUsage("1.2.3.4")
	require.False(t, state.CheckAnonLimit("1.2.3.4"))
	// Different IP still has quota
	require.True(t, state.CheckAnonLimit("5.6.7.8"))
}

func TestCallbackWithError(t *testing.T) {
	router := web.NewRouterWithAuth(&MockReportClient{}, t.TempDir(), web.NewAuthState())
	req := httptest.NewRequest("GET", "/api/auth/callback?error=access_denied", nil)
	req.Host = "localhost:8000"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 302, w.Code)
	require.Contains(t, w.Header().Get("Location"), "auth_error=access_denied")
}

func TestGetUserToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	require.Equal(t, "", web.GetUserToken(req))

	req.Header.Set("Authorization", "Bearer mytoken123")
	require.Equal(t, "mytoken123", web.GetUserToken(req))

	req.Header.Set("Authorization", "Basic abc")
	require.Equal(t, "", web.GetUserToken(req))
}
