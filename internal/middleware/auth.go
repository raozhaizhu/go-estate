package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// AuthMiddleware 身份认证中间件
// 校验用户是否携带了 token 进行访问
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取验证头
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) == 0 { // 认证头不存在
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, appError.ErrAuthNoHeader)
			return
		}

		// 解析认证头
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" { // 认证头格式错误
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, appError.ErrAuthBadHeader)
			return
		}

		// 获取令牌
		accessToken := fields[1]

		// 校验令牌
		payload, err := tokenMaker.VerifyToken(accessToken, token.TokenTypeAccessToken)
		// log.Println("payload: ", payload)
		if err != nil { // 令牌无效或过期
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		// 将荷载存入上下文
		ctx.Set(token.PayloadKey, payload)
		ctx.Next()
	}
}
