package router

import (
	"net/http"

	"ding_call_back/controller"
)

// SetupRouter 设置路由
func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// 注册钉钉回调处理器
	mux.HandleFunc("/ding/callback/", controller.HandleDingCallback)

	// 注册配置注册处理器
	mux.HandleFunc("/ding/config", controller.HandleConfigRegister)

	return mux
}
