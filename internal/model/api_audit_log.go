package model

import "time"

// APIAuditLog API 审计日志
type APIAuditLog struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     *int      `json:"user_id,omitempty"`
	HTTPMethod string    `gorm:"size:16;not null" json:"http_method"`
	Path       string    `gorm:"size:1024;not null" json:"path"`
	HTTPStatus int       `gorm:"not null" json:"http_status"`
	LatencyMS  int64     `gorm:"not null" json:"latency_ms"`
	ClientIP   string    `gorm:"size:64" json:"client_ip,omitempty"`
	UserAgent  string    `gorm:"size:255" json:"user_agent,omitempty"`
	Request    string    `gorm:"type:longtext" json:"request"`
	Response   string    `gorm:"type:longtext" json:"response"`
	Error      string    `gorm:"type:text" json:"error"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
}

func (APIAuditLog) TableName() string {
	return "api_audit_log"
}
