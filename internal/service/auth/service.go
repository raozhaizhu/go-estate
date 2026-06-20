package auth

import (
	"github.com/gin-gonic/gin"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// Login 用户登录
func (svc *AuthService) Login(ctx *gin.Context, input LoginInput) (LoginDTO, string, error) {
	// 查询用户
	user, err := svc.store.GetUser(ctx, input.Username)
	if err != nil { // 用户不存在
		return LoginDTO{}, "", appError.ErrWrongUsernamePassword
	}

	// 校对密码
	err = util.CheckPassword(input.Password, user.HashedPassword)
	if err != nil { // 哈希错误
		return LoginDTO{}, "", appError.ErrWrongUsernamePassword

	}

	// 发放访问令牌
	accessToken, accessPayload, err := svc.tokenMaker.CreateToken(user.Username, role.Role(user.Role),
		svc.config.AccessTokenDuration, token.TokenTypeAccessToken)
	if err != nil {
		return LoginDTO{}, "", err

	}

	// 发放刷新令牌
	refreshToken, _, err := svc.tokenMaker.CreateToken(user.Username, role.Role(user.Role),
		svc.config.RefreshTokenDuration, token.TokenTypeRefreshToken)
	if err != nil {
		return LoginDTO{}, "", err

	}

	// 返回 DTO
	return LoginDTO{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: accessPayload.ExpiredAt,
		UserInfo: UserInfo{
			Username: user.Username,
			Role:     role.Role(user.Role),
		},
	}, refreshToken, nil
}
