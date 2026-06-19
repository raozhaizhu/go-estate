package user

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterUser(r *gin.RouterGroup, store db.Store) {
	svc := user.NewUserService(store)
	ctrl := NewUserController(svc)

	g := r.Group("/user")
	{
		g.GET("/:username", response.Wrapper(ctrl.GetUser))
		g.POST("", response.Wrapper(ctrl.CreateNormalUser))
		g.PATCH("/:username", response.Wrapper(ctrl.UpdateUser))
	}
}
