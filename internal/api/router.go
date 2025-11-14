package api

import (
	"net/http"

	"wxbot-new/internal/config"
)

// Router 路由管理器
type Router struct {
	mux           *http.ServeMux
	configHandler *ConfigHandler
}

// NewRouter 创建路由管理器
func NewRouter(configManager *config.Manager) *Router {
	return &Router{
		mux:           http.NewServeMux(),
		configHandler: NewConfigHandler(configManager),
	}
}

// RegisterRoutes 注册所有路由
func (r *Router) RegisterRoutes() http.Handler {
	// ========== 健康检查 ==========
	r.mux.HandleFunc("/health", r.healthCheck)

	// ========== 配置管理 API ==========
	// 完整配置操作
	r.mux.HandleFunc("/api/config", r.handleConfig)

	// 单个配置项操作
	r.mux.HandleFunc("/api/config/", r.handleConfigItem)

	// 应用中间件链
	handler := Chain(
		RecoveryMiddleware(),    // 最外层: 捕获 panic
		LoggingMiddleware(),     // 日志记录
		CORSMiddleware(),        // CORS 支持
		ContentTypeMiddleware(), // 内容类型检查
	)(r.mux)

	return handler
}

// handleConfig 处理完整配置的 CRUD
func (r *Router) handleConfig(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.configHandler.GetConfig(w, req)
	case http.MethodPost:
		r.configHandler.CreateConfig(w, req)
	case http.MethodPut:
		r.configHandler.UpdateConfig(w, req)
	case http.MethodDelete:
		r.configHandler.DeleteConfig(w, req)
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

// handleConfigItem 处理单个配置项的操作
func (r *Router) handleConfigItem(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.configHandler.GetConfigItem(w, req)
	case http.MethodPut:
		r.configHandler.UpdateConfigItem(w, req)
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

// healthCheck 健康检查
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
