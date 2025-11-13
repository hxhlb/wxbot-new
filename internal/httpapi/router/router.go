package router

import (
	"net/http"

	"wxbot-new/internal/httpapi/controller"
)

// RegisterRoutes 仅负责注册路由到 mux
func RegisterRoutes(mux *http.ServeMux, sendCtl *controller.SendController) {
	mux.HandleFunc("/send", sendCtl.Send)
}
