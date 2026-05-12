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

	items := h.wrapWithNickname(userID, list)
	total, _ := h.store.GetHistoryCount(userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// SearchHistory 关键词搜索历史记录
func (h *HistoryHandler) SearchHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	keyword := r.URL.Query().Get("q")
	if keyword == "" {
		// 空关键词 = 返回全部
		h.GetHistory(w, r)
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	list, err := h.store.SearchHistory(userID, keyword, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "搜索失败"})
		return
	}

	items := h.wrapWithNickname(userID, list)
	total, _ := h.store.SearchHistoryCount(userID, keyword)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// GetLatestHistory 获取用户最新一条历史记录
func (h *HistoryHandler) GetLatestHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	history, err := h.store.GetLatestHistory(userID)
	if err != nil {
		if err == store.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "暂无记录"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	user, err := h.store.GetUserByID(userID)
	nickname := ""
	if err == nil && user != nil {
		nickname = user.Nickname
	}

	writeJSON(w, http.StatusOK, historyItem{History: *history, Nickname: nickname})
}

// GetRecentHistory 获取用户最近 N 条历史记录（用于侧边栏）
func (h *HistoryHandler) GetRecentHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	list, err := h.store.GetHistory(userID, limit, 0)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	items := h.wrapWithNickname(userID, list)

	writeJSON(w, http.StatusOK, items)
}

func (h *HistoryHandler) wrapWithNickname(userID int64, list []store.History) []historyItem {
	user, err := h.store.GetUserByID(userID)
	nickname := ""
	if err == nil && user != nil {
		nickname = user.Nickname
	}

	items := make([]historyItem, len(list))
	for i, hh := range list {
		items[i] = historyItem{History: hh, Nickname: nickname}
	}
	if items == nil {
		items = []historyItem{}
	}
	return items
}
