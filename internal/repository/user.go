package repository

import (
	"context"
	"sort"
	"strings"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id int) (*model.User, error)
	FindByIDIn(ctx context.Context, ids []int) ([]*model.User, error)
	FindByOpenID(ctx context.Context, openID string) (*model.User, error)
	FindAll(ctx context.Context) ([]*model.User, error)
	ApplyStatsDeltas(ctx context.Context, deltas map[int]model.UserStatsDelta) error
	RefreshStatsByUserIDs(ctx context.Context, userIDs []int) error
}

// userRepository 用户数据访问实现
type userRepository struct {
	db *gorm.DB
}

type userStatsValues struct {
	TotalPoints int
	TotalGames  int
	WinCount    int
	WinRate     float64
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByIDIn(ctx context.Context, ids []int) ([]*model.User, error) {
	if len(ids) == 0 {
		return []*model.User{}, nil
	}

	var users []*model.User
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&users).Error
	return users, err
}

func (r *userRepository) FindByOpenID(ctx context.Context, openID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("open_id = ?", openID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *userRepository) ApplyStatsDeltas(ctx context.Context, deltas map[int]model.UserStatsDelta) error {
	if len(deltas) == 0 {
		return nil
	}

	uniqueIDs := make([]int, 0, len(deltas))
	for userID := range deltas {
		uniqueIDs = append(uniqueIDs, userID)
	}
	sort.Ints(uniqueIDs)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var users []*model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", uniqueIDs).
			Find(&users).Error; err != nil {
			return err
		}
		if len(users) != len(uniqueIDs) {
			return gorm.ErrRecordNotFound
		}

		updates := make(map[int]userStatsValues, len(users))
		for _, user := range users {
			delta := deltas[user.ID]
			totalPoints := user.TotalPoints + delta.PointsDelta
			totalGames := user.TotalGames + delta.GamesDelta
			winCount := user.WinCount + delta.WinsDelta

			if totalGames < 0 || winCount < 0 {
				return gorm.ErrRecordNotFound
			}

			winRate := 0.0
			if totalGames > 0 {
				winRate = float64(winCount) / float64(totalGames)
			}

			updates[user.ID] = userStatsValues{
				TotalPoints: totalPoints,
				TotalGames:  totalGames,
				WinCount:    winCount,
				WinRate:     winRate,
			}
		}

		if err := r.bulkUpdateUserStats(tx, updates); err != nil {
			return err
		}
		return nil
	})
}

func (r *userRepository) RefreshStatsByUserIDs(ctx context.Context, userIDs []int) error {
	uniqueIDs := uniqueInts(userIDs)
	if len(uniqueIDs) == 0 {
		return nil
	}

	type userStatsRow struct {
		UserID      int
		TotalPoints int
		TotalGames  int
		WinCount    int
	}

	statsRows := make([]userStatsRow, 0, len(uniqueIDs))
	err := r.db.WithContext(ctx).
		Table("user AS u").
		Select(`
			u.id AS user_id,
			COALESCE(SUM(CASE WHEN g.id IS NOT NULL THEN gr.final_points ELSE 0 END), 0) AS total_points,
			COUNT(DISTINCT CASE WHEN g.id IS NOT NULL AND gr.role NOT IN ? THEN g.id END) AS total_games,
			COALESCE(SUM(CASE WHEN g.id IS NOT NULL AND gr.role = ? THEN 1 ELSE 0 END), 0) AS win_count
		`, []model.PlayerRole{model.RoleRecorder, model.RoleSquatRedeem}, model.RoleWinner).
		Joins("LEFT JOIN game_record AS gr ON gr.user_id = u.id").
		Joins("LEFT JOIN game AS g ON g.id = gr.game_id AND g.status = ?", model.GameStatusSettled).
		Where("u.id IN ?", uniqueIDs).
		Group("u.id").
		Scan(&statsRows).Error
	if err != nil {
		return err
	}

	statsByUserID := make(map[int]userStatsRow, len(statsRows))
	for _, row := range statsRows {
		if row.TotalGames < 0 {
			row.TotalGames = 0
		}
		if row.WinCount < 0 {
			row.WinCount = 0
		}
		statsByUserID[row.UserID] = row
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := make(map[int]userStatsValues, len(uniqueIDs))
		for _, userID := range uniqueIDs {
			row, ok := statsByUserID[userID]
			if !ok {
				row = userStatsRow{}
			}

			winRate := 0.0
			if row.TotalGames > 0 {
				winRate = float64(row.WinCount) / float64(row.TotalGames)
			}

			updates[userID] = userStatsValues{
				TotalPoints: row.TotalPoints,
				TotalGames:  row.TotalGames,
				WinCount:    row.WinCount,
				WinRate:     winRate,
			}
		}

		if err := r.bulkUpdateUserStats(tx, updates); err != nil {
			return err
		}
		return nil
	})
}

func (r *userRepository) bulkUpdateUserStats(tx *gorm.DB, updates map[int]userStatsValues) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]int, 0, len(updates))
	for userID := range updates {
		ids = append(ids, userID)
	}
	sort.Ints(ids)

	args := make([]any, 0, len(ids)*12)
	var sql strings.Builder
	sql.WriteString("UPDATE user SET ")
	appendCaseClause(&sql, &args, "total_points", ids, updates, func(v userStatsValues) any { return v.TotalPoints })
	sql.WriteString(", ")
	appendCaseClause(&sql, &args, "total_games", ids, updates, func(v userStatsValues) any { return v.TotalGames })
	sql.WriteString(", ")
	appendCaseClause(&sql, &args, "win_count", ids, updates, func(v userStatsValues) any { return v.WinCount })
	sql.WriteString(", ")
	appendCaseClause(&sql, &args, "win_rate", ids, updates, func(v userStatsValues) any { return v.WinRate })
	sql.WriteString(" WHERE id IN (")
	for idx, userID := range ids {
		if idx > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString("?")
		args = append(args, userID)
	}
	sql.WriteString(")")

	return tx.Exec(sql.String(), args...).Error
}

func appendCaseClause(builder *strings.Builder, args *[]any, column string, ids []int, updates map[int]userStatsValues, valueFn func(userStatsValues) any) {
	builder.WriteString(column)
	builder.WriteString(" = CASE id")
	for _, userID := range ids {
		builder.WriteString(" WHEN ? THEN ?")
		*args = append(*args, userID, valueFn(updates[userID]))
	}
	builder.WriteString(" ELSE ")
	builder.WriteString(column)
	builder.WriteString(" END")
}

func uniqueInts(ids []int) []int {
	if len(ids) == 0 {
		return []int{}
	}

	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
