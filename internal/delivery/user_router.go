package delivery

import (
	"github.com/gin-gonic/gin"
	user "github.com/raozhaizhu/go-estate/internal/controller/user"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

const (
	UserApi = "/api/v1/user"
)

// RegisterUser
func RegisterUser(metaGroup *gin.RouterGroup, authGroup *gin.RouterGroup, service user.Service) {
	if service == nil {
		return
	}

	controller := user.New(service)

	// 调用路由注册函数
	RegisterUserRoutes(metaGroup, authGroup, controller)
}

// RegisterUserRoutes
func RegisterUserRoutes(publicGroup *gin.RouterGroup, protectedGroup *gin.RouterGroup, controller *user.Controller) {
	// 创建普通用户可公开访问
	userPublicGroup := publicGroup.Group("/user")
	{
		userPublicGroup.POST("", response.Wrapper(controller.CreateNormalUser))
	}

	// 创建 vip 用户, 查询/更新用户信息, 需要身份验证
	userProtectedGroup := protectedGroup.Group("/user")
	{
		userProtectedGroup.POST("/vip", middleware.RequireRoles(userDomain.RoleAtLeastAdmin), response.Wrapper(controller.CreateVip))
		userProtectedGroup.GET("/:username", response.Wrapper(controller.GetUser))
		userProtectedGroup.PATCH("/:username", response.Wrapper(controller.UpdateUser))
	}
}
