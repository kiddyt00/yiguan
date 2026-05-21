package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kiddyt00/yiguan/internal/store"
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

// fillUserAvatar 计算有效头像：用户未上传自定义头像时回退到微信头像
func fillUserAvatar(u *store.User) {
	if u.Avatar == "" && u.WxAvatar != "" {
		u.Avatar = u.WxAvatar
	}
}
