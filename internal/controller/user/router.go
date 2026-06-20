package user

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterUser(publicGroup *gin.RouterGroup, protectedGroup *gin.RouterGroup, store db.Store) {
	svc := user.NewUserService(store)
	ctrl := NewUserController(svc)

	// 创建普通用户可公开访问
	userPublicGroup := publicGroup.Group("/user")
	{
		userPublicGroup.POST("", response.Wrapper(ctrl.CreateNormalUser))
	}

	// 创建 vip 用户, 查询/更新信息, 需要登录访问
	userProtectedGroup := protectedGroup.Group("/user")
	{
		userProtectedGroup.POST("/vip", response.Wrapper(ctrl.CreateVip))
		userProtectedGroup.GET("/:username", response.Wrapper(ctrl.GetUser))
		userProtectedGroup.PATCH("/:username", response.Wrapper(ctrl.UpdateUser))
	}

}
