package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/controller/auth"
	dailyData "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	"github.com/raozhaizhu/go-estate/internal/controller/user"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

func SetupRouter(store db.Store, config util.Config, tokenMaker token.Maker) *gin.Engine {
	// 初始化路由引擎
	r := gin.New()

	// 挂载全局中间件
	r.Use(gin.Recovery()) // 防崩溃
	r.Use(gin.Logger())   // 日志记录

	r.GET("/ping", func(c *gin.Context) { // 挂载探针路由
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 定义全局版本路由
	v1Api := "/api/v1"

	// 初始化公开路由
	publicGroup := r.Group(v1Api)

	// 初始化受保护路由
	authMiddleware := middleware.AuthMiddleware(tokenMaker) // 初始化认证中间件
	protectedGroup := r.Group(v1Api)
	protectedGroup.Use(authMiddleware)

	// 挂载模块
	dailyData.RegisterDailyData(protectedGroup, store)        // dailyData 模块必须登录才能访问
	user.RegisterUser(protectedGroup, store)                  // user 模块必须登录才能访问
	auth.RegisterAuth(publicGroup, store, config, tokenMaker) // auth 模块可公开访问(需要登录)

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

	return r
}
