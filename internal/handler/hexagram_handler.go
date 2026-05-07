package handler

import (
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/store"
)

type HexagramHandler struct {
	store store.Store
}

func NewHexagramHandler(st store.Store) *HexagramHandler {
	return &HexagramHandler{store: st}
}

func (h *HexagramHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	items, err := h.store.ListAllHistory(limit, offset, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取记录失败"})
		return
	}
	if items == nil {
		items = []store.History{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "total": len(items)})
}

func (h *HexagramHandler) GetHistoryDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	history, err := h.store.GetHistoryByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "记录不存在"})
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (h *HexagramHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.DeleteHistory(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}
