package model

import "time"

// APIAuditLog API 审计日志
type APIAuditLog struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID  string    `gorm:"size:64;not null;index:idx_api_audit_log_request_id" json:"request_id"`
	UserID     *int      `gorm:"index:idx_api_audit_log_user_id" json:"user_id,omitempty"`
	HTTPMethod string    `gorm:"size:16;not null;index:idx_api_audit_log_http_method" json:"http_method"`
	Path       string    `gorm:"size:255;not null;index:idx_api_audit_log_path" json:"path"`
	HTTPStatus int       `gorm:"not null;index:idx_api_audit_log_http_status" json:"http_status"`
	BizCode    *int      `gorm:"index:idx_api_audit_log_biz_code" json:"biz_code,omitempty"`
	Success    bool      `gorm:"not null;index:idx_api_audit_log_success" json:"success"`
	LatencyMS  int64     `gorm:"not null;index:idx_api_audit_log_latency_ms" json:"latency_ms"`
	ClientIP   string    `gorm:"size:64;index:idx_api_audit_log_client_ip" json:"client_ip,omitempty"`
	UserAgent  string    `gorm:"size:255" json:"user_agent,omitempty"`
	Request    string    `gorm:"type:longtext" json:"request"`
	Response   string    `gorm:"type:longtext" json:"response"`
	Error      string    `gorm:"type:text" json:"error"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime;index:idx_api_audit_log_created_at" json:"created_at"`
}

func (APIAuditLog) TableName() string {
	return "api_audit_log"
}
