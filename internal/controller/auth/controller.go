package auth

import (
	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

/** ====================================================================================
 * 🏁 Post: Login
 * =====================================================================================
 */

// Login 账户登录
// Post: /api/v1/auth/login
func (c *AuthController) Login(ctx *gin.Context) (interface{}, error) {
	var req LoginRequest
	// 参数错误
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil { // 解析 Json
		return nil, response.MarkBindError(err)
	}

	// 参数转换
	input := req.toSvcInput()

	// -> svc 获得登录信息
	data, refreshToken, err := c.service.Login(ctx, input)
	if err != nil {
		return nil, err
	}

	// 将刷新令牌放入 cookie
	c.setRefreshTokenCookie(ctx, refreshToken)

	return data, nil
}

const refreshTokenKey = "refresh_token"
const loginCookiePath = "/api/vi/auth/login"

// setRefreshTokenCookie 设置 refreshToken 到 cookie
func (c *AuthController) setRefreshTokenCookie(ctx *gin.Context, refreshToken string) {
	ctx.SetCookie(
		refreshTokenKey,                  // key
		refreshToken,                     // value
		int(c.refreshDuration.Seconds()), // maxAge
		loginCookiePath,                  // path 只有在访问这个路径的时候才会发送该 cookie
		"",                               // domain 作用域(默认当前域名)
		false,                            // https 允许 http 传输
		true,                             // httpOnly 防js 窃取
	)
}
