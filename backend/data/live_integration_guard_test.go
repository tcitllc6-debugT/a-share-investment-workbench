package data

import (
	"os"
	"strings"
	"testing"
)

const liveIntegrationEnv = "RUN_LIVE_INTEGRATION_TESTS"

func liveIntegrationEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(liveIntegrationEnv))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func requireLiveIntegration(t testing.TB) {
	t.Helper()
	if !liveIntegrationEnabled() {
		t.Skipf("set %s=1 to run real browser and network integration tests", liveIntegrationEnv)
	}
}

func TestLiveIntegrationEnabled(t *testing.T) {
	t.Setenv(liveIntegrationEnv, "")
	if liveIntegrationEnabled() {
		t.Fatal("live integration tests must be disabled by default")
	}

	t.Setenv(liveIntegrationEnv, "true")
	if !liveIntegrationEnabled() {
		t.Fatal("live integration tests should be enabled explicitly")
	}
}

func TestBrowserIntegrationTestsDoNotLaunchVisibleWindows(t *testing.T) {
	source, err := os.ReadFile("crawler_api_test.go")
	if err != nil {
		t.Fatalf("read crawler integration test: %v", err)
	}
	if strings.Contains(string(source), "GetHtmlWithActions(&actions, false)") {
		t.Fatal("crawler integration tests must run headless to avoid opening a browser window")
	}
	if strings.Contains(string(source), `chromedp.Navigate("http`) {
		t.Fatal("crawler integration tests must read target URLs from the environment instead of hard-coding stock pages")
	}
}

func TestRuntimeSourcesContainNoEmbeddedSessionCredentials(t *testing.T) {
	source, err := os.ReadFile("market_news_api.go")
	if err != nil {
		t.Fatalf("read market news source: %v", err)
	}
	for _, forbidden := range []string{`SetHeader("token",`, "SESSION="} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("market news runtime source still embeds credential marker %q", forbidden)
		}
	}
}

func TestFollowStocksHasNoPaidLimit(t *testing.T) {
	source, err := os.ReadFile("stock_data_api.go")
	if err != nil {
		t.Fatalf("read stock data source: %v", err)
	}
	for _, forbidden := range []string{"升级VIP", "count >= 63"} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("stock follow flow still contains paid limit marker %q", forbidden)
		}
	}
}
