package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"crypto/rand"
	"encoding/base64"

	"github.com/go-chi/chi/v5"
	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
)

const anonAnalyzeLimit = 2

// AuthState manages OAuth CSRF tokens and anonymous usage tracking.
type AuthState struct {
	mu            sync.Mutex
	pendingStates map[string]bool
	anonUsage     map[string]int // ip -> count
}

func NewAuthState() *AuthState {
	return &AuthState{
		pendingStates: make(map[string]bool),
		anonUsage:     make(map[string]int),
	}
}

func (a *AuthState) AddState(state string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pendingStates[state] = true
}

func (a *AuthState) ConsumeState(state string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.pendingStates[state] {
		return false
	}
	delete(a.pendingStates, state)
	return true
}

func (a *AuthState) CheckAnonLimit(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.anonUsage[ip] < anonAnalyzeLimit
}

func (a *AuthState) RecordAnonUsage(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.anonUsage[ip]++
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getRedirectURI(r *http.Request) string {
	if override := os.Getenv("WCL_REDIRECT_URI"); override != "" {
		return override
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/api/auth/callback"
}

func getFrontendURL(r *http.Request) string {
	if override := os.Getenv("FRONTEND_URL"); override != "" {
		return override
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// GetUserToken extracts Bearer token from Authorization header.
func GetUserToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}

// GetClientIP extracts the client IP from the request.
func GetClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.Split(fwd, ",")[0]
	}
	return r.RemoteAddr
}

// MountAuthRoutes registers /api/auth/login and /api/auth/callback on the router.
func MountAuthRoutes(r chi.Router, authState *AuthState) {
	r.Get("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		state := generateState()
		authState.AddState(state)

		params := url.Values{
			"client_id":     {os.Getenv("WCL_CLIENT_ID")},
			"redirect_uri":  {getRedirectURI(r)},
			"response_type": {"code"},
			"state":         {state},
		}
		http.Redirect(w, r, wcl.AuthorizeURL+"?"+params.Encode(), http.StatusFound)
	})

	r.Get("/api/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		frontendURL := getFrontendURL(r)

		if errParam := r.URL.Query().Get("error"); errParam != "" {
			http.Redirect(w, r, frontendURL+"/?auth_error="+url.QueryEscape(errParam), http.StatusFound)
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code == "" || state == "" {
			http.Error(w, `{"detail":"Missing code or state"}`, 400)
			return
		}

		if !authState.ConsumeState(state) {
			http.Error(w, `{"detail":"Invalid state parameter. Try logging in again."}`, 400)
			return
		}

		// Exchange code for token
		clientID := os.Getenv("WCL_CLIENT_ID")
		clientSecret := os.Getenv("WCL_CLIENT_SECRET")
		redirectURI := getRedirectURI(r)

		data := url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {code},
			"redirect_uri": {redirectURI},
		}

		req, _ := http.NewRequest("POST", wcl.OAuthTokenURL, strings.NewReader(data.Encode()))
		req.SetBasicAuth(clientID, clientSecret)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			http.Redirect(w, r, frontendURL+"/?auth_error=token_exchange_failed", http.StatusFound)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var tokenData map[string]any
		json.Unmarshal(body, &tokenData)
		accessToken, _ := tokenData["access_token"].(string)

		http.Redirect(w, r, frontendURL+"/?wcl_token="+url.QueryEscape(accessToken), http.StatusFound)
	})
}
