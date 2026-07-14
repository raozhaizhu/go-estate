package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/hibiken/asynq"
	_ "github.com/raozhaizhu/go-estate/docs"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	"github.com/raozhaizhu/go-estate/pkg/async"
	"github.com/raozhaizhu/go-estate/pkg/token"

	_ "github.com/go-sql-driver/mysql"
)

// @title           Go Estate API 文档
// @version         1.0
// @description     房地产项目的后端 API 接口文档
// @host            localhost:8080
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
	// 初始化服务器
	srv, err := server.NewServer(deps)
	if err != nil {
		log.Fatal("服务器初始化失败: %w", err)
		os.Exit(1)
	}

	// 运行服务器在指定端口
	err = srv.Start(":" + deps.Config.ServerPort)
	if err != nil {
		log.Fatal("服务器运行失败: %w", err)
		os.Exit(1)
	}
}

func prepareDeps() app.Deps {
	// 加载配置
	config := util.InitConfig(".")
	// 初始化数据库
	store := db.InitStore(config.DBSource)
	// 初始化 redis
	redisCache := cache.NewSessionCache(config.RedisAddress, config.RedisPassword)
	// 启动后台工作服务器
	opt := asynq.RedisClientOpt{Addr: config.RedisAddress, Password: config.RedisPassword, DB: 0}
	distributor := worker.NewRedisTaskDistributor(opt)
	go goRunTaskProcessor(redisCache, opt)
	// 初始化 Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	// 初始化 TokenMaker
	tokenMaker, err := token.NewJwtMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("tokenMaker初始化失败: %w", err)
	}
	// 准备好所有依赖
	deps := app.Deps{
		Config:      config,
		Store:       store,
		Cache:       redisCache,
		TokenMaker:  tokenMaker,
		Distributor: distributor,
		Logger:      logger,
		AsyncRunner: async.Go,
	}

	return deps
}

func goRunTaskProcessor(redisCache cache.Cache, opt asynq.RedisClientOpt) {
	taskProcessor := worker.NewRedisTaskProcessor(redisCache)
	taskServer := asynq.NewServer(opt, asynq.Config{Concurrency: 10})

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TaskDeleteSessions, taskProcessor.HandleDeleteSessionsTask)

	log.Println("开启启动 Asynq Worker 服务器")
	if err := taskServer.Run(mux); err != nil {
		log.Fatal("Asynq Worker 启动失败: ", err)
	}
}

// tryMigrateExit 执行数据库版本升级
func tryMigrateExit(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatalf("无法创建migration示例")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("数据库版本合并失败")
	}

	log.Println("数据库版本合并成功")
}
