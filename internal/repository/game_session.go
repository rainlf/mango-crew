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
	FindLatestActive(ctx context.Context) (*model.GameSession, error)
	FindActiveSessions(ctx context.Context) ([]*model.GameSession, error)
	FindAllSessions(ctx context.Context, limit, offset int) ([]*model.GameSession, error)
	EndSession(ctx context.Context, id int) error
	CountGames(ctx context.Context, sessionID int) (int64, error)
	ReplacePlayers(ctx context.Context, sessionID int, userIDs []int) error
	FindPlayers(ctx context.Context, sessionID int) ([]*model.SessionPlayer, error)
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

func (r *gameSessionRepository) FindLatestActive(ctx context.Context) (*model.GameSession, error) {
	var session model.GameSession
	err := r.db.WithContext(ctx).
		Where("status = ?", 0).
		Order("created_at DESC").
		First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
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
		Updates(map[string]any{
			"status":   1,
			"ended_at": now,
		}).Error
}

func (r *gameSessionRepository) CountGames(ctx context.Context, sessionID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Game{}).
		Where("session_id = ?", sessionID).
		Where("status = ?", model.GameStatusSettled).
		Count(&count).Error
	return count, err
}

func (r *gameSessionRepository) ReplacePlayers(ctx context.Context, sessionID int, userIDs []int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.SessionPlayer{}).Error; err != nil {
			return err
		}

		if len(userIDs) == 0 {
			return nil
		}

		players := make([]*model.SessionPlayer, 0, len(userIDs))
		for idx, userID := range userIDs {
			players = append(players, &model.SessionPlayer{
				SessionID: sessionID,
				UserID:    userID,
				Seat:      idx + 1,
			})
		}
		return tx.Create(players).Error
	})
}

func (r *gameSessionRepository) FindPlayers(ctx context.Context, sessionID int) ([]*model.SessionPlayer, error) {
	var players []*model.SessionPlayer
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("seat ASC").
		Find(&players).Error
	return players, err
}
