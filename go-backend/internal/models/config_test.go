package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
mastery:
  dr_table: [0, 1.0, 1.7, 2.3, 2.8, 3.2, 3.6, 4.0, 4.4, 4.8]

soul_of_the_forest:
  skip: false
  multiplier: 0.6
wild_growth:
  skip: true
  skip_reason: "always take"
`
	dir := t.TempDir()
	p := filepath.Join(dir, "talents.yaml")
	err := os.WriteFile(p, []byte(yamlContent), 0644)
	require.NoError(t, err)

	config, err := models.LoadConfig(p)
	require.NoError(t, err)
	require.Equal(t, []float64{0, 1.0, 1.7, 2.3, 2.8, 3.2, 3.6, 4.0, 4.4, 4.8}, config.Mastery.DRTable)
	require.False(t, config.Talents["soul_of_the_forest"].Skip)
	require.NotNil(t, config.Talents["soul_of_the_forest"].Multiplier)
	require.Equal(t, 0.6, *config.Talents["soul_of_the_forest"].Multiplier)
	require.True(t, config.Talents["wild_growth"].Skip)
}

func TestMissingTalentUsesDefaults(t *testing.T) {
	tc := models.TalentConfig{}
	require.False(t, tc.Skip)
	require.Nil(t, tc.Multiplier)
}
