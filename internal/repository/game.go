package repository

import (
	"context"
	"time"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
)

// GameRepository 游戏数据访问接口
type GameRepository interface {
	Create(ctx context.Context, game *model.Game) error
	Update(ctx context.Context, game *model.Game) error
	FindByID(ctx context.Context, id int) (*model.Game, error)
	FindRecentGames(ctx context.Context, limit, offset int) ([]*model.Game, error)
	FindGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.Game, error)
	SettleGame(ctx context.Context, id int) error
	CancelGame(ctx context.Context, id int) error

	// 玩家相关
	CreatePlayers(ctx context.Context, players []*model.GamePlayer) error
	FindPlayersByGameID(ctx context.Context, gameID int) ([]*model.GamePlayer, error)
	FindPlayersByUserID(ctx context.Context, userID int, limit int) ([]*model.GamePlayer, error)

	// 番型相关
	CreateWinTypes(ctx context.Context, winTypes []*model.GamePlayerWinType) error
	FindWinTypesByPlayerID(ctx context.Context, playerID int) ([]*model.GamePlayerWinType, error)

	// 统计
	CountPlayerGames(ctx context.Context, userID int) (int64, error)
	CountPlayerWins(ctx context.Context, userID int) (int64, error)
	SumPlayerPoints(ctx context.Context, userID int) (int, error)
}

// gameRepository 游戏数据访问实现
type gameRepository struct {
	db *gorm.DB
}

// NewGameRepository 创建游戏仓库实例
func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(ctx context.Context, game *model.Game) error {
	return r.db.WithContext(ctx).Create(game).Error
}

func (r *gameRepository) Update(ctx context.Context, game *model.Game) error {
	return r.db.WithContext(ctx).Save(game).Error
}

func (r *gameRepository) FindByID(ctx context.Context, id int) (*model.Game, error) {
	var game model.Game
	err := r.db.WithContext(ctx).First(&game, id).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) FindRecentGames(ctx context.Context, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	err := r.db.WithContext(ctx).
		Where("status = ?", model.GameStatusSettled).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) FindGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	err := r.db.WithContext(ctx).
		Model(&model.Game{}).
		Distinct("game.*").
		Joins("JOIN game_player ON game_player.game_id = game.id").
		Where("game_player.user_id = ?", userID).
		Where("game.status = ?", model.GameStatusSettled).
		Order("game.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) SettleGame(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Game{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"status":     model.GameStatusSettled,
				"settled_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&model.GamePlayer{}).
			Where("game_id = ?", id).
			Update("is_settled", true).Error
	})
}

func (r *gameRepository) CancelGame(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&model.Game{}).
		Where("id = ?", id).
		Update("status", model.GameStatusCanceled).Error
}

// 玩家相关

func (r *gameRepository) CreatePlayers(ctx context.Context, players []*model.GamePlayer) error {
	return r.db.WithContext(ctx).Create(players).Error
}

func (r *gameRepository) FindPlayersByGameID(ctx context.Context, gameID int) ([]*model.GamePlayer, error) {
	var players []*model.GamePlayer
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Find(&players).Error
	return players, err
}

func (r *gameRepository) FindPlayersByUserID(ctx context.Context, userID int, limit int) ([]*model.GamePlayer, error) {
	var players []*model.GamePlayer
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&players).Error
	return players, err
}

// 番型相关

func (r *gameRepository) CreateWinTypes(ctx context.Context, winTypes []*model.GamePlayerWinType) error {
	return r.db.WithContext(ctx).Create(winTypes).Error
}

func (r *gameRepository) FindWinTypesByPlayerID(ctx context.Context, playerID int) ([]*model.GamePlayerWinType, error) {
	var winTypes []*model.GamePlayerWinType
	err := r.db.WithContext(ctx).
		Where("game_player_id = ?", playerID).
		Find(&winTypes).Error
	return winTypes, err
}

// 统计

func (r *gameRepository) CountPlayerGames(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GamePlayer{}).
		Joins("JOIN game ON game.id = game_player.game_id").
		Where("game_player.user_id = ?", userID).
		Where("game.status = ?", model.GameStatusSettled).
		Distinct("game_player.game_id").
		Count(&count).Error
	return count, err
}

func (r *gameRepository) CountPlayerWins(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GamePlayer{}).
		Joins("JOIN game ON game.id = game_player.game_id").
		Where("game_player.user_id = ?", userID).
		Where("game_player.role = ?", model.RoleWinner).
		Where("game.status = ?", model.GameStatusSettled).
		Count(&count).Error
	return count, err
}

func (r *gameRepository) SumPlayerPoints(ctx context.Context, userID int) (int, error) {
	var result struct {
		Total int
	}
	err := r.db.WithContext(ctx).
		Model(&model.GamePlayer{}).
		Joins("JOIN game ON game.id = game_player.game_id").
		Select("COALESCE(SUM(final_points), 0) as total").
		Where("game_player.user_id = ?", userID).
		Where("game.status = ?", model.GameStatusSettled).
		Scan(&result).Error
	return result.Total, err
}
