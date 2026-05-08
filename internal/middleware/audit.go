package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/rainlf/mango-crew/pkg/response"
)

const maxAuditPayloadSize = 64 * 1024
const defaultAuditQueueSize = 1024
const auditWriteTimeout = 3 * time.Second

type auditResponseWriter struct {
	gin.ResponseWriter
	body       strings.Builder
	truncated  bool
	writeLimit int
}

type AuditProcessor struct {
	auditRepo repository.APIAuditLogRepository
	queue     chan *model.APIAuditLog
	wg        sync.WaitGroup
}

func NewAuditProcessor(auditRepo repository.APIAuditLogRepository, queueSize int) *AuditProcessor {
	if queueSize <= 0 {
		queueSize = defaultAuditQueueSize
	}

	processor := &AuditProcessor{
		auditRepo: auditRepo,
		queue:     make(chan *model.APIAuditLog, queueSize),
	}

	processor.wg.Add(1)
	go processor.run()

	return processor
}

func newAuditResponseWriter(writer gin.ResponseWriter, limit int) *auditResponseWriter {
	return &auditResponseWriter{
		ResponseWriter: writer,
		writeLimit:     limit,
	}
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	w.appendBody(string(data))
	return w.ResponseWriter.Write(data)
}

func (w *auditResponseWriter) WriteString(s string) (int, error) {
	w.appendBody(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *auditResponseWriter) appendBody(content string) {
	if w.writeLimit <= 0 || w.truncated {
		return
	}

	remaining := w.writeLimit - w.body.Len()
	if remaining <= 0 {
		w.truncated = true
		return
	}

	if len(content) > remaining {
		w.body.WriteString(content[:remaining])
		w.truncated = true
		return
	}

	w.body.WriteString(content)
}

func (w *auditResponseWriter) Body() string {
	if !w.truncated {
		return w.body.String()
	}
	return w.body.String() + "...(truncated)"
}

// Middleware 统一记录 API 请求审计日志
func (p *AuditProcessor) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		startAt := time.Now()
		requestJSONPayload, hasJSONRequest := captureRequestPayload(c)
		writer := newAuditResponseWriter(c.Writer, maxAuditPayloadSize)
		c.Writer = writer

		c.Next()

		latencyMS := time.Since(startAt).Milliseconds()
		httpStatus := c.Writer.Status()
		responseBody := writer.Body()
		requestPath := buildAuditPath(c.Request)
		userID := resolveAuditUserID(c.Request, responseBody)

		requestText := buildRequestJSONText(requestJSONPayload, hasJSONRequest)
		auditErr := buildAuditError(c, responseBody)
		auditLog := &model.APIAuditLog{
			UserID:     userID,
			HTTPMethod: c.Request.Method,
			Path:       truncateTextToSize(requestPath, 1024),
			HTTPStatus: httpStatus,
			LatencyMS:  latencyMS,
			ClientIP:   truncateTextToSize(c.ClientIP(), 64),
			UserAgent:  truncateTextToSize(c.Request.UserAgent(), 255),
			Request:    requestText,
			Response:   responseBody,
			Error:      auditErr,
		}

		p.enqueue(auditLog)
	}
}

func (p *AuditProcessor) Shutdown(ctx context.Context) error {
	close(p.queue)

	done := make(chan struct{})
	go func() {
		defer close(done)
		p.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (p *AuditProcessor) enqueue(auditLog *model.APIAuditLog) {
	select {
	case p.queue <- auditLog:
	default:
		logger.Warn("audit queue full, drop api audit log", logger.String("path", auditLog.Path))
	}
}

func (p *AuditProcessor) run() {
	defer p.wg.Done()

	for auditLog := range p.queue {
		writeCtx, cancel := context.WithTimeout(context.Background(), auditWriteTimeout)
		err := p.auditRepo.Create(writeCtx, auditLog)
		cancel()

		if err != nil {
			logger.Error("create api audit log failed", logger.Err(err), logger.String("path", auditLog.Path))
		}
	}
}

func captureRequestPayload(c *gin.Context) (any, bool) {
	contentType := c.ContentType()

	if strings.HasPrefix(contentType, "application/json") {
		return captureJSONBody(c.Request)
	}

	return nil, false
}

func captureJSONBody(r *http.Request) (any, bool) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(strings.NewReader(""))
		return nil, false
	}

	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	if len(bodyBytes) == 0 {
		return nil, false
	}

	var parsed any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, false
	}

	return parsed, true
}

func buildRequestJSONText(requestJSONPayload any, hasJSONRequest bool) string {
	if !hasJSONRequest {
		return ""
	}

	payload, err := json.Marshal(requestJSONPayload)
	if err != nil {
		return ""
	}

	if len(payload) > maxAuditPayloadSize {
		return string(payload[:maxAuditPayloadSize]) + "...(truncated)"
	}

	return string(payload)
}

func buildAuditError(c *gin.Context, responseBody string) string {
	errorTexts := make([]string, 0, len(c.Errors)+1)
	for _, err := range c.Errors {
		if err == nil {
			continue
		}
		errorTexts = append(errorTexts, err.Error())
	}

	if responseError := extractErrorFromResponse(responseBody); responseError != "" {
		errorTexts = append(errorTexts, responseError)
	}

	return strings.Join(uniqueStrings(errorTexts), "; ")
}

func extractErrorFromResponse(responseBody string) string {
	if responseBody == "" {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(responseBody), &payload); err != nil {
		return ""
	}

	codeValue, ok := payload["code"]
	if !ok {
		return ""
	}

	code, ok := numberToInt(codeValue)
	if !ok || code == response.CodeSuccess {
		return ""
	}

	message, _ := payload["message"].(string)
	return message
}

func buildAuditPath(r *http.Request) string {
	path := r.URL.Path
	if r.URL.RawQuery == "" {
		return path
	}
	return path + "?" + r.URL.RawQuery
}

func resolveAuditUserID(r *http.Request, responseBody string) *int {
	if isLoginAPI(r) {
		if userID := extractLoginUserIDFromResponse(responseBody); userID != nil {
			return userID
		}
	}

	return extractUserIDFromHeader(r)
}

func isLoginAPI(r *http.Request) bool {
	return r.URL.Path == "/api/user/login"
}

func extractLoginUserIDFromResponse(responseBody string) *int {
	if responseBody == "" {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(responseBody), &payload); err != nil {
		return nil
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		return nil
	}

	rawUserID, exists := data["user_id"]
	if !exists {
		return nil
	}

	userID, ok := numberToInt(rawUserID)
	if !ok {
		return nil
	}

	return &userID
}

func extractUserIDFromHeader(r *http.Request) *int {
	for _, key := range []string{"X-User-ID", "X-User-Id", "x-user-id"} {
		value := strings.TrimSpace(r.Header.Get(key))
		if value == "" {
			continue
		}

		userID, err := strconv.Atoi(value)
		if err == nil {
			return &userID
		}
	}

	return nil
}

func truncateTextToSize(text string, limit int) string {
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit]
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func numberToInt(value any) (int, bool) {
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	case int:
		return typed, true
	default:
		return 0, false
	}
}
