package wxsvc

// Handler 定义服务回调接口（等价 Python 版本的 CONNECT/RECV/CLOSE 回调）
type Handler interface {
    OnConnect(clientID uint32)
    OnReceive(clientID uint32, messageType int, data map[string]any)
    OnClose(clientID uint32)
}

