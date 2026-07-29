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
	totalOrders, _ := h.store.CountAllOrders()
	activeMemberships, _ := h.store.CountActiveMemberships()
	dailyDivineTrend, _ := h.store.GetDailyDivineTrend()
	if dailyDivineTrend == nil {
		dailyDivineTrend = map[string]int64{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_users":         totalUsers,
		"active_users":        activeUsers,
		"today_divines":       todayDivines,
		"total_divines":       totalDivines,
		"ad_watches_today":    adWatchesToday,
		"total_ads_watched":   totalAdWatches,
		"total_orders":        totalOrders,
		"active_memberships":  activeMemberships,
		"daily_divine_trend":  dailyDivineTrend,
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
		fillUserAvatar(&u)
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

// ListOrders 管理员订单列表（GET /api/admin/orders）
func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.store.ListAllOrders(limit, offset)
	if err != nil {
		log.Printf("查询全部订单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	if orders == nil {
		orders = []store.Order{}
	}
	total, _ := h.store.CountAllOrders()

	// 为订单补充用户昵称
	type orderWithUser struct {
		store.Order
		UserName string `json:"user_name"`
	}
	items := make([]orderWithUser, 0, len(orders))
	for _, o := range orders {
		u, err := h.store.GetUserByID(o.UserID)
		uname := ""
		if err == nil {
			uname = u.Nickname
		}
		items = append(items, orderWithUser{Order: o, UserName: uname})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// ListMemberships 管理员会员列表（GET /api/admin/memberships）
func (h *AdminHandler) ListMemberships(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	memberships, err := h.store.ListAllMemberships(limit, offset)
	if err != nil {
		log.Printf("查询全部会员失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	if memberships == nil {
		memberships = []store.Membership{}
	}

	// 补充用户信息
	type membershipWithUser struct {
		store.Membership
		UserName string `json:"user_name"`
	}
	items := make([]membershipWithUser, 0, len(memberships))
	for _, m := range memberships {
		u, err := h.store.GetUserByID(m.UserID)
		uname := ""
		if err == nil {
			uname = u.Nickname
		}
		items = append(items, membershipWithUser{Membership: m, UserName: uname})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}
