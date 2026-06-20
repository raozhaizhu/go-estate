package middleware

import (
	"context"
	"log"

	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// GetPayload 从上下文中获取荷载
func GetPayload(ctx context.Context) (*token.Payload, error) {
	log.Println("ctx: ", ctx)
	// 提取 payload
	val := ctx.Value(PayloadKey)

	// 提取失败, 返回错误
	if val == nil {
		return nil, appError.ErrAuthRequired
	}
	payload := val.(*token.Payload)

	// 返回 payload
	return payload, nil
}
