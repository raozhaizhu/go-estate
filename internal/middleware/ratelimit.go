package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

func RateLimiter(redisCache cache.Cache, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 IP
		ip := c.GetString(CtxKeyClientIP)
		ctx := context.Background()

		// 得到当前访问次数
		currCnt, err := redisCache.IncrIPCnt(ctx, ip, duration)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}

		// 访问次数过多, 拒绝访问
		if currCnt > int64(limit) {
			response.Fail(c, appError.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
