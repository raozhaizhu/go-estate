package delivery

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/controller/auth"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterAuth(metaGroup *gin.RouterGroup, authGroup *gin.RouterGroup, service auth.Service, config util.Config, redisCache cache.Cache) {
	if service == nil {
		return
	}
	ctrl := auth.New(service, config.RefreshTokenDuration, config.IsProduction())

	// 定义公共路由
	authPublicGroup := metaGroup.Group("/auth")
	// 挂载限流中间件
	authPublicGroup.Use(middleware.RateLimiter(redisCache, 5, time.Minute))

	{
		authPublicGroup.POST("/login", response.Wrapper(ctrl.Login))
		authPublicGroup.POST("/refresh", response.Wrapper(ctrl.Refresh))
	}

	// 定义保护路由
	authProtectedGroup := authGroup.Group("/auth")
	{
		authProtectedGroup.POST("/logout", response.Wrapper(ctrl.Logout))
	}
}
