package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	shareprovider "go-stock/backend/provider/share"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultAddr          = "127.0.0.1:18888"
	assistantTokenCookie = "a_share_assistant_token"
	maxRequestBodyBytes  = 4 * 1024 * 1024
	maxQuestionBytes     = 32 * 1024
)

//go:embed static
var staticFS embed.FS

type app struct{}

type Config struct {
	Addr  string
	Token string
}

type aiConfigResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	BaseURL   string `json:"baseUrl"`
	ModelName string `json:"modelName"`
}

type chatRequest struct {
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID int    `json:"sysPromptId"`
	Thinking    bool   `json:"thinking"`
	EnableTools bool   `json:"enableTools"`
	HistoryJSON string `json:"historyJSON"`
}

type sessionSaveRequest struct {
	SessionId string                      `json:"sessionId"`
	Messages  []models.AiAssistantMessage `json:"messages"`
}

type shareRequest struct {
	Text  string `json:"text"`
	Title string `json:"title"`
}

// Start 在当前进程内启动 ai-assistant-web 服务（阻塞，适合放在 goroutine 中）。
func Start() error {
	return StartContext(context.Background(), Config{})
}

func StartContext(ctx context.Context, config Config) error {
	checkDir("data")
	checkDir("logs")

	// When launched as a desktop companion, db may already be initialized.
	if db.Dao == nil {
		db.Init("")
	}
	autoMigrate()
	data.InitAnalyzeSentiment()

	a := &app{}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", a.health)
	mux.HandleFunc("/api/vip-status", a.vipStatus)
	mux.HandleFunc("/api/ai-configs", a.getAIConfigs)
	mux.HandleFunc("/api/prompts", a.getPrompts)
	mux.HandleFunc("/api/session", a.session)
	mux.HandleFunc("/api/chat/summary-stream", a.summaryChatStream)
	mux.HandleFunc("/api/share", a.shareText)

	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		logger.SugaredLogger.Fatalf("load static files failed: %v", err)
	}
	staticServer := http.FileServer(http.FS(subFS))
	token := strings.TrimSpace(config.Token)
	if token == "" {
		token = resolveToken()
	}
	mux.Handle("/", withSessionCookie(token, staticServer))

	addr := normalizeLoopbackAddr(config.Addr)
	if strings.TrimSpace(config.Addr) == "" {
		addr = getAddr()
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           withRequestBodyLimit(maxRequestBodyBytes, withSecurity(token, mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	logger.SugaredLogger.Infof("ai-assistant-web started at: %s", addr)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *app) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"time": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (a *app) vipStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	level, active := data.EffectiveSponsorVipLevel()
	ok := data.IsLocalFeatureEnabled("ai_assistant_web") || (active && level >= 2)
	payload := map[string]any{
		"ok":       ok,
		"vipLevel": level,
		"active":   active,
	}
	if !ok {
		payload["message"] = vipDeniedMessage(level, active)
	}
	writeJSON(w, http.StatusOK, payload)
}

func vipDeniedMessage(level int, active bool) string {
	if !active && level > 0 {
		return "本地功能当前未启用，请检查 A股智投工作台 配置。"
	}
	return "A股智投工作台 AI 助手（Web）当前未启用，请使用与桌面端相同的 data 目录启动服务。"
}

func requireVip2(w http.ResponseWriter) bool {
	if data.IsLocalFeatureEnabled("ai_assistant_web") {
		return true
	}
	level, active := data.EffectiveSponsorVipLevel()
	if active && level >= 2 {
		return true
	}
	writeJSON(w, http.StatusForbidden, map[string]any{
		"code":     "VIP2_REQUIRED",
		"message":  vipDeniedMessage(level, active),
		"vipLevel": level,
		"active":   active,
	})
	return false
}

func (a *app) getAIConfigs(w http.ResponseWriter, _ *http.Request) {
	if !requireVip2(w) {
		return
	}
	cfgs := data.GetSettingConfig().AiConfigs
	resp := make([]aiConfigResp, 0, len(cfgs))
	for _, c := range cfgs {
		resp = append(resp, aiConfigResp{
			ID:        c.ID,
			Name:      c.Name,
			BaseURL:   c.BaseUrl,
			ModelName: c.ModelName,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *app) getPrompts(w http.ResponseWriter, r *http.Request) {
	if !requireVip2(w) {
		return
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	promptType := strings.TrimSpace(r.URL.Query().Get("type"))
	res := data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
	writeJSON(w, http.StatusOK, res)
}

func (a *app) session(w http.ResponseWriter, r *http.Request) {
	if !requireVip2(w) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		sessionId := r.URL.Query().Get("sessionId")
		res, err := data.GetAiAssistantSession(sessionId)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	case http.MethodPost:
		var req sessionSaveRequest
		if _, err := readJSONBody(r, &req); err != nil {
			writeRequestError(w, err)
			return
		}
		if err := data.SaveAiAssistantSession(req.SessionId, req.Messages); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *app) summaryChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !requireVip2(w) {
		return
	}

	var req chatRequest
	if _, err := readJSONBody(r, &req); err != nil {
		writeRequestError(w, err)
		return
	}
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "question is required"})
		return
	}
	if len(req.Question) > maxQuestionBytes {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "question is too large"})
		return
	}

	if req.AIConfigID <= 0 {
		cfgs := data.GetSettingConfig().AiConfigs
		if len(cfgs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no ai config found"})
			return
		}
		req.AIConfigID = int(cfgs[0].ID)
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "stream not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var sysPromptID *int
	if req.SysPromptID > 0 {
		sysPromptID = &req.SysPromptID
	}

	history := parseHistory(req.HistoryJSON)
	tools := make([]data.Tool, 0)
	if req.EnableTools {
		tools = data.Tools(tools)
	}
	o := data.NewDeepSeekOpenAi(ctx, req.AIConfigID)
	var ch <-chan map[string]any
	if req.EnableTools {
		ch = o.NewSummaryStockNewsStreamWithTools(req.Question, sysPromptID, tools, req.Thinking, history)
	} else {
		ch = o.NewSummaryStockNewsStream(req.Question, sysPromptID, req.Thinking, history)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				_, _ = w.Write([]byte("event: done\ndata: [DONE]\n\n"))
				flusher.Flush()
				return
			}
			raw, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			_, _ = w.Write([]byte("data: " + string(raw) + "\n\n"))
			flusher.Flush()
		}
	}
}

