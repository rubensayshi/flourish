package web_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
	"github.com/stretchr/testify/require"
)

func TestGetReturnsNilWhenMissing(t *testing.T) {
	cache := web.NewResultCache(t.TempDir())
	result := cache.Get("ABC123", 1, "Player")
	require.Nil(t, result)
}

func TestSetThenGetRoundtrips(t *testing.T) {
	cache := web.NewResultCache(t.TempDir())
	data := map[string]any{"total_healing": 1000.0, "talents": []any{map[string]any{"name": "SotF", "attributed": 500.0}}}
	cache.Set("ABC123", 1, "Player", data)
	result := cache.Get("ABC123", 1, "Player")
	require.NotNil(t, result)
	require.Equal(t, 1000.0, result["total_healing"])
}

func TestCacheKeyIsCaseInsensitiveForPlayer(t *testing.T) {
	cache := web.NewResultCache(t.TempDir())
	data := map[string]any{"total_healing": 1000.0}
	cache.Set("ABC123", 1, "Saikó", data)
	result := cache.Get("ABC123", 1, "saikó")
	require.NotNil(t, result)
	require.Equal(t, 1000.0, result["total_healing"])
}
