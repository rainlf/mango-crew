package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rainlf/mango-crew/internal/cache"
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
	RebuildUserStats(ctx context.Context, userIDs []int) (int, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	gameRepo   repository.GameRepository
	cache      *cache.Store
	cfg        *config.Config
	httpClient *http.Client
	wxConfig   config.WechatConfig
	appID      string
	appSecret  string
}

const maxNicknameLength = 4

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, gameRepo repository.GameRepository, cacheStore *cache.Store, cfg *config.Config, wxConfig config.WechatConfig, appID, appSecret string) UserService {
	return &userService{
		userRepo:   userRepo,
		gameRepo:   gameRepo,
		cache:      cacheStore,
		cfg:        cfg,
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
		s.invalidateUserCaches(ctx, user.ID)
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
	cacheKey := s.userStatsCacheKey(id)
	var cached model.UserWithStatsDTO
	if ok, err := s.getCache(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := (&model.UserWithStatsDTO{}).FromUser(user)
	s.setCache(ctx, cacheKey, result, s.cfg.Redis.UserTTL())
	return result, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID int, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Nickname != "" {
		nickname := strings.TrimSpace(req.Nickname)
		if nickname == "" {
			return nil, fmt.Errorf("昵称不能为空")
		}
		if utf8.RuneCountInString(nickname) > maxNicknameLength {
			return nil, fmt.Errorf("昵称最多%d个字", maxNicknameLength)
		}
		user.Nickname = nickname
	}

	if req.Avatar != "" {
		user.AvatarURL = req.Avatar
	}

	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.invalidateUserCaches(ctx, userID)
	return user, nil
}

func (s *userService) GetUserRank(ctx context.Context) ([]*model.UserWithStatsDTO, error) {
	cacheKey := s.rankCacheKey()
	var cached []*model.UserWithStatsDTO
	if ok, err := s.getCache(ctx, cacheKey, &cached); err == nil && ok {
		return cached, nil
	}

	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.UserWithStatsDTO
	for _, user := range users {
		result = append(result, (&model.UserWithStatsDTO{}).FromUser(user))
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalPoints == result[j].TotalPoints {
			return result[i].WinRate > result[j].WinRate
		}
		return result[i].TotalPoints > result[j].TotalPoints
	})

	if err := s.attachRankTags(ctx, result); err != nil {
		return nil, err
	}

	s.setCache(ctx, cacheKey, result, s.cfg.Redis.RankTTL())
	return result, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*model.UserDTO, error) {
	cacheKey := s.allUsersCacheKey()
	var cached []*model.UserDTO
	if ok, err := s.getCache(ctx, cacheKey, &cached); err == nil && ok {
		return cached, nil
	}

	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*model.UserDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, (&model.UserDTO{}).FromUser(user))
	}

	s.setCache(ctx, cacheKey, dtos, s.cfg.Redis.UserTTL())
	return dtos, nil
}

func (s *userService) RebuildUserStats(ctx context.Context, userIDs []int) (int, error) {
	targetIDs := uniqueInts(userIDs)
	if len(targetIDs) == 0 {
		users, err := s.userRepo.FindAll(ctx)
		if err != nil {
			return 0, err
		}
		targetIDs = make([]int, 0, len(users))
		for _, user := range users {
			targetIDs = append(targetIDs, user.ID)
		}
	}

	if len(targetIDs) == 0 {
		return 0, nil
	}

	if err := s.userRepo.RefreshStatsByUserIDs(ctx, targetIDs); err != nil {
		return 0, err
	}

	s.deleteCache(ctx, s.rankCacheKey(), s.allUsersCacheKey())
	s.deleteCacheByPrefix(ctx, "user:stats:")

	return len(targetIDs), nil
}

// decodeBase64Avatar 解码base64头像（如果需要）
func decodeBase64Avatar(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}

func (s *userService) userStatsCacheKey(userID int) string {
	return "user:stats:" + strconv.Itoa(userID)
}

func (s *userService) rankCacheKey() string {
	return "users:rank:v2"
}

