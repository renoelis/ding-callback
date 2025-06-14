package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"ding_call_back/model"
	"ding_call_back/service"
	"ding_call_back/utils"
)

var (
	dbStore     *service.DBStore
	dbStoreOnce sync.Once
)

// 确保dbStore被正确初始化
func getDBStore() *service.DBStore {
	dbStoreOnce.Do(func() {
		dbStore = service.NewDBStore()
		if dbStore == nil {
			log.Fatal("数据库存储服务初始化失败")
		}
		log.Println("数据库存储服务初始化成功")
	})
	return dbStore
}

// HandleConfigRegister 处理配置注册请求
func HandleConfigRegister(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("读取请求体失败: %v", err))
		return
	}

	// 解析请求体
	var config model.CallbackParams
	if err := json.Unmarshal(body, &config); err != nil {
		log.Printf("解析请求体失败: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("解析请求体失败: %v", err))
		return
	}

	// 验证必要参数
	if config.AESKey == "" || config.Token == "" || config.CorpID == "" {
		missingParams := []string{}
		if config.AESKey == "" {
			missingParams = append(missingParams, "aes_key")
		}
		if config.Token == "" {
			missingParams = append(missingParams, "token")
		}
		if config.CorpID == "" {
			missingParams = append(missingParams, "corpId")
		}
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("缺少必要参数: %s", strings.Join(missingParams, ", ")))
		return
	}

	// 存储配置到数据库
	uuid, err := getDBStore().StoreConfig(config)
	if err != nil {
		log.Printf("存储配置失败: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("存储配置失败: %v", err))
		return
	}

	// 返回成功响应
	data := map[string]string{
		"uuid":         uuid,
		"callback_url": "/ding/callback/" + uuid,
	}

	utils.RespondWithJSON(w, http.StatusOK, "配置注册成功", data)
}
