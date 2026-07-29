package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// RequireAuth 身份认证中间件
// 校验用户是否携带了 token 进行访问
func RequireAuth(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取验证头
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) == 0 { // 认证头不存在
			response.Fail(c, appError.ErrAuthNoHeader)
			c.Abort()
			return
		}
		// 解析认证头
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" { // 认证头格式错误
			response.Fail(c, appError.ErrAuthBadHeader)
			c.Abort()
			return
		}

		// 获取令牌
		accessToken := fields[1]

		// 校验令牌
		payload, err := tokenMaker.VerifyToken(accessToken, token.TokenTypeAccessToken)
		if err != nil { // 令牌无效或过期
			response.Fail(c, err)
			c.Abort()
			return
		}
		// 将荷载, 用户名存入上下文
		ctx := context.WithValue(c.Request.Context(), ctxKey.CtxKeyPayload, payload)
		ctx = context.WithValue(ctx, ctxKey.CtxKeyUserName, payload.Username)
		c.Request = c.Request.WithContext(ctx)
		c.Set(ctxKey.CtxKeyPayload, payload)
		c.Set(ctxKey.CtxKeyUserName, payload.Username)
		c.Next()
	}
}
