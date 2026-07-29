package handler

import (
	"log"
	"net/http"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// InviteHandler 拉新裂变处理器
type InviteHandler struct {
	store store.Store
}

func NewInviteHandler(st store.Store) *InviteHandler {
	return &InviteHandler{store: st}
}

// GetInviteCode 获取/生成邀请码
func (h *InviteHandler) GetInviteCode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	code, err := h.store.GenerateInviteCode(userID)
	if err != nil {
		log.Printf("生成邀请码失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"invite_code": code,
		"share_url":   "https://zgjz.insightj.cn?invite=" + code,
	})
}

// GetInviteProgress 获取拉新进度
func (h *InviteHandler) GetInviteProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	p, err := h.store.GetInviteProgress(userID)
	if err != nil {
		log.Printf("查询拉新进度失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// ClaimReward 手动领取奖励
func (h *InviteHandler) ClaimReward(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	rewarded, err := h.store.CheckAndReward(userID)
	if err != nil {
		log.Printf("发放奖励失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "发放失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rewarded": rewarded,
	})
}

// ServeHTTP 路由分发
func (h *InviteHandler) ServeMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/invite/code", h.GetInviteCode)
	mux.HandleFunc("GET /api/invite/progress", h.GetInviteProgress)
	mux.HandleFunc("POST /api/invite/claim", h.ClaimReward)
	return mux
}
