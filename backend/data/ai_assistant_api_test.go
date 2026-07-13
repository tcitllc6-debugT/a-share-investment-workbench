package data

import (
	"strings"
	"testing"

	"go-stock/backend/models"
)

func TestNormalizeAssistantMessagesLimitsCountAndContent(t *testing.T) {
	messages := make([]models.AiAssistantMessage, maxSavedMessages+25)
	for i := range messages {
		messages[i] = models.AiAssistantMessage{Role: "user", Content: strings.Repeat("x", maxSavedMessageContentBytes+100)}
	}

	got := normalizeAssistantMessages(messages)
	if len(got) != maxSavedMessages {
		t.Fatalf("message count = %d, want %d", len(got), maxSavedMessages)
	}
	for _, message := range got {
		if len(message.Content) > maxSavedMessageContentBytes {
			t.Fatalf("message content was not truncated: %d", len(message.Content))
		}
	}
}
