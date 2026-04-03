package wcl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CachedClient wraps a Querier with disk-based JSON caching.
type CachedClient struct {
	inner    Querier
	cacheDir string
}

func NewCachedClient(inner Querier, cacheDir string) *CachedClient {
	os.MkdirAll(cacheDir, 0755)
	return &CachedClient{inner: inner, cacheDir: cacheDir}
}

func (c *CachedClient) readCache(name string) ([]byte, bool) {
	path := filepath.Join(c.cacheDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func (c *CachedClient) writeCache(name string, data []byte) {
	path := filepath.Join(c.cacheDir, name)
	os.WriteFile(path, data, 0644)
}

func (c *CachedClient) GetReport(code string) (map[string]any, error) {
	key := code + "_report.json"
	if data, ok := c.readCache(key); ok {
		var result map[string]any
		json.Unmarshal(data, &result)
		return result, nil
	}
	result, err := c.inner.GetReport(code)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(result)
	c.writeCache(key, data)
	return result, nil
}

func (c *CachedClient) GetEvents(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
	key := fmt.Sprintf("%s_%d_%d_events.json", code, fightID, sourceID)
	if data, ok := c.readCache(key); ok {
		var result []map[string]any
		json.Unmarshal(data, &result)
		return result, nil
	}
	result, err := c.inner.GetEvents(code, fightID, sourceID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(result)
	c.writeCache(key, data)
	return result, nil
}

func (c *CachedClient) GetDamageTaken(code string, fightID, sourceID int, startTime, endTime float64, filter string) (int, error) {
	return c.inner.GetDamageTaken(code, fightID, sourceID, startTime, endTime, filter)
}
