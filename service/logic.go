package service

import (
	"ding_call_back/model"
	"ding_call_back/utils"
	"encoding/json"
)

// ProcessCallback 处理钉钉回调
func ProcessCallback(params model.CallbackParams, body []byte, signature, timestamp, nonce, encrypt string) (string, error) {
	// 创建钉钉加密实例
	crypto := utils.NewDingTalkCrypto(params.Token, params.AESKey, params.CorpID)

	// 如果没有加密参数，可能是明文请求
	var decryptedMsg string
	var err error

	// 如果URL中没有encrypt参数，尝试从请求体中获取
	if encrypt == "" {
		// 尝试从请求体中解析encrypt字段
		var requestBody struct {
			Encrypt string `json:"encrypt"`
			Config  string `json:"config,omitempty"`
		}
		if err := json.Unmarshal(body, &requestBody); err == nil && requestBody.Encrypt != "" {
			encrypt = requestBody.Encrypt
		}
	}

	if encrypt != "" {
		// 解密消息
		decryptedMsg, err = crypto.GetDecryptMsg(signature, timestamp, nonce, encrypt)
		if err != nil {
			return "", err
		}
	} else {
		// 明文消息
		decryptedMsg = string(body)
	}

	return decryptedMsg, nil
}
