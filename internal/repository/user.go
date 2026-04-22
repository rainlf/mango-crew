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
