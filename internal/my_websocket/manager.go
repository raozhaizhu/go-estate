package myWebsocket

import (
	"log/slog"
	"sync"
)

// Manager 管理者, 负责管理客户组连接
type Manager struct {
	sync.RWMutex

	// clients username-> clients -> true
	// 用于记录用户的连接群, 和各连接的状态(开或关)
	clients map[string]map[*Client]bool

	logger *slog.Logger
}

// NewManager 创建新管理者
func NewManager() *Manager {
	logger := slog.Default().With("layer", "wsManager")

	return &Manager{
		clients: make(map[string]map[*Client]bool),
		logger:  logger,
	}
}

// UserConnecting 该用户是否处在连接状态
func (m *Manager) UserConnecting(username string) bool {
	_, ok := m.clients[username]
	return ok
}

// AddClient 新增客户连接到管理器
func (m *Manager) AddClient(username string, client *Client) {
	m.Lock()
	defer m.Unlock()

	// 若该用户未曾注册, 则新建客户频道
	if m.clients[username] == nil {
		m.clients[username] = make(map[*Client]bool)
	}
	m.clients[username][client] = true
}

// RemoveClient 移除特定用户的特定连接, 并做内存清理(如果这是该用户最后一条连接)
func (m *Manager) RemoveClient(username string, client *Client) {
	m.Lock()
	defer m.Unlock()

	m.RemoveClientLocked(username, client)
}

// RemoveClient 移除特定用户的多个连接, 并做内存清理(如果清理后用户再无连接)
func (m *Manager) RemoveClients(username string, clients []*Client) {
	m.Lock()
	defer m.Unlock()

	for _, client := range clients {
		m.RemoveClientLocked(username, client)
	}
}

// RemoveClientLocked 带锁执行, 移除特定用户的特定连接, 并做内存清理(如果这是该用户最后一条连接)
func (m *Manager) RemoveClientLocked(username string, client *Client) {
	// 若该用户已注册该连接, 将该连接移除
	if clients, ok := m.clients[username]; ok {
		if _, exits := clients[client]; exits {
			// 先关闭,后删除连接
			close(client.send)
			delete(clients, client)
		}
		// 删除该连接后, 如果用户所有连接都断开, 释放内存
		if len(clients) == 0 {
			delete(m.clients, username)
		}
	}
}

// SendToUser 向特定用户的所有在线设备发送信息, 若有断开的连接, 将其清理
func (m *Manager) SendToUser(username string, message []byte) {
	// 向特定用户的所有在线设备发送信息
	m.RLock()
	clientsToRemove := m.SendToUserLocked(username, message)
	m.RUnlock()

	// 若有断开的连接, 将其清理
	if len(clientsToRemove) > 0 {
		m.RemoveClients(username, clientsToRemove)
	}
}

// SendToUserLocked 带锁执行, 向特定用户的所有在线设备发送信息
func (m *Manager) SendToUserLocked(username string, message []byte) []*Client {
	// 预备清理的连接(若有)
	var clientsToRemove []*Client

	// 若该用户已注册该连接, 发送消息
	if clients, ok := m.clients[username]; ok {
		for client := range clients {
			select {
			// 发送信息
			case client.send <- message:
			default:
				// 如果发送通道阻塞, 记录下来, 后续统一清理
				clientsToRemove = append(clientsToRemove, client)
			}
		}
	}

	return clientsToRemove
}
