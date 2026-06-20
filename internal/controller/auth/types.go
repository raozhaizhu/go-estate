package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
)

/** ====================================================================================
 * 🏁 AuthController
 * =====================================================================================
 *
 */
type AuthController struct {
	service         AuthService
	refreshDuration time.Duration
}

type AuthService interface {
	Login(ctx *gin.Context, input auth.LoginInput) (auth.LoginDTO, string, error)
}

func NewAuthController(service AuthService, refreshDuration time.Duration) *AuthController {
	return &AuthController{service: service, refreshDuration: refreshDuration}
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
