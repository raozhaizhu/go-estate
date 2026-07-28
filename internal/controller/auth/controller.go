package auth

import (
	"github.com/gin-gonic/gin"
	domain "github.com/raozhaizhu/go-estate/internal/domain/auth"
	"github.com/raozhaizhu/go-estate/internal/middleware"
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
// @Summary      账户登录
// @Description  用户通过账号密码登录。成功后返回 AccessToken，并将 RefreshToken 写入 Cookie。
// @Tags         Auth (身份认证)
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录参数"
// @Success      200  {object}  response.SuccessResult[auth.DTO]  "登录成功 (返回 auth.DTO 载荷)"
// @Failure      400  {object}  response.ClientErrorResult        "参数解析或校验错误 (HTTP 返回 200, code: 400xx)"
// @Failure      401  {object}  response.AuthErrorResult          "账户名或密码错误 (HTTP 返回 200, code: 401xx)"
// @Failure      404  {object}  response.NotFoundErrorResult      "账户不存在 (HTTP 返回 200, code: 404xx)"
// @Failure      500  {object}  response.ServerErrorResult        "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/auth/login [post]
func (ctrl *controller) Login(c *gin.Context) (interface{}, error) {
	var req LoginRequest
	// 参数错误
	if err := c.ShouldBindBodyWithJSON(&req); err != nil { // 解析 Json
		return nil, response.MarkBindError(err)
	}
	// 提取上下文
	ctx := c.Request.Context()
	meta, err := middleware.GetCtxClientMeta(ctx)
	if err != nil {
		return nil, err
	}
	// 参数转换
	input := req.toSvcInput(meta)

	// -> svc 获得登录信息
	data, refreshToken, err := ctrl.service.Login(c, input)
	if err != nil {
		return nil, err
	}

	// 将刷新令牌放入 cookie
	ctrl.setRefreshTokenCookie(c, refreshToken)

	return data, nil
}

// setRefreshTokenCookie 设置 refreshToken 到 cookie
// TODO 设置 env 里的 production_mode
func (ctrl *controller) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	c.SetCookie(
		token.RefreshTokenKey,               // key
		refreshToken,                        // value
		int(ctrl.refreshDuration.Seconds()), // maxAge
		domain.RefreshPath,                  // path 只有在访问这个路径的时候才会发送该 cookie
		"",                                  // domain 作用域(默认当前域名)
		ctrl.isProduction,                   // https 生产环境下 https 传输, 开发环境下 http 即可
		true,                                // httpOnly 防js 窃取
	)
}

/** ====================================================================================
 * 🏁 Post: Refresh
 * =====================================================================================
 */

// Refresh 刷新访问令牌
// @Summary      刷新访问令牌
// @Description  通过 Cookie 中的 RefreshToken 换取新的 AccessToken (注意：需要浏览器自动携带名为 refresh_token 的 Cookie)
// @Tags         Auth (身份认证)
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Produce      json
// @Success      200  {object}  response.SuccessResult[auth.DTO]  "刷新成功"
// @Failure      401  {object}  response.AuthErrorResult          "刷新令牌不存在或已失效 (HTTP 返回 200, code: 401xx)"
// @Failure      500  {object}  response.ServerErrorResult        "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/auth/refresh [post]
func (ctrl *controller) Refresh(c *gin.Context) (interface{}, error) {
	// 从 cookie 获取刷新令牌
	refreshTokenStr, err := c.Cookie(token.RefreshTokenKey)
	if err != nil { // 刷新令牌不存在, 返错
		return nil, appError.ErrCookieNoRefreshToken
	}
	// 提取上下文
	ctx := c.Request.Context()

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

// Logout 用户登出当前设备
// @Summary      用户登出当前设备
// @Description  清理当前设备的登录状态。需要携带 AccessToken 和设备标识 ID。
// @Tags         Auth (身份认证)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Produce      json
// @Success      200  {object}  response.SuccessResult[any] "登出成功 (data 为 null)"
// @Failure      401  {object}  response.AuthErrorResult            "未登录或 Token 失效 (HTTP 返回 200, code: 401xx)"
// @Failure      500  {object}  response.ServerErrorResult          "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/auth/logout [post]
func (ctrl *controller) Logout(c *gin.Context) (interface{}, error) {
	// 获取荷载
	payload, err := token.GetPayload(c)
	if err != nil {
		return nil, err
	}

	// 提取上下文
	ctx := c.Request.Context()
	meta, err := middleware.GetCtxClientMeta(ctx)
	if err != nil {
		return nil, err
	}

	// 筹备参数
	input := auth.LogoutInput{Username: payload.Username, DeviceID: meta.DeviceID}

	// -> svc 退出登录
	err = ctrl.service.Logout(c, input)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
