package data

import (
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/db"

	"github.com/stretchr/testify/require"
)

func TestSettingsExportDoesNotLeakLegacySponsorCode(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "settings-export.db"))
	t.Cleanup(func() {
		sqlDB, err := db.Dao.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())
	})
	require.NoError(t, db.Dao.AutoMigrate(&Settings{}, &AIConfig{}))
	require.NoError(t, db.Dao.Create(&Settings{SponsorCode: "legacy-private-code"}).Error)

	exported := NewSettingsApi().Export()
	require.False(t, strings.Contains(exported, "legacy-private-code"), exported)
}

func TestReplicaSettingsDoNotRequireOriginalAuthorServices(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "settings.db"))
	t.Cleanup(func() {
		sqlDB, err := db.Dao.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())
	})

	cfg := GetSettingConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Settings)
	require.NotContains(t, cfg.PromptPlazaApiBase, "go-stock.sparkmemory.top")
	require.Empty(t, cfg.PromptPlazaApiBase)
}

func TestReplicaSettingsDefaultToAShareMarketFocus(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "settings.db"))
	t.Cleanup(func() {
		sqlDB, err := db.Dao.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())
	})

	cfg := GetSettingConfig()

	require.Equal(t, MarketFocusAShare, cfg.MarketFocusMode)
	require.True(t, cfg.IsAShareFocus())
}

func TestReplicaSettingsPreserveExplicitMarketFocus(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "settings.db"))
	t.Cleanup(func() {
		sqlDB, err := db.Dao.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())
	})
	require.NoError(t, db.Dao.AutoMigrate(&Settings{}))
	require.NoError(t, db.Dao.Create(&Settings{MarketFocusMode: MarketFocusAll}).Error)

	cfg := GetSettingConfig()

	require.Equal(t, MarketFocusAll, cfg.MarketFocusMode)
	require.False(t, cfg.IsAShareFocus())
}
