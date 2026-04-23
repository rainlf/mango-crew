package repository

import (
	"context"

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
	CancelGame(ctx context.Context, id int) error

	// 对局记录相关
	CreateRecords(ctx context.Context, records []*model.GameRecord) error
	FindRecordsByGameID(ctx context.Context, gameID int) ([]*model.GameRecord, error)
	FindRecordsByUserID(ctx context.Context, userID int, limit int) ([]*model.GameRecord, error)

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
		Joins("JOIN game_record ON game_record.game_id = game.id").
		Where("game_record.user_id = ?", userID).
		Where("game.status = ?", model.GameStatusSettled).
		Order("game.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) CancelGame(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&model.Game{}).
		Where("id = ?", id).
		Update("status", model.GameStatusCanceled).Error
}

// 对局记录相关

func (r *gameRepository) CreateRecords(ctx context.Context, records []*model.GameRecord) error {
	for _, record := range records {
		if err := record.SyncWinTypesRaw(); err != nil {
			return err
		}
	}
	return r.db.WithContext(ctx).Create(records).Error
}

func (r *gameRepository) FindRecordsByGameID(ctx context.Context, gameID int) ([]*model.GameRecord, error) {
	var records []*model.GameRecord
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if err := record.LoadWinTypesFromRaw(); err != nil {
			return nil, err
		}
	}
	return records, err
}

func (r *gameRepository) FindRecordsByUserID(ctx context.Context, userID int, limit int) ([]*model.GameRecord, error) {
	var records []*model.GameRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if err := record.LoadWinTypesFromRaw(); err != nil {
			return nil, err
		}
	}
	return records, err
}

// 统计

func (r *gameRepository) CountPlayerGames(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GameRecord{}).
		Joins("JOIN game ON game.id = game_record.game_id").
		Where("game_record.user_id = ?", userID).
		Where("game_record.role <> ?", model.RoleRecorder).
		Where("game.status = ?", model.GameStatusSettled).
		Distinct("game_record.game_id").
		Count(&count).Error
	return count, err
}

func (r *gameRepository) CountPlayerWins(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GameRecord{}).
		Joins("JOIN game ON game.id = game_record.game_id").
		Where("game_record.user_id = ?", userID).
		Where("game_record.role = ?", model.RoleWinner).
		Where("game.status = ?", model.GameStatusSettled).
		Count(&count).Error
	return count, err
}

func (r *gameRepository) SumPlayerPoints(ctx context.Context, userID int) (int, error) {
	var result struct {
		Total int
	}
	err := r.db.WithContext(ctx).
		Model(&model.GameRecord{}).
		Joins("JOIN game ON game.id = game_record.game_id").
		Select("COALESCE(SUM(final_points), 0) as total").
		Where("game_record.user_id = ?", userID).
		Where("game.status = ?", model.GameStatusSettled).
		Scan(&result).Error
	return result.Total, err
}
