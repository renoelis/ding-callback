package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"ding_call_back/model"
	"ding_call_back/service"
	"ding_call_back/utils"
)

// HandleDingCallback 处理钉钉回调请求
func HandleDingCallback(w http.ResponseWriter, r *http.Request) {
	var params model.CallbackParams
	var err error
	var configFound bool

	// 获取路径（移除前缀 /ding/callback/）
	path := strings.TrimPrefix(r.URL.Path, "/ding/callback/")
	if path == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "缺少必要参数")
		return
	}

	// 首先尝试从数据库获取配置
	params, configFound = getDBStore().GetConfig(path)

	// 如果从数据库找到配置
	if configFound {
		log.Printf("使用数据库配置: UUID=%s, URL=%s, AESKey=%s, Token=%s, CorpID=%s",
			path, params.URL, params.AESKey, params.Token, params.CorpID)
	} else {
		// 尝试从URL路径解析（兼容旧方式）
		decodedPath, err := url.QueryUnescape(path)
		if err != nil {
			log.Printf("URL解码失败: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("URL解码失败: %v", err))
			return
		}

		// Base64解码
		jsonData, err := base64.StdEncoding.DecodeString(decodedPath)
		if err != nil {
			log.Printf("Base64解码失败: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Base64解码失败: %v", err))
			return
		}

		// 解析JSON
		if err := json.Unmarshal(jsonData, &params); err != nil {
			log.Printf("JSON解析失败: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("JSON解析失败: %v", err))
			return
		}

		log.Printf("使用URL路径配置: URL=%s, AESKey=%s, Token=%s, CorpID=%s",
			params.URL, params.AESKey, params.Token, params.CorpID)
	}

	// 检查必要参数
	if params.AESKey == "" || params.Token == "" || params.CorpID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "缺少必要参数: AESKey, Token 或 CorpID")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("读取请求体失败: %v", err))
		return
	}

	// 解析钉钉请求
	signature := r.URL.Query().Get("signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	encrypt := r.URL.Query().Get("encrypt")
	msgSignature := r.URL.Query().Get("msg_signature")

	// 如果msg_signature存在，优先使用它
	if msgSignature != "" {
		signature = msgSignature
	}

	// 处理回调
	decryptedMsg, err := service.ProcessCallback(params, body, signature, timestamp, nonce, encrypt)
	if err != nil {
		log.Printf("处理回调失败: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("处理回调失败: %v", err))
		return
	}

	log.Printf("解密后的消息: %s", decryptedMsg)

	// 创建钉钉加密实例
	crypto := utils.NewDingTalkCrypto(params.Token, params.AESKey, params.CorpID)

	// 加密"success"字符串作为响应
	encryptedMap, err := crypto.GetEncryptMsg("success")
	if err != nil {
		log.Printf("加密响应失败: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("加密响应失败: %v", err))
		return
	}

	// 如果有URL参数，则转发消息并获取转发响应
	var forwardResponseData interface{}
	if params.URL != "" {
		log.Printf("转发消息到: %s", params.URL)

		// 创建POST请求
		resp, err := http.Post(params.URL, "application/json", bytes.NewBuffer([]byte(decryptedMsg)))
		if err != nil {
			log.Printf("转发请求失败: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("转发请求失败: %v", err))
			return
		}
		defer resp.Body.Close()

		// 读取转发响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取转发响应失败: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("读取转发响应失败: %v", err))
			return
		}

		// 尝试解析转发响应为JSON对象
		var jsonData interface{}
		if err := json.Unmarshal(respBody, &jsonData); err != nil {
			// 如果不是JSON，使用字符串
			forwardResponseData = string(respBody)
		} else {
			forwardResponseData = jsonData
		}

		log.Printf("转发完成，接收到转发响应: %s", string(respBody))
	} else {
		log.Printf("无需转发，直接返回加密响应")
	}

	// 创建完整响应，包含钉钉要求的加密响应和转发接口的响应
	fullResponse := map[string]interface{}{
		"encrypt":       encryptedMap["encrypt"],
		"msg_signature": encryptedMap["msg_signature"],
		"nonce":         encryptedMap["nonce"],
		"timeStamp":     encryptedMap["timeStamp"],
	}

	// 如果有转发响应，添加到data字段
	if forwardResponseData != nil {
		fullResponse["data"] = forwardResponseData
	}

	// 返回完整响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fullResponse); err != nil {
		log.Printf("序列化响应失败: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("序列化响应失败: %v", err))
		return
	}
}
