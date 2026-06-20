package auth

import (
	"context"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

/** ====================================================================================
 * 🏁 AuthService
 * =====================================================================================
 */

// AuthService 用户服务
type AuthService struct {
	store      AuthStore
	config     util.Config
	tokenMaker token.Maker
}

// AuthStore 用户数据库
type AuthStore interface {
	GetUser(ctx context.Context, username string) (db.User, error)
}

// NewAuthService 返回用户服务指针
func NewAuthService(store AuthStore, config util.Config, tokenMaker token.Maker) *AuthService {
	return &AuthService{store: store, config: config, tokenMaker: tokenMaker}
}

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */

type LoginInput struct {
	Username string
	Password string
}

type LoginDTO struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiredAt time.Time `json:"access_token_expired_at"`
	UserInfo             UserInfo  `json:"user_info"`
}

type UserInfo struct {
	Username string    `json:"username"`
	Role     role.Role `json:"role"`
}
