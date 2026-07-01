package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/controller/auth"
	dailyData "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	userController "github.com/raozhaizhu/go-estate/internal/controller/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// 定义全局版本路由
const CurrAPI = "/api/v1"

type Services struct {
	UserSvc      userController.Service
	AuthSvc      auth.Service
	DailyDataSvc dailyData.Service
}

func SetupRouter(services Services, config util.Config, tokenMaker token.Maker) *gin.Engine {
	// 初始化路由引擎
	router := gin.New()

	// 挂载全局中间件
	router.Use(gin.Recovery()) // 防崩溃
	router.Use(gin.Logger())   // 日志记录

	// 挂载探针路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 处理路径错误, 当用户访问不存在的 api 资源时, 返回自定义的错误格式(而不是直接 404, 不带 Code)
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.Result[any]{
			Code: appError.ErrPathNotFound.Code,
			Msg:  appError.ErrPathNotFound.Msg,
		})
	})

	// 基础路由
	api := router.Group(CurrAPI)

	// 元数据校验组 : 目前所有接口都需要元数据
	metaGroup := api.Group("/")
	metaGroup.Use(middleware.RequireMetadata())

	// 受保护组
	authGroup := metaGroup.Group("/")
	authGroup.Use(middleware.AuthMiddleware(tokenMaker))

	// 挂载模块
	RegisterUser(metaGroup, authGroup, services.UserSvc)         // user模块 部分需登录
	RegisterAuth(metaGroup, authGroup, services.AuthSvc, config) // auth模块 部分需登录
	RegisterDailyData(authGroup, services.DailyDataSvc)          // dailyData模块 必须登录

	return router
}
