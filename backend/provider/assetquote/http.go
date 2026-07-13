package assetquote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxResponseBytes = 2 << 20

func defaultHTTPClient() *http.Client {
	return &http.Client{}
}

func doRequest(ctx context.Context, client *http.Client, provider, endpoint string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", provider, err)
	}
	req.Header.Set("Accept", "application/json,text/plain,*/*")
	req.Header.Set("User-Agent", "A-Share-Workbench/1.0")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &HTTPError{Provider: provider, Retryable: true, Message: err.Error()}
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if readErr != nil {
		return nil, &HTTPError{Provider: provider, Retryable: true, Message: readErr.Error()}
	}
	if len(body) > maxResponseBytes {
		return nil, &HTTPError{Provider: provider, Retryable: false, Message: "response is too large"}
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(body))
		if len(message) > 240 {
			message = message[:240]
		}
		return nil, &HTTPError{
			Provider: provider, StatusCode: resp.StatusCode,
			Retryable: resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError,
			Message:   message,
		}
	}
	return body, nil
}

func doJSON(ctx context.Context, client *http.Client, provider, endpoint string, target any) error {
	return doJSONWithHeaders(ctx, client, provider, endpoint, nil, target)
}

func doJSONWithHeaders(ctx context.Context, client *http.Client, provider, endpoint string, headers map[string]string, target any) error {
	body, err := doRequest(ctx, client, provider, endpoint, headers)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return &HTTPError{Provider: provider, Retryable: false, Message: "decode response: " + err.Error()}
	}
	return nil
}

func fetchQuoteBatch(ctx context.Context, requests []QuoteRequest, fetch func(context.Context, QuoteRequest) (Quote, error)) ([]Quote, error) {
	quotes := make([]Quote, 0, len(requests))
	for _, request := range requests {
		quote, err := fetch(ctx, request)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}
