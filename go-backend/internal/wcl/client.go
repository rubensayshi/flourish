package wcl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

const (
	defaultOAuthURL = "https://www.warcraftlogs.com/oauth/token"
	defaultAPIURL   = "https://www.warcraftlogs.com/api/v2/client"
)

// Querier is the interface for WCL API clients.
type Querier interface {
	GetReport(code string) (map[string]any, error)
	GetEvents(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error)
	GetDamageTaken(code string, fightID, sourceID int, startTime, endTime float64, filter string) (int, error)
}

type Client struct {
	clientID     string
	clientSecret string
	token        string
	httpClient   *http.Client
	baseURL      string
	oauthURL     string
}

type ClientOption func(*Client)

func WithBaseURL(url string) ClientOption {
	return func(c *Client) { c.baseURL = url }
}

func WithOAuthURL(url string) ClientOption {
	return func(c *Client) { c.oauthURL = url }
}

func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

func NewClient(clientID, clientSecret string, opts ...ClientOption) *Client {
	c := &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
		baseURL:      defaultAPIURL,
		oauthURL:     defaultOAuthURL,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) authenticate() error {
	req, err := http.NewRequest("POST", c.oauthURL, bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	c.token = result["access_token"].(string)
	return nil
}

func (c *Client) query(query string, variables map[string]any) (map[string]any, error) {
	if c.token == "" {
		if err := c.authenticate(); err != nil {
			return nil, err
		}
	}

	body, _ := json.Marshal(map[string]any{"query": query, "variables": variables})
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if errs, ok := result["errors"]; ok {
		return nil, fmt.Errorf("WCL API errors: %v", errs)
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	return data, nil
}

func (c *Client) GetReport(code string) (map[string]any, error) {
	data, err := c.query(FightsQuery, map[string]any{"code": code})
	if err != nil {
		return nil, err
	}
	report := data["reportData"].(map[string]any)["report"].(map[string]any)
	return report, nil
}

func (c *Client) GetEvents(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
	var allEvents []map[string]any
	currentStart := startTime

	for {
		data, err := c.query(EventsQuery, map[string]any{
			"code":      code,
			"startTime": currentStart,
			"endTime":   endTime,
			"sourceID":  sourceID,
			"fightIDs":  []int{fightID},
		})
		if err != nil {
			return nil, err
		}

		eventsData := data["reportData"].(map[string]any)["report"].(map[string]any)["events"].(map[string]any)
		eventsList := eventsData["data"].([]any)
		for _, e := range eventsList {
			allEvents = append(allEvents, e.(map[string]any))
		}

		nextPage := eventsData["nextPageTimestamp"]
		if nextPage == nil {
			break
		}
		switch v := nextPage.(type) {
		case float64:
			currentStart = v
		default:
			break
		}
		if currentStart == 0 {
			break
		}
	}

	sort.Slice(allEvents, func(i, j int) bool {
		ti, _ := allEvents[i]["timestamp"].(float64)
		tj, _ := allEvents[j]["timestamp"].(float64)
		return ti < tj
	})
	return allEvents, nil
}

func (c *Client) GetDamageTaken(code string, fightID, sourceID int, startTime, endTime float64, filter string) (int, error) {
	data, err := c.query(DamageTakenTableQuery, map[string]any{
		"code":             code,
		"startTime":        startTime,
		"endTime":          endTime,
		"sourceID":         sourceID,
		"fightIDs":         []int{fightID},
		"filterExpression": filter,
	})
	if err != nil {
		return 0, err
	}
	tableData := data["reportData"].(map[string]any)["report"].(map[string]any)["table"].(map[string]any)["data"].(map[string]any)
	entries, ok := tableData["entries"].([]any)
	if !ok {
		return 0, nil
	}
	total := 0
	for _, entry := range entries {
		e := entry.(map[string]any)
		if t, ok := e["total"].(float64); ok {
			total += int(t)
		}
	}
	return total, nil
}
