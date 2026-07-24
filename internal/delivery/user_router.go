package delivery

import (
	"time"

	"github.com/gin-gonic/gin"
	user "github.com/raozhaizhu/go-estate/internal/controller/user"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

const (
	UserApi = "/api/v1/user"
)

// RegisterUser
func RegisterUser(metaGroup *gin.RouterGroup, authGroup *gin.RouterGroup, service user.Service, config util.Config, redisCache cache.Cache) {
	if service == nil {
		return
	}

	controller := user.New(service)

	// 定义公共路由
	userPublicGroup := metaGroup.Group("/user")
	// 挂载限流中间件
	userPublicGroup.Use(middleware.RateLimiter(redisCache, 5, time.Minute))

	{
		userPublicGroup.POST("", response.Wrapper(controller.CreateNormalUser))
	}

	// 定义保护路由
	userProtectedGroup := authGroup.Group("/user")
	{
		userProtectedGroup.POST("/vip", middleware.RequireRoles(userDomain.RoleAtLeastAdmin), response.Wrapper(controller.CreateVip))
		userProtectedGroup.GET("/:username", response.Wrapper(controller.GetUser))
		userProtectedGroup.PATCH("/:username", response.Wrapper(controller.UpdateUser))
	}
}
