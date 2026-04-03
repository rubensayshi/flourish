package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ResultCache struct {
	dir string
}

func NewResultCache(dir string) *ResultCache {
	os.MkdirAll(dir, 0755)
	return &ResultCache{dir: dir}
}

func (c *ResultCache) keyPath(code string, fightID int, player string) string {
	safePlayer := strings.ToLower(strings.ReplaceAll(player, " ", "_"))
	return filepath.Join(c.dir, fmt.Sprintf("%s_%d_%s.json", code, fightID, safePlayer))
}

func (c *ResultCache) Get(code string, fightID int, player string) map[string]any {
	path := c.keyPath(code, fightID, player)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

func (c *ResultCache) Set(code string, fightID int, player string, data map[string]any) {
	path := c.keyPath(code, fightID, player)
	bytes, _ := json.Marshal(data)
	os.WriteFile(path, bytes, 0644)
}
