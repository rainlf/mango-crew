package repository

import (
	"context"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
)

// CurrentPlayerRepository 当前牌桌玩家数据访问接口
type CurrentPlayerRepository interface {
	ReplacePlayers(ctx context.Context, userIDs []int) error
	FindPlayers(ctx context.Context) ([]*model.SessionPlayer, error)
}

// currentPlayerRepository 当前牌桌玩家数据访问实现
type currentPlayerRepository struct {
	db *gorm.DB
}

// NewCurrentPlayerRepository 创建当前牌桌玩家仓库实例
func NewCurrentPlayerRepository(db *gorm.DB) CurrentPlayerRepository {
	return &currentPlayerRepository{db: db}
}

func (r *currentPlayerRepository) ReplacePlayers(ctx context.Context, userIDs []int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.SessionPlayer{}).Error; err != nil {
			return err
		}

		if len(userIDs) == 0 {
			return nil
		}

		players := make([]*model.SessionPlayer, 0, len(userIDs))
		for idx, userID := range userIDs {
			players = append(players, &model.SessionPlayer{
				UserID: userID,
				Seat:   idx + 1,
			})
		}
		return tx.Create(players).Error
	})
}

func (r *currentPlayerRepository) FindPlayers(ctx context.Context) ([]*model.SessionPlayer, error) {
	var players []*model.SessionPlayer
	err := r.db.WithContext(ctx).
		Order("seat ASC").
		Find(&players).Error
	return players, err
}
