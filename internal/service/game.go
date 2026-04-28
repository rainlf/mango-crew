package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/rainlf/mango-crew/internal/cache"
	"github.com/rainlf/mango-crew/internal/config"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/pkg/logger"
)

// GameService 游戏服务接口
type GameService interface {
	UpdateCurrentPlayers(ctx context.Context, req *model.UpdateCurrentPlayersRequest) (*model.PlayerSummaryDTO, error)

	RecordMaJiangGame(ctx context.Context, req *model.RecordMaJiangGameRequest) (*model.Game, error)
	CancelGame(ctx context.Context, gameID int) error
	GetGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.GameDTO, error)
	GetRecentGames(ctx context.Context, limit, offset int) ([]*model.GameDTO, error)

	GetPlayers(ctx context.Context) (*model.PlayerSummaryDTO, error)
}

// gameService 游戏服务实现
type gameService struct {
	currentPlayerRepo repository.CurrentPlayerRepository
	gameRepo          repository.GameRepository
	userRepo          repository.UserRepository
	cache             *cache.Store
	cfg               *config.Config
	rand              *rand.Rand
}

// NewGameService 创建游戏服务实例
func NewGameService(currentPlayerRepo repository.CurrentPlayerRepository, gameRepo repository.GameRepository, userRepo repository.UserRepository, cacheStore *cache.Store, cfg *config.Config) GameService {
	return &gameService{
		currentPlayerRepo: currentPlayerRepo,
		gameRepo:          gameRepo,
		userRepo:          userRepo,
		cache:             cacheStore,
		cfg:               cfg,
		rand:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *gameService) UpdateCurrentPlayers(ctx context.Context, req *model.UpdateCurrentPlayersRequest) (*model.PlayerSummaryDTO, error) {
	if err := s.ensureUsersExist(ctx, []int{req.UserID}); err != nil {
		return nil, err
	}
	userIDs, err := s.normalizeCurrentPlayerIDs(req.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := s.ensureUsersExist(ctx, userIDs); err != nil {
		return nil, err
	}
	if err := s.currentPlayerRepo.ReplacePlayers(ctx, userIDs); err != nil {
		return nil, err
	}
	s.invalidatePlayerCaches(ctx)
	return s.GetPlayers(ctx)
}

func (s *gameService) RecordMaJiangGame(ctx context.Context, req *model.RecordMaJiangGameRequest) (*model.Game, error) {
	gameType := model.GameTypeFromCode(req.GameType)
	var err error

	if req.RecorderID <= 0 {
		return nil, errors.New("recorderId不能为空")
	}
	if err := s.ensureUsersExist(ctx, []int{req.RecorderID}); err != nil {
		return nil, err
	}

	currentPlayerIDs := req.Players
	if len(currentPlayerIDs) > 0 {
		currentPlayerIDs, err = s.normalizeCurrentPlayerIDs(currentPlayerIDs)
		if err != nil {
			return nil, err
		}
		if err := s.ensureUsersExist(ctx, currentPlayerIDs); err != nil {
			return nil, err
		}
		if err := s.currentPlayerRepo.ReplacePlayers(ctx, currentPlayerIDs); err != nil {
			return nil, err
		}
		s.invalidatePlayerCaches(ctx)
	} else {
		currentPlayerIDs, err = s.loadCurrentPlayerIDs(ctx)
		if err != nil {
			return nil, err
		}
	}

	if err := s.validateRecordGameRequest(req, currentPlayerIDs, gameType); err != nil {
		return nil, err
	}

	now := time.Now()
	game := &model.Game{
		Type:      gameType,
		Status:    model.GameStatusSettled,
		Remark:    req.Remark,
		CreatedBy: req.RecorderID,
		CreatedAt: now,
		SettledAt: &now,
	}
	if err := s.gameRepo.Create(ctx, game); err != nil {
		return nil, fmt.Errorf("create game failed: %w", err)
	}

	records, err := s.buildRecordedPlayers(game.ID, req, currentPlayerIDs, gameType)
	if err != nil {
		return nil, err
	}
	if err := s.gameRepo.CreateRecords(ctx, records); err != nil {
		return nil, fmt.Errorf("create game records failed: %w", err)
	}

	affectedUserIDs := append([]int{req.RecorderID}, currentPlayerIDs...)
	deltas := buildUserStatsDeltas(records)
	if err := s.userRepo.ApplyStatsDeltas(ctx, deltas); err != nil {
		return nil, fmt.Errorf("apply user stats delta failed: %w", err)
	}
	s.invalidateGameCaches(ctx, affectedUserIDs...)

	return game, nil
}

func (s *gameService) CancelGame(ctx context.Context, gameID int) error {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return err
	}
	if game.Status == model.GameStatusCanceled {
		return nil
	}

	affectedUserIDs, err := s.loadGameRelatedUserIDs(ctx, gameID)
	if err != nil {
		logger.Warn("load game users before cancel failed", logger.Int("game_id", gameID), logger.Err(err))
	}
	records, err := s.gameRepo.FindRecordsByGameID(ctx, gameID)
	if err != nil {
		return err
	}
	if err := s.gameRepo.CancelGame(ctx, gameID); err != nil {
		return err
	}
	if game.Status == model.GameStatusSettled {
		deltas := negateUserStatsDeltas(buildUserStatsDeltas(records))
		if err := s.userRepo.ApplyStatsDeltas(ctx, deltas); err != nil {
			return fmt.Errorf("apply user stats delta failed: %w", err)
		}
	}
	s.invalidateGameCaches(ctx, affectedUserIDs...)
	return nil
}

func (s *gameService) GetGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.GameDTO, error) {
	cacheKey := s.gamesByUserCacheKey(userID, limit, offset)
	var cached []*model.GameDTO
	if ok, err := s.getCache(ctx, cacheKey, &cached); err == nil && ok {
		return cached, nil
	}

	games, err := s.gameRepo.FindGamesByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	result, err := s.buildGameDTOs(ctx, games)
	if err != nil {
		return nil, err
	}
	s.setCache(ctx, cacheKey, result, s.cfg.Redis.GameListTTL())
	return result, nil
}

