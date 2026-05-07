package handler

import (
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/store"
)

// AdminHandler 后台管理处理器
type AdminHandler struct {
	store store.Store
}

// NewAdminHandler 创建后台处理器
func NewAdminHandler(st store.Store) *AdminHandler {
	return &AdminHandler{store: st}
}

// Dashboard 仪表盘数据
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	totalUsers, _ := h.store.GetTotalUsers()
	todayDivines, _ := h.store.GetTodayDivineCount()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_users":   totalUsers,
		"today_divines": todayDivines,
	})
}

// ListUsers 用户列表
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	users, _ := h.store.ListUsers(limit, offset)
	total, _ := h.store.GetTotalUsers()
	if users == nil {
		users = []store.User{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": users,
		"total": total,
	})
}
