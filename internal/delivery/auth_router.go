package delivery

import (
	"github.com/gin-gonic/gin"
	authController "github.com/raozhaizhu/go-estate/internal/controller/auth"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterAuth(router *gin.RouterGroup, service authController.Service, config util.Config) {
	if service == nil {
		return
	}
	ctrl := authController.New(service, config.RefreshTokenDuration)

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", response.Wrapper(ctrl.Login))
	}
}
