package update

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultManifestURL      = "https://github.com/tcitllc6-debugT/a-share-investment-workbench/releases/latest/download/manifest.json"
	githubReleaseHost       = "github.com"
	githubReleasePathPrefix = "/tcitllc6-debugT/a-share-investment-workbench/releases/"
)

// ManifestPublicKeyBase64 is the public half of the release signing key. The
// matching private key is deliberately kept outside the repository.
const ManifestPublicKeyBase64 = "uEJnCxT8yz1DhDU2YyE0zZDWblXSwXWb/ThkDzWzxi4="

type Asset struct {
	Name   string `json:"name"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type Manifest struct {
	Version     string  `json:"version"`
	PublishedAt string  `json:"publishedAt"`
	Notes       string  `json:"notes"`
	Assets      []Asset `json:"assets"`
	Signature   string  `json:"signature"`
}

type ManifestProvider struct {
	ManifestURL string
	PublicKey   ed25519.PublicKey
	Client      *http.Client
}

func NewManifestProvider(manifestURL string, publicKey ed25519.PublicKey) *ManifestProvider {
	return &ManifestProvider{
		ManifestURL: strings.TrimSpace(manifestURL),
		PublicKey:   append(ed25519.PublicKey(nil), publicKey...),
		Client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func NewDefaultManifestProvider(manifestURL string) (*ManifestProvider, error) {
	manifestURL = normalizeManagedManifestURL(manifestURL)
	publicKey, err := base64.StdEncoding.DecodeString(ManifestPublicKeyBase64)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid embedded update public key")
	}
	return NewManifestProvider(manifestURL, ed25519.PublicKey(publicKey)), nil
}

func normalizeManagedManifestURL(manifestURL string) string {
	manifestURL = strings.TrimSpace(manifestURL)
	parsed, err := url.Parse(manifestURL)
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), githubReleaseHost) ||
		!strings.HasPrefix(strings.ToLower(parsed.EscapedPath()), strings.ToLower(githubReleasePathPrefix)) {
		return DefaultManifestURL
	}
	return parsed.String()
}

func (p *ManifestProvider) Check(ctx context.Context, currentVersion string) (Result, error) {
	if p == nil || strings.TrimSpace(p.ManifestURL) == "" {
		return Result{Source: "managed-manifest"}, fmt.Errorf("update manifest URL is not configured")
	}
	if len(p.PublicKey) != ed25519.PublicKeySize {
		return Result{Source: "managed-manifest"}, fmt.Errorf("embedded update public key is invalid")
	}
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.ManifestURL, nil)
	if err != nil {
		return Result{Source: "managed-manifest"}, fmt.Errorf("create update manifest request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Source: "managed-manifest"}, fmt.Errorf("fetch update manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{Source: "managed-manifest"}, fmt.Errorf("fetch update manifest: unexpected HTTP status %d", resp.StatusCode)
	}
	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return Result{Source: "managed-manifest"}, fmt.Errorf("decode update manifest: %w", err)
	}
	if err := VerifyManifest(manifest, p.PublicKey); err != nil {
		return Result{Source: "managed-manifest"}, err
	}
	result := Result{Version: manifest.Version, Notes: manifest.Notes, Source: "managed-manifest"}
	if compareVersions(manifest.Version, currentVersion) <= 0 {
		return result, nil
	}
	asset, ok := manifest.platformAsset(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return result, nil
	}
	result.Available = true
	result.URL = asset.URL
	result.SHA256 = asset.SHA256
	result.Size = asset.Size
	result.AssetName = asset.Name
	return result, nil
}

func (m Manifest) platformAsset(goos, goarch string) (Asset, bool) {
	for _, asset := range m.Assets {
		if strings.EqualFold(strings.TrimSpace(asset.OS), goos) && strings.EqualFold(strings.TrimSpace(asset.Arch), goarch) {
			parsed, err := url.Parse(strings.TrimSpace(asset.URL))
			if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				return Asset{}, false
			}
			return asset, true
		}
	}
	return Asset{}, false
}

func SignManifest(manifest *Manifest, privateKey ed25519.PrivateKey) error {
	if manifest == nil || len(privateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid manifest signing key")
	}
	payload, err := manifestPayload(*manifest)
	if err != nil {
		return err
	}
	manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	return nil
}

func VerifyManifest(manifest Manifest, publicKey ed25519.PublicKey) error {
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("update manifest does not contain a version")
	}
	signature, err := base64.StdEncoding.DecodeString(manifest.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("update manifest signature is invalid")
	}
	payload, err := manifestPayload(manifest)
	if err != nil {
		return fmt.Errorf("encode update manifest for verification: %w", err)
	}
	if !ed25519.Verify(publicKey, payload, signature) {
		return fmt.Errorf("update manifest signature verification failed")
	}
	return nil
}

func manifestPayload(manifest Manifest) ([]byte, error) {
	manifest.Signature = ""
	return json.Marshal(manifest)
}

func compareVersions(left, right string) int {
	leftParts, leftPre := versionParts(left)
	rightParts, rightPre := versionParts(right)
	for index := 0; index < 3; index++ {
		if leftParts[index] > rightParts[index] {
			return 1
		}
		if leftParts[index] < rightParts[index] {
			return -1
		}
	}
	if leftPre == rightPre {
		return 0
	}
	if leftPre == "" {
		return 1
	}
	if rightPre == "" {
		return -1
	}
	return strings.Compare(leftPre, rightPre)
}

func versionParts(version string) ([3]int, string) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	base, pre, hasPre := strings.Cut(version, "-")
	if !hasPre {
		pre = ""
	}
	parts := strings.Split(base, ".")
	result := [3]int{}
	for index := 0; index < len(parts) && index < len(result); index++ {
		value, err := strconv.Atoi(parts[index])
		if err == nil && value >= 0 {
			result[index] = value
		}
	}
	return result, pre
}
