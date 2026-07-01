package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/hibiken/asynq"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 加载配置
	config := util.InitConfig(".../..")

	// 初始化数据库
	store := db.InitStore(config.DBSource)

	// 初始化 redis
	redisCache := cache.NewSessionCache(config.RedisAddress, config.RedisPassword)

	// 启动后台工作服务器
	opt := asynq.RedisClientOpt{Addr: config.RedisAddress, Password: config.RedisPassword, DB: 0}
	asynqClnt := asynq.NewClient(opt)
	go goRunTaskProcessor(redisCache, opt)

	// 初始化服务器
	srv, err := server.NewServer(config, store, redisCache, asynqClnt)
	if err != nil {
		log.Fatal("服务器初始化失败: %w", err)
	}

	// 运行服务器在指定端口
	err = srv.Start(":" + config.ServerPort)
	if err != nil {
		log.Fatal("服务器运行失败: %w", err)
	}
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
