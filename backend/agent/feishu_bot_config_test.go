package agent

import (
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/db"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewFeishuBotKeepsDynamicDefaultBinding(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	db.Dao = dao
	require.NoError(t, dao.AutoMigrate(&data.Settings{}, &data.AIConfig{}))
	require.NoError(t, dao.Create(&data.Settings{
		OpenAiEnable: true,
		FeishuAppID:  "app-id", FeishuAppSecret: "app-secret",
		FeishuBotAiConfigId: 0,
	}).Error)
	require.NoError(t, dao.Create(&data.AIConfig{ModelName: "agent-model", IsDefaultAgent: true}).Error)

	bot := NewFeishuBot()
	require.NotNil(t, bot)
	require.Zero(t, bot.aiConfigId)
}

func TestNewFeishuBotKeepsExplicitModelBinding(t *testing.T) {
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	db.Dao = dao
	require.NoError(t, dao.AutoMigrate(&data.Settings{}, &data.AIConfig{}))
	explicit := &data.AIConfig{ModelName: "explicit-model"}
	require.NoError(t, dao.Create(explicit).Error)
	require.NoError(t, dao.Create(&data.Settings{
		OpenAiEnable: true,
		FeishuAppID:  "app-id", FeishuAppSecret: "app-secret",
		FeishuBotAiConfigId: int(explicit.ID),
	}).Error)

	bot := NewFeishuBot()
	require.NotNil(t, bot)
	require.Equal(t, int(explicit.ID), bot.aiConfigId)
}
