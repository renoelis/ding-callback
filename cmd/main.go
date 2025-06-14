package main

import (
	"fmt"
	"log"
	"net/http"

	"ding_call_back/config"
	"ding_call_back/router"
)

func main() {
	// 初始化数据库
	if err := config.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 设置路由
	r := router.SetupRouter()

	// 获取配置的端口
	port := config.GetPort()

	// 启动服务
	fmt.Printf("服务已启动，监听端口：%s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
