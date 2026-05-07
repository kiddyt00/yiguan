package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

type AdHandler struct {
	store store.Store
}

func NewAdHandler(st store.Store) *AdHandler {
	return &AdHandler{store: st}
}

// === Admin API ===

func (h *AdHandler) ListAds(w http.ResponseWriter, r *http.Request) {
	ads, err := h.store.ListAds()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取广告列表失败"})
		return
	}
	if ads == nil {
		ads = []store.Ad{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": ads, "total": len(ads)})
}

func (h *AdHandler) CreateAd(w http.ResponseWriter, r *http.Request) {
	var ad store.Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if ad.Name == "" || ad.ContentURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "名称和 content_url 不能为空"})
		return
	}
	if ad.WatchDuration <= 0 {
		ad.WatchDuration = 30
	}
	if ad.RewardQuota <= 0 {
		ad.RewardQuota = 1
	}
	if err := h.store.CreateAd(&ad); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	writeJSON(w, http.StatusCreated, ad)
}

func (h *AdHandler) UpdateAd(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var ad store.Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	ad.ID = id
	if err := h.store.UpdateAd(&ad); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	writeJSON(w, http.StatusOK, ad)
}

func (h *AdHandler) DeleteAd(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.DeleteAd(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}

func (h *AdHandler) ToggleAd(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	enabled := r.URL.Query().Get("enabled") == "true"
	if err := h.store.ToggleAd(id, enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "操作失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
}

func (h *AdHandler) GetAdStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetAdStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取统计失败"})
		return
	}
	if stats == nil {
		stats = []store.AdStat{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": stats})
}

// === 用户端 API ===

func (h *AdHandler) ListActiveAds(w http.ResponseWriter, r *http.Request) {
	ads, err := h.store.ListActiveAds()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取广告失败"})
		return
	}
	if ads == nil {
		ads = []store.Ad{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": ads})
}

func (h *AdHandler) StartWatch(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	adID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	ad, err := h.store.GetAdByID(adID)
	if err != nil || ad.IsEnabled == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "广告不存在"})
		return
	}

	rec := &store.AdRecord{
		UserID:        userID,
		AdID:          adID,
		WatchDuration: 0,
		Status:        "watching",
		Rewarded:      0,
	}
	if err := h.store.CreateAdRecord(rec); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建记录失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"record_id": rec.ID, "watch_duration": ad.WatchDuration})
}

func (h *AdHandler) CompleteWatch(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	adID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	var req struct {
		Duration int `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	ad, err := h.store.GetAdByID(adID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "广告不存在"})
		return
	}

	if req.Duration < ad.WatchDuration {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("观看时长不足，还需 %d 秒", ad.WatchDuration-req.Duration),
		})
		return
	}

	rec, err := h.store.GetAdRecord(userID, adID)
	if err != nil || rec.Status != "watching" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无进行中的观看记录"})
		return
	}

	rec.WatchDuration = req.Duration
	rec.Status = "completed"
	rec.Rewarded = 1
	if err := h.store.UpdateAdRecord(rec); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新记录失败"})
		return
	}

	for i := 0; i < ad.RewardQuota; i++ {
		h.store.AddQuota(userID, "ad")
	}

	remaining, _ := h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rewarded":        ad.RewardQuota,
		"remaining_quota": remaining,
	})
}
