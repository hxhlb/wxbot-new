package service

import (
	"fmt"
	"sync"
	"time"

	"wxbot-new/internal/message"
)

// PendingRequest 待处理的请求
type PendingRequest struct {
	ResponseChan chan map[string]interface{}
	Timeout      time.Time
}

// ResponseManager 异步响应管理器
// 用于处理 DLL 异步返回的消息
type ResponseManager struct {
	pendingRequests map[message.MessageType]*PendingRequest
	mu              sync.RWMutex
	defaultTimeout  time.Duration
}

// NewResponseManager 创建响应管理器
func NewResponseManager(timeout time.Duration) *ResponseManager {
	return &ResponseManager{
		pendingRequests: make(map[message.MessageType]*PendingRequest),
		defaultTimeout:  timeout,
	}
}

// RegisterRequest 注册一个待处理的请求
func (rm *ResponseManager) RegisterRequest(msgType message.MessageType, timeout time.Duration) chan map[string]interface{} {
	if timeout == 0 {
		timeout = rm.defaultTimeout
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 如果已存在,先清理旧的
	if old, exists := rm.pendingRequests[msgType]; exists {
		close(old.ResponseChan)
	}

	responseChan := make(chan map[string]interface{}, 1)
	rm.pendingRequests[msgType] = &PendingRequest{
		ResponseChan: responseChan,
		Timeout:      time.Now().Add(timeout),
	}

	return responseChan
}

// HandleResponse 处理收到的响应
func (rm *ResponseManager) HandleResponse(msgType message.MessageType, data map[string]interface{}) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	request, exists := rm.pendingRequests[msgType]
	if !exists {
		return false
	}

	// 检查是否超时
	if time.Now().After(request.Timeout) {
		close(request.ResponseChan)
		delete(rm.pendingRequests, msgType)
		return false
	}

	// 发送响应
	select {
	case request.ResponseChan <- data:
		delete(rm.pendingRequests, msgType)
		return true
	default:
		// Channel 已满或已关闭
		delete(rm.pendingRequests, msgType)
		return false
	}
}

// WaitForResponse 等待响应
func (rm *ResponseManager) WaitForResponse(responseChan chan map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	select {
	case data := <-responseChan:
		return data, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("等待响应超时")
	}
}

// CleanupExpiredRequests 清理过期的请求
func (rm *ResponseManager) CleanupExpiredRequests() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	for msgType, request := range rm.pendingRequests {
		if now.After(request.Timeout) {
			close(request.ResponseChan)
			delete(rm.pendingRequests, msgType)
		}
	}
}

// StartCleanupRoutine 启动定期清理过期请求的协程
func (rm *ResponseManager) StartCleanupRoutine(interval time.Duration, stopChan chan bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.CleanupExpiredRequests()
		case <-stopChan:
			return
		}
	}
}
