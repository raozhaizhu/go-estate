package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/middleware"
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
	isProduction    bool
}

type Service interface {
	Login(ctx context.Context, input auth.LoginInput) (*auth.DTO, string, error)
	Refresh(ctx context.Context, refreshTokenStr string) (*auth.DTO, error)
	Logout(ctx context.Context, input auth.LogoutInput) error
}

func New(service Service, refreshDuration time.Duration, isProduction bool) *controller {
	return &controller{service: service, refreshDuration: refreshDuration, isProduction: isProduction}
}

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */

// LoginRequest 登录请求格式
type LoginRequest struct {
	Username string `uri:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=8,max=16"`
}

// toSvcInput 转换: LoginRequest -> LoginInput
func (r *LoginRequest) toSvcInput(ctx *gin.Context) auth.LoginInput {
	return auth.LoginInput{
		Username:  r.Username,
		Password:  r.Password,
		DeviceID:  ctx.GetString(middleware.CtxKeyDeviceID),
		UserAgent: ctx.GetString(middleware.CtxKeyUserAgent),
		ClientIp:  ctx.GetString(middleware.CtxKeyClientIP),
	}
}
