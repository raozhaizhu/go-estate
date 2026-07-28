package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	"github.com/raozhaizhu/go-estate/pkg/async"
	"github.com/raozhaizhu/go-estate/pkg/logger"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

type Deps struct {
	Config        util.Config
	Store         db.Store
	Cache         cache.Cache
	ObjectStore   objectStore.StorageService
	TokenMaker    token.Maker
	Distributor   worker.TaskDistributor
	TaskProcessor worker.TaskProcessor
	AsyncRunner   async.AsyncRunner
}

func PrepareDeps(config util.Config) *Deps {
	// 初始化 Logger
	logger.InitSlogger()
	lifecycleLogger := slog.Default().With("layer", "build", "category", "lifecycle")
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
	deps := &Deps{
		Config:        config,
		Store:         store,
		Cache:         redisCache,
		ObjectStore:   minioStorage,
		TokenMaker:    tokenMaker,
		Distributor:   distributor,
		TaskProcessor: taskProcessor,
		AsyncRunner:   async.Go,
	}

	return deps
}

func PrepareTestDeps() *Deps {
	config := util.InitTestConfig()
	if config.IsProduction() {
		return nil
	}
	config.MinioEndpoint = config.MinioEndpointLocal
	fmt.Printf("config: %+v", config)

	return PrepareDeps(config)
}
