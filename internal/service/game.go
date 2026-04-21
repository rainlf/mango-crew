package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/pkg/logger"
)

// GameService 游戏服务接口
type GameService interface {
	CreateSession(ctx context.Context, userID int, req *model.CreateSessionRequest) (*model.GameSession, error)
	EndSession(ctx context.Context, sessionID int) error
	GetSessions(ctx context.Context, limit, offset int) ([]*model.GameSessionDTO, error)
	GetActiveSessions(ctx context.Context) ([]*model.GameSessionDTO, error)
	UpdateCurrentPlayers(ctx context.Context, userID int, req *model.UpdateCurrentPlayersRequest) (*model.PlayerSummaryDTO, error)

	CreateGame(ctx context.Context, userID int, req *model.CreateGameRequest) (*model.Game, error)
	RecordMaJiangGame(ctx context.Context, req *model.RecordMaJiangGameRequest) (*model.Game, error)
	SettleGame(ctx context.Context, gameID int) error
	CancelGame(ctx context.Context, gameID int) error
	GetGamesBySession(ctx context.Context, sessionID int, limit, offset int) ([]*model.GameDTO, error)
	GetGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.GameDTO, error)
	GetRecentGames(ctx context.Context, limit, offset int) ([]*model.GameDTO, error)

	GetPlayers(ctx context.Context) (*model.PlayerSummaryDTO, error)
}

// gameService 游戏服务实现
type gameService struct {
	sessionRepo repository.GameSessionRepository
	gameRepo    repository.GameRepository
	userRepo    repository.UserRepository
	rand        *rand.Rand
}

