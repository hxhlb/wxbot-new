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
