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
	FindBySessionID(ctx context.Context, sessionID int, limit, offset int) ([]*model.Game, error)
	FindRecentGames(ctx context.Context, limit, offset int) ([]*model.Game, error)
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

func (r *gameRepository) FindBySessionID(ctx context.Context, sessionID int, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Where("status != ?", model.GameStatusCanceled).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) FindRecentGames(ctx context.Context, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	err := r.db.WithContext(ctx).
		Where("status != ?", model.GameStatusCanceled).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) SettleGame(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Game{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.GameStatusSettled,
			"settled_at": now,
		}).Error
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
	err := r.db.WithContext(ctx).Model(&model.GamePlayer{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *gameRepository) CountPlayerWins(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.GamePlayer{}).
		Where("user_id = ?", userID).
		Where("role = ?", model.RoleWinner).
		Count(&count).Error
	return count, err
}

func (r *gameRepository) SumPlayerPoints(ctx context.Context, userID int) (int, error) {
	var result struct {
		Total int
	}
	err := r.db.WithContext(ctx).Model(&model.GamePlayer{}).
		Select("COALESCE(SUM(final_points), 0) as total").
		Where("user_id = ?", userID).
		Scan(&result).Error
	return result.Total, err
}
