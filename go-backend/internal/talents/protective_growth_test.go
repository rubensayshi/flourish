package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestProtectiveGrowthAttributesDrAsHealing(t *testing.T) {
	attr := talents.NewProtectiveGrowthAttributor(100000)
	result := attr.Finalize()
	expected := 100000.0 * 0.08 / 0.92
	require.InDelta(t, expected, result, 1.0)
}

func TestProtectiveGrowthZeroDamage(t *testing.T) {
	attr := talents.NewProtectiveGrowthAttributor(0)
	require.Equal(t, 0.0, attr.Finalize())
}

func TestProtectiveGrowthNilDamage(t *testing.T) {
	attr := talents.NewProtectiveGrowthAttributorNil()
	require.Equal(t, 0.0, attr.Finalize())
}
