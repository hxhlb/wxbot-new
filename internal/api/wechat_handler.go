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

// LogoutCurrent 注销当前微信账号
func (h *WeChatHandler) LogoutCurrent(w http.ResponseWriter, r *http.Request) {
	if err := h.wechatService.HelperLogoutCurrent(); err != nil {
		log.Printf("注销当前微信账号失败: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "注销当前微信账号失败: "+err.Error())
		return
	}

	// 无需返回业务数据,仅返回统一成功响应
	SuccessResponse(w, "注销请求已发送", nil)
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
	loginInfo, err := h.wechatService.HelperGetCurrentLoginInfo()
	if err != nil {
		log.Printf("获取登录信息失败: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "获取登录信息失败: "+err.Error())
		return
	}

	SuccessResponse(w, "获取登录信息成功", loginInfo)
}

// RefreshQRCode 刷新二维码
func (h *WeChatHandler) RefreshQRCode(w http.ResponseWriter, r *http.Request) {
	qrData, err := h.wechatService.HelperRefreshQRCode()
	if err != nil {
		log.Printf("刷新二维码失败: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "刷新二维码失败: "+err.Error())
		return
	}

	SuccessResponse(w, "刷新二维码成功", qrData)
}

// GetFriendList 获取好友列表
func (h *WeChatHandler) GetFriendList(w http.ResponseWriter, r *http.Request) {
	friends, err := h.wechatService.HelperGetFriendList()
	if err != nil {
		log.Printf("获取好友列表失败: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "获取好友列表失败: "+err.Error())
		return
	}

	// 直接返回好友数组，保持与其他接口一致（data 类型可为任意）
	SuccessResponse(w, "获取好友列表成功", friends)
}

// GetFriendInfo 获取指定好友信息
func (h *WeChatHandler) GetFriendInfo(w http.ResponseWriter, r *http.Request) {
	wxid := r.URL.Query().Get("wxid")
	if wxid == "" {
		ErrorResponse(w, http.StatusBadRequest, "wxid不能为空")
		return
	}

	friend, err := h.wechatService.HelperGetFriendInfo(wxid)
	if err != nil {
		log.Printf("获取好友信息失败: %v", err)
		ErrorResponse(w, http.StatusInternalServerError, "获取好友信息失败: "+err.Error())
		return
	}

	// 直接返回好友对象
	SuccessResponse(w, "获取好友信息成功", friend)
}
