package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

// 最小要求
const MinSecretSize = 32

// JWTMaker 即 Json Web Token Maker
type JWTMaker struct {
	secretKey string
}

// NewJwtMaker 返回新 JWTMaker
func NewJwtMaker(secretKey string) (Maker, error) {
	if len(secretKey) < MinSecretSize {
		return nil, appError.NewInvalidKeySizeError(len(secretKey), MinSecretSize)
	}

	return &JWTMaker{secretKey: secretKey}, nil
}

// CreateToken 返回新令牌
func (maker *JWTMaker) CreateToken(username string, role role.Role, duration time.Duration, tokenType TokenType) (string, *Payload, error) {
	// 构造荷载
	payload, err := NewPayload(username, role, duration, tokenType)
	if err != nil {
		return "", payload, err
	}
	// 构造 JWT 令牌
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload) // 声明算法 HS256, 对称加密
	token, err := jwtToken.SignedString([]byte(maker.secretKey))   // 传入 secretKey 序列化&哈希

	return token, payload, err
}

// VerifyToken 校验令牌是否合规
func (maker *JWTMaker) VerifyToken(token string, tokenType TokenType) (*Payload, error) {
	// 传入函数, 用于校验 jwt 类型是否正确(是对称加密), 正确则返回私钥
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, appError.ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	// 若 token 有效(格式正确,未过期,未篡改), 得到 *jwt.Token 数据结构
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) { // 令牌过期
			return nil, appError.ErrExpiredToken
		}
		return nil, appError.ErrInvalidToken // 令牌无效
	}

	// 得到 payload
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, appError.ErrInvalidToken // 令牌荷载无效(格式错误)
	}

	// 校验令牌类型和是否过期
	err = payload.Valid(tokenType)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
