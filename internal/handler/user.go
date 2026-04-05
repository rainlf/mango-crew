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

	response.Success(c, (&model.UserDTO{}).FromUser(user))
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

	response.Success(c, (&model.UserDTO{}).FromUser(user))
}

// UpdateUserInfo 更新用户信息（包含头像）
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userIDStr := c.PostForm("userId")
	username := c.PostForm("username")

	if userIDStr == "" {
		response.BadRequest(c, "userId不能为空")
		return
	}
	if username == "" {
		response.BadRequest(c, "username不能为空")
		return
	}
	if len(username) > 16 {
		response.BadRequest(c, "username长度不能超过16")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.BadRequest(c, "userId格式错误")
		return
	}

	// 获取头像文件
	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "avatar不能为空")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.InternalError(c, "读取头像失败")
		return
	}
	defer f.Close()

	avatar := make([]byte, file.Size)
	if _, err := f.Read(avatar); err != nil {
		response.InternalError(c, "读取头像失败")
		return
	}

	user, err := h.userService.UpdateUserInfo(c.Request.Context(), userID, username, avatar)
	if err != nil {
		logger.Error("update user info failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, (&model.UserDTO{}).FromUser(user))
}

// UpdateUsername 更新用户名
func (h *UserHandler) UpdateUsername(c *gin.Context) {
	userIDStr := c.PostForm("userId")
	username := c.PostForm("username")

	if userIDStr == "" {
		response.BadRequest(c, "userId不能为空")
		return
	}
	if username == "" {
		response.BadRequest(c, "username不能为空")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.BadRequest(c, "userId格式错误")
		return
	}

	if err := h.userService.UpdateUsername(c.Request.Context(), userID, username); err != nil {
		logger.Error("update username failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetUserRank 获取用户排名
func (h *UserHandler) GetUserRank(c *gin.Context) {
	users, err := h.userService.GetUserRank(c.Request.Context())
	if err != nil {
		logger.Error("get user rank failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	dtos := make([]*model.UserDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, (&model.UserDTO{}).FromUser(user))
	}

	response.Success(c, dtos)
}

// RegisterUserRoutes 注册用户路由
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler) {
	userGroup := r.Group("/user")
	{
		userGroup.GET("/login", handler.Login)
		userGroup.GET("/info", handler.GetUserInfo)
		userGroup.POST("/info", handler.UpdateUserInfo)
		userGroup.POST("/username", handler.UpdateUsername)
		userGroup.GET("/rank", handler.GetUserRank)
	}
}
