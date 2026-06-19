package user

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/service/user"
)

func RegisterUser(r *gin.RouterGroup, store db.Store) {
	svc := user.NewUserService(store)
	ctrl := NewUserController(svc)

	g := r.Group("/user")
	{
		g.GET("/:username", ctrl.GetUser)
		g.POST("", ctrl.CreateNormalUser)
		g.PATCH("/:username", ctrl.UpdateUser)
	}
}
