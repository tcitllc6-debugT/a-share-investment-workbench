package main

func monitorWorkAreaSize(left, top, right, bottom int32) (int, int) {
	return int(right - left), int(bottom - top)
}

// resolveInitialWindowSize keeps a usable first-run window on high-DPI displays
// while preserving any previously saved size that still fits the active screen.
func resolveInitialWindowSize(screenWidth, screenHeight, savedWidth, savedHeight int) (int, int) {
	if screenWidth <= 0 || screenHeight <= 0 {
		return 1000, 800
	}

	if savedWidth > 0 && savedHeight > 0 {
		return min(savedWidth, screenWidth), min(savedHeight, screenHeight)
	}

	return screenWidth, screenHeight
}
