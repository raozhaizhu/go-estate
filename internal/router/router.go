package router

import (
	"log"

	"github.com/gin-gonic/gin"
	dailyData "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	"github.com/raozhaizhu/go-estate/internal/controller/user"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
)

func SetupRouter(store db.Store) *gin.Engine {
	// 初始化路由引擎
	r := gin.New()

	// 挂载全局中间件
	r.Use(gin.Recovery()) // 防崩溃
	r.Use(gin.Logger())   // 日志记录

	r.GET("/ping", func(c *gin.Context) { // 挂载探针路由
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 定义全局版本路由
	v1 := r.Group("/api/v1")

	// 挂载模块
	dailyData.RegisterDailyData(v1, store)
	user.RegisterUser(v1, store)

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

	return r
}
