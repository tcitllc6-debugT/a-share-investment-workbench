package main

import "testing"

func TestResolveInitialWindowSizeUsesFullLogicalScreenOnFirstRun(t *testing.T) {
	width, height := resolveInitialWindowSize(640, 360, 0, 0)
	if width != 640 || height != 360 {
		t.Fatalf("got %dx%d, want 640x360", width, height)
	}
}

func TestResolveInitialWindowSizeClampsSavedWindowToCurrentScreen(t *testing.T) {
	width, height := resolveInitialWindowSize(1280, 720, 2200, 1500)
	if width != 1280 || height != 720 {
		t.Fatalf("got %dx%d, want 1280x720", width, height)
	}
}

func TestResolveInitialWindowSizePreservesSavedWindowThatFits(t *testing.T) {
	width, height := resolveInitialWindowSize(1280, 720, 1080, 680)
	if width != 1080 || height != 680 {
		t.Fatalf("got %dx%d, want 1080x680", width, height)
	}
}

func TestMonitorWorkAreaSizeUsesWorkRectangleInsteadOfFullMonitor(t *testing.T) {
	width, height := monitorWorkAreaSize(0, 0, 3840, 2100)
	if width != 3840 || height != 2100 {
		t.Fatalf("got %dx%d, want 3840x2100", width, height)
	}
}
