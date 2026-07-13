package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestDownloadVerifiedAssetRejectsChecksumMismatchWithoutLeavingOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("signed update payload"))
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "update.zip")
	if err := DownloadVerifiedAsset(context.Background(), server.URL, "wrong", destination); err == nil {
		t.Fatal("expected checksum mismatch")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("expected no verified output after checksum mismatch, got %v", err)
	}
}

func TestDownloadVerifiedAssetWritesMatchingPayload(t *testing.T) {
	payload := []byte("signed update payload")
	digest := sha256.Sum256(payload)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(payload)
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "update.zip")
	if err := DownloadVerifiedAsset(context.Background(), server.URL, hex.EncodeToString(digest[:]), destination); err != nil {
		t.Fatalf("download verified asset: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil || string(got) != string(payload) {
		t.Fatalf("unexpected downloaded payload %q, error %v", string(got), err)
	}
}

func TestDownloadVerifiedAssetSerializesConcurrentRequestsForOneDestination(t *testing.T) {
	payload := []byte("signed update payload")
	digest := sha256.Sum256(payload)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(payload)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "update.zip")
	start := make(chan struct{})
	errors := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			errors <- DownloadVerifiedAsset(context.Background(), server.URL, hex.EncodeToString(digest[:]), destination)
		}()
	}
	close(start)
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent verified download failed: %v", err)
		}
	}
}
