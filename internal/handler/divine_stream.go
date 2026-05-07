package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

type DivineStreamHandler struct {
	store  store.Store
	router *llm.Router
}

func NewDivineStreamHandler(st store.Store, router *llm.Router) *DivineStreamHandler {
	return &DivineStreamHandler{store: st, router: router}
}

func (h *DivineStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	remaining, err := h.store.GetRemainingQuota(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询配额失败"})
		return
	}
	if remaining <= 0 {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":           "次数不足",
			"remaining_quota": 0,
		})
		return
	}

	var req divineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入问题"})
		return
	}

	if err := h.store.ConsumeQuota(userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "扣减配额失败"})
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "不支持流式响应"})
		return
	}

	writeSSE := func(event string, data interface{}, flush http.Flusher) {
		jsonData, _ := json.Marshal(map[string]interface{}{"event": event, "data": data})
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
		flush.Flush()
	}

	// 阶段 1: 铜钱抛掷
	lines := engine.CastSixLines()
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	for i, line := range lines {
		result := lineType(line)
		writeSSE("phase", map[string]interface{}{
			"phase": "coins",
			"data":  map[string]interface{}{"throw": i + 1, "label": names[i], "result": result, "yang": line%2 != 0},
		}, flusher)
		time.Sleep(200 * time.Millisecond)
	}

	// 阶段 2: 构建卦象
	primary, changing, positions, master := engine.BuildHexagrams(lines)
	yaoPositions := buildYaoPositions(positions, master)
	yaoDesc := engine.FormatYaoPositions(positions, master)

	writeSSE("phase", map[string]interface{}{
		"phase": "hexagram",
		"data": map[string]interface{}{
			"primary_gua":   primary.Name,
			"changing_gua":  changing.Name,
			"yao_positions": yaoPositions,
		},
	}, flusher)

	// 阶段 3: AI 解卦流式
	llmClient := h.router.Get()
	prompt := llm.BuildPrompt(req.Question, primary.Name, changing.Name, yaoDesc)

	var interpretation strings.Builder
	llmErr := llmClient.DivineStream(prompt, func(chunk string) error {
		interpretation.WriteString(chunk)
		writeSSE("ai", map[string]interface{}{"chunk": chunk}, flusher)
		return nil
	})

	if llmErr != nil {
		writeSSE("error", map[string]interface{}{"error": "解卦服务暂不可用: " + llmErr.Error()}, flusher)
	}

	// 保存历史记录
	history := &store.History{
		UserID:         userID,
		Question:       req.Question,
		PrimaryGua:     primary.Name,
		ChangingGua:    changing.Name,
		YaoPositions:   yaoDesc,
		Interpretation: interpretation.String(),
	}
	if saveErr := h.store.SaveHistory(history); saveErr != nil {
		writeSSE("error", map[string]interface{}{"error": "保存记录失败"}, flusher)
	}

	// 完成
	remaining, _ = h.store.GetRemainingQuota(userID)
	writeSSE("done", map[string]interface{}{
		"id":              history.ID,
		"interpretation":  interpretation.String(),
		"remaining_quota": remaining,
		"created_at":      time.Now().Format(time.RFC3339),
	}, flusher)
}

func lineType(line int) string {
	switch line {
	case 6:
		return "old_yin"
	case 7:
		return "young_yang"
	case 8:
		return "young_yin"
	case 9:
		return "old_yang"
	default:
		return "unknown"
	}
}
