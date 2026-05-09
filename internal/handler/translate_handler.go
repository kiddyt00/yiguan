package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// TranslateHandler 翻译 AI 解读内容
type TranslateHandler struct {
	store  store.Store
	router *llm.Router
}

// NewTranslateHandler 创建翻译处理器
func NewTranslateHandler(st store.Store, router *llm.Router) *TranslateHandler {
	return &TranslateHandler{store: st, router: router}
}

// ServeHTTP GET 查缓存的翻译，POST 生成/返回翻译
func (h *TranslateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的历史记录ID"})
		return
	}

	target := r.URL.Query().Get("target")
	if target != "zh" && target != "en" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "target 必须为 zh 或 en"})
		return
	}

	// 验证历史记录所有权
	history, err := h.store.GetHistoryByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "历史记录不存在"})
		return
	}
	if history.UserID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权访问此记录"})
		return
	}

	if history.Lang == target {
		// 目标语言和原文一致，直接返回原文
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"content": history.Interpretation,
			"cached":  true,
			"source":  "original",
		})
		return
	}

	// 查 DB 缓存
	cached, err := h.store.GetTranslation(id, target)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"content": cached.Content,
			"cached":  true,
		})
		return
	}

	// GET 请求：仅查缓存，不生成
	if r.Method == "GET" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "翻译不存在，请使用 POST 请求生成"})
		return
	}

	// POST 请求：调 LLM 生成翻译
	prompt := llm.BuildTranslatePrompt(history.Interpretation, target)

	clients := h.router.GetAllEnabled()
	var translation string
	var lastErr error
	for _, client := range clients {
		translation, lastErr = client.DivineWithRetry(prompt, 1)
		if lastErr == nil {
			break
		}
		log.Printf("翻译调用失败: %v, 尝试下一个", lastErr)
	}

	if lastErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "翻译服务暂不可用: " + lastErr.Error()})
		return
	}

	// 存入 DB
	if saveErr := h.store.SaveTranslation(&store.Translation{
		HistoryID: id,
		Lang:      target,
		Content:   translation,
	}); saveErr != nil {
		log.Printf("保存翻译缓存失败: %v", saveErr)
		// 已生成内容，仍然返回
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"content": translation,
		"cached":  false,
	})
}
