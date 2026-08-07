package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// AvatarHandler 用户头像上传与静态服务
// 小程序端使用微信「头像昵称填写能力」：<button open-type="chooseAvatar"> 拿临时路径 → 上传本接口
type AvatarHandler struct {
	store store.Store
	dir   string // 头像存储目录（默认 /data/avatars，docker 挂载 ./data）
}

// NewAvatarHandler 创建头像处理器
func NewAvatarHandler(st store.Store, dir string) *AvatarHandler {
	if dir == "" {
		dir = "/data/avatars"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("创建头像目录失败: %v", err)
	}
	return &AvatarHandler{store: st, dir: dir}
}

// allowedAvatarExt 头像文件扩展名白名单
func allowedAvatarExt(name string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg":
		return ".jpg", true
	case ".png":
		return ".png", true
	case ".webp":
		return ".webp", true
	}
	return "", false
}

// UploadAvatar 上传头像（POST /api/upload/avatar，multipart 字段 avatar）
func (h *AvatarHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	// 限制请求体 3MB（防内存滥用）
	r.Body = http.MaxBytesReader(w, r.Body, 3<<20)
	if err := r.ParseMultipartForm(3 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "图片过大或格式错误"})
		return
	}
	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少头像文件"})
		return
	}
	defer file.Close()

	ext, ok := allowedAvatarExt(header.Filename)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 jpg/png/webp 图片"})
		return
	}
	if header.Size > 2<<20 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "图片不能超过 2MB"})
		return
	}

	// 覆盖式保存：同一用户固定文件名，避免残留旧文件
	filename := fmt.Sprintf("u%d%s", userID, ext)
	dst := filepath.Join(h.dir, filename)
	out, err := os.Create(dst)
	if err != nil {
		log.Printf("创建头像文件失败 user=%d: %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		log.Printf("写入头像文件失败 user=%d: %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
		return
	}

	// 返回可访问 URL（按请求域名拼，避免写死）
	scheme := "https"
	if r.TLS == nil && strings.HasPrefix(r.Host, "localhost") {
		scheme = "http"
	}
	avatarURL := fmt.Sprintf("%s://%s/api/avatars/%s", scheme, r.Host, filename)
	if err := h.store.UpdateUserAvatar(userID, avatarURL); err != nil {
		log.Printf("更新头像字段失败 user=%d: %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"avatar": avatarURL})
}

// ServeAvatar 静态服务头像文件（GET /api/avatars/{name}）
func (h *AvatarHandler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	// 防路径穿越：只允许纯文件名
	if name == "" || strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(h.dir, name))
}
