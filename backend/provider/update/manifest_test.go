package update

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultManifestUsesGitHubReleaseOverHTTPS(t *testing.T) {
	if !strings.HasPrefix(DefaultManifestURL, "https://github.com/tcitllc6-debugT/a-share-investment-workbench/releases/latest/download/") {
		t.Fatalf("default manifest must use the GitHub release channel: %s", DefaultManifestURL)
	}
}

func TestDefaultManifestProviderMigratesLegacyServerConfiguration(t *testing.T) {
	provider, err := NewDefaultManifestProvider("http://legacy-updates.invalid/a-share-desk/manifest.json")
	if err != nil {
		t.Fatalf("create default provider: %v", err)
	}
	if provider.ManifestURL != DefaultManifestURL {
		t.Fatalf("legacy update URL should migrate to GitHub Releases: %s", provider.ManifestURL)
	}
}

func TestDefaultManifestProviderKeepsRepositoryReleaseURL(t *testing.T) {
	want := "https://github.com/tcitllc6-debugT/a-share-investment-workbench/releases/download/v1.2.3/manifest.json"
	provider, err := NewDefaultManifestProvider(want)
	if err != nil {
		t.Fatalf("create default provider: %v", err)
	}
	if provider.ManifestURL != want {
		t.Fatalf("trusted repository release URL changed: %s", provider.ManifestURL)
	}
}

func TestManifestProviderReturnsSignedNewerPlatformAsset(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	manifest := Manifest{
		Version: "1.2.0",
		Notes:   "test update",
		Assets: []Asset{{
			OS:     runtime.GOOS,
			Arch:   runtime.GOARCH,
			URL:    "https://updates.example.test/a-share-desk.zip",
			SHA256: "abc123",
		}},
	}
	if err := SignManifest(&manifest, privateKey); err != nil {
		t.Fatalf("sign manifest: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(manifest)
	}))
	defer server.Close()

	provider := NewManifestProvider(server.URL, publicKey)
	result, err := provider.Check(t.Context(), "1.1.0")
	if err != nil {
		t.Fatalf("check signed manifest: %v", err)
	}
	if !result.Available || result.Version != "1.2.0" || result.URL != manifest.Assets[0].URL || result.SHA256 != "abc123" {
		t.Fatalf("unexpected update result: %+v", result)
	}
}

func TestManifestProviderRejectsInvalidSignature(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	manifest := Manifest{Version: "1.2.0", Signature: "invalid", Assets: []Asset{{OS: runtime.GOOS, Arch: runtime.GOARCH, URL: "https://updates.example.test/app.zip", SHA256: "abc"}}}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(manifest)
	}))
	defer server.Close()

	_, err = NewManifestProvider(server.URL, publicKey).Check(t.Context(), "1.1.0")
	if err == nil {
		t.Fatal("expected invalid manifest signature to be rejected")
	}
}

func TestCompareVersions(t *testing.T) {
	if compareVersions("v1.10.0", "1.9.9") <= 0 {
		t.Fatal("expected 1.10.0 to be newer than 1.9.9")
	}
	if compareVersions("1.0.0", "1.0.0") != 0 {
		t.Fatal("expected equal versions")
	}
}

func TestReleaseManifestArtifactWhenConfigured(t *testing.T) {
	manifestPath := strings.TrimSpace(os.Getenv("A_SHARE_RELEASE_MANIFEST"))
	if manifestPath == "" {
		t.Skip("A_SHARE_RELEASE_MANIFEST is not configured")
	}
	encoded, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read release manifest: %v", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(encoded, &manifest); err != nil {
		t.Fatalf("decode release manifest: %v", err)
	}
	publicKey, err := base64.StdEncoding.DecodeString(ManifestPublicKeyBase64)
	if err != nil {
		t.Fatalf("decode embedded public key: %v", err)
	}
	if err := VerifyManifest(manifest, ed25519.PublicKey(publicKey)); err != nil {
		t.Fatalf("verify release manifest: %v", err)
	}
	for _, asset := range manifest.Assets {
		assetPath := filepath.Join(filepath.Dir(manifestPath), asset.Name)
		info, err := os.Stat(assetPath)
		if err != nil {
			t.Fatalf("stat release asset %s: %v", asset.Name, err)
		}
		if info.Size() != asset.Size {
			t.Fatalf("release asset %s size=%d want=%d", asset.Name, info.Size(), asset.Size)
		}
		matches, err := fileMatchesSHA256(assetPath, asset.SHA256)
		if err != nil || !matches {
			t.Fatalf("release asset %s checksum mismatch: matches=%v err=%v", asset.Name, matches, err)
		}
	}
}
