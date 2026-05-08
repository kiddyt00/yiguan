package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/kiddyt00/yiguan/internal/store"
)

// AdminHandler 后台管理处理器
type AdminHandler struct {
	store  store.Store
	llmMgr LLMManager
}

// LLMManager LLM 管理器接口（解耦）
type LLMManager interface {
	InvalidateCache()
}

// NewAdminHandler 创建后台处理器
func NewAdminHandler(st store.Store, llmMgr LLMManager) *AdminHandler {
	return &AdminHandler{store: st, llmMgr: llmMgr}
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

// ---- LLM Provider 管理 ----

// ListLLMProviders GET /api/admin/llm
func (h *AdminHandler) ListLLMProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.store.ListLLMProviders()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	if providers == nil {
		providers = []store.LLMProvider{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": providers})
}

// CreateLLMProvider POST /api/admin/llm
func (h *AdminHandler) CreateLLMProvider(w http.ResponseWriter, r *http.Request) {
	var p store.LLMProvider
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if p.Provider == "" || p.Endpoint == "" || p.Model == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider/endpoint/model 为必填"})
		return
	}
	if err := h.store.CreateLLMProvider(&p); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "provider 已存在"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	h.llmMgr.InvalidateCache()
	writeJSON(w, http.StatusCreated, p)
}

// UpdateLLMProvider PUT /api/admin/llm/{id}
func (h *AdminHandler) UpdateLLMProvider(w http.ResponseWriter, r *http.Request) {
	// 检查是否为 /default 结尾 → 转发到 SetDefault
	if strings.HasSuffix(strings.TrimSuffix(r.URL.Path, "/"), "/default") {
		h.SetDefaultLLMProvider(w, r)
		return
	}

	id, err := parseLLMID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 ID"})
		return
	}

	var p store.LLMProvider
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	p.ID = id

	if err := h.store.UpdateLLMProvider(&p); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	h.llmMgr.InvalidateCache()
	writeJSON(w, http.StatusOK, p)
}

// DeleteLLMProvider DELETE /api/admin/llm/{id}
func (h *AdminHandler) DeleteLLMProvider(w http.ResponseWriter, r *http.Request) {
	id, err := parseLLMID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 ID"})
		return
	}
	if err := h.store.DeleteLLMProvider(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	h.llmMgr.InvalidateCache()
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}

// SetDefaultLLMProvider PUT /api/admin/llm/{id}/default
func (h *AdminHandler) SetDefaultLLMProvider(w http.ResponseWriter, r *http.Request) {
	id, err := parseLLMID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 ID"})
		return
	}
	if err := h.store.SetDefaultLLMProvider(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "设置失败"})
		return
	}
	h.llmMgr.InvalidateCache()
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已设为默认"})
}

// parseLLMID 从路径中提取 LLM provider ID
// /api/admin/llm/123 → 123
// /api/admin/llm/123/default → 123
func parseLLMID(path string) (int64, error) {
	trimmed := strings.TrimPrefix(path, "/api/admin/llm/")
	trimmed = strings.TrimSuffix(trimmed, "/default")
	trimmed = strings.TrimSuffix(trimmed, "/")
	return strconv.ParseInt(trimmed, 10, 64)
}
