package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/raozhaizhu/go-estate/internal/myWebsocket"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
)

// upgrader 升级 HTTP 为 websocket 的配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin 开发环境下允许跨域, 生产环境下需修改
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS 处理 Websocket 请求
func ServeWS(manager *myWebsocket.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default().With("layer", "middleware", "module", "serve_ws")
		// 获取用户名
		username := c.GetString(ctxKey.CtxKeyUserName)
		// 用户未登录
		if username == "" {
			response.Fail(c, appError.ErrAuthRequired)
			c.Abort()
			return
		}

		// 升级协议
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			response.Fail(c, appError.NewSrvErr(err))
			logger.ErrorContext(c.Request.Context(), "websocket 升级失败", "error", err)
		}

		// 初始化客户端
		client := myWebsocket.NewClient(manager, username, conn)
		// 注册到管理者
		manager.AddClient(username, client)

		// 启动读写器
		client.Start()
	}
}
