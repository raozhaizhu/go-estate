package auth

import (
	"log"

	"github.com/gin-gonic/gin"
	domain "github.com/raozhaizhu/go-estate/internal/domain/auth"
	"github.com/raozhaizhu/go-estate/internal/service/auth"

	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

/** ====================================================================================
 * 🏁 Post: Login
 * =====================================================================================
 */

// Login 账户登录
// Post: /api/v1/auth/login
func (ctrl *controller) Login(ctx *gin.Context) (interface{}, error) {
	var req LoginRequest
	// 参数错误
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil { // 解析 Json
		return nil, response.MarkBindError(err)
	}

	// 参数转换
	input := req.toSvcInput(ctx)

	// -> svc 获得登录信息
	data, refreshToken, err := ctrl.service.Login(ctx, input)
	if err != nil {
		return nil, err
	}

	// 将刷新令牌放入 cookie
	ctrl.setRefreshTokenCookie(ctx, refreshToken)

	return data, nil
}

// setRefreshTokenCookie 设置 refreshToken 到 cookie
func (ctrl *controller) setRefreshTokenCookie(ctx *gin.Context, refreshToken string) {
	ctx.SetCookie(
		token.RefreshTokenKey,               // key
		refreshToken,                        // value
		int(ctrl.refreshDuration.Seconds()), // maxAge
		domain.RefreshPath,                  // path 只有在访问这个路径的时候才会发送该 cookie
		"",                                  // domain 作用域(默认当前域名)
		false,                               // https 允许 http 传输
		true,                                // httpOnly 防js 窃取
	)
}

/** ====================================================================================
 * 🏁 Post: Refresh
 * =====================================================================================
 */

// Refresh 刷新访问令牌
// Post: /api/v1/auth/refresh
func (ctrl *controller) Refresh(ctx *gin.Context) (interface{}, error) {
	// 从 cookie 获取刷新令牌
	refreshTokenStr, err := ctx.Cookie(token.RefreshTokenKey)
	if err != nil { // 刷新令牌不存在, 返错
		return nil, appError.ErrCookieNoRefreshToken
	}

	// -> svc 校验刷新令牌合规
	data, err := ctrl.service.Refresh(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	// 返回访问令牌
	return data, nil
}

/** ====================================================================================
 * 🏁 Logout
 * =====================================================================================
 */

// 用户登出当前设备
func (ctrl *controller) Logout(ctx *gin.Context) (interface{}, error) {
	// 获取荷载
	payload, err := token.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	// 提取用户名, 设备 ID
	username, deviceID := payload.Username, ctx.GetHeader("X-Device-ID")
	if deviceID == "" {
		return nil, appError.ErrEmptyDeviceID
	}
	log.Println("用户登出, 设备 ID 为: ", deviceID)

	// 筹备参数
	input := auth.LogoutInput{Username: username, DeviceID: deviceID}

	// -> svc 退出登录
	err = ctrl.service.Logout(ctx, input)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