func (s *userService) allUsersCacheKey() string {
	return "users:all"
}

func (s *userService) invalidateUserCaches(ctx context.Context, userID int) {
	s.deleteCache(ctx, s.userStatsCacheKey(userID), s.rankCacheKey(), "users:rank", s.allUsersCacheKey(), "players:summary")
	s.deleteCacheByPrefix(ctx, "games:recent:", "games:user:")
}

func (s *userService) getCache(ctx context.Context, key string, dest any) (bool, error) {
	if s.cache == nil || s.cfg == nil {
		return false, nil
	}
	ok, err := s.cache.GetJSON(ctx, key, dest)
	if err != nil {
		logger.Warn("read cache failed", logger.String("key", key), logger.Err(err))
	}
	return ok, err
}

func (s *userService) setCache(ctx context.Context, key string, value any, ttl time.Duration) {
	if s.cache == nil || s.cfg == nil {
		return
	}
	if err := s.cache.SetJSON(ctx, key, value, ttl); err != nil {
		logger.Warn("write cache failed", logger.String("key", key), logger.Err(err))
	}
}

func (s *userService) deleteCache(ctx context.Context, keys ...string) {
	if s.cache == nil || len(keys) == 0 {
		return
	}
	if err := s.cache.Delete(ctx, keys...); err != nil {
		logger.Warn("delete cache failed", logger.Any("keys", keys), logger.Err(err))
	}
}

func (s *userService) deleteCacheByPrefix(ctx context.Context, prefixes ...string) {
	if s.cache == nil || len(prefixes) == 0 {
		return
	}
	if err := s.cache.DeleteByPrefix(ctx, prefixes...); err != nil {
		logger.Warn("delete cache by prefix failed", logger.Any("prefixes", prefixes), logger.Err(err))
	}
}

func (s *userService) attachRankTags(ctx context.Context, users []*model.UserWithStatsDTO) error {
	if len(users) == 0 || s.gameRepo == nil {
		return nil
	}

	userIDs := make([]int, 0, len(users))
	for _, user := range users {
		if user == nil || user.UserDTO == nil {
			continue
		}
		userIDs = append(userIDs, user.ID)
	}

	recordsByUserID, err := s.gameRepo.FindRecentWinningRecordsByUserIDs(ctx, userIDs, 10)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user == nil || user.UserDTO == nil {
			continue
		}
		user.Tags = summarizeRankTags(recordsByUserID[user.ID], 0)
	}

	return nil
}

func summarizeRankTags(records []*model.GameRecord, limit int) []string {
	if len(records) == 0 {
		return nil
	}

	type tagStat struct {
		Code          string
		Name          string
		Count         int
		FirstSeenRank int
	}

	statsByCode := make(map[string]*tagStat)
	for recordIdx, record := range records {
		if record == nil || len(record.WinTypes) == 0 {
			continue
		}

		for _, wt := range record.WinTypes {
			if wt == nil {
				continue
			}

			wtInfo, ok := model.GetWinTypeByCode(wt.WinTypeCode)
			if !ok || wtInfo.Name == "无花果" {
				continue
			}

			stat, exists := statsByCode[wt.WinTypeCode]
			if !exists {
				stat = &tagStat{
					Code:          wt.WinTypeCode,
					Name:          wtInfo.Name,
					FirstSeenRank: recordIdx,
				}
				statsByCode[wt.WinTypeCode] = stat
			}
			stat.Count++
		}
	}

	if len(statsByCode) == 0 {
		return nil
	}

	stats := make([]*tagStat, 0, len(statsByCode))
	for _, stat := range statsByCode {
		stats = append(stats, stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		if stats[i].FirstSeenRank != stats[j].FirstSeenRank {
			return stats[i].FirstSeenRank < stats[j].FirstSeenRank
		}
		return stats[i].Name < stats[j].Name
	})

	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}

	tags := make([]string, 0, len(stats))
	for _, stat := range stats {
		tags = append(tags, stat.Name)
	}
	return tags
}
