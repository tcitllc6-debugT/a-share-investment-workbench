package data

import "testing"

func TestIsAShareStockCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"000001.SZ", true},
		{"600000.SH", true},
		{"830799.BJ", true},
		{"sz000001", true},
		{"SH600000", true},
		{"bj830799", true},
		{"00700.HK", false},
		{"hk00700", false},
		{"AAPL.US", false},
		{"gb_aapl", false},
		{"513310.SH", true},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsAShareStockCode(tt.code); got != tt.want {
			t.Fatalf("IsAShareStockCode(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}
