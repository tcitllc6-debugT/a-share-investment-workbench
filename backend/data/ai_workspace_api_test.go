package data

import (
	"strings"
	"testing"

	"go-stock/backend/db"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupAIWorkspaceTestDB(t *testing.T) {
	t.Helper()
	dao, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	db.Dao = dao
	require.NoError(t, db.Dao.AutoMigrate(&Settings{}, &AIConfig{}))
}

func TestAIWorkspaceReadMasksStoredKeys(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	require.NoError(t, db.Dao.Create(&Settings{OpenAiEnable: true, Prompt: "system", QuestionTemplate: "{{stockName}}分析"}).Error)
	require.NoError(t, db.Dao.Create(&AIConfig{
		Name: "DeepSeek", BaseUrl: "https://api.deepseek.com", ApiKey: "secret-key",
		ModelName: "deepseek-chat", IsDefaultResearch: true, IsDefaultAgent: true,
	}).Error)

	got := GetAIWorkspaceConfig()

	require.True(t, got.Enabled)
	require.Len(t, got.Services, 1)
	require.Empty(t, got.Services[0].ApiKey)
	require.True(t, got.Services[0].HasApiKey)
	require.True(t, got.Services[0].IsDefaultResearch)
}

func TestAIWorkspaceSavePreservesBlankExistingKeyAndGeneralSettings(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	settings := &Settings{OpenAiEnable: true, DarkTheme: true, EnableNews: true, Prompt: "old"}
	require.NoError(t, db.Dao.Create(settings).Error)
	service := &AIConfig{
		Name: "旧名称", BaseUrl: "https://api.deepseek.com", ApiKey: "keep-me",
		ModelName: "deepseek-chat",
	}
	require.NoError(t, db.Dao.Create(service).Error)

	err := UpdateAIWorkspaceConfig(&AIWorkspaceConfig{
		Enabled:          true,
		Prompt:           "new",
		QuestionTemplate: "{{stockName}}研判",
		CrawlTimeOut:     45,
		KDays:            60,
		Services: []*AIConfig{
			{ID: service.ID, Name: "研究模型", BaseUrl: service.BaseUrl, ModelName: service.ModelName, IsDefaultResearch: true},
			{Name: "本地模型", BaseUrl: "http://127.0.0.1:11434/v1", ModelName: "qwen3", IsDefaultResearch: true, IsDefaultAgent: true},
		},
	})
	require.NoError(t, err)

	var savedSettings Settings
	require.NoError(t, db.Dao.First(&savedSettings, settings.ID).Error)
	require.True(t, savedSettings.DarkTheme)
	require.True(t, savedSettings.EnableNews)
	require.Equal(t, "new", savedSettings.Prompt)

	var services []AIConfig
	require.NoError(t, db.Dao.Order("id ASC").Find(&services).Error)
	require.Len(t, services, 2)
	require.Equal(t, "keep-me", services[0].ApiKey)
	require.True(t, services[0].IsDefaultResearch)
	require.False(t, services[1].IsDefaultResearch)
	require.True(t, services[1].IsDefaultAgent)
}

func TestAIWorkspaceSaveClearsBlankKeyWhenEndpointChanges(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	require.NoError(t, db.Dao.Create(&Settings{OpenAiEnable: true}).Error)
	service := &AIConfig{BaseUrl: "https://original.example/v1", ApiKey: "must-not-leak", ModelName: "model"}
	require.NoError(t, db.Dao.Create(service).Error)

	require.NoError(t, UpdateAIWorkspaceConfig(&AIWorkspaceConfig{
		Enabled: true,
		Services: []*AIConfig{{
			ID: service.ID, BaseUrl: "https://different.example/v1", ModelName: "model",
		}},
	}))

	var saved AIConfig
	require.NoError(t, db.Dao.First(&saved, service.ID).Error)
	require.Empty(t, saved.ApiKey)
}

func TestSanitizedAIConfigsNeverExposeKeys(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	require.NoError(t, db.Dao.Create(&AIConfig{Name: "模型", ApiKey: "top-secret", ModelName: "m"}).Error)

	configs := SanitizedAIConfigs()

	require.Len(t, configs, 1)
	require.Empty(t, configs[0].ApiKey)
	require.True(t, configs[0].HasApiKey)
}

func TestUpdateGeneralConfigDoesNotOverwriteAISettingsOrServices(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	settings := &Settings{OpenAiEnable: true, Prompt: "keep-prompt", DarkTheme: false}
	require.NoError(t, db.Dao.Create(settings).Error)
	require.NoError(t, db.Dao.Create(&AIConfig{Name: "keep-service", ApiKey: "keep-key", ModelName: "model"}).Error)

	incoming := &SettingConfig{Settings: &Settings{
		Model:        gorm.Model{ID: settings.ID},
		DarkTheme:    true,
		OpenAiEnable: false,
		Prompt:       "must-not-overwrite",
	}}
	require.Equal(t, "保存成功！", UpdateGeneralConfig(incoming))

	var got Settings
	require.NoError(t, db.Dao.First(&got, settings.ID).Error)
	require.True(t, got.DarkTheme)
	require.True(t, got.OpenAiEnable)
	require.Equal(t, "keep-prompt", got.Prompt)
	var services []AIConfig
	require.NoError(t, db.Dao.Find(&services).Error)
	require.Len(t, services, 1)
	require.Equal(t, "keep-key", services[0].ApiKey)
}

func TestLegacyUpdateConfigCannotOverwriteAISettingsOrServices(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	settings := &Settings{OpenAiEnable: true, Prompt: "keep-prompt", DarkTheme: false}
	require.NoError(t, db.Dao.Create(settings).Error)
	require.NoError(t, db.Dao.Create(&AIConfig{Name: "keep-service", ApiKey: "keep-key", ModelName: "model"}).Error)

	incoming := &SettingConfig{Settings: &Settings{
		Model:        gorm.Model{ID: settings.ID},
		DarkTheme:    true,
		OpenAiEnable: false,
		Prompt:       "must-not-overwrite",
	}, AiConfigs: nil}
	require.Equal(t, "保存成功！", UpdateConfig(incoming))

	var got Settings
	require.NoError(t, db.Dao.First(&got, settings.ID).Error)
	require.True(t, got.DarkTheme)
	require.True(t, got.OpenAiEnable)
	require.Equal(t, "keep-prompt", got.Prompt)
	var services []AIConfig
	require.NoError(t, db.Dao.Find(&services).Error)
	require.Len(t, services, 1)
	require.Equal(t, "keep-key", services[0].ApiKey)
}

func TestSettingsExportMasksAIKeys(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	require.NoError(t, db.Dao.Create(&Settings{OpenAiEnable: true}).Error)
	require.NoError(t, db.Dao.Create(&AIConfig{Name: "model", ApiKey: "top-secret", ModelName: "m"}).Error)

	exported := NewSettingsApi().Export()
	require.False(t, strings.Contains(exported, "top-secret"))
	require.Contains(t, exported, `"hasApiKey": true`)
}

func TestResolveAIConfigHonorsExplicitAndPurposeDefaults(t *testing.T) {
	first := &AIConfig{ID: 1, Name: "first"}
	research := &AIConfig{ID: 2, Name: "research", IsDefaultResearch: true}
	agent := &AIConfig{ID: 3, Name: "agent", IsDefaultAgent: true}
	configs := []*AIConfig{first, research, agent}

	got, ok := ResolveAIConfig(configs, 1, AIConfigPurposeAgent)
	require.True(t, ok)
	require.Same(t, first, got)

	got, ok = ResolveAIConfig(configs, 0, AIConfigPurposeResearch)
	require.True(t, ok)
	require.Same(t, research, got)

	got, ok = ResolveAIConfig(configs, 0, AIConfigPurposeAgent)
	require.True(t, ok)
	require.Same(t, agent, got)

	got, ok = ResolveAIConfig(configs, 0, "unknown")
	require.True(t, ok)
	require.Same(t, first, got)

	got, ok = ResolveAIConfig(nil, 0, AIConfigPurposeResearch)
	require.False(t, ok)
	require.Nil(t, got)
}

func TestGetAIConfigThinkingUsesDefaultResearchModel(t *testing.T) {
	config := &SettingConfig{AiConfigs: []*AIConfig{
		{ID: 1, Thinking: false},
		{ID: 2, Thinking: true, IsDefaultResearch: true},
	}}
	require.True(t, config.GetAIConfigThinking(0))
	require.False(t, config.GetAIConfigThinking(1))
}
