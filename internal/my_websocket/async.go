package myWebsocket

import (
	"context"
	"encoding/json"
	"fmt"
)

const typeQueryConsumption = "QUERY_CONSUMPTION"

// GoSendPointsMsgToUser 异步发送扣费通知
func (m *Manager) GoSendPointsMsgToUser(ctx context.Context, username string, amount int) {
	asyncCtx := context.WithoutCancel(ctx)

	// 开启协程, 异步通知用户
	go func() {
		// 将信息序列化
		content := fmt.Sprintf("查询成功，扣除 %v 个查询点数", amount)
		msg, err := json.Marshal(map[string]interface{}{
			"type":    typeQueryConsumption,
			"content": content,
		})
		if err != nil { // 序列化失败
			m.logger.ErrorContext(asyncCtx, "WebSocket 消息序列化失败", "err", err.Error())
			return
		}

		// 发送 WS 消息
		m.SendToUser(username, msg)
	}()
}
