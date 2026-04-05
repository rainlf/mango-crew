package repository

import (
	"context"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
)

// GameRepository 游戏数据访问接口
type GameRepository interface {
	CreateGame(ctx context.Context, game *model.MaJiangGame) error
	UpdateGame(ctx context.Context, game *model.MaJiangGame) error
	FindGameByID(ctx context.Context, id int) (*model.MaJiangGame, error)
	FindLastGames(ctx context.Context, limit, offset int) ([]*model.MaJiangGame, error)
	FindGamesByIDs(ctx context.Context, ids []int) ([]*model.MaJiangGame, error)
	FindLastGamesByUser(ctx context.Context, userID, limit int) ([]*model.MaJiangGame, error)

	CreateGameItems(ctx context.Context, items []*model.MaJiangGameItem) error
	FindGameItemsByGameID(ctx context.Context, gameID int) ([]*model.MaJiangGameItem, error)
	FindGameItemsByGameIDAndType(ctx context.Context, gameID int, userType model.MaJiangUserType) ([]*model.MaJiangGameItem, error)
	FindLastGameIDsByUserID(ctx context.Context, userID int, userTypes []model.MaJiangUserType, limit, offset int) ([]int, error)
}

// gameRepository 游戏数据访问实现
type gameRepository struct {
	db *gorm.DB
}

// NewGameRepository 创建游戏仓库实例
func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) CreateGame(ctx context.Context, game *model.MaJiangGame) error {
	return r.db.WithContext(ctx).Create(game).Error
}

func (r *gameRepository) UpdateGame(ctx context.Context, game *model.MaJiangGame) error {
	return r.db.WithContext(ctx).Save(game).Error
}

func (r *gameRepository) FindGameByID(ctx context.Context, id int) (*model.MaJiangGame, error) {
	var game model.MaJiangGame
	err := r.db.WithContext(ctx).First(&game, id).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) FindLastGames(ctx context.Context, limit, offset int) ([]*model.MaJiangGame, error) {
	var games []*model.MaJiangGame
	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("created_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) FindGamesByIDs(ctx context.Context, ids []int) ([]*model.MaJiangGame, error) {
	var games []*model.MaJiangGame
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Where("is_deleted = ?", false).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) FindLastGamesByUser(ctx context.Context, userID, limit int) ([]*model.MaJiangGame, error) {
	var games []*model.MaJiangGame
	err := r.db.WithContext(ctx).
		Joins("JOIN mgtt_majiang_game_item ON mgtt_majiang_game_item.game_id = mgtt_majiang_game.id").
		Where("mgtt_majiang_game_item.user_id = ?", userID).
		Where("mgtt_majiang_game.is_deleted = ?", false).
		Group("mgtt_majiang_game.id").
		Order("mgtt_majiang_game.created_time DESC").
		Limit(limit).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) CreateGameItems(ctx context.Context, items []*model.MaJiangGameItem) error {
	return r.db.WithContext(ctx).Create(items).Error
}

func (r *gameRepository) FindGameItemsByGameID(ctx context.Context, gameID int) ([]*model.MaJiangGameItem, error) {
	var items []*model.MaJiangGameItem
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Find(&items).Error
	return items, err
}

func (r *gameRepository) FindGameItemsByGameIDAndType(ctx context.Context, gameID int, userType model.MaJiangUserType) ([]*model.MaJiangGameItem, error) {
	var items []*model.MaJiangGameItem
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Where("type = ?", userType).
		Find(&items).Error
	return items, err
}

func (r *gameRepository) FindLastGameIDsByUserID(ctx context.Context, userID int, userTypes []model.MaJiangUserType, limit, offset int) ([]int, error) {
	var ids []int
	err := r.db.WithContext(ctx).
		Model(&model.MaJiangGameItem{}).
		Select("DISTINCT game_id").
		Where("user_id = ?", userID).
		Where("type IN ?", userTypes).
		Order("created_time DESC").
		Limit(limit).
		Offset(offset).
		Pluck("game_id", &ids).Error
	return ids, err
}
