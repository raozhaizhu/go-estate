package middleware

import (
	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

const (
	CtxKeyDeviceID  = "ctx_device_id"
	CtxKeyUserAgent = "ctx_user_agent"
	CtxKeyClientIP  = "ctx_client_ip"
)

// RequireMetadata 校验必要元数据, 并将其加入上下文
func RequireMetadata() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取必要元数据
		deviceID := ctx.GetHeader("X-Device-ID")
		userAgent := ctx.Request.UserAgent()
		clientIP := ctx.ClientIP()

		if deviceID == "" {
			response.Fail(ctx, appError.ErrEmptyDeviceID)
			ctx.Abort()
			return
		}
		if userAgent == "" {
			response.Fail(ctx, appError.ErrEmptyUserAgent)
			ctx.Abort()
			return
		}

		ctx.Set(CtxKeyDeviceID, deviceID)
		ctx.Set(CtxKeyUserAgent, userAgent)
		ctx.Set(CtxKeyClientIP, clientIP)

		ctx.Next()
	}
}
