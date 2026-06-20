package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

// TokenType 令牌类型
type TokenType byte

const (
	// TokenTypeAccessToken 访问令牌(短期使用)
	TokenTypeAccessToken = 1
	// TokenTypeRefreshToken 刷新令牌(长期使用)
	TokenTypeRefreshToken = 2
)

// Payload 令牌荷载
type Payload struct {
	TokenID   uuid.UUID `json:"id"`
	TokenType TokenType `json:"token_type"`
	Username  string    `json:"username"`
	Role      role.Role `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload 构造令牌荷载
func NewPayload(username string, role role.Role, duration time.Duration, tokenType TokenType) (*Payload, error) {
	// 获取 uuid
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	// 构造荷载
	payload := &Payload{
		TokenID:   tokenID,
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
