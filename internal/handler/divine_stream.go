package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/llm"
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
	result := divineCore(w, r, h.store)
	if result == nil {
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

	writeSSE := func(event string, data interface{}) {
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
		flusher.Flush()
	}

	// 阶段 1: 铜钱抛掷
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	for i, line := range result.Lines {
		writeSSE("phase", map[string]interface{}{
			"phase": "coins",
			"data":  map[string]interface{}{"throw": i + 1, "label": names[i], "result": lineType(line), "yang": line%2 != 0},
		})
		time.Sleep(200 * time.Millisecond)
	}

	// 阶段 2: 卦象
	writeSSE("phase", map[string]interface{}{
		"phase": "hexagram",
		"data": map[string]interface{}{
			"primary_gua":   result.Primary.Name,
			"changing_gua":  result.Changing.Name,
			"yao_positions": result.YaoPositions,
		},
	})

	// 阶段 3: AI 解卦流式（容错链）
	var interpretation strings.Builder
	var llmErr error
	clients := h.router.GetAllEnabled()
	for _, client := range clients {
		prompt := llm.BuildPrompt(result.Question, result.Primary.Name, result.Changing.Name, result.YaoDesc)
		llmErr = client.DivineStreamWithRetry(prompt, func(chunk string) error {
			interpretation.WriteString(chunk)
			writeSSE("ai", map[string]interface{}{"chunk": chunk})
			return nil
		}, 1)
		if llmErr == nil {
			break
		}
		writeSSE("status", map[string]interface{}{"msg": "正在切换线路..."})
	}

	if llmErr != nil {
		writeSSE("error", map[string]interface{}{"error": "解卦服务暂不可用: " + llmErr.Error()})
	}

	// 保存历史记录
	history := &store.History{
		UserID:         result.UserID,
		Question:       result.Question,
		PrimaryGua:     result.Primary.Name,
		ChangingGua:    result.Changing.Name,
		YaoPositions:   result.YaoDesc,
		Interpretation: interpretation.String(),
	}
	if saveErr := h.store.SaveHistory(history); saveErr != nil {
		writeSSE("error", map[string]interface{}{"error": "保存记录失败"})
	}

	// 完成
	remaining, _ := h.store.GetRemainingQuota(result.UserID)
	writeSSE("done", map[string]interface{}{
		"id":              history.ID,
		"interpretation":  interpretation.String(),
		"remaining_quota": remaining,
		"created_at":      time.Now().Format(time.RFC3339),
	})
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
