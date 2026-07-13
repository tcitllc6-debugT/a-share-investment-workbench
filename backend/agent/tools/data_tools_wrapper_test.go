package tools

import (
	"encoding/json"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"path/filepath"
	"testing"
)

func TestInjectRecommendMetaSingle(t *testing.T) {
	meta := AgentMeta{
		ModelName:    "deepseek-chat",
		SystemPrompt: "system prompt",
		UserPrompt:   "user prompt",
	}
	got := injectRecommendMeta("CreateAiRecommendStocks", `{"stockCode":"600000","modelName":"ai-filled"}`, meta)
	if got == "" {
		t.Fatal("expected injected JSON, got empty string")
	}
	var row map[string]any
	if err := json.Unmarshal([]byte(got), &row); err != nil {
		t.Fatalf("invalid injected JSON: %v", err)
	}
	if row["modelName"] != meta.ModelName {
		t.Fatalf("modelName = %v, want %s", row["modelName"], meta.ModelName)
	}
	if row["systemPrompt"] != meta.SystemPrompt {
		t.Fatalf("systemPrompt = %v, want %s", row["systemPrompt"], meta.SystemPrompt)
	}
	if row["userPrompt"] != meta.UserPrompt {
		t.Fatalf("userPrompt = %v, want %s", row["userPrompt"], meta.UserPrompt)
	}
}

func TestFlattenInvestCalendarReadsTimelineGroups(t *testing.T) {
	groups := []any{
		map[string]any{
			"date": "2026-07-13",
			"list": []any{
				map[string]any{
					"title": "中国 · 社融数据",
					"type":  "economic",
					"stock": map[string]any{"code": "CN.TSF", "name": ""},
				},
			},
		},
	}

	rows := flattenInvestCalendar(groups)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].Date != "2026-07-13" || rows[0].Title != "中国 · 社融数据" || rows[0].StockCode != "CN.TSF" {
		t.Fatalf("row = %#v", rows[0])
	}
}

func TestInjectRecommendMetaBatch(t *testing.T) {
	meta := AgentMeta{
		ModelName:    "qwen-plus",
		SystemPrompt: "system prompt",
		UserPrompt:   "user prompt",
	}
	got := injectRecommendMeta("BatchCreateAiRecommendStocks", `{"stocks":[{"stockCode":"600000"},{"stockCode":"000001","modelName":"ai-filled"}]}`, meta)
	if got == "" {
		t.Fatal("expected injected JSON, got empty string")
	}
	var payload struct {
		Stocks []map[string]any `json:"stocks"`
	}
	if err := json.Unmarshal([]byte(got), &payload); err != nil {
		t.Fatalf("invalid injected JSON: %v", err)
	}
	if len(payload.Stocks) != 2 {
		t.Fatalf("len(stocks) = %d, want 2", len(payload.Stocks))
	}
	for i, row := range payload.Stocks {
		if row["modelName"] != meta.ModelName {
			t.Fatalf("stocks[%d].modelName = %v, want %s", i, row["modelName"], meta.ModelName)
		}
		if row["systemPrompt"] != meta.SystemPrompt {
			t.Fatalf("stocks[%d].systemPrompt = %v, want %s", i, row["systemPrompt"], meta.SystemPrompt)
		}
		if row["userPrompt"] != meta.UserPrompt {
			t.Fatalf("stocks[%d].userPrompt = %v, want %s", i, row["userPrompt"], meta.UserPrompt)
		}
	}
}

func TestGetAllDataTools(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "agent-tools.db"))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open agent tool database: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := db.Dao.AutoMigrate(&data.Settings{}, &data.AIConfig{}); err != nil {
		t.Fatalf("migrate agent tool settings: %v", err)
	}
	tools := GetAllDataTools()
	t.Logf("Total tools count: %d", len(tools))

	toolNames := make(map[string]int)
	for i, tool := range tools {
		info, err := tool.Info(nil)
		if err != nil {
			t.Errorf("Tool %d: failed to get info: %v", i, err)
			continue
		}
		t.Logf("Tool %d: %s - %s", i+1, info.Name, info.Desc)

		if count, exists := toolNames[info.Name]; exists {
			t.Errorf("Duplicate tool name found: %s (count: %d)", info.Name, count+1)
		}
		toolNames[info.Name]++
	}

	t.Log("\n=== Tool List ===")
	for name, count := range toolNames {
		if count > 1 {
			t.Errorf("Duplicate tool: %s (count: %d)", name, count)
		}
	}
}
