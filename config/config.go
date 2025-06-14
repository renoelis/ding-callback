package config

import (
	"os"
)

// Config 应用配置结构
type Config struct {
	Port string
}

var config *Config

// 初始化默认配置
func init() {
	config = &Config{
		Port: "3014",
	}

	// 如果环境变量中有PORT，则使用环境变量的值
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}
}

// GetPort 获取服务端口
func GetPort() string {
	return config.Port
}
