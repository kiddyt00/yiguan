package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// FetchModels 代理请求供应商 /v1/models 接口，返回可用模型列表
func (h *ModelHandler) FetchModels(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		APIKey   string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Endpoint == "" || req.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "endpoint 和 api_key 不能为空"})
		return
	}

	// 构造 models 接口 URL：去掉 endpoint 尾部斜杠后拼接 /models
	base := strings.TrimRight(req.Endpoint, "/")
	url := base + "/models"

	httpReq, err := http.NewRequestWithContext(r.Context(), "GET", url, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "构造请求失败"})
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "请求供应商失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": fmt.Sprintf("供应商返回 %d: %s", resp.StatusCode, string(body)),
		})
		return
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "解析模型列表失败"})
		return
	}

	ids := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"models": ids})
}

// TestConnection 测试供应商连接 — 调用 /models 接口验证 endpoint + api_key 是否有效
func (h *ModelHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		APIKey   string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Endpoint == "" || req.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "endpoint 和 api_key 不能为空"})
		return
	}

	base := strings.TrimRight(req.Endpoint, "/")
	url := base + "/models"

	httpReq, _ := http.NewRequestWithContext(r.Context(), "GET", url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": false, "error": err.Error(), "latency_ms": elapsed.Milliseconds(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	ok := resp.StatusCode == http.StatusOK

	result := map[string]interface{}{
		"ok":         ok,
		"status":     resp.StatusCode,
		"latency_ms": elapsed.Milliseconds(),
	}
	if !ok {
		result["error"] = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	writeJSON(w, http.StatusOK, result)
}
