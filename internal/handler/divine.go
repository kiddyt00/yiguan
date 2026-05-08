package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// DivineHandler 算卦处理器
type DivineHandler struct {
	store store.Store
	llm   *llm.Client
}

// NewDivineHandler 创建算卦处理器
func NewDivineHandler(st store.Store, llmClient *llm.Client) *DivineHandler {
	return &DivineHandler{store: st, llm: llmClient}
}

type divineReq struct {
	Question string `json:"question"`
}

type divineResp struct {
	Primary        *engine.GuaInfo `json:"primary"`
	Changing       *engine.GuaInfo `json:"changing"`
	YaoPositions   []yaoPos        `json:"yao_positions"`
	Interpretation string          `json:"interpretation"`
	RemainingQuota int             `json:"remaining_quota"`
}

type yaoPos struct {
	Position int    `json:"position"` // 0-5
	Label    string `json:"label"`    // 初爻~上爻
	IsMaster bool   `json:"is_master"`
}

// ServeHTTP 处理算卦请求
func (h *DivineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	// 检查 quota
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

	// 扣减 quota
	if err := h.store.ConsumeQuota(userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "扣减配额失败"})
		return
	}

	// 起卦
	lines := engine.CastSixLines()
	primary, changing, positions, master := engine.BuildHexagrams(lines)

	// 格式化变爻
	yaoPositions := buildYaoPositions(positions, master)
	yaoDesc := engine.FormatYaoPositions(positions, master)

	// 调用 LLM 解卦
	prompt := llm.BuildPrompt(req.Question, primary.Name, changing.Name, yaoDesc)
	interpretation, err := h.llm.DivineWithRetry(prompt, 2)
	if err != nil {
		interpretation = "解卦服务暂不可用：" + err.Error()
	}

	// 保存历史
	h.store.SaveHistory(&store.History{
		UserID:         userID,
		Question:       req.Question,
		PrimaryGua:     primary.Name,
		ChangingGua:    changing.Name,
		YaoPositions:   yaoDesc,
		Interpretation: interpretation,
	})

	remaining, _ = h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, divineResp{
		Primary:        primary,
		Changing:       changing,
		YaoPositions:   yaoPositions,
		Interpretation: interpretation,
		RemainingQuota: remaining,
	})
}

func buildYaoPositions(positions []int, master int) []yaoPos {
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	var result []yaoPos
	for _, p := range positions {
		result = append(result, yaoPos{
			Position: p,
			Label:    names[p],
			IsMaster: p == master,
		})
	}
	return result
}
