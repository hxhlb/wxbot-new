package service

import (
	"fmt"
	"log"
	"os"
	"time"

	"wxbot-new/internal/loader"
)

// WeChatServiceEnhanced 增强型微信服务 (支持Manual Mapping)
type WeChatServiceEnhanced struct {
	*WeChatService
	injectionMethod loader.InjectionMethod
	injector        *loader.Injector
}

// NewWeChatServiceEnhanced 创建增强型微信服务
// useManualMap: true=使用Manual Mapping(隐蔽), false=使用经典LoadLibrary(兼容)
func NewWeChatServiceEnhanced(loaderPath, dllPath string, logRecvCallback int, callbackURLs []string, useManualMap bool) *WeChatServiceEnhanced {
	method := loader.MethodClassic
	if useManualMap {
		method = loader.MethodManualMap
	}

	base := NewWeChatService(loaderPath, dllPath, logRecvCallback, callbackURLs)

	return &WeChatServiceEnhanced{
		WeChatService:   base,
		injectionMethod: method,
	}
}

// Initialize 初始化服务
func (s *WeChatServiceEnhanced) Initialize() error {
	log.Printf("正在初始化微信服务 (注入模式: %s)...", s.getMethodName())

	// 如果使用嵌入模式,不需要检查文件
	hasEmbedded := loader.HasEmbeddedDLL("NoveHelper.dll")
	if !hasEmbedded {
		// 检查文件是否存在
		if _, err := os.Stat(s.loaderPath); os.IsNotExist(err) {
			return fmt.Errorf("Loader DLL文件不存在: %s", s.loaderPath)
		}

		if _, err := os.Stat(s.dllPath); os.IsNotExist(err) {
			return fmt.Errorf("Helper DLL文件不存在: %s", s.dllPath)
		}
	} else {
		log.Println("检测到嵌入的DLL资源")
	}

	// 创建注入器
	inj, err := loader.NewInjector(s.loaderPath, s.dllPath, s.injectionMethod)
	if err != nil {
		return fmt.Errorf("创建注入器失败: %v", err)
	}

	if err := inj.Initialize(); err != nil {
		return fmt.Errorf("初始化注入器失败: %v", err)
	}

	s.injector = inj
	s.loader = inj.GetLoader()

	// 注册回调
	s.registerCallbacks()

	// 初始化Socket
	if err := s.loader.InitWeChatSocket(); err != nil {
		return fmt.Errorf("初始化微信Socket失败: %v", err)
	}

	log.Println("微信服务初始化成功")
	return nil
}

// Start 启动服务
func (s *WeChatServiceEnhanced) Start() error {
	if err := s.Initialize(); err != nil {
		return err
	}

	s.isRunning = true
	s.shouldStop = false

	// 执行注入
	log.Printf("正在注入微信 (方式: %s)...", s.getMethodName())
	clientID, err := s.injector.InjectWeChat()
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

// getMethodName 获取注入方式名称
func (s *WeChatServiceEnhanced) getMethodName() string {
	if s.injectionMethod == loader.MethodManualMap {
		return "ManualMapping"
	}
	return "Classic"
}

// Stop 停止服务
func (s *WeChatServiceEnhanced) Stop() {
	log.Println("正在停止微信服务...")
	s.shouldStop = true
	s.isRunning = false

	// 停止清理协程
	close(s.cleanupStopChan)

	if s.injector != nil {
		if err := s.injector.Release(); err != nil {
			log.Printf("释放资源时发生异常: %v", err)
		} else {
			log.Println("微信连接已断开")
		}
	}

	log.Println("微信服务已停止")
}
