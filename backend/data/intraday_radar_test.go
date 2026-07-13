package data

import (
	"testing"
	"time"
)

func TestClassifyRadarEvent(t *testing.T) {
	event := classifyRadarEvent(StockChangeItem{
		Code:       "300418",
		ChangeType: 8201,
		TypeName:   "火箭发射",
	}, true, false)

	if event.Category != RadarCategoryMomentum {
		t.Fatalf("category = %q, want %q", event.Category, RadarCategoryMomentum)
	}
	if event.Priority != RadarPriorityP1 {
		t.Fatalf("priority = %q, want %q", event.Priority, RadarPriorityP1)
	}
	if !event.IsFollowed || event.IsHolding {
		t.Fatalf("follow/holding flags = %t/%t, want true/false", event.IsFollowed, event.IsHolding)
	}
}

func TestRadarNotificationCooldown(t *testing.T) {
	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.Local)
	service := NewIntradayRadarServiceWithClock(func() time.Time { return now })
	event := RadarEvent{
		StockChangeItem: StockChangeItem{Code: "300418"},
		Category:        RadarCategoryMomentum,
		Priority:        RadarPriorityP1,
	}

	if !service.shouldNotify(event) {
		t.Fatal("first event should notify")
	}
	if service.shouldNotify(event) {
		t.Fatal("same event should be cooled down")
	}
	now = now.Add(16 * time.Minute)
	if !service.shouldNotify(event) {
		t.Fatal("event should notify after the 15-minute cooldown")
	}
}
