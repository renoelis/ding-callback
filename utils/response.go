package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"ding_call_back/model"
)

// RespondWithError 返回统一格式的错误响应
func RespondWithError(w http.ResponseWriter, code int, message string) {
	response := model.ErrorResponse{
		Success: false,
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码错误响应失败: %v", err)
		http.Error(w, `{"success":false,"code":500,"message":"服务器内部错误"}`, http.StatusInternalServerError)
	}
}

// RespondWithJSON 返回统一格式的成功响应
func RespondWithJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	response := model.SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码成功响应失败: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "服务器内部错误")
	}
}
