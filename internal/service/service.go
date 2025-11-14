package service

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"wxbot-new/internal/loader"
	"wxbot-new/internal/message"
)

// WeChatService 微信服务管理器
type WeChatService struct {
	loaderPath           string
	dllPath              string
	loader               *loader.NoveLoader
	isRunning            bool
	shouldStop           bool
	clientID             uint32
	lastHeartbeat        time.Time
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration
	connectedClients     map[uintptr]bool
	responseManager      *ResponseManager
	cleanupStopChan      chan bool
	mu                   sync.RWMutex
}

// NewWeChatService 创建微信服务
func NewWeChatService(loaderPath, dllPath string) *WeChatService {
	return &WeChatService{
		loaderPath:           loaderPath,
		dllPath:              dllPath,
		maxReconnectAttempts: 5,
		reconnectDelay:       10 * time.Second,
		connectedClients:     make(map[uintptr]bool),
		responseManager:      NewResponseManager(10 * time.Second),
		cleanupStopChan:      make(chan bool),
	}
}

// Initialize 初始化服务
func (s *WeChatService) Initialize() error {
	log.Println("正在初始化微信服务...")

	// 检查文件是否存在
	if _, err := os.Stat(s.loaderPath); os.IsNotExist(err) {
		return fmt.Errorf("Loader DLL文件不存在: %s", s.loaderPath)
	}

	if _, err := os.Stat(s.dllPath); os.IsNotExist(err) {
		return fmt.Errorf("Helper DLL文件不存在: %s", s.dllPath)
	}

	// 创建加载器
	l, err := loader.NewNoveLoader(s.loaderPath)
	if err != nil {
		return fmt.Errorf("创建NoveLoader失败: %v", err)
	}
	s.loader = l

	// 注册回调
	s.registerCallbacks()

	// 初始化Socket
	if err := s.loader.InitWeChatSocket(); err != nil {
		return fmt.Errorf("初始化微信Socket失败: %v", err)
	}

	log.Println("微信服务初始化成功")
	return nil
}

// registerCallbacks 注册回调函数
func (s *WeChatService) registerCallbacks() {
	// 连接回调
	s.loader.AddConnectCallback(func(clientID uintptr) {
		s.mu.Lock()
		s.connectedClients[clientID] = true
		// 更新为真实的客户端ID
		s.clientID = uint32(clientID)
		clientCount := len(s.connectedClients)
		s.mu.Unlock()

		log.Printf("客户端 %d 已连接，当前连接数: %d (已更新 ClientID)", clientID, clientCount)
	})

	// 接收消息回调
	s.loader.AddRecvCallback(func(clientID uintptr, msgType int, data map[string]interface{}) {
		log.Printf("收到来自客户端 %d 的消息 - 类型: %d, 数据: %v", clientID, msgType, data)

		msgTypeEnum := message.MessageType(msgType)

		// 尝试将响应传递给响应管理器(基于消息类型+客户端ID匹配)
		if s.responseManager.HandleResponse(msgType, uint32(clientID), data) {
			log.Printf("响应已发送给等待的请求: msgType=%d, clientID=%d", msgType, clientID)
		}

		// 处理不同类型的消息
		switch msgTypeEnum {
		case message.MTUserLogin:
			log.Printf("用户登录: %v", data)
		case message.MTUserLogout:
			log.Printf("用户登出: %v", data)
		case message.MTDebugLog:
			log.Printf("调试日志: %v", data)
		case message.MTFriendList:
			log.Printf("收取好友列表数据: %v", data)
		case message.MTGroupList:
			log.Printf("收取群列表数据: %v", data)
		case message.MTGroupMemberList:
			log.Printf("收取群成员列表数据: %v", data)
		case message.MTCurrentLoginInfo:
			log.Printf("收取当前登录信息: data=%v", data)
		case message.MTChatMessage:
			log.Printf("收取聊天消息数据: %v", data)
			// 示例：向文件传输助手发送消息
			// s.HelperSendText("filehelper", "收到消息")
		}
	})

	// 关闭回调
	s.loader.AddCloseCallback(func(clientID uintptr) {
		s.mu.Lock()
		delete(s.connectedClients, clientID)
		clientCount := len(s.connectedClients)
		s.mu.Unlock()

		log.Printf("客户端 %d 已断开，当前连接数: %d", clientID, clientCount)
	})
}

