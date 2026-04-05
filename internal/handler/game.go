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

// CreateSession 创建场次
func (h *GameHandler) CreateSession(c *gin.Context) {
	var req model.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 从上下文获取当前用户ID
	userID := getCurrentUserID(c)

	session, err := h.gameService.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("create session failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, session)
}

// EndSession 结束场次
func (h *GameHandler) EndSession(c *gin.Context) {
	sessionIDStr := c.PostForm("sessionId")
	if sessionIDStr == "" {
		response.BadRequest(c, "sessionId不能为空")
		return
	}

	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		response.BadRequest(c, "sessionId格式错误")
		return
	}

	if err := h.gameService.EndSession(c.Request.Context(), sessionID); err != nil {
		logger.Error("end session failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetSessions 获取场次列表
func (h *GameHandler) GetSessions(c *gin.Context) {
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

	sessions, err := h.gameService.GetSessions(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("get sessions failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, sessions)
}

// GetActiveSessions 获取进行中的场次
func (h *GameHandler) GetActiveSessions(c *gin.Context) {
	sessions, err := h.gameService.GetActiveSessions(c.Request.Context())
	if err != nil {
		logger.Error("get active sessions failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, sessions)
}

// CreateGame 创建游戏
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req model.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := getCurrentUserID(c)

	game, err := h.gameService.CreateGame(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("create game failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, game)
}

// SettleGame 结算游戏
func (h *GameHandler) SettleGame(c *gin.Context) {
	var req model.SettleGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.gameService.SettleGame(c.Request.Context(), req.GameID); err != nil {
		logger.Error("settle game failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
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

// GetGamesBySession 获取场次的游戏列表
func (h *GameHandler) GetGamesBySession(c *gin.Context) {
	sessionIDStr := c.Query("sessionId")
	if sessionIDStr == "" {
		response.BadRequest(c, "sessionId不能为空")
		return
	}

	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		response.BadRequest(c, "sessionId格式错误")
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

	games, err := h.gameService.GetGamesBySession(c.Request.Context(), sessionID, limit, offset)
	if err != nil {
		logger.Error("get games by session failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, games)
}

// GetRecentGames 获取最近的游戏列表
func (h *GameHandler) GetRecentGames(c *gin.Context) {
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

// RegisterGameRoutes 注册游戏路由
func RegisterGameRoutes(r *gin.RouterGroup, handler *GameHandler) {
	// 场次相关
	sessionGroup := r.Group("/session")
	{
		sessionGroup.POST("", handler.CreateSession)
		sessionGroup.POST("/end", handler.EndSession)
		sessionGroup.GET("/list", handler.GetSessions)
		sessionGroup.GET("/active", handler.GetActiveSessions)
	}

	// 游戏相关
	gameGroup := r.Group("/game")
	{
		gameGroup.POST("", handler.CreateGame)
		gameGroup.POST("/settle", handler.SettleGame)
		gameGroup.POST("/cancel", handler.CancelGame)
		gameGroup.GET("/list", handler.GetGamesBySession)
		gameGroup.GET("/recent", handler.GetRecentGames)
		gameGroup.GET("/players", handler.GetPlayers)
	}
}

// getCurrentUserID 从上下文获取当前用户ID
// 实际应该从JWT或Session中获取，这里简化处理
func getCurrentUserID(c *gin.Context) int {
	// 从查询参数或Header中获取
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		userIDStr = c.Query("userId")
	}
	if userIDStr == "" {
		return 0
	}
	userID, _ := strconv.Atoi(userIDStr)
	return userID
}
