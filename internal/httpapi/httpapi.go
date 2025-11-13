package httpapi

import (
	"net/http"

	"wxbot-new/internal/httpapi/controller"
	"wxbot-new/internal/httpapi/router"
	"wxbot-new/internal/wxsvc"
)

// RegisterRoutes 组装路由：router 只做注册，controller 承担处理逻辑
func RegisterRoutes(mux *http.ServeMux, svc *wxsvc.Service) {
	// 构造 controller，并交给 router 注册
	sendCtl := controller.NewSendController(svc)
	router.RegisterRoutes(mux, sendCtl)
}
