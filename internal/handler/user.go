package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/service"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/rainlf/mango-crew/pkg/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
	uploadDir   string
	publicPath  string
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService, uploadDir, publicPath string) *UserHandler {
	return &UserHandler{
		userService: userService,
		uploadDir:   uploadDir,
		publicPath:  publicPath,
	}
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
	}

	// 处理头像文件上传
	file, err := c.FormFile("avatar")
	if err != nil {
		logger.Info("no avatar file in request", logger.String("err", err.Error()))
	}
	if err == nil && file != nil {
		ext := filepath.Ext(file.Filename)
		if ext == "" {
			ext = ".png"
		}
		filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixMilli(), ext)

		// 头像直接保存在配置的根目录下。
		uploadDir := h.uploadDir
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			logger.Error("create upload dir failed", logger.Err(err))
			response.Error(c, 1, "头像保存失败")
			return
		}

		savePath := filepath.Join(uploadDir, filename)
		logger.Info("saving avatar", logger.String("path", savePath), logger.String("filename", file.Filename))
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			logger.Error("save avatar failed", logger.Err(err))
			response.Error(c, 1, "头像保存失败")
			return
		}
		req.Avatar = strings.TrimRight(h.publicPath, "/") + "/" + filename
		logger.Info("avatar saved", logger.String("url", req.Avatar))
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

// GetFitnessRank 获取健身榜
func (h *UserHandler) GetFitnessRank(c *gin.Context) {
	users, err := h.userService.GetFitnessRank(c.Request.Context())
	if err != nil {
		logger.Error("get fitness rank failed", logger.Err(err))
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

// RebuildUserStats 重建用户统计信息
func (h *UserHandler) RebuildUserStats(c *gin.Context) {
	var req model.RebuildUserStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	count, err := h.userService.RebuildUserStats(c.Request.Context(), req.UserIDs)
	if err != nil {
		logger.Error("rebuild user stats failed", logger.Err(err))
		response.Error(c, 1, err.Error())
		return
	}

	response.Success(c, gin.H{
		"count": count,
	})
}

// RegisterUserRoutes 注册用户路由
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler) {
	userGroup := r.Group("/user")
	{
		userGroup.GET("/login", handler.Login)
		userGroup.GET("/info", handler.GetUserInfo)
		userGroup.POST("/update", handler.UpdateUser)
		userGroup.GET("/rank", handler.GetUserRank)
		userGroup.GET("/fitness/rank", handler.GetFitnessRank)
		userGroup.GET("/list", handler.GetAllUsers)
		userGroup.POST("/stats/rebuild", handler.RebuildUserStats)
	}
}
