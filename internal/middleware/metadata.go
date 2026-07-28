package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
)

type ClientMeta struct {
	DeviceID  string
	UserAgent string
	ClientIP  string
}

// RequireMetadata 校验必要元数据, 并将其加入上下文
func RequireMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取必要元数据
		deviceID := c.GetHeader("X-Device-ID")
		userAgent := c.Request.UserAgent()
		clientIP := c.ClientIP()

		if deviceID == "" {
			response.Fail(c, appError.ErrEmptyDeviceID)
			c.Abort()
			return
		}
		if userAgent == "" {
			response.Fail(c, appError.ErrEmptyUserAgent)
			c.Abort()
			return
		}

		meta := ClientMeta{
			DeviceID:  deviceID,
			UserAgent: userAgent,
			ClientIP:  clientIP,
		}
		ctx := context.WithValue(c.Request.Context(), ctxKey.CtxKeyClientMeta, meta)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func GetCtxClientMeta(ctx context.Context) (ClientMeta, error) {
	meta, ok := ctx.Value(ctxKey.CtxKeyClientMeta).(ClientMeta)
	if !ok {
		return ClientMeta{}, appError.NewSrvErr(fmt.Errorf("提取 metadata 失败"))
	}
	return meta, nil
}