func (s *gameService) GetRecentGames(ctx context.Context, limit, offset int) ([]*model.GameDTO, error) {
	cacheKey := s.recentGamesCacheKey(limit, offset)
	var cached []*model.GameDTO
	if ok, err := s.getCache(ctx, cacheKey, &cached); err == nil && ok {
		return cached, nil
	}

	games, err := s.gameRepo.FindRecentGames(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	result, err := s.buildGameDTOs(ctx, games)
	if err != nil {
		return nil, err
	}
	s.setCache(ctx, cacheKey, result, s.cfg.Redis.GameListTTL())
	return result, nil
}

func (s *gameService) GetPlayers(ctx context.Context) (*model.PlayerSummaryDTO, error) {
	cacheKey := s.playersCacheKey()
	var cached model.PlayerSummaryDTO
	if ok, err := s.getCache(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	}

	dto := &model.PlayerSummaryDTO{
		CurrentPlayers: make([]*model.UserDTO, 0),
		AllPlayers:     make([]*model.UserDTO, 0),
	}

	currentPlayers, err := s.currentPlayerRepo.FindPlayers(ctx)
	if err != nil {
		return nil, err
	}

	currentUserIDs := make([]int, 0, len(currentPlayers))
	for _, player := range currentPlayers {
		currentUserIDs = append(currentUserIDs, player.UserID)
	}

	usersByID, err := s.findUsersByIDMap(ctx, currentUserIDs)
	if err != nil {
		return nil, err
	}
	for _, player := range currentPlayers {
		if user, ok := usersByID[player.UserID]; ok {
			dto.CurrentPlayers = append(dto.CurrentPlayers, (&model.UserDTO{}).FromUser(user))
		}
	}

	// 获取所有用户
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		dto.AllPlayers = append(dto.AllPlayers, (&model.UserDTO{}).FromUser(user))
	}

	s.setCache(ctx, cacheKey, dto, s.cfg.Redis.PlayerTTL())
	return dto, nil
}

