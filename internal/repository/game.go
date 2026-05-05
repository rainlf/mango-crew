package repository

import (
	"context"
	"errors"
	"time"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GameRepository 游戏数据访问接口
type GameRepository interface {
	Create(ctx context.Context, game *model.Game) error
	Update(ctx context.Context, game *model.Game) error
	FindByID(ctx context.Context, id int) (*model.Game, error)
	FindRecentGames(ctx context.Context, limit, offset int) ([]*model.Game, error)
	FindGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.Game, error)
	FindRecentFitnessGames(ctx context.Context, limit, offset int) ([]*model.Game, error)
	FindFitnessGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.Game, error)
	ListFitnessStatsByUser(ctx context.Context) (map[int]model.UserFitnessStats, error)
	CancelGame(ctx context.Context, id int) error
	GetRecorderPrizePool(ctx context.Context) (int, error)
	AdjustRecorderPrizePool(ctx context.Context, delta int) error
	ListRecorderPrizePoolContributors(ctx context.Context) (map[int]int, error)
	ListRecorderPrizePoolJackpotRecords(ctx context.Context) ([]*model.GameRecord, error)

	// 对局记录相关
	CreateRecords(ctx context.Context, records []*model.GameRecord) error
	FindRecordsByGameID(ctx context.Context, gameID int) ([]*model.GameRecord, error)
	FindRecordsByGameIDs(ctx context.Context, gameIDs []int) (map[int][]*model.GameRecord, error)
	FindRecordsByUserID(ctx context.Context, userID int, limit int) ([]*model.GameRecord, error)
	FindRecentWinningRecordsByUserIDs(ctx context.Context, userIDs []int, limitPerUser int) (map[int][]*model.GameRecord, error)

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
		Where("type <> ?", model.SquatRedeem).
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
		Where("game_record.role IN ?", []model.PlayerRole{model.RoleWinner, model.RoleLoser}).
		Where("game_record.final_points <> 0").
		Where("game.status = ?", model.GameStatusSettled).
		Where("game.type <> ?", model.SquatRedeem).
		Order("game.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) FindRecentFitnessGames(ctx context.Context, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	err := r.db.WithContext(ctx).
		Where("status = ?", model.GameStatusSettled).
		Where("type = ?", model.SquatRedeem).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) FindFitnessGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	err := r.db.WithContext(ctx).
		Model(&model.Game{}).
		Distinct("game.*").
		Joins("JOIN game_record ON game_record.game_id = game.id").
		Where("game_record.user_id = ?", userID).
		Where("game_record.role = ?", model.RoleSquatRedeem).
		Where("game.type = ?", model.SquatRedeem).
		Where("game.status = ?", model.GameStatusSettled).
		Order("game.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) ListFitnessStatsByUser(ctx context.Context) (map[int]model.UserFitnessStats, error) {
	type fitnessStatsRow struct {
		UserID      int `gorm:"column:user_id"`
		TotalPoints int `gorm:"column:total_points"`
		TotalGames  int `gorm:"column:total_games"`
	}

	var rows []fitnessStatsRow
	err := r.db.WithContext(ctx).
		Table("game_record").
		Select(`
			game_record.user_id AS user_id,
			COALESCE(SUM(game_record.base_points), 0) AS total_points,
			COUNT(*) AS total_games
		`).
		Joins("JOIN game ON game.id = game_record.game_id").
		Where("game_record.role = ?", model.RoleSquatRedeem).
		Where("game.type = ?", model.SquatRedeem).
		Where("game.status = ?", model.GameStatusSettled).
		Group("game_record.user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	statsByUserID := make(map[int]model.UserFitnessStats, len(rows))
	for _, row := range rows {
		statsByUserID[row.UserID] = model.UserFitnessStats{
			UserID:      row.UserID,
			TotalPoints: row.TotalPoints,
			TotalGames:  row.TotalGames,
		}
	}
	return statsByUserID, nil
}

func (r *gameRepository) CancelGame(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&model.Game{}).
		Where("id = ?", id).
		Update("status", model.GameStatusCanceled).Error
}

