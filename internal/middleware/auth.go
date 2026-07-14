package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// RequireAuth 身份认证中间件
// 校验用户是否携带了 token 进行访问
func RequireAuth(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取验证头
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) == 0 { // 认证头不存在
			response.Fail(ctx, appError.ErrAuthNoHeader)
			ctx.Abort()
			return
		}
		// 解析认证头
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" { // 认证头格式错误
			response.Fail(ctx, appError.ErrAuthBadHeader)
			ctx.Abort()
			return
		}

		// 获取令牌
		accessToken := fields[1]

		// 校验令牌
		payload, err := tokenMaker.VerifyToken(accessToken, token.TokenTypeAccessToken)
		if err != nil { // 令牌无效或过期
			response.Fail(ctx, err)
			ctx.Abort()
			return
		}

		// 将荷载存入上下文
		ctx.Set(token.PayloadKey, payload)
		ctx.Next()
	}
}
