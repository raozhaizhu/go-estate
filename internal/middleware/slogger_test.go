package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestSlogMiddleware(t *testing.T) {
	// 准备 Buffer 用于捕获日志
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	ctxHandler := logger.ContextHandler{Handler: jsonHandler}
	slog.SetDefault(slog.New(ctxHandler))

	// 初始化 Gin 环境
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.SlogMiddleware())

	// 注册测试路由
	router.GET("/success", func(c *gin.Context) { response.Success(c, "success") })
	router.GET("/warn", func(c *gin.Context) { response.Fail(c, appError.ErrAuthBadHeader) })
	router.GET("/error", func(c *gin.Context) { response.Fail(c, appError.ErrServerErr) })

	// 执行请求
	tests := []struct {
		name          string
		path          string
		expectedLevel string
	}{
		{"Success", "/success", "INFO"},
		{"Warn", "/warn", "WARN"},
		{"Error", "/error", "ERROR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset() // 每次测试前清空日志缓冲区
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
			router.ServeHTTP(w, req)

			// 解析 JSON 日志
			logOutput := buf.String()
			var logEntry map[string]any
			err := json.Unmarshal([]byte(logOutput), &logEntry)
			require.NoError(t, err)

			// 验证级别/状态/Msg
			require.Equal(t, tc.expectedLevel, strings.ToUpper(logEntry["level"].(string)))
			require.NotEmpty(t, logEntry["status"])
			require.Equal(t, "HTTP 请求已处理", logEntry["msg"])
		})
	}
}
