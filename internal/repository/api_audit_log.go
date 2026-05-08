package repository

import (
	"context"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
)

// APIAuditLogRepository API 审计日志仓储
type APIAuditLogRepository interface {
	Create(ctx context.Context, auditLog *model.APIAuditLog) error
}

type apiAuditLogRepository struct {
	db *gorm.DB
}

// NewAPIAuditLogRepository 创建 API 审计日志仓储实例
func NewAPIAuditLogRepository(db *gorm.DB) APIAuditLogRepository {
	return &apiAuditLogRepository{db: db}
}

func (r *apiAuditLogRepository) Create(ctx context.Context, auditLog *model.APIAuditLog) error {
	return r.db.WithContext(ctx).Create(auditLog).Error
}
