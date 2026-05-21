package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

type UserHandler struct {
	store    store.Store
	wxAppID  string
	wxSecret string
}

func NewUserHandler(st store.Store) *UserHandler {
	return &UserHandler{store: st}
}

func (h *UserHandler) SetWechatConfig(appID, appSecret string) {
	h.wxAppID = appID
	h.wxSecret = appSecret
}

type updateUserReq struct {
	Nickname string `json:"nickname"`
	Address  string `json:"address"`
}

type bindWechatReq struct {
	Code string `json:"code"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	user, err := h.store.GetUserByID(userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	fillUserAvatar(user)
	user.Password = ""
	remaining, _ := h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":            user,
		"remaining_quota": remaining,
	})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if err := h.store.UpdateUser(userID, req.Nickname, req.Address); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	user, _ := h.store.GetUserByID(userID)
	fillUserAvatar(user)
	user.Password = ""
	remaining, _ := h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":            user,
		"remaining_quota": remaining,
	})
}

func (h *UserHandler) BindWechat(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	var req bindWechatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 code"})
		return
	}
	if h.wxAppID == "" || h.wxSecret == "" {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "微信登录未配置"})
		return
	}
	openid, err := exchangeWechatCode(h.wxAppID, h.wxSecret, req.Code)
	if err != nil {
		log.Printf("微信 code 换 openid 失败: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "微信登录失败: " + err.Error()})
		return
	}
	existing, _ := h.store.GetUserByOpenID(openid)
	if existing != nil && existing.ID != userID {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "该微信已被其他账号绑定"})
		return
	}
	if err := h.store.UpdateUserOpenID(userID, openid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "绑定失败"})
		return
	}
	user, _ := h.store.GetUserByID(userID)
	user.Password = ""
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user, "bound": true, "openid": openid})
}

func exchangeWechatCode(appID, secret, code string) (string, error) {
	if appID == "" || secret == "" {
		return "", fmt.Errorf("微信小程序未配置")
	}
	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, secret, code))
	if err != nil {
		return "", fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()
	var wr struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return "", fmt.Errorf("解析微信响应失败: %w", err)
	}
	if wr.ErrCode != 0 {
		return "", fmt.Errorf("微信错误 %d: %s", wr.ErrCode, wr.ErrMsg)
	}
	return wr.OpenID, nil
}