func (a *app) shareText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !requireVip2(w) {
		return
	}
	var req shareRequest
	if _, err := readJSONBody(r, &req); err != nil {
		writeRequestError(w, err)
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	req.Title = strings.TrimSpace(req.Title)
	if req.Text == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "内容为空"})
		return
	}
	if req.Title == "" {
		req.Title = "AI助手"
	}
	path, err := shareprovider.NewLocalProvider(filepath.Join("data", "shares")).ShareText(req.Text, req.Title)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "saved to " + path})
}

func parseHistory(historyJSON string) []map[string]interface{} {
	historyJSON = strings.TrimSpace(historyJSON)
	if historyJSON == "" {
		return nil
	}
	var list []models.AiAssistantMessage
	if err := json.Unmarshal([]byte(historyJSON), &list); err != nil {
		return nil
	}
	history := make([]map[string]interface{}, 0, len(list))
	for _, m := range list {
		item := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if m.Reasoning != "" {
			item["reasoning_content"] = m.Reasoning
		}
		history = append(history, item)
	}
	return history
}

func (a *app) chatStream(w http.ResponseWriter, r *http.Request) {
	// 保留旧接口兼容，转发到新的 summary 接口能力
	a.summaryChatStream(w, r)
}

func autoMigrate() {
	db.Dao.AutoMigrate(&data.Settings{})
	db.Dao.AutoMigrate(&data.AIConfig{})
	db.Dao.AutoMigrate(&models.PromptTemplate{})
	db.Dao.AutoMigrate(&models.AiAssistantSession{})
}

func getAddr() string {
	return normalizeLoopbackAddr(os.Getenv("AI_ASSISTANT_WEB_ADDR"))
}

func normalizeLoopbackAddr(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultAddr
	}
	if strings.HasPrefix(raw, ":") {
		raw = "127.0.0.1" + raw
	}
	_, port, err := net.SplitHostPort(raw)
	if err != nil || strings.TrimSpace(port) == "" {
		return defaultAddr
	}
	return net.JoinHostPort("127.0.0.1", port)
}

func EnabledFromEnvironment() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AI_ASSISTANT_WEB_ENABLED"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func resolveToken() string {
	if configured := strings.TrimSpace(os.Getenv("AI_ASSISTANT_WEB_TOKEN")); configured != "" {
		return configured
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func checkDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		_ = os.Mkdir(dir, os.ModePerm)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withSessionCookie(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     assistantTokenCookie,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		next.ServeHTTP(w, r)
	})
}

func withSecurity(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")

		origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/")
		if origin != "" {
			expected := "http://" + r.Host
			if origin != expected {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "cross-origin request denied"})
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/health" && !requestHasToken(r, token) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "local assistant token required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestHasToken(r *http.Request, token string) bool {
	provided := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if provided == "" {
		if cookie, err := r.Cookie(assistantTokenCookie); err == nil {
			provided = cookie.Value
		}
	}
	return len(provided) == len(token) && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) == 1
}

func withRequestBodyLimit(limit int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}
		next.ServeHTTP(w, r)
	})
}

func readJSONBody(r *http.Request, dst any) (bool, error) {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return false, err
	}
	return true, nil
}

func writeRequestError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body is too large"})
		return
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
}
