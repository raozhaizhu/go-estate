package auth

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

func RegisterAuth(router *gin.RouterGroup, store db.Store, config util.Config, tokenMaker token.Maker) {
	service := auth.NewAuthService(store, config, tokenMaker)
	ctrl := NewAuthController(service, config.RefreshTokenDuration)

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", response.Wrapper(ctrl.Login))
	}
}
