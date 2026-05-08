package handler

import (
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// HistoryHandler 历史记录处理器
type HistoryHandler struct {
	store store.Store
}

// historyItem 历史记录响应（含用户昵称）
type historyItem struct {
	store.History
	Nickname string `json:"nickname"`
}

// NewHistoryHandler 创建历史处理器
func NewHistoryHandler(st store.Store) *HistoryHandler {
	return &HistoryHandler{store: st}
}

// GetHistory 获取当前用户历史记录（分页）
func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	list, err := h.store.GetHistory(userID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	// 获取用户昵称
	user, err := h.store.GetUserByID(userID)
	nickname := ""
	if err == nil && user != nil {
		nickname = user.Nickname
	}

	// 包装历史记录，附加昵称
	items := make([]historyItem, len(list))
	for i, hh := range list {
		items[i] = historyItem{History: hh, Nickname: nickname}
	}
	if items == nil {
		items = []historyItem{}
	}

	total, _ := h.store.GetHistoryCount(userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}
