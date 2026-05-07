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

	total, _ := h.store.GetHistoryCount(userID)
	if list == nil {
		list = []store.History{} // 返回空数组而非 null
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": total,
	})
}
