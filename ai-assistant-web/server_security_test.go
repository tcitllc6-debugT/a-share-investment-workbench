package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeLoopbackAddr(t *testing.T) {
	t.Setenv("AI_ASSISTANT_WEB_ADDR", "")
	if got := getAddr(); got != "127.0.0.1:18888" {
		t.Fatalf("default address = %q, want loopback", got)
	}

	for _, input := range []string{":19999", "0.0.0.0:19999", "192.168.1.8:19999"} {
		t.Setenv("AI_ASSISTANT_WEB_ADDR", input)
		if got := getAddr(); got != "127.0.0.1:19999" {
			t.Fatalf("address %q normalized to %q", input, got)
		}
	}
}

func TestAuthenticatedAPIRequiresStartupToken(t *testing.T) {
	handler := withSecurity("startup-token", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/prompts", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	authorizedReq := httptest.NewRequest(http.MethodGet, "/api/prompts", nil)
	authorizedReq.Header.Set("Authorization", "Bearer startup-token")
	authorized := httptest.NewRecorder()
	handler.ServeHTTP(authorized, authorizedReq)
	if authorized.Code != http.StatusNoContent {
		t.Fatalf("authorized status = %d", authorized.Code)
	}

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusNoContent {
		t.Fatalf("health status = %d", health.Code)
	}
}

func TestRequestBodyLimit(t *testing.T) {
	handler := withRequestBodyLimit(16, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := readJSONBody(r, &map[string]any{})
		if err != nil {
			writeRequestError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(`{"payload":"this is too large"}`))
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body status = %d", recorder.Code)
	}
}
