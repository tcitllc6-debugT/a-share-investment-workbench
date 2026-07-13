package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var updateDownloadMu sync.Mutex

type PendingUpdate struct {
	ExpectedVersion string    `json:"expectedVersion"`
	PackagePath     string    `json:"packagePath"`
	InstallDir      string    `json:"installDir"`
	Executable      string    `json:"executable"`
	StartedAt       time.Time `json:"startedAt"`
	Healthy         bool      `json:"healthy"`
	HealthyAt       time.Time `json:"healthyAt,omitempty"`
}

func DownloadVerifiedAsset(ctx context.Context, assetURL, expectedSHA256, destination string) error {
	expectedSHA256 = strings.TrimSpace(expectedSHA256)
	if expectedSHA256 == "" {
		return fmt.Errorf("update asset checksum is missing")
	}
	updateDownloadMu.Lock()
	defer updateDownloadMu.Unlock()

	if matches, err := fileMatchesSHA256(destination, expectedSHA256); err != nil {
		return err
	} else if matches {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download update asset: unexpected HTTP status %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(destination), filepath.Base(destination)+".*.download")
	if err != nil {
		return err
	}
	temporaryPath := file.Name()
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(file, hash), resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	actualSHA256 := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualSHA256, expectedSHA256) {
		return fmt.Errorf("update asset checksum mismatch")
	}
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace existing update asset: %w", err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return err
	}
	removeTemporary = false
	return nil
}

func fileMatchesSHA256(path, expectedSHA256 string) (bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open existing update asset: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("hash existing update asset: %w", err)
	}
	return strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expectedSHA256), nil
}

func WritePendingUpdate(statePath string, state PendingUpdate) error {
	if state.StartedAt.IsZero() {
		state.StartedAt = time.Now()
	}
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath, append(encoded, '\n'), 0600)
}

func MarkPendingUpdateHealthy(statePath, currentVersion string) error {
	encoded, err := os.ReadFile(statePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var state PendingUpdate
	if err := json.Unmarshal(encoded, &state); err != nil {
		return err
	}
	if state.ExpectedVersion != "" && compareVersions(currentVersion, state.ExpectedVersion) != 0 {
		return fmt.Errorf("pending update version does not match current version")
	}
	state.Healthy = true
	state.HealthyAt = time.Now()
	return WritePendingUpdate(statePath, state)
}
