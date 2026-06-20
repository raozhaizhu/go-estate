package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 加载配置
	cfg := util.InitConfig(".../..")

	// 初始化数据库
	store := db.InitStore(cfg.DBSource)

	// 初始化服务器
	srv, err := server.NewServer(cfg, store)
	if err != nil {
		log.Fatal("服务器初始化失败: %w", err)
	}

	// 运行服务器在指定端口
	err = srv.Start(cfg.ServerPort)
	if err != nil {
		log.Fatal("服务器运行失败: %w", err)
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
