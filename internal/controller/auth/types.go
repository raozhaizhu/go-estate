package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
)

/** ====================================================================================
 * 🏁 controller
 * =====================================================================================
 *
 */
type controller struct {
	service         Service
	refreshDuration time.Duration
}

type Service interface {
	Login(ctx *gin.Context, input auth.LoginInput) (auth.LoginDTO, string, error)
}

func New(service Service, refreshDuration time.Duration) *controller {
	return &controller{service: service, refreshDuration: refreshDuration}
}

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */
type LoginRequest struct {
	Username string  `uri:"username" binding:"required,min=1"`
	Password *string `json:"password" binding:"omitempty,min=8,max=16"`
}

func (r *LoginRequest) toSvcInput() auth.LoginInput {
	return auth.LoginInput{
		Username: r.Username,
		Password: *r.Password,
	}
}
