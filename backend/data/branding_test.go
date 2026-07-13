package data

import (
	"os"
	"strings"
	"testing"
)

func TestUserVisibleNotificationsUseCurrentBrand(t *testing.T) {
	for _, path := range []string{"dingding_api.go", "feishu_api.go", "intraday_radar.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(source), "A-Share Desk") {
			t.Fatalf("%s still contains the retired user-visible brand", path)
		}
	}
}
