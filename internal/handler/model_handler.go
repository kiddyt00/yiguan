package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/store"
)

type ModelHandler struct {
	store    store.Store
	onReload func()
}

func NewModelHandler(st store.Store, onReload func()) *ModelHandler {
	return &ModelHandler{store: st, onReload: onReload}
}

func (h *ModelHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.store.ListModels()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取模型列表失败"})
		return
	}
	if models == nil {
		models = []store.LLMModel{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": models, "total": len(models)})
}

func (h *ModelHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var m store.LLMModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if m.Name == "" || m.Endpoint == "" || m.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "名称、endpoint、api_key 不能为空"})
		return
	}
	if m.DisplayName == "" {
		m.DisplayName = m.Provider + " " + m.Name
	}
	m.IsEnabled = 1 // 默认启用
	if err := h.store.CreateModel(&m); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *ModelHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var m store.LLMModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	m.ID = id
	if err := h.store.UpdateModel(&m); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *ModelHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.DeleteModel(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}

func (h *ModelHandler) SetDefaultModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.SetDefaultModel(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "设置默认失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已设为默认"})
}

func (h *ModelHandler) ToggleModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	enabled := r.URL.Query().Get("enabled") == "true"
	if err := h.store.ToggleModel(id, enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "操作失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
}
