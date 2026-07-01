package delivery

import (
	"github.com/gin-gonic/gin"
	authController "github.com/raozhaizhu/go-estate/internal/controller/auth"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterAuth(metaGroup *gin.RouterGroup, authGroup *gin.RouterGroup, service authController.Service, config util.Config) {
	if service == nil {
		return
	}
	ctrl := authController.New(service, config.RefreshTokenDuration)

	authPublicGroup := metaGroup.Group("/auth")
	{
		authPublicGroup.POST("/login", response.Wrapper(ctrl.Login))
		authPublicGroup.POST("/refresh", response.Wrapper(ctrl.Refresh))
	}

	authProtectedGroup := authGroup.Group("/auth")
	{
		authProtectedGroup.POST("/logout", response.Wrapper(ctrl.Logout))
	}
}
