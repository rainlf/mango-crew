package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/internal/config"
	"github.com/rainlf/mango-crew/internal/handler"
	"github.com/rainlf/mango-crew/internal/middleware"
	"github.com/rainlf/mango-crew/internal/model"
	"github.com/rainlf/mango-crew/internal/repository"
	"github.com/rainlf/mango-crew/internal/service"
	"github.com/rainlf/mango-crew/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	// 获取配置文件路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load config from %s: %v\n", configPath, err)
		os.Exit(1)
	}

	fmt.Printf("✅ 配置文件加载成功: %s\n", configPath)

	wechatAppID := os.Getenv("WECHAT_APP_ID")
	wechatAppSecret := os.Getenv("WECHAT_APP_SECRET")
	if wechatAppID == "" || wechatAppSecret == "" {
		fmt.Println("WECHAT_APP_ID and WECHAT_APP_SECRET must be set")
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting Mango Crew server...")

	// 连接数据库
	db, err := initDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect database", logger.Err(err))
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		logger.Fatal("Failed to migrate database", logger.Err(err))
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewGameSessionRepository(db)
	gameRepo := repository.NewGameRepository(db)

	// 初始化服务
	userService := service.NewUserService(userRepo, gameRepo, cfg.Wechat, wechatAppID, wechatAppSecret)
	gameService := service.NewGameService(sessionRepo, gameRepo, userRepo)

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService, cfg.Storage.UploadDir)
	gameHandler := handler.NewGameHandler(gameService)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 静态文件服务（头像等上传文件）
	r.Static("/uploads", cfg.Storage.UploadDir)

	// 注册路由
	api := r.Group("/api")
	{
		handler.RegisterHealthRoutes(api)
		handler.RegisterUserRoutes(api, userHandler)
		handler.RegisterGameRoutes(api, gameHandler)
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 启动服务器
	go func() {
		logger.Info("Server is running", logger.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Err(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", logger.Err(err))
	}

	logger.Info("Server exited")
}

func initDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.GameSession{},
		&model.SessionPlayer{},
		&model.Game{},
		&model.GamePlayer{},
		&model.GamePlayerWinType{},
		&model.WinType{},
	)
}
