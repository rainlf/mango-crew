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

	CreateGame(ctx context.Context, userID int, req *model.CreateGameRequest) (*model.Game, error)
	SettleGame(ctx context.Context, gameID int) error
	CancelGame(ctx context.Context, gameID int) error
	GetGamesBySession(ctx context.Context, sessionID int, limit, offset int) ([]*model.GameDTO, error)
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

// Game 相关

func (s *gameService) CreateGame(ctx context.Context, userID int, req *model.CreateGameRequest) (*model.Game, error) {
	// 验证请求
	if err := s.validateCreateGameRequest(req); err != nil {
		return nil, err
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
				if wt, ok := model.GetWinTypeByCode(wtCode); ok {
					player.WinTypes = append(player.WinTypes, &model.GamePlayerWinType{
						WinTypeCode: wtCode,
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

func (s *gameService) GetRecentGames(ctx context.Context, limit, offset int) ([]*model.GameDTO, error) {
	games, err := s.gameRepo.FindRecentGames(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.buildGameDTOs(ctx, games)
}

func (s *gameService) GetPlayers(ctx context.Context) (*model.PlayerSummaryDTO, error) {
	// 获取最近一场非运动类型的游戏
	games, err := s.gameRepo.FindRecentGames(ctx, 10, 0)
	if err != nil {
		return nil, err
	}

	dto := &model.PlayerSummaryDTO{
		CurrentPlayers: make([]*model.UserDTO, 0),
		AllPlayers:     make([]*model.UserDTO, 0),
	}

	// 找到最近一场非运动类型的玩家
	for _, game := range games {
		if game.Type != model.YunDong {
			players, err := s.gameRepo.FindPlayersByGameID(ctx, game.ID)
			if err != nil {
				break
			}
			for _, player := range players {
				user, err := s.userRepo.FindByID(ctx, player.UserID)
				if err == nil && user != nil {
					dto.CurrentPlayers = append(dto.CurrentPlayers, (&model.UserDTO{}).FromUser(user))
				}
			}
			break
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
