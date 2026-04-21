package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rainlf/mango-crew/internal/config"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/pkg/logger"
)

// UserService 用户服务接口
type UserService interface {
	Login(ctx context.Context, code string) (*model.User, error)
	GetUserByID(ctx context.Context, id int) (*model.UserWithStatsDTO, error)
	UpdateUser(ctx context.Context, userID int, req *model.UpdateUserRequest) (*model.User, error)
	GetUserRank(ctx context.Context) ([]*model.UserWithStatsDTO, error)
	GetAllUsers(ctx context.Context) ([]*model.UserDTO, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	gameRepo   repository.GameRepository
	httpClient *http.Client
	wxConfig   config.WechatConfig
	appID      string
	appSecret  string
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, gameRepo repository.GameRepository, wxConfig config.WechatConfig, appID, appSecret string) UserService {
	return &userService{
		userRepo:   userRepo,
		gameRepo:   gameRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		wxConfig:   wxConfig,
		appID:      appID,
		appSecret:  appSecret,
	}
}

func (s *userService) Login(ctx context.Context, code string) (*model.User, error) {
	// 调用微信接口获取 openid 和 session_key
	url := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.wxConfig.LoginURL, s.appID, s.appSecret, code)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		logger.Error("wechat login request failed", logger.Err(err))
		return nil, fmt.Errorf("wechat login request failed: %w", err)
	}
	defer resp.Body.Close()

	var wxSession model.WeixinSession
	if err := json.NewDecoder(resp.Body).Decode(&wxSession); err != nil {
		logger.Error("decode wechat response failed", logger.Err(err))
		return nil, fmt.Errorf("decode wechat response failed: %w", err)
	}

	if !wxSession.IsValid() {
		logger.Error("wechat login failed", logger.String("errmsg", wxSession.ErrMsg))
		return nil, fmt.Errorf("wechat login failed: %s", wxSession.ErrMsg)
	}

	// 查找或创建用户
	user, err := s.userRepo.FindByOpenID(ctx, wxSession.OpenID)
	if err != nil {
		logger.Error("find user by openid failed", logger.Err(err))
		return nil, err
	}

	now := time.Now()
	if user == nil {
		// 新用户注册
		user = &model.User{
			OpenID:     wxSession.OpenID,
			SessionKey: wxSession.SessionKey,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			logger.Error("create user failed", logger.Err(err))
			return nil, err
		}
		logger.Info("new user registered", logger.Int("user_id", user.ID))
	} else {
		// 更新现有用户
		user.SessionKey = wxSession.SessionKey
		user.UpdatedAt = now
		if err := s.userRepo.Update(ctx, user); err != nil {
			logger.Error("update user failed", logger.Err(err))
			return nil, err
		}
		logger.Info("user login success", logger.Int("user_id", user.ID), logger.String("nickname", user.Nickname))
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int) (*model.UserWithStatsDTO, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 实时计算统计数据
	totalPoints, err := s.gameRepo.SumPlayerPoints(ctx, id)
	if err != nil {
		logger.Warn("sum player points failed", logger.Err(err), logger.Int("user_id", id))
	}

	totalGames, err := s.gameRepo.CountPlayerGames(ctx, id)
	if err != nil {
		logger.Warn("count player games failed", logger.Err(err), logger.Int("user_id", id))
	}

	winCount, err := s.gameRepo.CountPlayerWins(ctx, id)
	if err != nil {
		logger.Warn("count player wins failed", logger.Err(err), logger.Int("user_id", id))
	}

	return &model.UserWithStatsDTO{
		UserDTO:     (&model.UserDTO{}).FromUser(user),
		TotalPoints: totalPoints,
		TotalGames:  int(totalGames),
		WinCount:    int(winCount),
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID int, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	if req.Avatar != "" {
		user.AvatarURL = req.Avatar
	}

	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserRank(ctx context.Context) ([]*model.UserWithStatsDTO, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.UserWithStatsDTO
	for _, user := range users {
		totalPoints, _ := s.gameRepo.SumPlayerPoints(ctx, user.ID)
		totalGames, _ := s.gameRepo.CountPlayerGames(ctx, user.ID)
		winCount, _ := s.gameRepo.CountPlayerWins(ctx, user.ID)

		result = append(result, &model.UserWithStatsDTO{
			UserDTO:     (&model.UserDTO{}).FromUser(user),
			TotalPoints: totalPoints,
			TotalGames:  int(totalGames),
			WinCount:    int(winCount),
		})
	}

	// 按积分降序排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].TotalPoints < result[j].TotalPoints {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*model.UserDTO, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*model.UserDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, (&model.UserDTO{}).FromUser(user))
	}
	return dtos, nil
}

// decodeBase64Avatar 解码base64头像（如果需要）
func decodeBase64Avatar(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}
