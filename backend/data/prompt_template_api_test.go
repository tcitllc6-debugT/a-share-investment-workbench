package data

import (
	"strings"
	"testing"
)

func TestDefaultPromptTemplates(t *testing.T) {
	templates := DefaultPromptTemplates()
	if len(templates) < 5 {
		t.Fatalf("expected at least 5 default prompt templates, got %d", len(templates))
	}

	seen := map[string]bool{}
	types := map[string]bool{}
	hasAShareContent := false
	hasStockRecommendationPrompt := false
	hasHoldingAdvicePrompt := false
	for _, template := range templates {
		if strings.TrimSpace(template.Name) == "" {
			t.Fatalf("default prompt template has empty name: %#v", template)
		}
		if strings.TrimSpace(template.Type) == "" {
			t.Fatalf("default prompt template %q has empty type", template.Name)
		}
		if strings.TrimSpace(template.Content) == "" {
			t.Fatalf("default prompt template %q has empty content", template.Name)
		}
		key := template.Name + "\x00" + template.Type
		if seen[key] {
			t.Fatalf("duplicate default prompt template name/type: %s/%s", template.Name, template.Type)
		}
		seen[key] = true
		types[template.Type] = true
		if strings.Contains(template.Content, "A股") {
			hasAShareContent = true
		}
		if template.Name == "AI自动推荐A股候选股" && strings.Contains(template.Content, "股票推荐记录") {
			hasStockRecommendationPrompt = true
		}
		if template.Name == "自选持仓成本复盘与投资建议" &&
			strings.Contains(template.Content, "costPrice") &&
			strings.Contains(template.Content, "costVolume") &&
			strings.Contains(template.Content, "交易日志") {
			hasHoldingAdvicePrompt = true
		}
	}

	if !types["模型系统Prompt"] {
		t.Fatal("expected a default system prompt")
	}
	if !types["模型用户Prompt"] {
		t.Fatal("expected a default user prompt")
	}
	if !hasAShareContent {
		t.Fatal("expected default prompts to be A-share focused")
	}
	if !hasStockRecommendationPrompt {
		t.Fatal("expected a built-in AI stock recommendation prompt")
	}
	if !hasHoldingAdvicePrompt {
		t.Fatal("expected a built-in holding cost advice prompt")
	}
}
