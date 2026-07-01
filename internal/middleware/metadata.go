package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

// RequireMetadata 校验必要元数据, 并将其加入上下文
func RequireMetadata() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取必要元数据
		deviceID := ctx.GetHeader("X-Device-ID")
		userAgent := ctx.Request.UserAgent()

		if deviceID == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, appError.ErrEmptyDeviceID)
			return
		}
		if userAgent == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, appError.ErrEmptyUserAgent)
			return
		}

		ctx.Next()
	}
}
