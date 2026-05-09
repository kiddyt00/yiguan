package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// checkPassword 验证 bcrypt 密码
func checkPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// getLang 从 Accept-Language header 提取语言（zh 或 en）
func getLang(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")
	if strings.HasPrefix(strings.ToLower(lang), "en") {
		return "en"
	}
	return "zh"
}
