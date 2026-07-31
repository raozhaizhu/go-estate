package myWebsocket

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

/** ====================================================================================
 * 🏁 Constants
 * =====================================================================================
 */

const (
	// writeWait 写超过时间
	writeWait = 10 * time.Second

	// pongWait 允许等待心跳响应的最长时间
	// 超过该时间, 意味着 Manager 无法收到来自 Client 的答复, 链接失效
	pongWait = 60 * time.Second

	// pingPeriod 发送心跳包的周期
	// 假设 pongWait = 60秒, 则 pingPeriod = 54 秒
	// 以 pingPeriod 为周期, Manager 向 Client 发送消息保持联系
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize Client 发消息给 Manager 时的最大尺寸
	// 单位: Byte
	maxMessageSize = 512

	// maxSendMessages client.send 最多容纳的消息数量
	maxSendMessages
)

/** ====================================================================================
 * 🏁 Type
 * =====================================================================================
 */

// Client 用户连接, Client 与 Manager 通信并接收信息
// Client 对应设备(Device), 一个用户(User) 可能在同一时间多台设备建立多个Client
type Client struct {
	// manager 管理者
	manager  *Manager
	username string
	// conn 用于和客户端联系, 收发信息
	conn *websocket.Conn
	// send Manager 发消息给 Client 的专用通道
	send chan []byte
}

func NewClient(manager *Manager, username string, conn *websocket.Conn) *Client {
	return &Client{
		manager:  manager,
		username: username,
		conn:     conn,
		// send 给缓冲, 防阻塞, 最多 256 条消息
		send: make(chan []byte, maxSendMessages),
	}
}

// Start 开始工作, 监听
func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}

// readPump 读取来自 Client 的消息并处理
// 若 Manager 发来心跳, 重置心跳周期, 若发现断网即时清理资源
func (c *Client) readPump() {
	// 退出时清理资源
	defer func() {
		c.manager.RemoveClient(c.username, c)
		c.conn.Close()
	}()

	// manager 最多可读该尺寸消息, client 最多可发该尺寸消息
	c.conn.SetReadLimit(maxMessageSize)
	// 设置初始的读超时时间
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}
	// 收到 Pong 回应后, 刷新超时时间
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return nil
	})

	for {
		// 读取来自 Client 的消息, 并进行处理(当前仅处理心跳)
		_, _, err := c.conn.ReadMessage()
		// 出现错误, 说明连接断开, 会执行 defer 资源清理
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket 读取错误", slog.String("error", err.Error()))
			}
			break
		}
	}

}

// writePump 从服务端接收消息, 若没有消息则发送心跳保持联系
// 条件: 在心跳周期内, 是否接收到来自 Manager 的消息
// 1. True. 写入消息
// 2. False. 发送心跳保持联系
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	// 退出时清理资源
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		// 从 Manager 收到信息, 执行写入
		case message, ok := <-c.send:
			// 设置读超时
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if !ok {
				// manager 关闭了通道, client 必须退出并清理资源
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// 实例化 writer, 将来自 Manager 的信息写入, 然后关闭 writer
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, err = w.Write(message)
			if err != nil {
				return
			}
			if err := w.Close(); err != nil {
				return
			}
		// 倒计时到期, Manager 主动联系 Client
		case <-ticker.C:
			// 定时发送 Ping 心跳包给客户端
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
