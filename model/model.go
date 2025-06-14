package model

import "encoding/json"

// CallbackParams 钉钉回调参数结构
type CallbackParams struct {
	URL    string `json:"url"`
	AESKey string `json:"aes_key"`
	Token  string `json:"token"`
	CorpID string `json:"corpId"`
}

// ConfigRequest 配置请求结构
type ConfigRequest struct {
	Config string          `json:"config"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// CallbackResponse 回调响应结构
type CallbackResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ErrorResponse 统一错误响应格式
type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse 统一成功响应格式
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
