package api

import (
	"log"
	"net/http"

	"wxbot-new/internal/service"
)

// WeChatHandler 微信相关接口处理器
type WeChatHandler struct {
	wechatService *service.WeChatService
}

// NewWeChatHandler 创建微信处理器
func NewWeChatHandler(wechatService *service.WeChatService) *WeChatHandler {
	return &WeChatHandler{
		wechatService: wechatService,
	}
}

// CheckServiceStatus 检查微信服务状态
func (h *WeChatHandler) CheckServiceStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"running":         false,
		"message":         "微信服务未初始化",
		"client_id":       0,
		"connected_count": 0,
	}

	if h.wechatService != nil {
		isRunning := h.wechatService.IsRunning()
		clientID := h.wechatService.GetClientID()
		connectedCount := h.wechatService.GetConnectedClientsCount()

		status["running"] = isRunning
		status["client_id"] = clientID
		status["connected_count"] = connectedCount

		if !isRunning {
			status["message"] = "微信服务已停止"
		} else if clientID == 0 {
			status["message"] = "微信服务运行中，但未成功注入或已断开连接"
		} else {
			status["message"] = "微信服务正常运行"
		}
	}

	SuccessResponse(w, "状态检查完成", status)
}

// GetCurrentLoginInfo 获取当前登录信息
func (h *WeChatHandler) GetCurrentLoginInfo(w http.ResponseWriter, r *http.Request) {
	if h.wechatService == nil || !h.wechatService.IsRunning() {
		ErrorResponse(w, http.StatusServiceUnavailable, "微信服务未运行")
		return
	}

	loginInfo, err := h.wechatService.HelperGetCurrentLoginInfo()
	if err != nil {
		log.Printf("获取登录信息失败: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "获取登录信息失败: "+err.Error())
		return
	}

	SuccessResponse(w, "获取登录信息成功", loginInfo)
}
