package auth

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
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
	taskClient   *asynq.Client
}

// New 返回用户服务指针
func New(store db.AuthStore, sessionCache cache.SessionCache, config util.Config, tokenMaker token.Maker, asynqClnt *asynq.Client) *service {
	return &service{store: store, sessionCache: sessionCache, config: config, tokenMaker: tokenMaker, taskClient: asynqClnt}
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
	Username string    `json:"username"`
	Role     role.Role `json:"role"`
}

/** ====================================================================================
 * 🏁 Logout
 * =====================================================================================
 */

type LogoutInput struct {
	Username string
	DeviceID string
}

func (input *LogoutInput) toDBParams() db.GetActiveSessionIDsByUserDeviceParams {
	return db.GetActiveSessionIDsByUserDeviceParams{
		Username: input.Username,
		DeviceID: input.DeviceID,
	}
}
