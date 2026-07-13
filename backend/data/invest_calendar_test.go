package data

import (
	"testing"
	"time"
)

func TestCalendarMonthRangeUsesWholeLocalMonth(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	start, end := calendarMonthRange("2028-02", time.Date(2026, 7, 13, 10, 0, 0, 0, location))

	if got := start.In(location).Format("2006-01-02 15:04:05"); got != "2028-02-01 00:00:00" {
		t.Fatalf("start = %s", got)
	}
	if got := end.In(location).Format("2006-01-02 15:04:05"); got != "2028-03-01 00:00:00" {
		t.Fatalf("end = %s", got)
	}
}

func TestWallstreetCalendarTimelineGroupsAndSortsEvents(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	items := []WSCNCalendarItem{
		{ID: 2, PublicDate: time.Date(2026, 7, 2, 9, 0, 0, 0, location).Unix(), Country: "中国", Title: "制造业数据", Importance: 2},
		{ID: 1, PublicDate: time.Date(2026, 7, 1, 21, 30, 0, 0, location).Unix(), Country: "美国", Event: "非农就业", Importance: 3, CalendarType: "economic", WscnTicker: "US.NFP"},
		{ID: 3, PublicDate: time.Date(2026, 7, 1, 8, 0, 0, 0, location).Unix(), Country: "中国", Event: "财新 PMI", Importance: 1},
	}

	timeline := wallstreetCalendarTimeline(items, location)
	if len(timeline) != 2 {
		t.Fatalf("timeline groups = %d, want 2", len(timeline))
	}
	first := timeline[0].(map[string]any)
	if first["date"] != "2026-07-01" {
		t.Fatalf("first date = %#v", first["date"])
	}
	events := first["list"].([]any)
	if len(events) != 2 {
		t.Fatalf("first day events = %d, want 2", len(events))
	}
	firstEvent := events[0].(map[string]any)
	if firstEvent["article_id"] != int64(3) || firstEvent["title"] != "中国 · 财新 PMI" {
		t.Fatalf("first event = %#v", firstEvent)
	}
	secondGroup := timeline[1].(map[string]any)
	secondEvent := secondGroup["list"].([]any)[0].(map[string]any)
	if secondEvent["title"] != "中国 · 制造业数据" {
		t.Fatalf("title fallback = %#v", secondEvent["title"])
	}
}

func TestWallstreetCalendarTimelineReturnsNonNilEmptySlice(t *testing.T) {
	timeline := wallstreetCalendarTimeline(nil, time.Local)
	if timeline == nil || len(timeline) != 0 {
		t.Fatalf("timeline = %#v, want non-nil empty slice", timeline)
	}
}

func TestCollectCalendarPagesFollowsCursorAndDeduplicates(t *testing.T) {
	calls := make([]int64, 0)
	result := collectCalendarPages(func(cursor int64) *WSCNCalendarResponse {
		calls = append(calls, cursor)
		page := &WSCNCalendarResponse{Code: 20000}
		page.Data.TotalCount = 3
		switch cursor {
		case 0:
			page.Data.Items = []WSCNCalendarItem{{ID: 1}, {ID: 2}}
			page.Data.NextCursor = 100
		case 100:
			page.Data.Items = []WSCNCalendarItem{{ID: 2}, {ID: 3}}
			page.Data.NextCursor = 200
		default:
			t.Fatalf("unexpected cursor %d", cursor)
		}
		return page
	})

	if result == nil || len(result.Data.Items) != 3 {
		t.Fatalf("result = %#v, want three unique items", result)
	}
	if len(calls) != 2 || calls[0] != 0 || calls[1] != 100 {
		t.Fatalf("cursors = %#v", calls)
	}
}

func TestCollectCalendarPagesStopsOnRepeatedCursor(t *testing.T) {
	calls := 0
	result := collectCalendarPages(func(cursor int64) *WSCNCalendarResponse {
		calls++
		page := &WSCNCalendarResponse{Code: 20000}
		page.Data.TotalCount = 99
		page.Data.Items = []WSCNCalendarItem{{ID: int64(calls)}}
		page.Data.NextCursor = 100
		return page
	})

	if result == nil || len(result.Data.Items) != 2 || calls != 2 {
		t.Fatalf("result=%#v calls=%d, want two pages before repeated cursor", result, calls)
	}
}
