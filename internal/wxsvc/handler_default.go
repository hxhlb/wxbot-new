package wxsvc

import (
    "log"
    "sync"
)

// defaultHandler 默认回调处理器：
// - 记录连接数量
// - 处理登录/登出/调试日志
// - 登录后发送一次启动 payload
type defaultHandler struct {
    svc *Service
    mu  sync.Mutex
    connected map[uint32]struct{}
}

func newDefaultHandler(s *Service) *defaultHandler {
    return &defaultHandler{svc: s, connected: make(map[uint32]struct{})}
}

func (h *defaultHandler) OnConnect(clientID uint32) {
    h.mu.Lock()
    h.connected[clientID] = struct{}{}
    size := len(h.connected)
    h.mu.Unlock()
    log.Printf("客户端 %d 已连接，当前连接数: %d", clientID, size)
}

func (h *defaultHandler) OnReceive(clientID uint32, messageType int, data map[string]any) {
    log.Printf("收到来自客户端 %d 的消息 - 类型: %d, 数据: %v", clientID, messageType, data)
    switch messageType {
    case MT_USER_LOGIN:
        log.Printf("用户登录: %v", data)
        _ = h.svc.SendStartupPayload("47945916190@chatroom", 0)
    case MT_USER_LOGOUT:
        log.Printf("用户登出: %v", data)
    case MT_DEBUG_LOG:
        log.Printf("调试日志: %v", data)
    }
}

func (h *defaultHandler) OnClose(clientID uint32) {
    h.mu.Lock()
    delete(h.connected, clientID)
    size := len(h.connected)
    h.mu.Unlock()
    log.Printf("客户端 %d 已断开，当前连接数: %d", clientID, size)
}

