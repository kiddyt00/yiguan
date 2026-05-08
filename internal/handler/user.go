package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// UserHandler 用户信息处理器
type UserHandler struct {
	store store.Store
}

// NewUserHandler 创建用户处理器
func NewUserHandler(st store.Store) *UserHandler {
	return &UserHandler{store: st}
}

type updateUserReq struct {
	Nickname string `json:"nickname"`
	Address  string `json:"address"`
}

// GetUser 获取当前用户信息
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	user, err := h.store.GetUserByID(userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	user.Password = ""
	remaining, _ := h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":            user,
		"remaining_quota": remaining,
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	if err := h.store.UpdateUser(userID, req.Nickname, req.Address); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}

	user, _ := h.store.GetUserByID(userID)
	user.Password = ""
	remaining, _ := h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":            user,
		"remaining_quota": remaining,
	})
}
