package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/raozhaizhu/go-estate/controller"
	db "github.com/raozhaizhu/go-estate/db/sqlc"
	"github.com/raozhaizhu/go-estate/router"
	"github.com/raozhaizhu/go-estate/service"
	"github.com/raozhaizhu/go-estate/util"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()
	tryMigrateExit()
	startServer(r)

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

// 开始服务
func startServer(r *gin.Engine) {
	// 引入数据库
	cfg := util.InitConfig(".")
	store := db.InitStore(cfg.DBSource)
	// 初始化daily_data服务
	srv := service.NewDailyDataService(store)
	ctrl := controller.NewDailyDataController(srv)
	router.Setup(r, ctrl)
}

// 检测参数, 若意图为升级数据库版本, 升级后退出, 不启动 api
func tryMigrateExit() {
	cfg := util.InitConfig(".")
	runDBMigration(cfg.MigrationURL, cfg.DBSource)
}

// 执行数据库版本升级
func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatalf("无法创建migration示例")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("数据库版本合并失败")
	}

	log.Println("数据库版本合并成功")
}
