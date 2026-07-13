package data

import (
	"encoding/json"
	"errors"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"strings"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

const (
	MarketFocusAShare = "a-share"
	MarketFocusAll    = "all"

	AIConfigPurposeResearch = "research"
	AIConfigPurposeAgent    = "agent"
)

type Settings struct {
	gorm.Model
	TushareToken     string `json:"tushareToken"`
	LocalPushEnable  bool   `json:"localPushEnable"`
	DingPushEnable   bool   `json:"dingPushEnable"`
	DingRobot        string `json:"dingRobot"`
	FeishuPushEnable bool   `json:"feishuPushEnable"`
	FeishuRobot      string `json:"feishuRobot"`
	FeishuSecret     string `json:"feishuSecret" gorm:"column:feishu_secret"`
	// 飞书应用机器人（接收消息+AI回复，长连接模式，与 FeishuPush 自定义机器人推送独立）
	FeishuBotEnable        bool   `json:"feishuBotEnable"`
	FeishuAppID            string `json:"feishuAppId" gorm:"column:feishu_app_id"`
	FeishuAppSecret        string `json:"feishuAppSecret" gorm:"column:feishu_app_secret"`
	FeishuBotAiConfigId    int    `json:"feishuBotAiConfigId" gorm:"column:feishu_bot_ai_config_id"`
	FeishuBotSysPromptId   int    `json:"feishuBotSysPromptId" gorm:"column:feishu_bot_sys_prompt_id"`
	FeishuBotEnableTools   bool   `json:"feishuBotEnableTools"`
	FeishuBotThinking      bool   `json:"feishuBotThinking"`
	FeishuBotAgentMode     string `json:"feishuBotAgentMode" gorm:"column:feishu_bot_agent_mode"`
	UpdateBasicInfoOnStart bool   `json:"updateBasicInfoOnStart"`
	RefreshInterval        int64  `json:"refreshInterval"`
	OpenAiEnable           bool   `json:"openAiEnable"`
	Prompt                 string `json:"prompt"`
	CheckUpdate            bool   `json:"checkUpdate"`
	UpdateChannel          string `json:"updateChannel"`
	UpdateManifestURL      string `json:"updateManifestUrl" gorm:"column:update_manifest_url"`
	QuestionTemplate       string `json:"questionTemplate"`
	CrawlTimeOut           int64  `json:"crawlTimeOut"`
	KDays                  int64  `json:"kDays"`
	EnableDanmu            bool   `json:"enableDanmu"`
	BrowserPath            string `json:"browserPath"`
	EnableNews             bool   `json:"enableNews"`
	DarkTheme              bool   `json:"darkTheme"`
	BrowserPoolSize        int    `json:"browserPoolSize"`
	EnableFund             bool   `json:"enableFund"`
	EnablePushNews         bool   `json:"enablePushNews"`
	EnableOnlyPushRedNews  bool   `json:"enableOnlyPushRedNews"`
	SponsorCode            string `json:"-"`
	HttpProxy              string `json:"httpProxy"`
	HttpProxyEnabled       bool   `json:"httpProxyEnabled"`
	EnableAgent            bool   `json:"enableAgent"`
	QgqpBId                string `json:"qgqpBId" gorm:"column:qgqp_b_id"`
	IwencaiApiKey          string `json:"iwencaiApiKey" gorm:"column:iwencai_api_key"`
	EmApiKey               string `json:"emApiKey" gorm:"column:em_api_key"`
	WindowWidth            int    `json:"windowWidth"`
	WindowHeight           int    `json:"windowHeight"`
	PromptPlazaApiBase     string `json:"promptPlazaApiBase" gorm:"column:prompt_plaza_api_base"`
	MarketFocusMode        string `json:"marketFocusMode" gorm:"column:market_focus_mode"`
}

func (receiver Settings) TableName() string {
	return "settings"
}

type AIConfig struct {
	ID                uint `gorm:"primarykey"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Name              string  `json:"name"`
	BaseUrl           string  `json:"baseUrl"`
	ApiKey            string  `json:"apiKey" `
	ModelName         string  `json:"modelName"`
	MaxTokens         int     `json:"maxTokens"`
	Temperature       float64 `json:"temperature"`
	TimeOut           int     `json:"timeOut"`
	HttpProxy         string  `json:"httpProxy"`
	HttpProxyEnabled  bool    `json:"httpProxyEnabled"`
	SessionId         string  `json:"sessionId" gorm:"index;size:64"`
	Thinking          bool    `json:"thinking"`
	IsDefaultResearch bool    `json:"isDefaultResearch" gorm:"column:is_default_research"`
	IsDefaultAgent    bool    `json:"isDefaultAgent" gorm:"column:is_default_agent"`
	HasApiKey         bool    `json:"hasApiKey" gorm:"-"`
}

func (AIConfig) TableName() string {
	return "ai_config"
}

type SettingConfig struct {
	*Settings
	AiConfigs []*AIConfig `json:"aiConfigs"`
}

type AIWorkspaceConfig struct {
	Enabled          bool        `json:"enabled"`
	Prompt           string      `json:"prompt"`
	QuestionTemplate string      `json:"questionTemplate"`
	CrawlTimeOut     int64       `json:"crawlTimeOut"`
	KDays            int64       `json:"kDays"`
	HttpProxy        string      `json:"httpProxy"`
	HttpProxyEnabled bool        `json:"httpProxyEnabled"`
	Services         []*AIConfig `json:"services"`
}

// ResolveAIConfig keeps explicit user choices authoritative, then applies the
// configured purpose default. The first valid service remains the legacy
// fallback for installations created before purpose defaults existed.
func ResolveAIConfig(configs []*AIConfig, aiConfigID int, purpose string) (*AIConfig, bool) {
	if aiConfigID > 0 {
		for _, item := range configs {
			if item != nil && int(item.ID) == aiConfigID {
				return item, true
			}
		}
	}

	for _, item := range configs {
		if item == nil {
			continue
		}
		if purpose == AIConfigPurposeResearch && item.IsDefaultResearch {
			return item, true
		}
		if purpose == AIConfigPurposeAgent && item.IsDefaultAgent {
			return item, true
		}
	}

	for _, item := range configs {
		if item != nil {
			return item, true
		}
	}
	return nil, false
}

func sanitizeAIConfig(item *AIConfig) *AIConfig {
	if item == nil {
		return nil
	}
	copyItem := *item
	copyItem.HasApiKey = strings.TrimSpace(item.ApiKey) != ""
	copyItem.ApiKey = ""
	return &copyItem
}

func SanitizedAIConfigs() []*AIConfig {
	var configs []*AIConfig
	if err := db.Dao.Order("id ASC").Find(&configs).Error; err != nil {
		logger.SugaredLogger.Errorf("查询脱敏 AI 配置失败: %v", err)
		return []*AIConfig{}
	}
	result := make([]*AIConfig, 0, len(configs))
	for _, item := range configs {
		result = append(result, sanitizeAIConfig(item))
	}
	return result
}

func GetAIWorkspaceConfig() *AIWorkspaceConfig {
	settings := &Settings{}
	if err := db.Dao.First(settings).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.SugaredLogger.Errorf("查询 AI 工作区设置失败: %v", err)
	}
	crawlTimeout := settings.CrawlTimeOut
	if crawlTimeout <= 0 {
		crawlTimeout = 60
	}
	kDays := settings.KDays
	if kDays < 30 {
		kDays = 60
	}
	questionTemplate := strings.TrimSpace(settings.QuestionTemplate)
	if questionTemplate == "" {
		questionTemplate = "{{stockName}}[{{stockCode}}]分析和总结"
	}
	return &AIWorkspaceConfig{
		Enabled:          settings.OpenAiEnable,
		Prompt:           settings.Prompt,
		QuestionTemplate: questionTemplate,
		CrawlTimeOut:     crawlTimeout,
		KDays:            kDays,
		HttpProxy:        settings.HttpProxy,
		HttpProxyEnabled: settings.HttpProxyEnabled,
		Services:         SanitizedAIConfigs(),
	}
}

func normalizeAIWorkspaceServices(services []*AIConfig) []*AIConfig {
	result := make([]*AIConfig, 0, len(services))
	researchChosen := false
	agentChosen := false
	for _, item := range services {
		if item == nil {
			continue
		}
		copyItem := *item
		copyItem.Name = strings.TrimSpace(copyItem.Name)
		copyItem.BaseUrl = strings.TrimRight(strings.TrimSpace(copyItem.BaseUrl), "/")
		copyItem.ApiKey = strings.TrimSpace(copyItem.ApiKey)
		copyItem.ModelName = strings.TrimSpace(copyItem.ModelName)
		copyItem.HttpProxy = strings.TrimSpace(copyItem.HttpProxy)
		copyItem.HasApiKey = false
		if copyItem.MaxTokens <= 0 {
			copyItem.MaxTokens = 4096
		}
		if copyItem.TimeOut <= 0 {
			copyItem.TimeOut = 300
		}
		if copyItem.IsDefaultResearch {
			if researchChosen {
				copyItem.IsDefaultResearch = false
			} else {
				researchChosen = true
			}
		}
		if copyItem.IsDefaultAgent {
			if agentChosen {
				copyItem.IsDefaultAgent = false
			} else {
				agentChosen = true
			}
		}
		result = append(result, &copyItem)
	}
	if len(result) > 0 {
		if !researchChosen {
			result[0].IsDefaultResearch = true
		}
		if !agentChosen {
			result[0].IsDefaultAgent = true
		}
	}
	return result
}

func UpdateAIWorkspaceConfig(input *AIWorkspaceConfig) error {
	if input == nil {
		return errors.New("AI 配置数据为空")
	}
	services := normalizeAIWorkspaceServices(input.Services)
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		settings := &Settings{}
		result := tx.First(settings)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			settings.OpenAiEnable = input.Enabled
			settings.Prompt = input.Prompt
			settings.QuestionTemplate = input.QuestionTemplate
			settings.CrawlTimeOut = input.CrawlTimeOut
			settings.KDays = input.KDays
			settings.HttpProxy = strings.TrimSpace(input.HttpProxy)
			settings.HttpProxyEnabled = input.HttpProxyEnabled
			if err := tx.Create(settings).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		} else if err := tx.Model(settings).Updates(map[string]any{
			"open_ai_enable":     input.Enabled,
			"prompt":             input.Prompt,
			"question_template":  input.QuestionTemplate,
			"crawl_time_out":     input.CrawlTimeOut,
			"k_days":             input.KDays,
			"http_proxy":         strings.TrimSpace(input.HttpProxy),
			"http_proxy_enabled": input.HttpProxyEnabled,
		}).Error; err != nil {
			return err
		}

		var existing []*AIConfig
		if err := tx.Find(&existing).Error; err != nil {
			return err
		}
		existingByID := make(map[uint]*AIConfig, len(existing))
		for _, item := range existing {
			existingByID[item.ID] = item
		}

		savedIDs := make([]uint, 0, len(services))
		for _, item := range services {
			if old := existingByID[item.ID]; item.ID > 0 && old != nil {
				if item.ApiKey == "" && sameAIEndpoint(item.BaseUrl, old.BaseUrl) {
					item.ApiKey = old.ApiKey
				}
				if err := tx.Model(&AIConfig{}).Where("id = ?", item.ID).Updates(map[string]any{
					"name": item.Name, "base_url": item.BaseUrl, "api_key": item.ApiKey,
					"model_name": item.ModelName, "max_tokens": item.MaxTokens,
					"temperature": item.Temperature, "time_out": item.TimeOut,
					"http_proxy": item.HttpProxy, "http_proxy_enabled": item.HttpProxyEnabled,
					"session_id": item.SessionId, "thinking": item.Thinking,
					"is_default_research": item.IsDefaultResearch,
					"is_default_agent":    item.IsDefaultAgent,
				}).Error; err != nil {
					return err
				}
			} else {
				item.ID = 0
				if err := tx.Create(item).Error; err != nil {
					return err
				}
			}
			savedIDs = append(savedIDs, item.ID)
		}
		if len(savedIDs) == 0 {
			return tx.Exec("DELETE FROM ai_config").Error
		}
		return tx.Where("id NOT IN ?", savedIDs).Delete(&AIConfig{}).Error
	})
	if err == nil {
		ConfigureFromSettings(GetSettingConfig())
	}
	return err
}

func (c *SettingConfig) GetAIConfigThinking(aiConfigId int) bool {
	config, ok := ResolveAIConfig(c.AiConfigs, aiConfigId, AIConfigPurposeResearch)
	if ok {
		return config.Thinking
	}
	return false
}

func (c *SettingConfig) IsAShareFocus() bool {
	if c == nil || c.Settings == nil {
		return true
	}
	return NormalizeMarketFocusMode(c.MarketFocusMode) == MarketFocusAShare
}

func NormalizeMarketFocusMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case MarketFocusAll:
		return MarketFocusAll
	default:
		return MarketFocusAShare
	}
}

type SettingsApi struct {
	Config *SettingConfig
}

func NewSettingsApi() *SettingsApi {
	return &SettingsApi{
		Config: GetSettingConfig(),
	}
}

func (s *SettingsApi) Export() string {
	settingsCopy := *s.Config.Settings
	settingsCopy.SponsorCode = ""
	exportConfig := &SettingConfig{Settings: &settingsCopy, AiConfigs: SanitizedAIConfigs()}
	d, _ := json.MarshalIndent(exportConfig, "", "    ")
	return string(d)
}

func updateLegacyConfigWithAI(s *SettingConfig) string {
	if s.Settings == nil {
		return "保存失败: 配置数据为空"
	}
	s.MarketFocusMode = NormalizeMarketFocusMode(s.MarketFocusMode)
	count := int64(0)
	db.Dao.Model(&Settings{}).Count(&count)
	if count > 0 {
		result := db.Dao.Model(&Settings{}).Where("id=?", s.ID).Updates(map[string]any{
			"local_push_enable":          s.LocalPushEnable,
			"ding_push_enable":           s.DingPushEnable,
			"ding_robot":                 s.DingRobot,
			"feishu_push_enable":         s.FeishuPushEnable,
			"feishu_robot":               s.FeishuRobot,
			"feishu_secret":              s.FeishuSecret,
			"feishu_bot_enable":          s.FeishuBotEnable,
			"feishu_app_id":              s.FeishuAppID,
			"feishu_app_secret":          s.FeishuAppSecret,
			"feishu_bot_ai_config_id":    s.FeishuBotAiConfigId,
			"feishu_bot_sys_prompt_id":   s.FeishuBotSysPromptId,
			"feishu_bot_enable_tools":    s.FeishuBotEnableTools,
			"feishu_bot_thinking":        s.FeishuBotThinking,
			"feishu_bot_agent_mode":      s.FeishuBotAgentMode,
			"update_basic_info_on_start": s.UpdateBasicInfoOnStart,
			"refresh_interval":           s.RefreshInterval,
			"open_ai_enable":             s.OpenAiEnable,
			"tushare_token":              s.TushareToken,
			"prompt":                     s.Prompt,
			"check_update":               s.CheckUpdate,
			"update_channel":             s.UpdateChannel,
			"update_manifest_url":        s.UpdateManifestURL,
			"question_template":          s.QuestionTemplate,
			"crawl_time_out":             s.CrawlTimeOut,
			"k_days":                     s.KDays,
			"enable_danmu":               s.EnableDanmu,
			"browser_path":               s.BrowserPath,
			"enable_news":                s.EnableNews,
			"dark_theme":                 s.DarkTheme,
			"enable_fund":                s.EnableFund,
			"enable_push_news":           s.EnablePushNews,
			"enable_only_push_red_news":  s.EnableOnlyPushRedNews,
			"http_proxy":                 s.HttpProxy,
			"http_proxy_enabled":         s.HttpProxyEnabled,
			"enable_agent":               s.EnableAgent,
			"qgqp_b_id":                  s.QgqpBId,
			"iwencai_api_key":            s.IwencaiApiKey,
			"em_api_key":                 s.EmApiKey,
			"window_width":               s.WindowWidth,
			"window_height":              s.WindowHeight,
			"prompt_plaza_api_base":      s.PromptPlazaApiBase,
			"market_focus_mode":          s.MarketFocusMode,
		})
		if result.Error != nil {
			logger.SugaredLogger.Errorf("更新配置失败: %v", result.Error)
			return "保存失败: " + result.Error.Error()
		}

		err := updateAiConfigs(s.AiConfigs)
		if err != nil {
			logger.SugaredLogger.Errorf("更新AI模型服务配置失败: %v", err)
			return "更新AI模型服务配置失败: " + err.Error()
		}
	} else {
		result := db.Dao.Model(&Settings{}).Create(s.Settings)
		if result.Error != nil {
			logger.SugaredLogger.Error("创建配置失败:", result.Error)
			return "创建配置失败: " + result.Error.Error()
		}
		err := updateAiConfigs(s.AiConfigs)
		if err != nil {
			logger.SugaredLogger.Errorf("更新AI模型服务配置失败: %v", err)
			return "更新AI模型服务配置失败: " + err.Error()
		}
	}

	ConfigureFromSettings(s)

	return "保存成功！"
}

// UpdateConfig remains for compatibility with older callers, but AI settings
// and model credentials are owned exclusively by UpdateAIWorkspaceConfig.
func UpdateConfig(s *SettingConfig) string {
	return UpdateGeneralConfig(s)
}

func generalSettingsUpdateMap(s *SettingConfig) map[string]any {
	return map[string]any{
		"local_push_enable":          s.LocalPushEnable,
		"ding_push_enable":           s.DingPushEnable,
		"ding_robot":                 s.DingRobot,
		"feishu_push_enable":         s.FeishuPushEnable,
		"feishu_robot":               s.FeishuRobot,
		"feishu_secret":              s.FeishuSecret,
		"feishu_bot_enable":          s.FeishuBotEnable,
		"feishu_app_id":              s.FeishuAppID,
		"feishu_app_secret":          s.FeishuAppSecret,
		"feishu_bot_ai_config_id":    s.FeishuBotAiConfigId,
		"feishu_bot_sys_prompt_id":   s.FeishuBotSysPromptId,
		"feishu_bot_enable_tools":    s.FeishuBotEnableTools,
		"feishu_bot_thinking":        s.FeishuBotThinking,
		"feishu_bot_agent_mode":      s.FeishuBotAgentMode,
		"update_basic_info_on_start": s.UpdateBasicInfoOnStart,
		"refresh_interval":           s.RefreshInterval,
		"tushare_token":              s.TushareToken,
		"check_update":               s.CheckUpdate,
		"update_channel":             s.UpdateChannel,
		"update_manifest_url":        s.UpdateManifestURL,
		"enable_danmu":               s.EnableDanmu,
		"browser_path":               s.BrowserPath,
		"enable_news":                s.EnableNews,
		"dark_theme":                 s.DarkTheme,
		"enable_fund":                s.EnableFund,
		"enable_push_news":           s.EnablePushNews,
		"enable_only_push_red_news":  s.EnableOnlyPushRedNews,
		"enable_agent":               s.EnableAgent,
		"qgqp_b_id":                  s.QgqpBId,
		"iwencai_api_key":            s.IwencaiApiKey,
		"em_api_key":                 s.EmApiKey,
		"window_width":               s.WindowWidth,
		"window_height":              s.WindowHeight,
		"prompt_plaza_api_base":      s.PromptPlazaApiBase,
		"market_focus_mode":          NormalizeMarketFocusMode(s.MarketFocusMode),
	}
}

// UpdateGeneralConfig updates only non-AI settings. AI fields and ai_config rows
// are exclusively owned by UpdateAIWorkspaceConfig.
func UpdateGeneralConfig(s *SettingConfig) string {
	if s == nil || s.Settings == nil {
		return "保存失败: 配置数据为空"
	}
	values := generalSettingsUpdateMap(s)
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		var existing Settings
		result := tx.Order("id ASC").First(&existing)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if s.ID > 0 {
				values["id"] = s.ID
			}
			return tx.Model(&Settings{}).Create(values).Error
		}
		if result.Error != nil {
			return result.Error
		}
		return tx.Model(&Settings{}).Where("id = ?", existing.ID).Updates(values).Error
	})
	if err != nil {
		logger.SugaredLogger.Errorf("更新普通设置失败: %v", err)
		return "保存失败: " + err.Error()
	}
	ConfigureFromSettings(GetSettingConfig())
	return "保存成功！"
}

func updateAiConfigs(aiConfigs []*AIConfig) error {
	if len(aiConfigs) == 0 {
		err := db.Dao.Exec("DELETE FROM ai_config").Error
		if err != nil {
			return err
		}
		return db.Dao.Exec("DELETE FROM sqlite_sequence WHERE name='ai_config'").Error
	}
	// 仅收集大于 0 的 ID，用于识别已存在的配置；
	// ID<=0 视为“新配置”，强制走插入逻辑，避免多个 ID 为 0 的配置互相覆盖。
	var ids []uint
	lo.ForEach(aiConfigs, func(item *AIConfig, index int) {
		if item.ID > 0 {
			ids = append(ids, item.ID)
		}
	})
	var existAiConfigs []*AIConfig
	err := db.Dao.Model(&AIConfig{}).Select("id").Where("id in (?) ", ids).Find(&existAiConfigs).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	idMap := make(map[uint]bool)
	lo.ForEach(existAiConfigs, func(item *AIConfig, index int) {
		idMap[item.ID] = true
	})
	var addAiConfigs []*AIConfig
	var notDeleteIds []uint
	var e error
	lo.ForEach(aiConfigs, func(item *AIConfig, index int) {
		if e != nil {
			return
		}
		// ID<=0 一律视为新配置，走插入逻辑；否则根据是否已存在决定更新或新增
		if item.ID <= 0 || !idMap[item.ID] {
			addAiConfigs = append(addAiConfigs, item)
		} else {
			notDeleteIds = append(notDeleteIds, item.ID)
			e = db.Dao.Model(&AIConfig{}).Where("id=?", item.ID).Updates(map[string]interface{}{
				"name":               item.Name,
				"base_url":           item.BaseUrl,
				"api_key":            item.ApiKey,
				"model_name":         item.ModelName,
				"max_tokens":         item.MaxTokens,
				"temperature":        item.Temperature,
				"time_out":           item.TimeOut,
				"http_proxy":         item.HttpProxy,
				"http_proxy_enabled": item.HttpProxyEnabled,
				"session_id":         item.SessionId,
				"thinking":           item.Thinking,
			}).Error
			if e != nil {
				return
			}
		}
	})
	if e != nil {
		return e
	}
	//删除旧的配置
	if len(notDeleteIds) > 0 {
		err = db.Dao.Exec("DELETE FROM ai_config WHERE id NOT IN ?", notDeleteIds).Error
		if err != nil {
			return err
		}
	}
	//logger.SugaredLogger.Infof("更新aiConfigs +%d", len(addAiConfigs))
	//批量新增的配置
	err = db.Dao.CreateInBatches(addAiConfigs, len(addAiConfigs)).Error
	return err
}

func GetSettingConfig() *SettingConfig {
	settingConfig := &SettingConfig{}
	settings := &Settings{}
	aiConfigs := make([]*AIConfig, 0)
	// 处理数据库查询可能返回的空结果
	result := db.Dao.Model(&Settings{}).First(settings)
	if settings.OpenAiEnable {
		// 处理AI配置查询可能出现的错误
		result = db.Dao.Model(&AIConfig{}).Find(&aiConfigs)
		if result.Error != nil {
			logger.SugaredLogger.Error("查询AI配置失败:", result.Error)
		} else if len(aiConfigs) > 0 {
			lo.ForEach(aiConfigs, func(item *AIConfig, index int) {
				if item.TimeOut <= 0 {
					item.TimeOut = 60 * 5
				}
			})
		}
		if settings.CrawlTimeOut <= 0 {
			settings.CrawlTimeOut = 60
		}
		if settings.KDays < 30 {
			settings.KDays = 60
		}
	}
	if settings.BrowserPath == "" {
		settings.BrowserPath, _ = CheckBrowser()
	}
	if settings.BrowserPoolSize <= 0 {
		settings.BrowserPoolSize = 1
	}
	settings.MarketFocusMode = NormalizeMarketFocusMode(settings.MarketFocusMode)
	settings.EnableAgent = false

	settingConfig.Settings = settings
	settingConfig.AiConfigs = aiConfigs

	return settingConfig
}
