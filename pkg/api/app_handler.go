package response

import (
	"github.com/gin-gonic/gin"
)

type AppHandler func(c *gin.Context) (interface{}, error)

func Wrapper(handler AppHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 执行业务逻辑
		data, err := handler(ctx)

		// 统一错误处理
		if err != nil {
			Fail(ctx, err)
			return
		}

		// 统一成功响应
		Success(ctx, data)
	}
}