func (s *gameService) validateRecordGameRequest(req *model.RecordMaJiangGameRequest, currentPlayerIDs []int, gameType model.GameType) error {
	if len(currentPlayerIDs) == 0 {
		return errors.New("当前牌桌没有玩家")
	}

	if len(currentPlayerIDs) != 4 {
		return errors.New("麻将对局必须先确定4名当前玩家")
	}

	playerSet := make(map[int]struct{}, len(currentPlayerIDs))
	for _, id := range currentPlayerIDs {
		playerSet[id] = struct{}{}
	}

	if len(req.Winners) == 0 {
		return errors.New("至少需要一个赢家")
	}

	winnerSet := make(map[int]struct{}, len(req.Winners))
	for _, winner := range req.Winners {
		if winner.UserID <= 0 {
			return errors.New("赢家用户不能为空")
		}
		if _, ok := playerSet[winner.UserID]; !ok {
			return errors.New("赢家必须在当前牌桌中")
		}
		if _, exists := winnerSet[winner.UserID]; exists {
			return errors.New("赢家不能重复")
		}
		if winner.BasePoints <= 0 {
			return errors.New("赢家底分必须大于0")
		}
		winnerSet[winner.UserID] = struct{}{}
	}

	loserSet := make(map[int]struct{}, len(req.Losers))
	for _, loserID := range req.Losers {
		if _, ok := playerSet[loserID]; !ok {
			return errors.New("输家必须在当前牌桌中")
		}
		if _, ok := winnerSet[loserID]; ok {
			return errors.New("赢家和输家不能重复")
		}
		if _, exists := loserSet[loserID]; exists {
			return errors.New("输家不能重复")
		}
		loserSet[loserID] = struct{}{}
	}

	switch gameType {
	case model.PingHu:
		if len(req.Winners) != 1 || len(req.Losers) != 1 {
			return errors.New("平胡必须是1个赢家和1个输家")
		}
	case model.ZiMo:
		if len(req.Winners) != 1 || len(req.Losers) != len(currentPlayerIDs)-1 {
			return errors.New("自摸必须是1个赢家和其余玩家全部输家")
		}
	case model.YiPaoShuangXiang:
		if len(req.Winners) != 2 || len(req.Losers) != 1 {
			return errors.New("一炮双响必须是2个赢家和1个输家")
		}
	case model.YiPaoSanXiang:
		if len(req.Winners) != 3 || len(req.Losers) != 1 {
			return errors.New("一炮三响必须是3个赢家和1个输家")
		}
	case model.XiangGong:
		if len(req.Winners) != 3 || len(req.Losers) != 1 {
			return errors.New("相公必须是3个赢家和1个输家")
		}
	}

	return nil
}

