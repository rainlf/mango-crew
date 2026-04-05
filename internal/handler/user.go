package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/service"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/rainlf/mango-crew/pkg/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.BadRequest(c, "code不能为空")
		return
	}

	user, err := h.userService.Login(c.Request.Context(), code)
	if err != nil {
		logger.Error("login failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{
		"user_id": user.ID,
	})
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
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

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, user)
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.PostForm("userId")
	if userIDStr == "" {
		response.BadRequest(c, "userId不能为空")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.BadRequest(c, "userId格式错误")
		return
	}

	req := &model.UpdateUserRequest{
		Nickname: c.PostForm("nickname"),
		Avatar:   c.PostForm("avatar"),
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		logger.Error("update user failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, (&model.UserDTO{}).FromUser(user))
}

// GetUserRank 获取用户排名
func (h *UserHandler) GetUserRank(c *gin.Context) {
	users, err := h.userService.GetUserRank(c.Request.Context())
	if err != nil {
		logger.Error("get user rank failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, users)
}

// GetAllUsers 获取所有用户
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers(c.Request.Context())
	if err != nil {
		logger.Error("get all users failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, users)
}

// RegisterUserRoutes 注册用户路由
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler) {
	userGroup := r.Group("/user")
	{
		userGroup.GET("/login", handler.Login)
		userGroup.GET("/info", handler.GetUserInfo)
		userGroup.POST("/update", handler.UpdateUser)
		userGroup.GET("/rank", handler.GetUserRank)
		userGroup.GET("/list", handler.GetAllUsers)
	}
}
