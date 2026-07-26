package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

// SlogMiddleware 将 Gin 的 HTTP 访问记录桥接到 slog 中
func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()
		// 不处理 swagger 请求
		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			return
		}

		// 请求结束，收集数据打日志
		cost := time.Since(start)
		// 获取 http 状态码
		httpStatus := c.Writer.Status()
		// 获取 bizCode 报错编号
		bizCode := c.GetInt(response.BizCodeKey)
		// 获取挂载的原始 error 信息
		var rawErrMsgs []string
		for _, e := range c.Errors.ByType(gin.ErrorTypePrivate) {
			rawErrMsgs = append(rawErrMsgs, e.Err.Error())
		}
		errs := strings.Join(rawErrMsgs, "; ")

		// 判断报错级别
		var logFn func(string, ...any)
		switch {
		case httpStatus >= 500 || bizCode >= 50000:
			logFn = logger.Error // 内部错误,灾难级
		case httpStatus >= 400 || (bizCode >= 40000 && bizCode < 50000):
			logFn = logger.Warn // 业务异常,警告级
		default:
			logFn = logger.Info // 请求成功,信息级
		}

		// 使用 slog 输出结构化日志
		logAttrs := []any{
			slog.Int("status", httpStatus),
			slog.Int("biz_code", bizCode),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("cost", cost.String()),
		}
		if errs != "" {
			logAttrs = append(logAttrs, slog.String("errors", errs))
		}

		logFn("HTTP 请求已处理", logAttrs...)
	}
}
