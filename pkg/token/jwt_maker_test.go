package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/stretchr/testify/require"
)

// TestJWTMakerBasic
// 基础测试: JWTMaker 能正常制造 token 和校验 token
func TestJWTMakerBasic(t *testing.T) {
	// 初始化 JWTMaker
	maker, err := NewJwtMaker(util.RandomString(32))
	require.NoError(t, err)

	// 设置 Payload 参数
	username := util.RandomUsername()
	role := role.RoleUser
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	// 得到 token,payload
	token, payload, err := maker.CreateToken(username, role, duration, TokenTypeAccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	// 校验 token 合规
	payload, err = maker.VerifyToken(token, TokenTypeAccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// 校验 payload 参数一致(或时间误差在允许范围内)
	require.NotZero(t, payload.TokenID)
	require.Equal(t, username, payload.Username)
	require.Equal(t, role, payload.Role)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

// TestExpiredToken
// 超时测试: JWTMaker 能发现令牌超时
func TestExpiredJWTToken(t *testing.T) {
	// 初始化 JWTMaker
	maker, err := NewJwtMaker(util.RandomString(32))
	require.NoError(t, err)

	// 设置 Payload 参数
	username := util.RandomUsername()
	role := role.RoleUser
	duration := -time.Minute

	// 得到 token,payload <- 哪怕超时也能正常铸造令牌
	token, payload, err := maker.CreateToken(username, role, duration, TokenTypeAccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	// 校验 payload, 因超时而失败
	payload, err = maker.VerifyToken(token, TokenTypeAccessToken)
	require.Error(t, err)
	require.EqualError(t, err, appError.ErrExpiredToken.Error())
	require.Nil(t, payload)
}

// TestMaliciousNoneJWTToken
// 安全测试: JWTMaker 能发现 Token 使用了错误或恶意的算法
func TestMaliciousNoneJWTToken(t *testing.T) {
	// 设置 Payload 参数
	username := util.RandomUsername()
	role := role.RoleUser
	duration := time.Minute

	// 初始化 Payload
	payload, err := NewPayload(username, role, duration, TokenTypeAccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	// 创造恶意 Token(基于 None 安全度)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	noneToken, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	require.NotEmpty(t, noneToken)

	// 初始化 JWTMaker
	maker, err := NewJwtMaker(util.RandomString(32))
	require.NoError(t, err)

	// 尝试使用恶意 NoneToken 通过 maker 校验
	payload, err = maker.VerifyToken(noneToken, TokenTypeAccessToken)
	require.Error(t, err)
	require.EqualError(t, err, appError.ErrInvalidToken.Error())
	require.Nil(t, payload)
}

// TestWrongTypeJWTToken
// 类型测试: JWTMaker 能发现 Token 类型错误
func TestWrongTypeJWTToken(t *testing.T) {
	// 初始化 JWTMaker
	maker, err := NewJwtMaker(util.RandomString(32))
	require.NoError(t, err)

	// 设置 Payload 参数
	username := util.RandomUsername()
	role := role.RoleUser
	duration := time.Minute

	// 得到 token,payload
	token, payload, err := maker.CreateToken(username, role, duration, TokenTypeAccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	// 校验 payload, 因类型错误而失败
	payload, err = maker.VerifyToken(token, TokenTypeRefreshToken)
	require.Error(t, err)
	require.EqualError(t, err, appError.ErrInvalidToken.Error())
	require.Nil(t, payload)
}

// TestWrongSizeKeyJWTToken
// 密钥长度测试: JWTMaker 能发现密钥长度不足(最小 32 位)
func TestWrongSizeKeyJWTToken(t *testing.T) {
	// 初始化 JWTMaker
	wrongSize := 31
	maker, err := NewJwtMaker(util.RandomString(wrongSize))
	require.Error(t, err)
	require.EqualError(t, err, appError.NewInvalidKeySizeError(wrongSize, MinSecretSize).Error())
	require.Nil(t, maker)
}
