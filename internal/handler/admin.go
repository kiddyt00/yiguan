package handler

import (
	"encoding/json"
	"log"
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
	dailyDivineTrend, _ := h.store.GetDailyDivineTrend()
	if dailyDivineTrend == nil {
		dailyDivineTrend = map[string]int64{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_users":        totalUsers,
		"active_users":       activeUsers,
		"today_divines":      todayDivines,
		"total_divines":      totalDivines,
		"ad_watches_today":   adWatchesToday,
		"total_ads_watched":  totalAdWatches,
		"daily_divine_trend": dailyDivineTrend,
	})
}

// ListUsers 用户列表
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	keyword := r.URL.Query().Get("keyword")
	var users []store.User
	var total int64
	if keyword != "" {
		users, _ = h.store.SearchUsers(keyword, limit, offset)
		total, _ = h.store.SearchUsersCount(keyword)
	} else {
		users, _ = h.store.ListUsers(limit, offset)
		total, _ = h.store.GetTotalUsers()
	}
	if users == nil {
		users = []store.User{}
	}
	// 为每个用户补充配额信息
	type userWithQuota struct {
		store.User
		RemainingQuota int `json:"remaining_quota"`
	}
	items := make([]userWithQuota, 0, len(users))
	for _, u := range users {
		q, _ := h.store.GetUserQuota(u.ID)
		items = append(items, userWithQuota{User: u, RemainingQuota: q})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
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
	if req.Delta == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "delta 不能为 0"})
		return
	}
	// 检查用户是否存在
	if _, err := h.store.GetUserByID(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	if err := h.store.UpdateUserQuota(id, req.Delta); err != nil {
		log.Printf("调整配额失败 user=%d delta=%d: %v", id, req.Delta, err)
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