func (r *gameRepository) GetRecorderPrizePool(ctx context.Context) (int, error) {
	pool, err := r.getOrCreateRecorderPrizePool(ctx)
	if err != nil {
		return 0, err
	}
	return pool.Balance, nil
}

func (r *gameRepository) AdjustRecorderPrizePool(ctx context.Context, delta int) error {
	if delta == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pool, err := r.getOrCreateRecorderPrizePoolTx(ctx, tx)
		if err != nil {
			return err
		}

		nextBalance := pool.Balance + delta
		if nextBalance < 0 {
			return errors.New("奖池余额不能为负数")
		}

		return tx.Model(&model.PrizePool{}).
			Where("id = ?", pool.ID).
			Update("balance", nextBalance).Error
	})
}

func (r *gameRepository) ListRecorderPrizePoolContributors(ctx context.Context) (map[int]int, error) {
	type contributorRow struct {
		UserID            int `gorm:"column:user_id"`
		ContributedPoints int `gorm:"column:contributed_points"`
	}

	var rows []contributorRow
	err := r.db.WithContext(ctx).
		Table("game_record").
		Select("game_record.user_id, COALESCE(SUM(game_record.base_points), 0) AS contributed_points").
		Joins("JOIN game ON game.id = game_record.game_id").
		Where("game_record.role = ?", model.RoleRecorder).
		Where("game.status = ?", model.GameStatusSettled).
		Group("game_record.user_id").
		Having("COALESCE(SUM(game_record.base_points), 0) > 0").
		Order("contributed_points DESC, game_record.user_id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	contributors := make(map[int]int, len(rows))
	for _, row := range rows {
		contributors[row.UserID] = row.ContributedPoints
	}
	return contributors, nil
}

func (r *gameRepository) ListRecorderPrizePoolJackpotRecords(ctx context.Context) ([]*model.GameRecord, error) {
	var records []*model.GameRecord
	err := r.db.WithContext(ctx).
		Model(&model.GameRecord{}).
		Joins("JOIN game ON game.id = game_record.game_id").
		Where("game_record.role = ?", model.RoleRecorder).
		Where("game_record.final_points > 0").
		Where("game.status = ?", model.GameStatusSettled).
		Order("game_record.created_at DESC, game_record.id DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *gameRepository) getOrCreateRecorderPrizePool(ctx context.Context) (*model.PrizePool, error) {
	var pool *model.PrizePool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		pool, err = r.getOrCreateRecorderPrizePoolTx(ctx, tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (r *gameRepository) getOrCreateRecorderPrizePoolTx(ctx context.Context, tx *gorm.DB) (*model.PrizePool, error) {
	var pool model.PrizePool
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("pool_type = ?", model.PrizePoolTypeRecorder).
		First(&pool).Error
	if err == nil {
		return &pool, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	pool = model.PrizePool{
		PoolType: model.PrizePoolTypeRecorder,
		Balance:  0,
	}
	if err := tx.WithContext(ctx).Create(&pool).Error; err != nil {
		return nil, err
	}
	return &pool, nil
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
		if loadErr := record.LoadWinTypesFromRaw(); loadErr != nil {
			return nil, loadErr
		}
	}
	return records, nil
}

func (r *gameRepository) FindRecordsByGameIDs(ctx context.Context, gameIDs []int) (map[int][]*model.GameRecord, error) {
	if len(gameIDs) == 0 {
		return map[int][]*model.GameRecord{}, nil
	}

	var records []*model.GameRecord
	err := r.db.WithContext(ctx).
		Where("game_id IN ?", uniqueInts(gameIDs)).
		Order("game_id ASC, seat ASC, id ASC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	recordsByGameID := make(map[int][]*model.GameRecord, len(gameIDs))
	for _, record := range records {
		if err := record.LoadWinTypesFromRaw(); err != nil {
			return nil, err
		}
		recordsByGameID[record.GameID] = append(recordsByGameID[record.GameID], record)
	}
	return recordsByGameID, nil
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
		if loadErr := record.LoadWinTypesFromRaw(); loadErr != nil {
			return nil, loadErr
		}
	}
	return records, nil
}

func (r *gameRepository) FindRecentWinningRecordsByUserIDs(ctx context.Context, userIDs []int, limitPerUser int) (map[int][]*model.GameRecord, error) {
	if len(userIDs) == 0 || limitPerUser <= 0 {
		return map[int][]*model.GameRecord{}, nil
	}

	type recentWinningRecordRow struct {
		ID            int              `gorm:"column:id"`
		GameID        int              `gorm:"column:game_id"`
		GameType      int              `gorm:"column:game_type"`
		UserID        int              `gorm:"column:user_id"`
		Seat          int              `gorm:"column:seat"`
		Role          model.PlayerRole `gorm:"column:role"`
		BasePoints    int              `gorm:"column:base_points"`
		FinalPoints   int              `gorm:"column:final_points"`
		IsSettled     bool             `gorm:"column:is_settled"`
		CreatedAt     time.Time        `gorm:"column:created_at"`
		UpdatedAt     time.Time        `gorm:"column:updated_at"`
		WinTypesRaw   string           `gorm:"column:win_types"`
		GameCreatedAt time.Time        `gorm:"column:game_created_at"`
	}

	var rows []recentWinningRecordRow
	query := `
SELECT
	id,
	game_id,
		game_type,
	user_id,
	seat,
	role,
	base_points,
	final_points,
	is_settled,
	created_at,
	updated_at,
	win_types,
	game_created_at
FROM (
	SELECT
		gr.id,
		gr.game_id,
			g.type AS game_type,
		gr.user_id,
		gr.seat,
		gr.role,
		gr.base_points,
		gr.final_points,
		gr.is_settled,
		gr.created_at,
		gr.updated_at,
		gr.win_types,
		g.created_at AS game_created_at,
		ROW_NUMBER() OVER (PARTITION BY gr.user_id ORDER BY g.created_at DESC, gr.id DESC) AS rn
	FROM game_record AS gr
	JOIN game AS g ON g.id = gr.game_id
	WHERE gr.user_id IN ?
	  AND gr.role = ?
	  AND g.status = ?
) AS ranked
WHERE rn <= ?
ORDER BY user_id ASC, game_created_at DESC, id DESC
`
	err := r.db.WithContext(ctx).Raw(query, uniqueInts(userIDs), model.RoleWinner, model.GameStatusSettled, limitPerUser).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	recordsByUserID := make(map[int][]*model.GameRecord, len(userIDs))
	for _, row := range rows {
		record := &model.GameRecord{
			ID:          row.ID,
			GameID:      row.GameID,
			GameType:    model.GameTypeFromCode(row.GameType),
			UserID:      row.UserID,
			Seat:        row.Seat,
			Role:        row.Role,
			BasePoints:  row.BasePoints,
			FinalPoints: row.FinalPoints,
			IsSettled:   row.IsSettled,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			WinTypesRaw: row.WinTypesRaw,
		}
		if err := record.LoadWinTypesFromRaw(); err != nil {
			return nil, err
		}
		recordsByUserID[row.UserID] = append(recordsByUserID[row.UserID], record)
	}

	return recordsByUserID, nil
}

// 统计

func (r *gameRepository) CountPlayerGames(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GameRecord{}).
		Joins("JOIN game ON game.id = game_record.game_id").
		Where("game_record.user_id = ?", userID).
		Where("game_record.role NOT IN ?", []model.PlayerRole{model.RoleRecorder, model.RoleSquatRedeem}).
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
