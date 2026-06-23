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

	// 处理路径错误, 当用户访问不存在的 api 资源时, 返回我们定义的错误格式
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.Result[any]{
			Code: appError.ErrPathNotFound.Code,
			Msg:  appError.ErrPathNotFound.Msg,
		})
	})

	// 初始化公开路由
	publicGroup := router.Group(CurrAPI)

	// 初始化受保护路由
	authMiddleware := middleware.AuthMiddleware(tokenMaker) // 初始化认证中间件
	protectedGroup := router.Group(CurrAPI)
	protectedGroup.Use(authMiddleware)

	// 挂载模块
	RegisterUser(publicGroup, protectedGroup, services.UserSvc) // user 模块可公开访问, 部分需登录
	RegisterAuth(publicGroup, services.AuthSvc, config)         // auth 模块可公开访问
	RegisterDailyData(protectedGroup, services.DailyDataSvc)    // dailyData 模块必须登录才能访问

	return router
}
