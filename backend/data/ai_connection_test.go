package data

import (
	"context"
	"go-stock/backend/db"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAIConnectionSuccessReturnsModelsAndLatency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/models", r.URL.Path)
		require.Equal(t, "Bearer valid-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"deepseek-chat"},{"id":"deepseek-reasoner"}]}`))
	}))
	defer server.Close()

	result := TestAIConnection(context.Background(), &AIConfig{
		BaseUrl: server.URL + "/v1", ApiKey: "valid-key", ModelName: "deepseek-chat", TimeOut: 2,
	})

	require.True(t, result.Success)
	require.True(t, result.Reachable)
	require.Equal(t, []string{"deepseek-chat", "deepseek-reasoner"}, result.Models)
	require.GreaterOrEqual(t, result.LatencyMs, int64(0))
}

func TestAIConnectionClassifiesInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	result := TestAIConnection(context.Background(), &AIConfig{BaseUrl: server.URL, ApiKey: "bad", TimeOut: 2})

	require.False(t, result.Success)
	require.True(t, result.Reachable)
	require.Equal(t, "invalid_key", result.Code)
	require.Contains(t, result.Message, "密钥")
}

func TestAIConnectionAllowsManualModelWhenModelsEndpointIsMissing(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	result := TestAIConnection(context.Background(), &AIConfig{BaseUrl: server.URL, ModelName: "manual-model", TimeOut: 2})

	require.False(t, result.Success)
	require.True(t, result.Reachable)
	require.True(t, result.ManualModelAllowed)
	require.Equal(t, "models_unsupported", result.Code)
}

func TestAIConnectionClassifiesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(80 * time.Millisecond)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	result := TestAIConnection(ctx, &AIConfig{BaseUrl: server.URL, TimeOut: 1})

	require.False(t, result.Success)
	require.False(t, result.Reachable)
	require.Equal(t, "timeout", result.Code)
	require.Contains(t, result.Message, "超时")
}

func TestAIConnectionDoesNotReuseStoredKeyForChangedEndpoint(t *testing.T) {
	setupAIWorkspaceTestDB(t)
	stored := &AIConfig{BaseUrl: "https://original.example/v1", ApiKey: "must-not-leak", ModelName: "model"}
	require.NoError(t, db.Dao.Create(stored).Error)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"model"}]}`))
	}))
	defer server.Close()

	result := TestAIConnection(context.Background(), &AIConfig{ID: stored.ID, BaseUrl: server.URL, TimeOut: 2})
	require.True(t, result.Success)
}