func (s *gameService) buildRecordedPlayers(gameID int, req *model.RecordMaJiangGameRequest, currentPlayerIDs []int, gameType model.GameType) ([]*model.GameRecord, error) {
	winnerMap := make(map[int]*model.RecordMaJiangWinnerDTO, len(req.Winners))
	for _, winner := range req.Winners {
		winnerMap[winner.UserID] = winner
	}

	loserSet := make(map[int]struct{}, len(req.Losers))
	for _, loserID := range req.Losers {
		loserSet[loserID] = struct{}{}
	}

	records := make([]*model.GameRecord, 0, len(currentPlayerIDs))
	recordMap := make(map[int]*model.GameRecord, len(currentPlayerIDs))
	for idx, userID := range currentPlayerIDs {
		// 对局参与者行只承载本局输赢分；记录者奖励分单独落一行 RoleRecorder。
		role := model.RoleNeutral
		basePoints := 0

		if winner, ok := winnerMap[userID]; ok {
			role = model.RoleWinner
			basePoints = winner.BasePoints
		} else if _, ok := loserSet[userID]; ok {
			role = model.RoleLoser
		}

		record := &model.GameRecord{
			GameID:     gameID,
			UserID:     userID,
			Seat:       idx + 1,
			Role:       role,
			BasePoints: basePoints,
			IsSettled:  true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if winner, ok := winnerMap[userID]; ok {
			for _, wtCode := range winner.WinTypes {
				wtInfo, found := model.ResolveWinType(wtCode)
				if !found {
					return nil, fmt.Errorf("未知番型: %s", wtCode)
				}
				record.WinTypes = append(record.WinTypes, &model.GameRecordWinType{
					WinTypeCode: wtInfo.Code,
					Multiplier:  wtInfo.BaseMulti,
				})
			}
			record.CalculatePoints()
		}

		records = append(records, record)
		recordMap[userID] = record
	}

	switch gameType {
	case model.PingHu:
		winner := recordMap[req.Winners[0].UserID]
		loser := recordMap[req.Losers[0]]
		loser.FinalPoints = -winner.FinalPoints
	case model.ZiMo:
		winner := recordMap[req.Winners[0].UserID]
		singleLosePoints := winner.FinalPoints
		winner.FinalPoints = singleLosePoints * len(req.Losers)
		for _, loserID := range req.Losers {
			recordMap[loserID].FinalPoints = -singleLosePoints
		}
	case model.YiPaoShuangXiang, model.YiPaoSanXiang, model.XiangGong:
		total := 0
		for _, winner := range req.Winners {
			total += recordMap[winner.UserID].FinalPoints
		}
		recordMap[req.Losers[0]].FinalPoints = -total
	}

	// 记录者奖励：无论记录者是否在牌桌内，都单独落一条 RoleRecorder 行，避免与输赢分混在一起。
	// seat 使用 99，避免影响前端按 seat(1-4) 渲染玩家座位。
	recorderBonus := 0
	if s.rand.Intn(100) < 1 {
		recorderBonus = 20
	} else {
		recorderBonus = 1
	}
	records = append(records, &model.GameRecord{
		GameID:      gameID,
		UserID:      req.RecorderID,
		Seat:        99,
		Role:        model.RoleRecorder,
		BasePoints:  0,
		FinalPoints: recorderBonus,
		IsSettled:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	return records, nil
}

func (s *gameService) loadCurrentPlayerIDs(ctx context.Context) ([]int, error) {
	players, err := s.currentPlayerRepo.FindPlayers(ctx)
	if err != nil {
		return nil, err
	}
	userIDs := make([]int, 0, len(players))
	for _, player := range players {
		userIDs = append(userIDs, player.UserID)
	}
	return userIDs, nil
}

func (s *gameService) normalizeCurrentPlayerIDs(userIDs []int) ([]int, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("当前牌桌玩家不能为空")
	}
	if len(userIDs) > 4 {
		return nil, errors.New("当前牌桌最多只能有4名玩家")
	}

	seen := make(map[int]struct{}, len(userIDs))
	normalized := make([]int, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			return nil, errors.New("用户ID必须大于0")
		}
		if _, ok := seen[userID]; ok {
			return nil, errors.New("当前牌桌玩家不能重复")
		}
		seen[userID] = struct{}{}
		normalized = append(normalized, userID)
	}
	return normalized, nil
}

func (s *gameService) ensureUsersExist(ctx context.Context, userIDs []int) error {
	uniqueIDs := uniqueInts(userIDs)
	users, err := s.userRepo.FindByIDIn(ctx, uniqueIDs)
	if err != nil {
		return err
	}

	exists := make(map[int]struct{}, len(users))
	for _, user := range users {
		exists[user.ID] = struct{}{}
	}

	for _, userID := range uniqueIDs {
		if _, ok := exists[userID]; !ok {
			return fmt.Errorf("用户不存在: %d", userID)
		}
	}
	return nil
}

func (s *gameService) buildGameDTOs(ctx context.Context, games []*model.Game) ([]*model.GameDTO, error) {
	var result []*model.GameDTO
	userIDs := make([]int, 0, len(games))
	gameIDs := make([]int, 0, len(games))
	for _, game := range games {
		userIDs = append(userIDs, game.CreatedBy)
		gameIDs = append(gameIDs, game.ID)
	}

	recordsByGameID, err := s.gameRepo.FindRecordsByGameIDs(ctx, gameIDs)
	if err != nil {
		return nil, err
	}

	for _, game := range games {
		records := recordsByGameID[game.ID]
		for _, record := range records {
			userIDs = append(userIDs, record.UserID)
		}
	}

	usersByID, err := s.findUsersByIDMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	for _, game := range games {
		dto := &model.GameDTO{
			ID:        game.ID,
			Type:      game.Type.Name(),
			TypeCode:  int(game.Type),
			Status:    int(game.Status),
			Remark:    game.Remark,
			CreatedAt: game.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if game.SettledAt != nil {
			formatted := game.SettledAt.Format("2006-01-02 15:04:05")
			dto.SettledAt = &formatted
		}

		// 获取创建者信息
		if creator, ok := usersByID[game.CreatedBy]; ok {
			dto.CreatedBy = (&model.UserDTO{}).FromUser(creator)
		}

		// 获取对局记录信息
		records := recordsByGameID[game.ID]
		if len(records) == 0 {
			logger.Warn("skip game without players", logger.Int("game_id", game.ID))
			continue
		}

		for _, record := range records {
			playerDTO := &model.GamePlayerDTO{
				ID:          record.ID,
				Seat:        record.Seat,
				Role:        record.Role.Name(),
				RoleCode:    int(record.Role),
				BasePoints:  record.BasePoints,
				FinalPoints: record.FinalPoints,
			}

			// 获取用户信息
			if user, ok := usersByID[record.UserID]; ok {
				playerDTO.User = (&model.UserDTO{}).FromUser(user)
			}

			// 获取番型
			for _, wt := range record.WinTypes {
				if wtInfo, ok := model.GetWinTypeByCode(wt.WinTypeCode); ok {
					playerDTO.WinTypes = append(playerDTO.WinTypes, &model.WinTypeDTO{
						Code:       wt.WinTypeCode,
						Name:       wtInfo.Name,
						Multiplier: wt.Multiplier,
					})
				}
			}

			dto.Players = append(dto.Players, playerDTO)
		}

		result = append(result, dto)
	}

	return result, nil
}

func (s *gameService) findUsersByIDMap(ctx context.Context, userIDs []int) (map[int]*model.User, error) {
	users, err := s.userRepo.FindByIDIn(ctx, uniqueInts(userIDs))
	if err != nil {
		return nil, err
	}

	usersByID := make(map[int]*model.User, len(users))
	for _, user := range users {
		usersByID[user.ID] = user
	}
	return usersByID, nil
}

func (s *gameService) loadGameRelatedUserIDs(ctx context.Context, gameID int) ([]int, error) {
	userIDs := make([]int, 0)

	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err == nil && game != nil {
		userIDs = append(userIDs, game.CreatedBy)
	}

	records, err := s.gameRepo.FindRecordsByGameID(ctx, gameID)
	if err != nil {
		if len(userIDs) > 0 {
			return uniqueInts(userIDs), nil
		}
		return nil, err
	}
	for _, record := range records {
		userIDs = append(userIDs, record.UserID)
	}
	return uniqueInts(userIDs), nil
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

func buildUserStatsDeltas(records []*model.GameRecord) map[int]model.UserStatsDelta {
	if len(records) == 0 {
		return map[int]model.UserStatsDelta{}
	}

	deltas := make(map[int]model.UserStatsDelta, len(records))
	gameCounted := make(map[int]struct{}, len(records))
	for _, record := range records {
		delta := deltas[record.UserID]
		delta.PointsDelta += record.FinalPoints
		if shouldCountRecordAsGameParticipation(record) {
			if _, ok := gameCounted[record.UserID]; !ok {
				delta.GamesDelta++
				gameCounted[record.UserID] = struct{}{}
			}
		}
		if record.Role == model.RoleWinner {
			delta.WinsDelta++
		}
		deltas[record.UserID] = delta
	}
	return deltas
}

func shouldCountRecordAsGameParticipation(record *model.GameRecord) bool {
	if record == nil {
		return false
	}
	// 记录人奖励行只记奖励分，不代表实际参赛。
	return record.Role != model.RoleRecorder
}

func negateUserStatsDeltas(deltas map[int]model.UserStatsDelta) map[int]model.UserStatsDelta {
	if len(deltas) == 0 {
		return map[int]model.UserStatsDelta{}
	}

	result := make(map[int]model.UserStatsDelta, len(deltas))
	for userID, delta := range deltas {
		result[userID] = delta.Negate()
	}
	return result
}

func (s *gameService) playersCacheKey() string {
	return "players:summary"
}

func (s *gameService) recentGamesCacheKey(limit, offset int) string {
	return "games:recent:limit:" + strconv.Itoa(limit) + ":offset:" + strconv.Itoa(offset)
}

func (s *gameService) gamesByUserCacheKey(userID, limit, offset int) string {
	return "games:user:" + strconv.Itoa(userID) + ":limit:" + strconv.Itoa(limit) + ":offset:" + strconv.Itoa(offset)
}

func (s *gameService) invalidatePlayerCaches(ctx context.Context) {
	s.deleteCache(ctx, s.playersCacheKey())
}

func (s *gameService) invalidateGameCaches(ctx context.Context, userIDs ...int) {
	keys := []string{s.playersCacheKey(), "users:all", "users:rank", "users:rank:v2"}
	for _, userID := range uniqueInts(userIDs) {
		keys = append(keys, "user:stats:"+strconv.Itoa(userID))
	}
	s.deleteCache(ctx, keys...)
	s.deleteCacheByPrefix(ctx, "games:recent:", "games:user:")
}

func (s *gameService) getCache(ctx context.Context, key string, dest any) (bool, error) {
	if s.cache == nil || s.cfg == nil {
		return false, nil
	}
	ok, err := s.cache.GetJSON(ctx, key, dest)
	if err != nil {
		logger.Warn("read cache failed", logger.String("key", key), logger.Err(err))
	}
	return ok, err
}

func (s *gameService) setCache(ctx context.Context, key string, value any, ttl time.Duration) {
	if s.cache == nil || s.cfg == nil {
		return
	}
	if err := s.cache.SetJSON(ctx, key, value, ttl); err != nil {
		logger.Warn("write cache failed", logger.String("key", key), logger.Err(err))
	}
}

func (s *gameService) deleteCache(ctx context.Context, keys ...string) {
	if s.cache == nil || len(keys) == 0 {
		return
	}
	if err := s.cache.Delete(ctx, keys...); err != nil {
		logger.Warn("delete cache failed", logger.Any("keys", keys), logger.Err(err))
	}
}

func (s *gameService) deleteCacheByPrefix(ctx context.Context, prefixes ...string) {
	if s.cache == nil || len(prefixes) == 0 {
		return
	}
	if err := s.cache.DeleteByPrefix(ctx, prefixes...); err != nil {
		logger.Warn("delete cache by prefix failed", logger.Any("prefixes", prefixes), logger.Err(err))
	}
}
