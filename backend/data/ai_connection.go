package data

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-stock/backend/db"
)

type AIConnectionResult struct {
	Success            bool     `json:"success"`
	Reachable          bool     `json:"reachable"`
	ManualModelAllowed bool     `json:"manualModelAllowed"`
	Code               string   `json:"code"`
	Message            string   `json:"message"`
	LatencyMs          int64    `json:"latencyMs"`
	StatusCode         int      `json:"statusCode"`
	Models             []string `json:"models"`
}

func connectionConfigWithStoredKey(input *AIConfig) *AIConfig {
	if input == nil {
		return &AIConfig{}
	}
	copyItem := *input
	if copyItem.ID > 0 && strings.TrimSpace(copyItem.ApiKey) == "" {
		var stored AIConfig
		if err := db.Dao.First(&stored, copyItem.ID).Error; err == nil {
			if sameAIEndpoint(copyItem.BaseUrl, stored.BaseUrl) {
				copyItem.ApiKey = stored.ApiKey
			}
		}
	}
	return &copyItem
}

func sameAIEndpoint(left, right string) bool {
	return strings.TrimRight(strings.TrimSpace(left), "/") == strings.TrimRight(strings.TrimSpace(right), "/")
}

func TestAIConnection(ctx context.Context, input *AIConfig) *AIConnectionResult {
	result := &AIConnectionResult{Models: []string{}}
	config := connectionConfigWithStoredKey(input)
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseUrl), "/")
	if baseURL == "" {
		result.Code = "invalid_url"
		result.Message = "请先选择服务商或填写接口地址"
		return result
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		result.Code = "invalid_url"
		result.Message = "接口地址格式不正确，请检查 http:// 或 https://"
		return result
	}

	timeout := time.Duration(config.TimeOut) * time.Second
	if timeout <= 0 || timeout > 30*time.Second {
		timeout = 15 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if config.HttpProxyEnabled && strings.TrimSpace(config.HttpProxy) != "" {
		proxyURL, proxyErr := url.Parse(strings.TrimSpace(config.HttpProxy))
		if proxyErr != nil {
			result.Code = "invalid_proxy"
			result.Message = "代理地址格式不正确"
			return result
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	client := &http.Client{Timeout: timeout, Transport: transport}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		result.Code = "invalid_url"
		result.Message = "无法创建连接请求，请检查接口地址"
		return result
	}
	if apiKey := strings.TrimSpace(config.ApiKey); apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
	request.Header.Set("Accept", "application/json")

	startedAt := time.Now()
	response, err := client.Do(request)
	result.LatencyMs = time.Since(startedAt).Milliseconds()
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
			result.Code = "timeout"
			result.Message = "连接超时，请检查网络、代理或适当增加响应超时"
		} else {
			result.Code = "unreachable"
			result.Message = "无法连接到模型服务，请检查接口地址和网络"
		}
		return result
	}
	defer response.Body.Close()
	result.Reachable = true
	result.StatusCode = response.StatusCode

	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		result.Code = "invalid_key"
		result.Message = "API 密钥无效或没有访问权限"
		return result
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		result.Code = "models_unsupported"
		result.Message = "服务可以访问，但不支持自动读取模型；请手动输入模型名称"
		result.ManualModelAllowed = true
		return result
	case http.StatusTooManyRequests:
		result.Code = "rate_limited"
		result.Message = "服务请求过于频繁或账户额度不足，请稍后重试"
		return result
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		result.Code = "provider_error"
		result.Message = "模型服务返回异常状态，请检查服务商控制台"
		return result
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		result.Code = "invalid_response"
		result.Message = "服务已经连接，但返回内容不是可识别的模型列表"
		result.ManualModelAllowed = true
		return result
	}
	seen := map[string]struct{}{}
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if id != "" {
			if _, ok := seen[id]; !ok {
				result.Models = append(result.Models, id)
				seen[id] = struct{}{}
			}
		}
	}
	for _, item := range payload.Models {
		name := strings.TrimSpace(item.Name)
		if name != "" {
			if _, ok := seen[name]; !ok {
				result.Models = append(result.Models, name)
				seen[name] = struct{}{}
			}
		}
	}
	result.Success = true
	result.Code = "ok"
	if len(result.Models) > 0 {
		result.Message = "连接成功，已读取可用模型"
	} else {
		result.Message = "连接成功，但服务没有返回模型；可以手动输入模型名称"
		result.ManualModelAllowed = true
	}
	return result
}
