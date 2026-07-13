package main

import (
	"testing"
	"time"
)

func TestNeedsFullSchemaMigration(t *testing.T) {
	tests := []struct {
		name    string
		applied int
		target  int
		want    bool
	}{
		{name: "new database", applied: 0, target: 3, want: true},
		{name: "older database", applied: 2, target: 3, want: true},
		{name: "current database", applied: 3, target: 3, want: false},
		{name: "newer database", applied: 4, target: 3, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsFullSchemaMigration(tt.applied, tt.target); got != tt.want {
				t.Fatalf("needsFullSchemaMigration(%d, %d) = %v, want %v", tt.applied, tt.target, got, tt.want)
			}
		})
	}
}

func TestStartupUpdateCheckDelayKeepsFirstPaintAheadOfNetwork(t *testing.T) {
	if startupUpdateCheckDelay < time.Second {
		t.Fatalf("startup update check delay = %s, want at least one second", startupUpdateCheckDelay)
	}
}

func TestStockCatalogBootstrapOnlyRunsForEmptyTable(t *testing.T) {
	if !needsStockCatalogBootstrap(0) {
		t.Fatal("empty catalog must be bootstrapped")
	}
	if needsStockCatalogBootstrap(1) {
		t.Fatal("existing catalog must be updated incrementally instead of rebuilt")
	}
}