// Start 启动服务
func (s *WeChatService) Start() error {
	if err := s.Initialize(); err != nil {
		return err
	}

	s.isRunning = true
	s.shouldStop = false

	// 注入微信
	log.Println("正在注入微信...")
	clientID, err := s.loader.InjectWeChat(s.dllPath)
	if err != nil {
		return fmt.Errorf("注入微信失败: %v", err)
	}

	if clientID == 0 {
		return fmt.Errorf("注入微信失败，客户端ID为0")
	}

	s.clientID = clientID
	log.Printf("成功注入微信，客户端ID: %d", clientID)
	s.reconnectAttempts = 0

	// 初始化心跳时间戳
	s.mu.Lock()
	s.lastHeartbeat = time.Now()
	s.mu.Unlock()

	// 启动心跳监控
	go s.startHeartbeat()

	// 启动响应管理器清理协程
	go s.responseManager.StartCleanupRoutine(5*time.Second, s.cleanupStopChan)

	// 启动主服务循环
	s.runService()

	return nil
}

// startHeartbeat 启动心跳监控
func (s *WeChatService) startHeartbeat() {
	log.Println("心跳监控已启动")

	for s.isRunning && !s.shouldStop {
		log.Println("心跳监控验证...")
		s.mu.Lock()
		s.lastHeartbeat = time.Now()
		s.mu.Unlock()

		time.Sleep(60 * time.Second)
	}
}

// runService 运行服务主循环
func (s *WeChatService) runService() {
	log.Println("微信服务已启动，正在运行...")

	defer s.Stop()

	for s.isRunning && !s.shouldStop {
		// 检查是否需要重连
		s.mu.RLock()
		lastHeartbeat := s.lastHeartbeat
		s.mu.RUnlock()

		if time.Since(lastHeartbeat) > 120*time.Second {
			log.Println("检测到连接超时，尝试重连...")
			if !s.reconnect() {
				break
			}
			continue
		}

		time.Sleep(1 * time.Second)
	}
}

// reconnect 重连服务
func (s *WeChatService) reconnect() bool {
	if s.reconnectAttempts >= s.maxReconnectAttempts {
		log.Printf("重连次数超过限制 (%d)，停止重连", s.maxReconnectAttempts)
		return false
	}

	s.reconnectAttempts++
	log.Printf("尝试重连 (%d/%d)...", s.reconnectAttempts, s.maxReconnectAttempts)

	// 清理当前连接
	if s.loader != nil {
		s.loader.DestroyWeChat()
	}

	time.Sleep(s.reconnectDelay)

	// 重新注入
	clientID, err := s.loader.InjectWeChat(s.dllPath)
	if err != nil || clientID == 0 {
		log.Println("重连失败")
		return false
	}

	s.clientID = clientID
	log.Printf("重连成功，客户端ID: %d", clientID)

	s.mu.Lock()
	s.lastHeartbeat = time.Now()
	s.mu.Unlock()

	s.reconnectAttempts = 0
	return true
}

// Stop 停止服务
func (s *WeChatService) Stop() {
	log.Println("正在停止微信服务...")
	s.shouldStop = true
	s.isRunning = false

	// 停止清理协程
	close(s.cleanupStopChan)

	if s.loader != nil {
		if err := s.loader.Release(); err != nil {
			log.Printf("释放资源时发生异常: %v", err)
		} else {
			log.Println("微信连接已断开")
		}
	}

	log.Println("微信服务已停止")
}

// SendMessage 发送消息
func (s *WeChatService) SendMessage(message string) error {
	if s.loader == nil {
		return fmt.Errorf("Loader 未初始化")
	}

	if s.clientID == 0 {
		return fmt.Errorf("客户端ID为0，微信可能未成功注入或已断开连接")
	}

	if !s.isRunning {
		return fmt.Errorf("微信服务未运行")
	}

	// 检查连接状态
	s.mu.RLock()
	isConnected := s.connectedClients[uintptr(s.clientID)]
	connectedCount := len(s.connectedClients)
	s.mu.RUnlock()

	log.Printf("发送消息前检查: ClientID=%d, IsConnected=%v, TotalConnections=%d, IsRunning=%v",
		s.clientID, isConnected, connectedCount, s.isRunning)

	if err := s.loader.SendWeChatData(s.clientID, message); err != nil {
		log.Printf("消息发送失败 (ClientID: %d): %v", s.clientID, err)
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功 (ClientID: %d): %s", s.clientID, message)
	return nil
}

// IsRunning 检查服务是否运行中
func (s *WeChatService) IsRunning() bool {
	return s.isRunning
}

// GetClientID 获取客户端ID
func (s *WeChatService) GetClientID() uint32 {
	return s.clientID
}

// GetConnectedClientsCount 获取已连接客户端数量
func (s *WeChatService) GetConnectedClientsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connectedClients)
}
