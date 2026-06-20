package user

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterUser(protectedGroup *gin.RouterGroup, store db.Store) {
	svc := user.NewUserService(store)
	ctrl := NewUserController(svc)

	userGroup := protectedGroup.Group("/user")
	{
		// User/Vip: 只能查本人数据; Admin: 可以查所有人数据
		userGroup.GET("/:username", response.Wrapper(ctrl.GetUser))
		// User: 任何人都可以创建; Vip: 只有 Admin 可以创建; Admin: 任何人都不得创建
		userGroup.POST("", response.Wrapper(ctrl.CreateNormalUser))
		// User/Vip: 只能更新本人数据; Admin: 可以更新所有人数据
		userGroup.PATCH("/:username", response.Wrapper(ctrl.UpdateUser))
	}
}
