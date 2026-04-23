package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/service"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/rainlf/mango-crew/pkg/response"
)

// GameHandler 游戏处理器
type GameHandler struct {
	gameService service.GameService
}

// NewGameHandler 创建游戏处理器
func NewGameHandler(gameService service.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// CancelGame 取消游戏
func (h *GameHandler) CancelGame(c *gin.Context) {
	var req model.CancelGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.gameService.CancelGame(c.Request.Context(), req.GameID); err != nil {
		logger.Error("cancel game failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetRecentGames 获取最近的游戏列表
func (h *GameHandler) GetRecentGames(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if val, parseErr := strconv.Atoi(l); parseErr == nil && val > 0 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, parseErr := strconv.Atoi(o); parseErr == nil && val >= 0 {
			offset = val
		}
	}

	games, err := h.gameService.GetRecentGames(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("get recent games failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, games)
}

// GetPlayers 获取玩家列表
func (h *GameHandler) GetPlayers(c *gin.Context) {
	players, err := h.gameService.GetPlayers(c.Request.Context())
	if err != nil {
		logger.Error("get players failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, players)
}

// UpdateCurrentPlayers 更新当前牌桌玩家
func (h *GameHandler) UpdateCurrentPlayers(c *gin.Context) {
	var req model.UpdateCurrentPlayersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	players, err := h.gameService.UpdateCurrentPlayers(c.Request.Context(), &req)
	if err != nil {
		logger.Error("update current players failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, players)
}

// RecordMaJiangGame 按麻将记牌场景直接记录一局已结算对局
func (h *GameHandler) RecordMaJiangGame(c *gin.Context) {
	var req model.RecordMaJiangGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	game, err := h.gameService.RecordMaJiangGame(c.Request.Context(), &req)
	if err != nil {
		logger.Error("record majiang game failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, game)
}

// GetGamesByUser 获取个人参与的游戏列表
func (h *GameHandler) GetGamesByUser(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		response.BadRequest(c, "userId不能为空")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.BadRequest(c, "userId格式错误")
		return
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if val, parseErr := strconv.Atoi(l); parseErr == nil && val > 0 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, parseErr := strconv.Atoi(o); parseErr == nil && val >= 0 {
			offset = val
		}
	}

	games, err := h.gameService.GetGamesByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("get games by user failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, games)
}

// RegisterGameRoutes 注册游戏路由
func RegisterGameRoutes(r *gin.RouterGroup, handler *GameHandler) {
	// 游戏相关
	gameGroup := r.Group("/game")
	{
		gameGroup.POST("/record", handler.RecordMaJiangGame)
		gameGroup.POST("/cancel", handler.CancelGame)
		gameGroup.POST("/players", handler.UpdateCurrentPlayers)
		gameGroup.GET("/user/list", handler.GetGamesByUser)
		gameGroup.GET("/recent", handler.GetRecentGames)
		gameGroup.GET("/players", handler.GetPlayers)
	}
}
