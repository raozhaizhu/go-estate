package auth

import (
	"context"
	"log/slog"
	"time"

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
	logger          *slog.Logger
}

type Service interface {
	Login(ctx context.Context, input auth.LoginInput) (*auth.DTO, string, error)
	Refresh(ctx context.Context, refreshTokenStr string) (*auth.DTO, error)
	Logout(ctx context.Context, input auth.LogoutInput) error
}

func New(service Service, refreshDuration time.Duration, isProduction bool) *controller {
	logger := slog.Default().With("layer", "controller", "module", "auth_controller")
	return &controller{service: service, refreshDuration: refreshDuration, isProduction: isProduction, logger: logger}
}

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */

// LoginRequest 登录请求格式
type LoginRequest struct {
	Username string `uri:"username" binding:"required,min=3" example:"Bob"`
	Password string `json:"password" binding:"required,min=8,max=16" example:"12345678"`
}

// toSvcInput 转换: LoginRequest -> LoginInput
func (r *LoginRequest) toSvcInput(meta middleware.ClientMeta) auth.LoginInput {
	return auth.LoginInput{
		Username:  r.Username,
		Password:  r.Password,
		DeviceID:  meta.DeviceID,
		UserAgent: meta.UserAgent,
		ClientIp:  meta.ClientIP,
	}
}
