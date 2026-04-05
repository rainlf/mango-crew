package repository

import (
	"context"
	"time"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
)

// GameSessionRepository 场次数据访问接口
type GameSessionRepository interface {
	Create(ctx context.Context, session *model.GameSession) error
	Update(ctx context.Context, session *model.GameSession) error
	FindByID(ctx context.Context, id int) (*model.GameSession, error)
	FindActiveSessions(ctx context.Context) ([]*model.GameSession, error)
	FindAllSessions(ctx context.Context, limit, offset int) ([]*model.GameSession, error)
	EndSession(ctx context.Context, id int) error
	CountGames(ctx context.Context, sessionID int) (int64, error)
}

// gameSessionRepository 场次数据访问实现
type gameSessionRepository struct {
	db *gorm.DB
}

// NewGameSessionRepository 创建场次仓库实例
func NewGameSessionRepository(db *gorm.DB) GameSessionRepository {
	return &gameSessionRepository{db: db}
}

func (r *gameSessionRepository) Create(ctx context.Context, session *model.GameSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *gameSessionRepository) Update(ctx context.Context, session *model.GameSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *gameSessionRepository) FindByID(ctx context.Context, id int) (*model.GameSession, error) {
	var session model.GameSession
	err := r.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gameSessionRepository) FindActiveSessions(ctx context.Context) ([]*model.GameSession, error) {
	var sessions []*model.GameSession
	err := r.db.WithContext(ctx).
		Where("status = ?", 0).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *gameSessionRepository) FindAllSessions(ctx context.Context, limit, offset int) ([]*model.GameSession, error) {
	var sessions []*model.GameSession
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error
	return sessions, err
}

func (r *gameSessionRepository) EndSession(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.GameSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   1,
			"ended_at": now,
		}).Error
}

func (r *gameSessionRepository) CountGames(ctx context.Context, sessionID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Game{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return count, err
}
