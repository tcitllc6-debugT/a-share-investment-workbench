package data

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalFeaturesAreAvailableWithoutSponsorCode(t *testing.T) {
	require.True(t, IsLocalFeatureEnabled("kline_analysis"))
	require.True(t, IsLocalFeatureEnabled("ai_assistant_web"))
}

func TestEffectiveSponsorVipLevelUsesLocalAccess(t *testing.T) {
	level, active := EffectiveSponsorVipLevel()
	require.Equal(t, 2, level)
	require.True(t, active)
}