// NewGameService 创建游戏服务实例
func NewGameService(sessionRepo repository.GameSessionRepository, gameRepo repository.GameRepository, userRepo repository.UserRepository) GameService {
	return &gameService{
		sessionRepo: sessionRepo,
		gameRepo:    gameRepo,
		userRepo:    userRepo,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Session 相关

func (s *gameService) CreateSession(ctx context.Context, userID int, req *model.CreateSessionRequest) (*model.GameSession, error) {
	session := &model.GameSession{
		Name:      req.Name,
		Status:    0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *gameService) EndSession(ctx context.Context, sessionID int) error {
	return s.sessionRepo.EndSession(ctx, sessionID)
}

func (s *gameService) GetSessions(ctx context.Context, limit, offset int) ([]*model.GameSessionDTO, error) {
	sessions, err := s.sessionRepo.FindAllSessions(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var result []*model.GameSessionDTO
	for _, session := range sessions {
		gameCount, _ := s.sessionRepo.CountGames(ctx, session.ID)
		creator, _ := s.userRepo.FindByID(ctx, session.CreatedBy)

		result = append(result, &model.GameSessionDTO{
			ID:        session.ID,
			Name:      session.Name,
			Status:    session.Status,
			CreatedBy: (&model.UserDTO{}).FromUser(creator),
			GameCount: int(gameCount),
			CreatedAt: session.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *gameService) GetActiveSessions(ctx context.Context) ([]*model.GameSessionDTO, error) {
	sessions, err := s.sessionRepo.FindActiveSessions(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.GameSessionDTO
	for _, session := range sessions {
		gameCount, _ := s.sessionRepo.CountGames(ctx, session.ID)
		creator, _ := s.userRepo.FindByID(ctx, session.CreatedBy)

		result = append(result, &model.GameSessionDTO{
			ID:        session.ID,
			Name:      session.Name,
			Status:    session.Status,
			CreatedBy: (&model.UserDTO{}).FromUser(creator),
			GameCount: int(gameCount),
			CreatedAt: session.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *gameService) UpdateCurrentPlayers(ctx context.Context, userID int, req *model.UpdateCurrentPlayersRequest) (*model.PlayerSummaryDTO, error) {
	userIDs, err := s.normalizeCurrentPlayerIDs(req.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := s.ensureUsersExist(ctx, userIDs); err != nil {
		return nil, err
	}

	session, err := s.getOrCreateActiveSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.ReplacePlayers(ctx, session.ID, userIDs); err != nil {
		return nil, err
	}
	return s.GetPlayers(ctx)
}

// Game 相关

func (s *gameService) CreateGame(ctx context.Context, userID int, req *model.CreateGameRequest) (*model.Game, error) {
	// 验证请求
	if err := s.validateCreateGameRequest(req); err != nil {
		return nil, err
	}

	if req.SessionID == 0 {
		activeSession, sessionErr := s.getOrCreateActiveSession(ctx, userID)
		if sessionErr != nil {
			return nil, sessionErr
		}
		req.SessionID = activeSession.ID
	}

	gameType := model.GameTypeFromCode(req.GameType)

	// 创建游戏记录
	game := &model.Game{
		SessionID: req.SessionID,
		Type:      gameType,
		Status:    model.GameStatusPending,
		Remark:    req.Remark,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}

	if err := s.gameRepo.Create(ctx, game); err != nil {
		return nil, fmt.Errorf("create game failed: %w", err)
	}

	// 创建玩家记录
	players := make([]*model.GamePlayer, 0, len(req.Players))
	for _, p := range req.Players {
		player := &model.GamePlayer{
			GameID:     game.ID,
			UserID:     p.UserID,
			Seat:       p.Seat,
			Role:       model.PlayerRole(p.Role),
			BasePoints: p.BasePoints,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// 处理番型
		if len(p.WinTypes) > 0 {
			player.WinTypes = make([]*model.GamePlayerWinType, 0, len(p.WinTypes))
			for _, wtCode := range p.WinTypes {
				if wt, ok := model.ResolveWinType(wtCode); ok {
					player.WinTypes = append(player.WinTypes, &model.GamePlayerWinType{
						WinTypeCode: wt.Code,
						Multiplier:  wt.BaseMulti,
					})
				}
			}
		}

		// 计算最终分数
		player.CalculatePoints()
		players = append(players, player)
	}

	// 记录者加分逻辑
	for _, player := range players {
		if player.Role == model.RoleRecorder && gameType != model.YunDong {
			// 1% 概率加 20 分，99% 概率加 1 分
			if s.rand.Intn(100) < 1 {
				player.FinalPoints = 20
			} else {
				player.FinalPoints = 1
			}
		}
	}

	if err := s.gameRepo.CreatePlayers(ctx, players); err != nil {
		return nil, fmt.Errorf("create players failed: %w", err)
	}

	// 保存番型
	for _, player := range players {
		if len(player.WinTypes) > 0 {
			for _, wt := range player.WinTypes {
				wt.GamePlayerID = player.ID
			}
			if err := s.gameRepo.CreateWinTypes(ctx, player.WinTypes); err != nil {
				logger.Warn("create win types failed", logger.Err(err))
			}
		}
	}

	return game, nil
}

func (s *gameService) RecordMaJiangGame(ctx context.Context, req *model.RecordMaJiangGameRequest) (*model.Game, error) {
	gameType := model.GameTypeFromCode(req.GameType)

	if req.RecorderID <= 0 {
		return nil, errors.New("recorderId不能为空")
	}
	if err := s.ensureUsersExist(ctx, []int{req.RecorderID}); err != nil {
		return nil, err
	}

	session, err := s.getOrCreateActiveSession(ctx, req.RecorderID)
	if err != nil {
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
		if err := s.sessionRepo.ReplacePlayers(ctx, session.ID, currentPlayerIDs); err != nil {
			return nil, err
		}
	} else {
		currentPlayerIDs, err = s.loadCurrentPlayerIDs(ctx, session.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := s.validateRecordGameRequest(req, currentPlayerIDs, gameType); err != nil {
		return nil, err
	}

	now := time.Now()
	game := &model.Game{
		SessionID: session.ID,
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

	players, err := s.buildRecordedPlayers(game.ID, req, currentPlayerIDs, gameType)
	if err != nil {
		return nil, err
	}
	if err := s.gameRepo.CreatePlayers(ctx, players); err != nil {
		return nil, fmt.Errorf("create players failed: %w", err)
	}

	for _, player := range players {
		if len(player.WinTypes) == 0 {
			continue
		}
		for _, wt := range player.WinTypes {
			wt.GamePlayerID = player.ID
		}
		if err := s.gameRepo.CreateWinTypes(ctx, player.WinTypes); err != nil {
			logger.Warn("create win types failed", logger.Err(err))
		}
	}

	return game, nil
}

func (s *gameService) SettleGame(ctx context.Context, gameID int) error {
	return s.gameRepo.SettleGame(ctx, gameID)
}

func (s *gameService) CancelGame(ctx context.Context, gameID int) error {
	return s.gameRepo.CancelGame(ctx, gameID)
}

func (s *gameService) GetGamesBySession(ctx context.Context, sessionID int, limit, offset int) ([]*model.GameDTO, error) {
	games, err := s.gameRepo.FindBySessionID(ctx, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.buildGameDTOs(ctx, games)
}

func (s *gameService) GetGamesByUser(ctx context.Context, userID int, limit, offset int) ([]*model.GameDTO, error) {
	games, err := s.gameRepo.FindGamesByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.buildGameDTOs(ctx, games)
}

func (s *gameService) GetRecentGames(ctx context.Context, limit, offset int) ([]*model.GameDTO, error) {
	games, err := s.gameRepo.FindRecentGames(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.buildGameDTOs(ctx, games)
}

func (s *gameService) GetPlayers(ctx context.Context) (*model.PlayerSummaryDTO, error) {
	dto := &model.PlayerSummaryDTO{
		CurrentPlayers: make([]*model.UserDTO, 0),
		AllPlayers:     make([]*model.UserDTO, 0),
	}

	session, err := s.sessionRepo.FindLatestActive(ctx)
	if err != nil {
		return nil, err
	}
	if session != nil {
		currentPlayers, findErr := s.sessionRepo.FindPlayers(ctx, session.ID)
		if findErr != nil {
			return nil, findErr
		}
		for _, player := range currentPlayers {
			user, findErr := s.userRepo.FindByID(ctx, player.UserID)
			if findErr == nil && user != nil {
				dto.CurrentPlayers = append(dto.CurrentPlayers, (&model.UserDTO{}).FromUser(user))
			}
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

	return dto, nil
}

// 辅助方法

func (s *gameService) validateCreateGameRequest(req *model.CreateGameRequest) error {
	gameType := model.GameTypeFromCode(req.GameType)

	// 验证玩家数量
	isSportType := gameType == model.YunDong
	if isSportType && len(req.Players) != 1 {
		return errors.New("运动类型只能有1个玩家")
	}
	if !isSportType && len(req.Players) < 2 {
		return errors.New("非运动类型至少需要2个玩家")
	}

	// 验证有赢家
	hasWinner := false
	for _, p := range req.Players {
		if p.Role == 1 { // Winner
			hasWinner = true
			break
		}
	}
	if !hasWinner {
		return errors.New("至少需要一个赢家")
	}

	return nil
}

func (s *gameService) validateRecordGameRequest(req *model.RecordMaJiangGameRequest, currentPlayerIDs []int, gameType model.GameType) error {
	if len(currentPlayerIDs) == 0 {
		return errors.New("当前牌桌没有玩家")
	}

	if gameType == model.YunDong {
		if len(currentPlayerIDs) != 1 {
			return errors.New("运动类型只能有1个玩家")
		}
	} else if len(currentPlayerIDs) != 4 {
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
		if gameType != model.YunDong && winner.BasePoints <= 0 {
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
	case model.YunDong:
		if len(req.Winners) != 1 || len(req.Losers) != 0 {
			return errors.New("运动类型只能记录1个赢家且没有输家")
		}
		if req.Winners[0].UserID != currentPlayerIDs[0] {
			return errors.New("运动类型只能记录当前牌桌中的唯一玩家")
		}
		if req.Winners[0].BasePoints < 0 {
			return errors.New("运动类型分数不能小于0")
		}
	}

	return nil
}

func (s *gameService) buildRecordedPlayers(gameID int, req *model.RecordMaJiangGameRequest, currentPlayerIDs []int, gameType model.GameType) ([]*model.GamePlayer, error) {
	winnerMap := make(map[int]*model.RecordMaJiangWinnerDTO, len(req.Winners))
	for _, winner := range req.Winners {
		winnerMap[winner.UserID] = winner
	}

	loserSet := make(map[int]struct{}, len(req.Losers))
	for _, loserID := range req.Losers {
		loserSet[loserID] = struct{}{}
	}

	players := make([]*model.GamePlayer, 0, len(currentPlayerIDs))
	playerMap := make(map[int]*model.GamePlayer, len(currentPlayerIDs))
	for idx, userID := range currentPlayerIDs {
		role := model.RoleNeutral
		basePoints := 0

		if winner, ok := winnerMap[userID]; ok {
			role = model.RoleWinner
			basePoints = winner.BasePoints
		} else if _, ok := loserSet[userID]; ok {
			role = model.RoleLoser
		} else if userID == req.RecorderID {
			role = model.RoleRecorder
		}

		player := &model.GamePlayer{
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
				player.WinTypes = append(player.WinTypes, &model.GamePlayerWinType{
					WinTypeCode: wtInfo.Code,
					Multiplier:  wtInfo.BaseMulti,
				})
			}
			player.CalculatePoints()
		}

		players = append(players, player)
		playerMap[userID] = player
	}

	switch gameType {
	case model.PingHu:
		winner := playerMap[req.Winners[0].UserID]
		loser := playerMap[req.Losers[0]]
		loser.FinalPoints = -winner.FinalPoints
	case model.ZiMo:
		winner := playerMap[req.Winners[0].UserID]
		singleLosePoints := winner.FinalPoints
		winner.FinalPoints = singleLosePoints * len(req.Losers)
		for _, loserID := range req.Losers {
			playerMap[loserID].FinalPoints = -singleLosePoints
		}
	case model.YiPaoShuangXiang, model.YiPaoSanXiang, model.XiangGong:
		total := 0
		for _, winner := range req.Winners {
			total += playerMap[winner.UserID].FinalPoints
		}
		playerMap[req.Losers[0]].FinalPoints = -total
	case model.YunDong:
		winner := playerMap[req.Winners[0].UserID]
		winner.FinalPoints = req.Winners[0].BasePoints
	}

	if gameType != model.YunDong {
		if recorder, ok := playerMap[req.RecorderID]; ok {
			if recorder.Role == model.RoleNeutral {
				recorder.Role = model.RoleRecorder
			}
			if s.rand.Intn(100) < 1 {
				recorder.FinalPoints += 20
			} else {
				recorder.FinalPoints++
			}
		}
	}

	return players, nil
}

func (s *gameService) getOrCreateActiveSession(ctx context.Context, userID int) (*model.GameSession, error) {
	session, err := s.sessionRepo.FindLatestActive(ctx)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	session = &model.GameSession{
		Name:      "当前牌桌",
		Status:    0,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *gameService) loadCurrentPlayerIDs(ctx context.Context, sessionID int) ([]int, error) {
	players, err := s.sessionRepo.FindPlayers(ctx, sessionID)
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
	seen := make(map[int]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil || user == nil {
			return fmt.Errorf("用户不存在: %d", userID)
		}
	}
	return nil
}

func (s *gameService) buildGameDTOs(ctx context.Context, games []*model.Game) ([]*model.GameDTO, error) {
	var result []*model.GameDTO

	for _, game := range games {
		dto := &model.GameDTO{
			ID:        game.ID,
			SessionID: game.SessionID,
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
		creator, err := s.userRepo.FindByID(ctx, game.CreatedBy)
		if err == nil {
			dto.CreatedBy = (&model.UserDTO{}).FromUser(creator)
		}

		// 获取玩家信息
		players, err := s.gameRepo.FindPlayersByGameID(ctx, game.ID)
		if err != nil {
			continue
		}
		if len(players) == 0 {
			logger.Warn("skip game without players", logger.Int("game_id", game.ID))
			continue
		}

		for _, player := range players {
			playerDTO := &model.GamePlayerDTO{
				ID:          player.ID,
				Seat:        player.Seat,
				Role:        player.Role.Name(),
				RoleCode:    int(player.Role),
				BasePoints:  player.BasePoints,
				FinalPoints: player.FinalPoints,
			}

			// 获取用户信息
			user, err := s.userRepo.FindByID(ctx, player.UserID)
			if err == nil {
				playerDTO.User = (&model.UserDTO{}).FromUser(user)
			}

			// 获取番型
			winTypes, err := s.gameRepo.FindWinTypesByPlayerID(ctx, player.ID)
			if err == nil {
				for _, wt := range winTypes {
					if wtInfo, ok := model.GetWinTypeByCode(wt.WinTypeCode); ok {
						playerDTO.WinTypes = append(playerDTO.WinTypes, &model.WinTypeDTO{
							Code:       wt.WinTypeCode,
							Name:       wtInfo.Name,
							Multiplier: wt.Multiplier,
						})
					}
				}
			}

			dto.Players = append(dto.Players, playerDTO)
		}

		result = append(result, dto)
	}

	return result, nil
}
