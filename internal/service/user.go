package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rainlf/mgtt-go/internal/config"
	"github.com/rainlf/mgtt-go/internal/model"
	"github.com/rainlf/mgtt-go/internal/repository"
	"github.com/rainlf/mgtt-go/pkg/logger"
)

// UserService 用户服务接口
type UserService interface {
	Login(ctx context.Context, code string) (*model.User, error)
	GetUserByID(ctx context.Context, id int) (*model.User, error)
	UpdateUserInfo(ctx context.Context, userID int, username string, avatar []byte) (*model.User, error)
	UpdateUsername(ctx context.Context, userID int, username string) error
	GetUserRank(ctx context.Context) ([]*model.User, error)
	GetAllUsers(ctx context.Context) ([]*model.User, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	gameRepo   repository.GameRepository
	httpClient *http.Client
	wxConfig   config.WechatConfig
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, gameRepo repository.GameRepository, wxConfig config.WechatConfig) UserService {
	return &userService{
		userRepo:   userRepo,
		gameRepo:   gameRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		wxConfig:   wxConfig,
	}
}

func (s *userService) Login(ctx context.Context, code string) (*model.User, error) {
	// 调用微信接口获取 openid 和 session_key
	url := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.wxConfig.LoginURL, s.wxConfig.AppID, s.wxConfig.AppSecret, code)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		logger.Error("wechat login request failed", logger.Error(err))
		return nil, fmt.Errorf("wechat login request failed: %w", err)
	}
	defer resp.Body.Close()

	var wxSession model.WeixinSession
	if err := json.NewDecoder(resp.Body).Decode(&wxSession); err != nil {
		logger.Error("decode wechat response failed", logger.Error(err))
		return nil, fmt.Errorf("decode wechat response failed: %w", err)
	}

	if !wxSession.IsValid() {
		logger.Error("wechat login failed", logger.String("errmsg", wxSession.ErrMsg))
		return nil, fmt.Errorf("wechat login failed: %s", wxSession.ErrMsg)
	}

	// 查找或创建用户
	user, err := s.userRepo.FindByOpenID(ctx, wxSession.OpenID)
	if err != nil {
		logger.Error("find user by openid failed", logger.Error(err))
		return nil, err
	}

	now := time.Now()
	if user == nil {
		// 新用户注册
		user = &model.User{
			OpenID:        wxSession.OpenID,
			SessionKey:    wxSession.SessionKey,
			LastLoginTime: now,
			Points:        0,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			logger.Error("create user failed", logger.Error(err))
			return nil, err
		}
		logger.Info("new user registered", logger.Int("user_id", user.ID))
	} else {
		// 更新现有用户
		user.SessionKey = wxSession.SessionKey
		user.LastLoginTime = now
		if err := s.userRepo.Update(ctx, user); err != nil {
			logger.Error("update user failed", logger.Error(err))
			return nil, err
		}
		logger.Info("user login success", logger.Int("user_id", user.ID), logger.String("username", user.Username))
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *userService) UpdateUserInfo(ctx context.Context, userID int, username string, avatar []byte) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Username = username
	if avatar != nil {
		user.Avatar = avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUsername(ctx context.Context, userID int, username string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Username = username
	return s.userRepo.Update(ctx, user)
}

func (s *userService) GetUserRank(ctx context.Context) ([]*model.User, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// 获取每个用户的最近标签
	for _, user := range users {
		games, err := s.gameRepo.FindLastGamesByUser(ctx, user.ID, 20)
		if err != nil {
			logger.Warn("find last games by user failed", logger.Error(err), logger.Int("user_id", user.ID))
			continue
		}

		tagSet := make(map[string]struct{})
		for _, game := range games {
			items, err := s.gameRepo.FindGameItemsByGameID(ctx, game.ID)
			if err != nil {
				continue
			}

			gameInfo := &model.MaJiangGameInfo{Game: game, Items: items}
			userItem := gameInfo.FindUserGameItem(user.ID)
			if userItem == nil {
				continue
			}

			if userItem.Type == model.Winner {
				if game.Type == model.ZiMo {
					tagSet[game.Type.Name()] = struct{}{}
				}
				// 添加番型标签
				if userItem.WinTypes != "" {
					tagSet[userItem.WinTypes] = struct{}{}
				}
			} else if userItem.Type == model.Loser {
				if game.Type == model.YiPaoShuangXiang || game.Type == model.YiPaoSanXiang || game.Type == model.XiangGong {
					tagSet[game.Type.Name()] = struct{}{}
				}
			}
		}

		tags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			tags = append(tags, tag)
		}
		user.LastTags = tags
	}

	// 排序：非零积分用户按积分降序，零积分用户放最后
	result := make([]*model.User, 0, len(users))
	zeroUsers := make([]*model.User, 0)
	nonZeroUsers := make([]*model.User, 0)

	for _, user := range users {
		if user.Points == 0 {
			zeroUsers = append(zeroUsers, user)
		} else {
			nonZeroUsers = append(nonZeroUsers, user)
		}
	}

	// 按积分降序排序
	for i := 0; i < len(nonZeroUsers)-1; i++ {
		for j := i + 1; j < len(nonZeroUsers); j++ {
			if nonZeroUsers[i].Points < nonZeroUsers[j].Points {
				nonZeroUsers[i], nonZeroUsers[j] = nonZeroUsers[j], nonZeroUsers[i]
			}
		}
	}

	result = append(result, nonZeroUsers...)
	result = append(result, zeroUsers...)

	return result, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.FindAll(ctx)
}
