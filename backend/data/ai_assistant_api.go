package data

import (
	"encoding/json"
	"go-stock/backend/db"
	"go-stock/backend/models"
	"strings"
	"time"
)

const (
	maxSavedMessages            = 120
	maxSavedMessageContentBytes = 32 * 1024
)

func normalizeAssistantMessages(messages []models.AiAssistantMessage) []models.AiAssistantMessage {
	if len(messages) > maxSavedMessages {
		messages = messages[len(messages)-maxSavedMessages:]
	}
	normalized := make([]models.AiAssistantMessage, 0, len(messages))
	for _, message := range messages {
		message.Role = strings.TrimSpace(message.Role)
		message.Content = truncateUTF8Bytes(message.Content, maxSavedMessageContentBytes)
		message.Reasoning = truncateUTF8Bytes(message.Reasoning, maxSavedMessageContentBytes)
		message.ToolCalls = truncateRawMessage(message.ToolCalls, maxSavedMessageContentBytes)
		message.ToolResults = truncateRawMessage(message.ToolResults, maxSavedMessageContentBytes)
		message.Timeline = truncateRawMessage(message.Timeline, maxSavedMessageContentBytes)
		normalized = append(normalized, message)
	}
	return normalized
}

func truncateUTF8Bytes(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return strings.ToValidUTF8(value[:limit], "")
}

func truncateRawMessage(value json.RawMessage, limit int) json.RawMessage {
	if len(value) <= limit {
		return value
	}
	return nil
}

// GetAiAssistantSession 获取指定 sessionId 的会话消息列表，若 sessionId 为空则获取最新的
func GetAiAssistantSession(sessionId string) (*models.AiAssistantSessionResp, error) {
	var row models.AiAssistantSession
	var err error
	if sessionId != "" {
		err = db.Dao.Model(&models.AiAssistantSession{}).Where("session_id = ?", sessionId).First(&row).Error
	} else {
		err = db.Dao.Model(&models.AiAssistantSession{}).Order("updated_at DESC").First(&row).Error
	}
	resp := &models.AiAssistantSessionResp{
		Messages:  []models.AiAssistantMessage{},
		SessionId: row.SessionId,
	}
	if err != nil {
		return resp, nil
	}
	if row.Messages == "" {
		return resp, nil
	}
	var list []models.AiAssistantMessage
	if err := json.Unmarshal([]byte(row.Messages), &list); err != nil {
		return resp, nil
	}
	resp.Messages = list
	return resp, nil
}

// SaveAiAssistantSession 保存会话消息到数据库，若 sessionId 已存在则更新，否则创建新记录
func SaveAiAssistantSession(sessionId string, messages []models.AiAssistantMessage) error {
	if len(messages) == 0 {
		return nil
	}
	toSave := normalizeAssistantMessages(messages)
	raw, err := json.Marshal(toSave)
	if err != nil {
		return err
	}
	payload := string(raw)

	var existing models.AiAssistantSession
	err = db.Dao.Model(&models.AiAssistantSession{}).Where("session_id = ?", sessionId).First(&existing).Error
	if err == nil {
		return db.Dao.Model(&models.AiAssistantSession{}).Where("session_id = ?", sessionId).Updates(map[string]interface{}{
			"messages":   payload,
			"updated_at": time.Now(),
		}).Error
	}
	return db.Dao.Create(&models.AiAssistantSession{SessionId: sessionId, Messages: payload}).Error
}
