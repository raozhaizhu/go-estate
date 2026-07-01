package token

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

// TokenType 令牌类型
type TokenType byte

const (
	// TokenTypeAccessToken 访问令牌(短期使用)
	TokenTypeAccessToken = 1
	// TokenTypeRefreshToken 刷新令牌(长期使用)
	TokenTypeRefreshToken = 2

	// PayloadKey SetKey, 用于从 Context 中提取 payload
	PayloadKey = "authorization_payload"
	// RefreshTokenKey 用于从 Cookie 中提取 refresh_token
	RefreshTokenKey = "refresh_token"
)

// Payload 令牌荷载
type Payload struct {
	ID        uuid.UUID `json:"id"`
	TokenType TokenType `json:"token_type"`
	Username  string    `json:"username"`
	Role      role.Role `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (p *Payload) ToDBParams(userAgent, clientIp, deviceID string) db.CreateSessionParams {
	params := db.CreateSessionParams{
		ID:        p.ID.String(),
		Username:  p.Username,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		ClientIp:  clientIp,
		IsBlocked: false,
		ExpiresAt: p.ExpiredAt,
	}

	return params
}

func (p *Payload) ToCacheParams() cache.AddNewSessionParams {
	params := cache.AddNewSessionParams{
		JTI: p.ID.String(),
		Session: cache.Session{
			Username:  p.Username,
			IsBlocked: false,
			ExpiresAt: p.ExpiredAt,
		},
	}

	return params
}

/** ====================================================================================
 * 🏁 Methods
 * =====================================================================================
 */

// NewPayload 构造令牌荷载
func NewPayload(username string, role role.Role, duration time.Duration, tokenType TokenType) (*Payload, error) {
	// 获取 uuid
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	// 构造荷载
	payload := &Payload{
		ID:        tokenID,
		TokenType: tokenType,
		Username:  username,
		Role:      role,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Valid 校验令牌荷载可用性
func (p *Payload) Valid(tokenType TokenType) error {
	// 校验令牌类型一致
	if p.TokenType != tokenType {
		return appError.ErrInvalidToken
	}
	// 校验令牌过期
	if time.Now().After(p.ExpiredAt) {
		return appError.ErrExpiredToken
	}

	return nil
}

// GetExpirationTime 返回令牌过期时间
func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: p.ExpiredAt,
	}, nil
}

// GetIssuedAt 返回令牌颁发时间
func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: p.IssuedAt,
	}, nil
}

// GetNotBefore 返回令牌 NBF(不可早于) 时间
func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: p.IssuedAt,
	}, nil
}

// GetIssuer 返回令牌颁发者
func (p *Payload) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject 返回令牌接收者
func (p *Payload) GetSubject() (string, error) {
	return "", nil
}

// GetAudience 返回令牌受众类型
func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

// GetPayload 从上下文中获取荷载
func GetPayload(ctx context.Context) (*Payload, error) {
	// 提取 payload
	val := ctx.Value(PayloadKey)

	// 提取失败, 返回错误
	if val == nil {
		return nil, appError.ErrAuthRequired
	}
	payload := val.(*Payload)

	// 返回 payload
	return payload, nil
}

func WithPayload(ctx context.Context, payload *Payload) context.Context {
	return context.WithValue(ctx, PayloadKey, payload)
}
