package handler

import (
	"net/http"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/llm"
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
	result := divineCore(w, r, h.store)
	if result == nil {
		return
	}

	// 调用 LLM 解卦
	prompt := llm.BuildPrompt(result.Question, result.Primary.Name, result.Changing.Name, result.YaoDesc)
	interpretation, err := h.llm.DivineWithRetry(prompt, 2)
	if err != nil {
		interpretation = "解卦服务暂不可用：" + err.Error()
	}

	// 保存历史
	h.store.SaveHistory(&store.History{
		UserID:         result.UserID,
		Question:       result.Question,
		PrimaryGua:     result.Primary.Name,
		ChangingGua:    result.Changing.Name,
		YaoPositions:   result.YaoDesc,
		Interpretation: interpretation,
	})

	remaining, _ := h.store.GetRemainingQuota(result.UserID)
	writeJSON(w, http.StatusOK, divineResp{
		Primary:        result.Primary,
		Changing:       result.Changing,
		YaoPositions:   result.YaoPositions,
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
