package handler

import (
	"bytes"
	"context"
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

// testHTTP 发起请求并返回 status, body, elapsed
func testHTTP(ctx context.Context, method, url, apiKey string, body []byte) (int, string, time.Duration) {
	start := time.Now()
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequestWithContext(ctx, method, url, nil)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	elapsed := time.Since(start).Round(time.Millisecond)
	if err != nil {
		return 0, err.Error(), elapsed
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return resp.StatusCode, strings.TrimSpace(string(respBody)), elapsed
}

// TestConnection 测试供应商连接 — 先试 /models，不支持则回退到 chat/completions
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

	// 方案 1: GET /models
	status, body, elapsed := testHTTP(r.Context(), "GET", base+"/models", req.APIKey, nil)
	if status == http.StatusOK {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true, "method": "/models", "status": status, "latency_ms": elapsed.Milliseconds(),
		})
		return
	}

	// 方案 2: 回退到 POST chat/completions（兼容不支持 /models 的端点）
	chatBody, _ := json.Marshal(map[string]interface{}{
		"model": "qwen-plus",
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
		"max_tokens": 1,
	})
	status, body, elapsed2 := testHTTP(r.Context(), "POST", base+"/chat/completions", req.APIKey, chatBody)

	ok := status == http.StatusOK
	result := map[string]interface{}{
		"ok": ok, "method": "/chat/completions", "status": status, "latency_ms": elapsed2.Milliseconds(),
	}
	if ok {
		result["note"] = "端点不支持 /models 但 chat/completions 可用"
	} else {
		result["error"] = fmt.Sprintf("HTTP %d: %s", status, body)
	}
	writeJSON(w, http.StatusOK, result)
}
