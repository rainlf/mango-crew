package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mgtt-go/internal/model"
	"github.com/rainlf/mgtt-go/internal/service"
	"github.com/rainlf/mgtt-go/pkg/logger"
	"github.com/rainlf/mgtt-go/pkg/response"
)

// GameHandler 游戏处理器
type GameHandler struct {
	gameService service.GameService
}

// NewGameHandler 创建游戏处理器
func NewGameHandler(gameService service.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// GetMaJiangGames 获取麻将游戏记录列表
func (h *GameHandler) GetMaJiangGames(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	logs, err := h.gameService.GetMaJiangGameLogs(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("get majiang games failed", logger.Error(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, logs)
}

// GetMaJiangGamesByUser 获取指定用户的麻将游戏记录
func (h *GameHandler) GetMaJiangGamesByUser(c *gin.Context) {
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
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	logs, err := h.gameService.GetMaJiangGamesByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("get majiang games by user failed", logger.Error(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, logs)
}

// SaveMaJiangGame 保存麻将游戏
func (h *GameHandler) SaveMaJiangGame(c *gin.Context) {
	var req model.SaveMaJiangGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	gameID, err := h.gameService.SaveMaJiangGame(c.Request.Context(), &req)
	if err != nil {
		logger.Error("save majiang game failed", logger.Error(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gameID)
}

// DeleteMaJiangGame 删除麻将游戏
func (h *GameHandler) DeleteMaJiangGame(c *gin.Context) {
	idStr := c.Query("id")
	userIDStr := c.Query("userId")

	if idStr == "" {
		response.BadRequest(c, "id不能为空")
		return
	}
	if userIDStr == "" {
		response.BadRequest(c, "userId不能为空")
		return
	}

	gameID, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "id格式错误")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.BadRequest(c, "userId格式错误")
		return
	}

	if err := h.gameService.DeleteMaJiangGame(c.Request.Context(), gameID, userID); err != nil {
		logger.Error("delete majiang game failed", logger.Error(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetMaJiangGamePlayers 获取麻将游戏玩家列表
func (h *GameHandler) GetMaJiangGamePlayers(c *gin.Context) {
	players, err := h.gameService.GetMaJiangGamePlayers(c.Request.Context())
	if err != nil {
		logger.Error("get majiang game players failed", logger.Error(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, players)
}

// RegisterGameRoutes 注册游戏路由
func RegisterGameRoutes(r *gin.RouterGroup, handler *GameHandler) {
	majiangGroup := r.Group("/majiang")
	{
		majiangGroup.GET("/games", handler.GetMaJiangGames)
		majiangGroup.GET("/user/games", handler.GetMaJiangGamesByUser)
		majiangGroup.POST("/game", handler.SaveMaJiangGame)
		majiangGroup.DELETE("/game", handler.DeleteMaJiangGame)
		majiangGroup.GET("/game/players", handler.GetMaJiangGamePlayers)
	}
}
