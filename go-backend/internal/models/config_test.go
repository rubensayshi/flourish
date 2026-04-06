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

func TestLoadSpellCoefficients(t *testing.T) {
	dir, _ := os.Getwd()
	path := filepath.Join(dir, "..", "..", "config", "spell_coefficients.yaml")
	sc, err := models.LoadSpellCoefficients(path)
	require.NoError(t, err)
	require.NotNil(t, sc)

	// Rejuvenation (774)
	rejuv, ok := sc.Spells[774]
	require.True(t, ok, "Rejuvenation should be in spell coefficients")
	require.Equal(t, "Rejuvenation", rejuv.Name)
	require.Equal(t, 12000, rejuv.DurationMS)
	require.Len(t, rejuv.Effects, 1)
	require.Equal(t, "periodic", rejuv.Effects[0].Type)
	require.InDelta(t, 0.803, rejuv.Effects[0].Coefficient, 0.001)
	require.Equal(t, 3000, rejuv.Effects[0].PeriodMS)

	// Regrowth (8936) — has both direct and periodic
	regrowth, ok := sc.Spells[8936]
	require.True(t, ok)
	require.Len(t, regrowth.Effects, 2)
	require.Equal(t, "direct", regrowth.Effects[0].Type)
	require.InDelta(t, 5.36, regrowth.Effects[0].Coefficient, 0.01)
	require.Equal(t, "periodic", regrowth.Effects[1].Type)

	// Swiftmend (18562)
	sm, ok := sc.Spells[18562]
	require.True(t, ok)
	require.InDelta(t, 10.37, sm.Effects[0].Coefficient, 0.01)
}
