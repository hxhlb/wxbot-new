package wxsvc

import (
    "encoding/json"
    "log"
    "sync"
    "time"
)

// Service 负责管理生命周期、重连与消息发送
type Service struct {
    loaderPath string
    dllPath    string

    loader     Loader
    clientID   uint32

    mu         sync.RWMutex
    running    bool
    shouldStop bool

    lastHeartbeat time.Time
    reconnectAttempts int
    maxReconnectAttempts int
    reconnectDelay time.Duration

    handlers []Handler
}

func NewService(loaderPath, dllPath string) *Service {
    s := &Service{
        loaderPath: loaderPath,
        dllPath:    dllPath,
        loader:     newLoader(loaderPath),
        lastHeartbeat: time.Now(),
        maxReconnectAttempts: 5,
        reconnectDelay: 10 * time.Second,
    }
    // 默认处理器
    s.handlers = append(s.handlers, newDefaultHandler(s))
    return s
}

// Start 初始化并启动服务
func (s *Service) Start() error {
    s.mu.Lock()
    if s.running {
        s.mu.Unlock()
        return nil
    }
    s.running = true
    s.shouldStop = false
    s.mu.Unlock()

    log.Println("正在初始化微信服务...")

    // 尝试获取版本信息
    if ver, err := s.loader.GetUserWeChatVersion(); err == nil && ver != "" {
        log.Printf("微信版本: %s", ver)
    }

    // 注册回调（当前 stub 环境仅模拟）
    s.loader.InitWeChatSocket(
        func(clientID uint32) {
            s.mu.Lock()
            s.clientID = clientID
            s.lastHeartbeat = time.Now()
            s.mu.Unlock()
            s.dispatchConnect(clientID)
        },
        func(clientID uint32, data []byte) {
            s.mu.Lock()
            s.lastHeartbeat = time.Now()
            s.mu.Unlock()

            // 解析 JSON 并分发
            var msg struct {
                Type int             `json:"type"`
                Data map[string]any `json:"data"`
            }
            if err := json.Unmarshal(data, &msg); err != nil {
                log.Printf("消息解析失败: %v, 原始: %s", err, string(data))
                return
            }
            s.dispatchReceive(clientID, msg.Type, msg.Data)
        },
        func(clientID uint32) {
            s.dispatchClose(clientID)
        },
    )

    // 注入/连接
    clientID, err := s.loader.InjectWeChat(s.dllPath)
    if err != nil {
        log.Printf("注入失败: %v", err)
        return err
    }
    s.mu.Lock()
    s.clientID = clientID
    s.lastHeartbeat = time.Now()
    s.mu.Unlock()

    log.Printf("成功注入，客户端ID: %d", clientID)

    go s.runLoop()
    return nil
}

// Stop 停止服务
func (s *Service) Stop() {
    s.mu.Lock()
    s.shouldStop = true
    s.running = false
    s.mu.Unlock()
    _ = s.loader.DestroyWeChat()
    log.Println("微信服务已停止")
}

func (s *Service) runLoop() {
    log.Println("微信服务已启动，正在运行...")
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    for {
        <-ticker.C
        s.mu.Lock()
        if s.shouldStop {
            s.mu.Unlock()
            return
        }
        // 2分钟无心跳则尝试重连
        if time.Since(s.lastHeartbeat) > 120*time.Second {
            log.Println("检测到连接超时，尝试重连...")
            if !s.reconnect() {
                s.mu.Unlock()
                s.Stop()
                return
            }
        }
        s.mu.Unlock()
    }
}

func (s *Service) reconnect() bool {
    if s.reconnectAttempts >= s.maxReconnectAttempts {
        log.Printf("重连次数超过限制(%d)，停止重连", s.maxReconnectAttempts)
        return false
    }
    s.reconnectAttempts++
    log.Printf("尝试重连 (%d/%d)...", s.reconnectAttempts, s.maxReconnectAttempts)
    _ = s.loader.DestroyWeChat()
    time.Sleep(s.reconnectDelay)
    clientID, err := s.loader.InjectWeChat(s.dllPath)
    if err != nil || clientID == 0 {
        log.Printf("重连失败: %v", err)
        return false
    }
    s.clientID = clientID
    s.lastHeartbeat = time.Now()
    s.reconnectAttempts = 0
    log.Printf("重连成功，客户端ID: %d", clientID)
    return true
}

// SendPayload 发送自定义 payload（会序列化为 JSON）
func (s *Service) SendPayload(payload map[string]any) bool {
    b, err := json.Marshal(payload)
    if err != nil {
        log.Printf("payload 序列化失败: %v", err)
        return false
    }
    msg := string(b)
    // 与 Python 版本保持一致：发送时使用 clientID=1
    ok, err := s.loader.SendWeChatData(1, msg)
    if err != nil {
        log.Printf("发送失败: %v", err)
        return false
    }
    if ok {
        log.Printf("消息发送成功: %s", msg)
    } else {
        log.Printf("消息发送失败: %s", msg)
    }
    return ok
}

// SendMessage 发送原始 JSON 字符串
func (s *Service) SendMessage(message string) bool {
    ok, err := s.loader.SendWeChatData(1, message)
    if err != nil {
        log.Printf("发送失败: %v", err)
        return false
    }
    return ok
}

// SendStartupPayload 启动时发送一次固定结构的消息
func (s *Service) SendStartupPayload(roomWxid string, status int) bool {
    payload := map[string]any{
        "data": map[string]any{
            "room_wxid": roomWxid,
            "status":    status,
        },
        "type": 11075,
    }
    return s.SendPayload(payload)
}

// 回调分发与注册
func (s *Service) AddHandler(h Handler) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.handlers = append(s.handlers, h)
}

func (s *Service) dispatchConnect(clientID uint32) {
    s.mu.RLock()
    hs := append([]Handler(nil), s.handlers...)
    s.mu.RUnlock()
    for _, h := range hs {
        h.OnConnect(clientID)
    }
}

func (s *Service) dispatchReceive(clientID uint32, messageType int, data map[string]any) {
    s.mu.RLock()
    hs := append([]Handler(nil), s.handlers...)
    s.mu.RUnlock()
    for _, h := range hs {
        h.OnReceive(clientID, messageType, data)
    }
}

func (s *Service) dispatchClose(clientID uint32) {
    s.mu.RLock()
    hs := append([]Handler(nil), s.handlers...)
    s.mu.RUnlock()
    for _, h := range hs {
        h.OnClose(clientID)
    }
}
