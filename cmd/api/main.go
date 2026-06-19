package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	dailyData "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	"github.com/raozhaizhu/go-estate/internal/controller/user"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/pkg/validator"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 加载配置
	cfg := util.InitConfig(".../..")
	// 初始化数据库
	store := db.InitStore(cfg.DBSource)
	// 初始化路由引擎
	r := gin.Default()
	// 初始化翻译器
	validator.InitTrans()
	// 定义全局版本路由
	v1 := r.Group("/api/v1")
	// 挂载模块
	dailyData.RegisterDailyData(v1, store)
	user.RegisterUser(v1, store)

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

// 执行数据库版本升级
// tryMigrateExit(cfg.MigrationURL, cfg.DBSource)
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
