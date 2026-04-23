package repository

import (
	"context"

	"github.com/rainlf/mango-crew/internal/model"
	"gorm.io/gorm"
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

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for userID, delta := range deltas {
			result := tx.Model(&model.User{}).
				Where("id = ?", userID).
				Where("total_games + ? >= 0", delta.GamesDelta).
				Where("win_count + ? >= 0", delta.WinsDelta).
				Updates(map[string]any{
					"total_points": gorm.Expr("total_points + ?", delta.PointsDelta),
					"total_games":  gorm.Expr("total_games + ?", delta.GamesDelta),
					"win_count":    gorm.Expr("win_count + ?", delta.WinsDelta),
					"win_rate": gorm.Expr(`
						CASE
							WHEN total_games + ? <= 0 THEN 0
							ELSE CAST(win_count + ? AS DECIMAL(8,4)) / CAST(total_games + ? AS DECIMAL(8,4))
						END
					`, delta.GamesDelta, delta.WinsDelta, delta.GamesDelta),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
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
			COUNT(DISTINCT CASE WHEN g.id IS NOT NULL AND gr.role <> ? THEN g.id END) AS total_games,
			COALESCE(SUM(CASE WHEN g.id IS NOT NULL AND gr.role = ? THEN 1 ELSE 0 END), 0) AS win_count
		`, model.RoleRecorder, model.RoleWinner).
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
		for _, userID := range uniqueIDs {
			row, ok := statsByUserID[userID]
			if !ok {
				row = userStatsRow{}
			}

			winRate := 0.0
			if row.TotalGames > 0 {
				winRate = float64(row.WinCount) / float64(row.TotalGames)
			}

			if err := tx.Model(&model.User{}).
				Where("id = ?", userID).
				Updates(map[string]any{
					"total_points": row.TotalPoints,
					"total_games":  row.TotalGames,
					"win_count":    row.WinCount,
					"win_rate":     winRate,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
