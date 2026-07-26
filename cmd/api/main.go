package main

import (
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	_ "github.com/raozhaizhu/go-estate/docs"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	"github.com/raozhaizhu/go-estate/pkg/async"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/raozhaizhu/go-estate/pkg/token"

	_ "github.com/go-sql-driver/mysql"
)

// @title           Go Estate API 文档
// @version         1.0
// @description     房地产项目的后端 API 接口文档
// @host            localhost
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey DeviceIDAuth
// @in                         header
// @name                       X-Device-ID

// @securityDefinitions.apikey UserAgentAuth
// @in                         header
// @name                       User-Agent
func main() {
	deps := prepareDeps()
	lifecycleLogger := deps.Logger.With(slog.String("category", "lifecycle"))

	// 初始化服务器
	srv, err := server.NewServer(deps)

	if err != nil {
		lifecycleLogger.Error("服务器初始化失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 运行服务器在指定端口
	err = srv.Start(":" + deps.Config.ServerPort)
	if err != nil {
		lifecycleLogger.Error("服务器运行失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func prepareDeps() app.Deps {
	// 加载配置
	config := util.InitConfig(".")
	// 初始化 Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	lifecycleLogger := logger.With(slog.String("category", "lifecycle"))
	// 初始化数据库
	store, err := db.InitStore(config.DBSource)
	if err != nil {
		lifecycleLogger.Error("数据库初始化失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// 初始化 redis
	redisCache, err := cache.NewCache(config.RedisAddress, config.RedisPassword)
	if err != nil {
		lifecycleLogger.Error("Redis 缓存初始化失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// 初始化 minIO 客户端
	minioStorage, err := objectStore.SetupMinIO(config.MinioEndpoint, config.MinioAccessKeyID, config.MinioSecretAccessKey)
	if err != nil {
		lifecycleLogger.Error("MinioClient 初始化失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// 启动后台工作服务器
	opt := asynq.RedisClientOpt{Addr: config.RedisAddress, Password: config.RedisPassword, DB: 0}
	distributor := worker.NewRedisTaskDistributor(opt)
	taskProcessor := worker.NewRedisTaskProcessor(opt, redisCache, store)

	// 初始化 TokenMaker
	tokenMaker, err := token.NewJwtMaker(config.TokenSymmetricKey)
	if err != nil {
		lifecycleLogger.Error("tokenMaker初始化失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// 准备好所有依赖
	deps := app.Deps{
		Config:        config,
		Store:         store,
		Cache:         redisCache,
		ObjectStore:   minioStorage,
		TokenMaker:    tokenMaker,
		Distributor:   distributor,
		TaskProcessor: taskProcessor,
		Logger:        logger,
		AsyncRunner:   async.Go,
	}

	return deps
}
