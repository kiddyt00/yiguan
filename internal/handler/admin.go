package handler

import (
	"encoding/json"
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
	totalDivines, _ := h.store.GetTotalDivineCount()
	activeUsers, _ := h.store.GetActiveUserCount()
	adWatchesToday, _ := h.store.GetTodayAdWatchCount()
	totalAdWatches, _ := h.store.GetTotalAdWatchCount()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_users":       totalUsers,
		"active_users":      activeUsers,
		"today_divines":     todayDivines,
		"total_divines":     totalDivines,
		"ad_watches_today":  adWatchesToday,
		"total_ads_watched": totalAdWatches,
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

// ToggleUser 启用/禁用用户
func (h *AdminHandler) ToggleUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.store.GetUserByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	if err := h.store.ToggleUser(id, user.IsActive == 0); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "操作失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
}

// AdjustUserQuota 手动调整配额
func (h *AdminHandler) AdjustUserQuota(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Delta int `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if err := h.store.UpdateUserQuota(id, req.Delta); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "调整配额失败"})
		return
	}
	remaining, _ := h.store.GetUserQuota(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{"remaining_quota": remaining})
}

// GetUserHistory 查看某用户的起卦记录
func (h *AdminHandler) GetUserHistory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.store.GetUserHistory(id, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取记录失败"})
		return
	}
	if items == nil {
		items = []store.History{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}
