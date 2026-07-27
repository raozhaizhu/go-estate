package auth

import (
	"log/slog"
	"time"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	"github.com/raozhaizhu/go-estate/pkg/async"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

/** ====================================================================================
 * 🏁 AuthService
 * =====================================================================================
 */

// service 用户服务
type service struct {
	store        db.AuthStore
	sessionCache cache.SessionCache
	config       util.Config
	tokenMaker   token.Maker
	distributor  worker.TaskDistributor
	logger       *slog.Logger
	asyncRunner  async.AsyncRunner
}

// New 返回用户服务指针
func New(deps *app.Deps) *service {
	return &service{store: deps.Store, sessionCache: deps.Cache, config: deps.Config, tokenMaker: deps.TokenMaker, distributor: deps.Distributor, logger: deps.Logger, asyncRunner: deps.AsyncRunner}
}

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */

type LoginInput struct {
	Username  string
	Password  string
	DeviceID  string
	UserAgent string
	ClientIp  string
}

type DTO struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiredAt time.Time `json:"access_token_expired_at"`
	UserInfo             UserInfo  `json:"user_info"`
}

type UserInfo struct {
	Username  string    `json:"username" example:"Bob"`
	Role      role.Role `json:"role" example:"1"`
	AvatarKey string    `json:"avatar_key" example:"default_avatar.png"`
}

/** ====================================================================================
 * 🏁 Logout
 * =====================================================================================
 */

type LogoutInput struct {
	Username string
	DeviceID string
}

func (input *LogoutInput) ToDBParams() db.GetActiveSessionIDsByUserDeviceForUpdateParams {
	return db.GetActiveSessionIDsByUserDeviceForUpdateParams{
		Username: input.Username,
		DeviceID: input.DeviceID,
	}
}
