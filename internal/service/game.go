package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/pkg/logger"
)

// GameService 游戏服务接口
type GameService interface {
	SaveMaJiangGame(ctx context.Context, req *model.SaveMaJiangGameRequest) (int, error)
	GetMaJiangGameLogs(ctx context.Context, limit, offset int) ([]*model.MaJiangGameLogDTO, error)
	GetMaJiangGamesByUser(ctx context.Context, userID, limit, offset int) ([]*model.MaJiangGameLogDTO, error)
	DeleteMaJiangGame(ctx context.Context, gameID, userID int) error
	GetMaJiangGamePlayers(ctx context.Context) (*model.PlayersDTO, error)
}

// gameService 游戏服务实现
type gameService struct {
	userRepo repository.UserRepository
	gameRepo repository.GameRepository
	rand     *rand.Rand
}

// NewGameService 创建游戏服务实例
func NewGameService(userRepo repository.UserRepository, gameRepo repository.GameRepository) GameService {
	return &gameService{
		userRepo: userRepo,
		gameRepo: gameRepo,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *gameService) SaveMaJiangGame(ctx context.Context, req *model.SaveMaJiangGameRequest) (int, error) {
	// 验证请求
	if err := s.validateSaveRequest(req); err != nil {
		return 0, err
	}

	gameType := model.MaJiangGameTypeFromCode(req.GameType)

	// 创建游戏记录
	game := &model.MaJiangGame{
		Type:    gameType,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}

	// 运动类型特殊处理
	if gameType == model.YunDong {
		game.Player1 = req.RecorderID
		game.Player2 = -1
		game.Player3 = -1
		game.Player4 = -1
	} else {
		game.Player1 = req.Players[0]
		game.Player2 = req.Players[1]
		game.Player3 = req.Players[2]
		game.Player4 = req.Players[3]
	}

	if err := s.gameRepo.CreateGame(ctx, game); err != nil {
		return 0, fmt.Errorf("create game failed: %w", err)
	}

	// 创建赢家记录
	winners := make([]*model.MaJiangGameItem, 0, len(req.Winners))
	for _, w := range req.Winners {
		item := &model.MaJiangGameItem{
			GameID:    game.ID,
			UserID:    w.UserID,
			Type:      model.Winner,
			BasePoint: w.BasePoints,
			WinTypes:  strings.Join(w.WinTypes, ","),
		}
		winners = append(winners, item)
	}

	// 创建输家记录
	losers := make([]*model.MaJiangGameItem, 0)
	if gameType != model.YunDong {
		for _, loserID := range req.Losers {
			if loserID > 0 {
				item := &model.MaJiangGameItem{
					GameID: game.ID,
					UserID: loserID,
					Type:   model.Loser,
				}
				losers = append(losers, item)
			}
		}
	}

	// 计算分数
	if err := s.calculatePoints(gameType, winners, losers); err != nil {
		return 0, err
	}

	// 创建记录者记录
	recorder := &model.MaJiangGameItem{
		GameID: game.ID,
		UserID: req.RecorderID,
		Type:   model.Recorder,
	}

	if gameType == model.YunDong {
		recorder.Points = 0
	} else {
		// 1% 概率加 20 分，99% 概率加 1 分
		if s.rand.Intn(100) < 1 {
			recorder.Points = 20
		} else {
			recorder.Points = 1
		}
	}

	// 保存所有游戏条目
	allItems := append(append(winners, losers...), recorder)
	if err := s.gameRepo.CreateGameItems(ctx, allItems); err != nil {
		return 0, fmt.Errorf("create game items failed: %w", err)
	}

	// 更新用户积分
	for _, item := range allItems {
		if err := s.userRepo.UpdatePoints(ctx, item.UserID, item.Points); err != nil {
			logger.Error("update user points failed", logger.Err(err), logger.Int("user_id", item.UserID))
		}
		user, _ := s.userRepo.FindByID(ctx, item.UserID)
		if user != nil {
			logger.Info("game saved",
				logger.Int("game_id", game.ID),
				logger.String("username", user.Username),
				logger.Int("points", item.Points))
		}
	}

	return game.ID, nil
}

func (s *gameService) calculatePoints(gameType model.MaJiangGameType, winners, losers []*model.MaJiangGameItem) error {
	switch gameType {
	case model.PingHu:
		if len(winners) != 1 || len(losers) != 1 {
			return errors.New("平胡需要1个赢家和1个输家")
		}
		winners[0].CalculatePoints()
		losers[0].Points = -winners[0].Points

	case model.ZiMo:
		if len(winners) != 1 || len(losers) != 3 {
			return errors.New("自摸需要1个赢家和3个输家")
		}
		winners[0].CalculatePoints()
		points := winners[0].Points
		winners[0].Points = points * 3
		for _, loser := range losers {
			loser.Points = -points
		}

	case model.YiPaoShuangXiang:
		if len(winners) != 2 || len(losers) != 1 {
			return errors.New("一炮双响需要2个赢家和1个输家")
		}
		totalPoints := 0
		for _, winner := range winners {
			winner.CalculatePoints()
			totalPoints += winner.Points
		}
		losers[0].Points = -totalPoints

	case model.YiPaoSanXiang:
		if len(winners) != 3 || len(losers) != 1 {
			return errors.New("一炮三响需要3个赢家和1个输家")
		}
		totalPoints := 0
		for _, winner := range winners {
			winner.CalculatePoints()
			totalPoints += winner.Points
		}
		losers[0].Points = -totalPoints

	case model.XiangGong:
		if len(winners) != 3 || len(losers) != 1 {
			return errors.New("相公需要3个赢家和1个输家")
		}
		totalPoints := 0
		for _, winner := range winners {
			winner.CalculatePoints()
			totalPoints += winner.Points
		}
		losers[0].Points = -totalPoints

	case model.YunDong:
		if len(winners) != 1 {
			return errors.New("运动类型只能记录1个用户")
		}
		// 运动类型使用用户填入的分数
		winners[0].Points = winners[0].BasePoint
	}

	return nil
}

func (s *gameService) validateSaveRequest(req *model.SaveMaJiangGameRequest) error {
	gameType := model.MaJiangGameTypeFromCode(req.GameType)
	isSportType := gameType == model.YunDong

	// 验证玩家数量
	if req.Players == nil ||
		(isSportType && len(req.Players) != 1) ||
		(!isSportType && len(req.Players) != 4) {
		return errors.New("玩家数量不正确")
	}

	if req.RecorderID == 0 {
		return errors.New("记录者ID不能为空")
	}

	if len(req.Winners) == 0 {
		return errors.New("赢家不能为空")
	}

	// 非运动类型需要输家
	if !isSportType && len(req.Losers) == 0 {
		return errors.New("输家不能为空")
	}

	// 验证赢家数据
	for _, winner := range req.Winners {
		if winner.UserID == 0 {
			return errors.New("赢家用户ID不能为空")
		}
		if winner.BasePoints <= 0 {
			return errors.New("赢家基础分必须大于0")
		}
	}

	// 验证赢家必须是玩家之一
	playerSet := make(map[int]struct{})
	for _, p := range req.Players {
		playerSet[p] = struct{}{}
	}
	for _, winner := range req.Winners {
		if _, ok := playerSet[winner.UserID]; !ok {
			return errors.New("赢家必须是玩家之一")
		}
	}

	// 运动类型特殊验证
	if isSportType {
		if req.Winners[0].UserID != req.RecorderID {
			return errors.New("运动类型只能记录自己")
		}
	} else {
		// 验证赢家和输家不能有交集
		winnerSet := make(map[int]struct{})
		for _, w := range req.Winners {
			winnerSet[w.UserID] = struct{}{}
		}
		for _, loserID := range req.Losers {
			if _, ok := winnerSet[loserID]; ok {
				return errors.New("赢家和输家不能有交集")
			}
			if _, ok := playerSet[loserID]; !ok {
				return errors.New("输家必须是玩家之一")
			}
		}
	}

	return nil
}

func (s *gameService) GetMaJiangGameLogs(ctx context.Context, limit, offset int) ([]*model.MaJiangGameLogDTO, error) {
	games, err := s.gameRepo.FindLastGames(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.buildGameLogDTOs(ctx, games, 0)
}

func (s *gameService) GetMaJiangGamesByUser(ctx context.Context, userID, limit, offset int) ([]*model.MaJiangGameLogDTO, error) {
	userTypes := []model.MaJiangUserType{model.Winner, model.Loser}
	gameIDs, err := s.gameRepo.FindLastGameIDsByUserID(ctx, userID, userTypes, limit, offset)
	if err != nil {
		return nil, err
	}

	games, err := s.gameRepo.FindGamesByIDs(ctx, gameIDs)
	if err != nil {
		return nil, err
	}

	dtos, err := s.buildGameLogDTOs(ctx, games, userID)
	if err != nil {
		return nil, err
	}

	// 标记为个人视角
	for _, dto := range dtos {
		dto.ForOnePlayer = true
		for _, winner := range dto.Winners {
			if winner.User.ID == userID {
				dto.PlayerWin = true
				break
			}
		}
		if !dto.PlayerWin {
			for _, loser := range dto.Losers {
				if loser.User.ID == userID {
					dto.PlayerWin = false
					break
				}
			}
		}
	}

	return dtos, nil
}

func (s *gameService) DeleteMaJiangGame(ctx context.Context, gameID, userID int) error {
	game, err := s.gameRepo.FindGameByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// 验证记录者身份
	recorders, err := s.gameRepo.FindGameItemsByGameIDAndType(ctx, gameID, model.Recorder)
	if err != nil || len(recorders) != 1 {
		return errors.New("无法验证记录者身份")
	}

	if recorders[0].UserID != userID {
		return errors.New("只有记录者可以删除游戏记录")
	}

	// 回滚用户积分
	items, err := s.gameRepo.FindGameItemsByGameID(ctx, gameID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := s.userRepo.UpdatePoints(ctx, item.UserID, -item.Points); err != nil {
			logger.Error("rollback user points failed", logger.Err(err))
		}
		user, _ := s.userRepo.FindByID(ctx, item.UserID)
		if user != nil {
			logger.Info("game rollback",
				logger.Int("game_id", gameID),
				logger.String("username", user.Username),
				logger.Int("points", -item.Points))
		}
	}

	// 软删除游戏
	game.IsDeleted = true
	return s.gameRepo.UpdateGame(ctx, game)
}

func (s *gameService) GetMaJiangGamePlayers(ctx context.Context) (*model.PlayersDTO, error) {
	games, err := s.gameRepo.FindLastGames(ctx, 10, 0)
	if err != nil {
		return nil, err
	}

	dto := &model.PlayersDTO{
		CurrentPlayers: make([]*model.UserDTO, 0),
		AllPlayers:     make([]*model.UserDTO, 0),
	}

	// 找到最近一条非运动类型的记录
	for _, game := range games {
		if game.Type != model.YunDong {
			players := []int{game.Player1, game.Player2, game.Player3, game.Player4}
			for _, pid := range players {
				if pid > 0 {
					user, err := s.userRepo.FindByID(ctx, pid)
					if err == nil && user != nil {
						dto.CurrentPlayers = append(dto.CurrentPlayers, (&model.UserDTO{}).FromUser(user))
					}
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

func (s *gameService) buildGameLogDTOs(ctx context.Context, games []*model.MaJiangGame, targetUserID int) ([]*model.MaJiangGameLogDTO, error) {
	dtos := make([]*model.MaJiangGameLogDTO, 0, len(games))

	for _, game := range games {
		dto := &model.MaJiangGameLogDTO{
			ID:          game.ID,
			Type:        game.Type.Name(),
			CreatedTime: game.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: game.UpdatedTime.Format("2006-01-02 15:04:05"),
		}

		// 设置玩家信息
		if game.Player1 > 0 {
			user, _ := s.userRepo.FindByID(ctx, game.Player1)
			dto.Player1 = (&model.UserDTO{}).FromUser(user)
		}
		if game.Player2 > 0 {
			user, _ := s.userRepo.FindByID(ctx, game.Player2)
			dto.Player2 = (&model.UserDTO{}).FromUser(user)
		}
		if game.Player3 > 0 {
			user, _ := s.userRepo.FindByID(ctx, game.Player3)
			dto.Player3 = (&model.UserDTO{}).FromUser(user)
		}
		if game.Player4 > 0 {
			user, _ := s.userRepo.FindByID(ctx, game.Player4)
			dto.Player4 = (&model.UserDTO{}).FromUser(user)
		}

		// 获取游戏条目
		items, err := s.gameRepo.FindGameItemsByGameID(ctx, game.ID)
		if err != nil {
			continue
		}

		// 分类赢家、输家、记录者
		winners := make([]*model.MaJiangGameItem, 0)
		losers := make([]*model.MaJiangGameItem, 0)
		var recorder *model.MaJiangGameItem

		for _, item := range items {
			switch item.Type {
			case model.Winner:
				winners = append(winners, item)
			case model.Loser:
				losers = append(losers, item)
			case model.Recorder:
				recorder = item
			}
		}

		// 构建赢家 DTO
		dto.Winners = make([]*model.MaJiangGameItemDTO, 0, len(winners))
		for _, w := range winners {
			user, _ := s.userRepo.FindByID(ctx, w.UserID)
			itemDTO := &model.MaJiangGameItemDTO{
				User:   (&model.UserDTO{}).FromUser(user),
				Points: w.Points,
				Tags:   make([]string, 0),
			}
			if game.Type == model.ZiMo {
				itemDTO.Tags = append(itemDTO.Tags, game.Type.Name())
			} else {
				itemDTO.Tags = append(itemDTO.Tags, w.Type.Name())
			}
			if w.WinTypes != "" {
				itemDTO.Tags = append(itemDTO.Tags, strings.Split(w.WinTypes, ",")...)
			}
			dto.Winners = append(dto.Winners, itemDTO)
		}

		// 构建输家 DTO
		dto.Losers = make([]*model.MaJiangGameItemDTO, 0, len(losers))
		for _, l := range losers {
			user, _ := s.userRepo.FindByID(ctx, l.UserID)
			itemDTO := &model.MaJiangGameItemDTO{
				User:   (&model.UserDTO{}).FromUser(user),
				Points: l.Points,
				Tags:   make([]string, 0),
			}
			if game.Type == model.YiPaoShuangXiang || game.Type == model.YiPaoSanXiang {
				itemDTO.Tags = append(itemDTO.Tags, game.Type.Name())
			} else {
				itemDTO.Tags = append(itemDTO.Tags, l.Type.Name())
			}
			dto.Losers = append(dto.Losers, itemDTO)
		}

		// 构建记录者 DTO
		if recorder != nil {
			user, _ := s.userRepo.FindByID(ctx, recorder.UserID)
			dto.Recorder = &model.MaJiangGameItemDTO{
				User:   (&model.UserDTO{}).FromUser(user),
				Points: recorder.Points,
			}
		}

		dtos = append(dtos, dto)
	}

	return dtos, nil
}
