package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/rainlf/mango-crew/pkg/response"
)

const maxAuditPayloadSize = 64 * 1024

type auditResponseWriter struct {
	gin.ResponseWriter
	body       strings.Builder
	truncated  bool
	writeLimit int
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

// Audit 统一记录 API 请求审计日志
func Audit(auditRepo repository.APIAuditLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		startAt := time.Now()
		requestID := ensureRequestID(c)
		requestJSONPayload, hasJSONRequest, requestUserID := captureRequestPayload(c)
		writer := newAuditResponseWriter(c.Writer, maxAuditPayloadSize)
		c.Writer = writer

		c.Next()

		latencyMS := time.Since(startAt).Milliseconds()
		httpStatus := c.Writer.Status()
		responseBody := writer.Body()
		userID := requestUserID
		if userID == nil {
			userID = extractUserIDFromJSONText(responseBody)
		}

		bizCode := extractBizCodeFromResponse(responseBody)
		requestText := buildRequestJSONText(requestJSONPayload, hasJSONRequest)
		auditErr := buildAuditError(c, responseBody)
		success := buildSuccess(httpStatus, bizCode, auditErr)
		auditLog := &model.APIAuditLog{
			RequestID:  requestID,
			UserID:     userID,
			HTTPMethod: c.Request.Method,
			Path:       c.Request.URL.Path,
			HTTPStatus: httpStatus,
			BizCode:    bizCode,
			Success:    success,
			LatencyMS:  latencyMS,
			ClientIP:   truncateTextToSize(c.ClientIP(), 64),
			UserAgent:  truncateTextToSize(c.Request.UserAgent(), 255),
			Request:    requestText,
			Response:   responseBody,
			Error:      auditErr,
		}

		if err := auditRepo.Create(c.Request.Context(), auditLog); err != nil {
			logger.Error("create api audit log failed", logger.Err(err), logger.String("path", auditLog.Path))
		}
	}
}

func captureRequestPayload(c *gin.Context) (any, bool, *int) {
	contentType := c.ContentType()

	if strings.HasPrefix(contentType, "application/json") {
		return captureJSONBody(c.Request)
	}

	return nil, false, extractUserIDFromRequest(c.Request)
}

func captureJSONBody(r *http.Request) (any, bool, *int) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(strings.NewReader(""))
		return nil, false, nil
	}

	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	if len(bodyBytes) == 0 {
		return nil, false, nil
	}

	var parsed any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, false, extractUserIDFromRequest(r)
	}

	return parsed, true, extractUserIDFromJSONObject(parsed)
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

func buildSuccess(httpStatus int, bizCode *int, auditErr string) bool {
	if httpStatus >= http.StatusBadRequest {
		return false
	}
	if bizCode != nil && *bizCode != response.CodeSuccess {
		return false
	}
	return auditErr == ""
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

func extractBizCodeFromResponse(responseBody string) *int {
	if responseBody == "" {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(responseBody), &payload); err != nil {
		return nil
	}

	codeValue, ok := payload["code"]
	if !ok {
		return nil
	}

	code, ok := numberToInt(codeValue)
	if !ok {
		return nil
	}

	return &code
}

func ensureRequestID(c *gin.Context) string {
	requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	if requestID == "" {
		requestID = strings.TrimSpace(c.GetHeader("X-Req-Id"))
	}
	if requestID == "" {
		requestID = generateRequestID()
	}

	c.Writer.Header().Set("X-Request-ID", requestID)
	return requestID
}

func generateRequestID() string {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + hex.EncodeToString(randomBytes)
}

func extractUserIDFromRequest(r *http.Request) *int {
	if userID := extractUserIDFromValues(r.URL.Query()); userID != nil {
		return userID
	}

	if err := r.ParseForm(); err == nil {
		if userID := extractUserIDFromValues(r.PostForm); userID != nil {
			return userID
		}
	}

	return nil
}

func extractUserIDFromJSONText(text string) *int {
	var payload any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil
	}

	return extractUserIDFromJSONObject(payload)
}

func extractUserIDFromJSONObject(payload any) *int {
	switch value := payload.(type) {
	case map[string]any:
		if userID := extractUserIDFromMap(value); userID != nil {
			return userID
		}
		for _, nested := range value {
			if userID := extractUserIDFromJSONObject(nested); userID != nil {
				return userID
			}
		}
	case []any:
		for _, nested := range value {
			if userID := extractUserIDFromJSONObject(nested); userID != nil {
				return userID
			}
		}
	}

	return nil
}

func extractUserIDFromMap(values map[string]any) *int {
	keys := []string{
		"user_id",
		"userId",
		"recorderId",
		"recorder_id",
		"created_by",
		"createdBy",
	}

	for _, key := range keys {
		rawValue, exists := values[key]
		if !exists {
			continue
		}
		if userID, ok := valueToInt(rawValue); ok {
			return &userID
		}
	}

	return nil
}

func extractUserIDFromValues(values map[string][]string) *int {
	for _, key := range []string{"user_id", "userId", "recorderId", "recorder_id"} {
		candidates, ok := values[key]
		if !ok || len(candidates) == 0 {
			continue
		}

		value := strings.TrimSpace(candidates[0])
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

func valueToInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
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
