package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// divineCoreResult 起卦中间结果，供同步和流式 handler 共享
type divineCoreResult struct {
	UserID       int64
	Question     string
	Lines        []int
	Primary      *engine.GuaInfo
	Changing     *engine.GuaInfo
	YaoPositions []yaoPos
	YaoDesc      string
}

// divineCore 执行算卦前置逻辑：配额检查 → 扣减 → 起卦 → 返回中间结果
// 返回 nil 时已通过 w 写入了错误响应
func divineCore(w http.ResponseWriter, r *http.Request, st store.Store) *divineCoreResult {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	remaining, err := st.GetRemainingQuota(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询配额失败"})
		return nil
	}
	if remaining <= 0 {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":           "次数不足",
			"remaining_quota": 0,
		})
		return nil
	}

	var req divineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return nil
	}
	if req.Question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入问题"})
		return nil
	}

	if err := st.ConsumeQuota(userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "扣减配额失败"})
		return nil
	}

	linesArr := engine.CastSixLines()
	lines := linesArr[:]
	primary, changing, positions, master := engine.BuildHexagrams(linesArr)
	yaoPositions := buildYaoPositions(positions, master)
	yaoDesc := engine.FormatYaoPositions(positions, master)

	return &divineCoreResult{
		UserID:       userID,
		Question:     req.Question,
		Lines:        lines,
		Primary:      primary,
		Changing:     changing,
		YaoPositions: yaoPositions,
		YaoDesc:      yaoDesc,
	}
}
