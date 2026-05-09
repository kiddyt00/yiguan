package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kiddyt00/yiguan/internal/store"
)

// AuthHandler 认证相关处理器
type AuthHandler struct {
	mux      *http.ServeMux
	store    store.Store
	secret   string
	wxAppID  string
	wxSecret string
}

// smsCode 内存存储短信验证码
var (
	smsCodes   = sync.Map{} // phone -> code
	smsExpires = sync.Map{} // phone -> expireTime
)

// NewAuthHandler 创建认证处理器
func NewAuthHandler(st store.Store, jwtSecret string) *AuthHandler {
	h := &AuthHandler{
		mux:    http.NewServeMux(),
		store:  st,
		secret: jwtSecret,
	}
	h.mux.HandleFunc("POST /api/auth/register", h.register)
	h.mux.HandleFunc("POST /api/auth/login", h.login)
	h.mux.HandleFunc("POST /api/auth/wechat-login", h.wechatLogin)
	h.mux.HandleFunc("POST /api/auth/sms-send", h.smsSend)
	h.mux.HandleFunc("POST /api/auth/sms-login", h.smsLogin)
	return h
}

// SetWechatConfig 设置微信小程序配置
func (h *AuthHandler) SetWechatConfig(appID, appSecret string) {
	h.wxAppID = appID
	h.wxSecret = appSecret
}

// ServeMux 返回内部 mux
func (h *AuthHandler) ServeMux() *http.ServeMux {
	return h.mux
}

type registerReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type authResp struct {
	User  *store.User `json:"user"`
	Token string      `json:"token"`
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if !validPhone(req.Phone) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "手机号格式不正确"})
		return
	}
	if len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "密码至少6位"})
		return
	}
	if req.Nickname == "" {
		req.Nickname = "易友" + req.Phone[len(req.Phone)-4:]
	}

	user, err := h.store.CreateUser(req.Phone, req.Password, req.Nickname)
	if err != nil {
		log.Printf("注册失败: %v", err)
		writeJSON(w, http.StatusConflict, map[string]string{"error": "该手机号已注册"})
		return
	}

	for i := 0; i < 3; i++ {
		h.store.AddQuota(user.ID, "free")
	}

	token, _ := h.generateToken(user.ID, "user")
	writeJSON(w, http.StatusCreated, authResp{User: user, Token: token})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	user, err := h.store.GetUserByPhone(req.Phone)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "手机号或密码错误"})
		return
	}
	if err := checkPassword(user.Password, req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "手机号或密码错误"})
		return
	}
	if user.IsActive == 0 {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "账号已被禁用"})
		return
	}

	token, _ := h.generateToken(user.ID, user.Role)
	user.Password = ""
	writeJSON(w, http.StatusOK, authResp{User: user, Token: token})
}

// wechatLogin 微信小程序登录
type wechatLoginReq struct {
	Code string `json:"code"`
}

type wechatResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (h *AuthHandler) wechatLogin(w http.ResponseWriter, r *http.Request) {
	var req wechatLoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 code"})
		return
	}

	openid, err := h.exchangeWechatCode(req.Code)
	if err != nil {
		log.Printf("微信 code 换 openid 失败: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "微信登录失败: " + err.Error()})
		return
	}

	user, err := h.store.GetUserByOpenID(openid)
	if err != nil {
		// 新用户，自动注册
		user, err = h.store.CreateUserByOpenID(openid, "微信用户")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建用户失败"})
			return
		}
	}

	if user.IsActive == 0 {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "账号已被禁用"})
		return
	}

	token, _ := h.generateToken(user.ID, user.Role)
	user.Password = ""
	writeJSON(w, http.StatusOK, authResp{User: user, Token: token})
}

func (h *AuthHandler) exchangeWechatCode(code string) (string, error) {
	if h.wxAppID == "" || h.wxSecret == "" {
		return "", fmt.Errorf("微信小程序未配置")
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		h.wxAppID, h.wxSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()

	var wr wechatResp
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return "", fmt.Errorf("解析微信响应失败: %w", err)
	}
	if wr.ErrCode != 0 {
		return "", fmt.Errorf("微信错误 %d: %s", wr.ErrCode, wr.ErrMsg)
	}
	return wr.OpenID, nil
}

// smsSend 发送短信验证码
func (h *AuthHandler) smsSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if !validPhone(req.Phone) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "手机号格式不正确"})
		return
	}

	// 60秒内不重复发送
	if exp, ok := smsExpires.Load(req.Phone); ok {
		if time.Now().Before(exp.(time.Time).Add(-4*time.Minute + 60*time.Second)) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "60秒后再试"})
			return
		}
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	smsCodes.Store(req.Phone, code)
	smsExpires.Store(req.Phone, time.Now().Add(5*time.Minute))

	// TODO: 接入真实短信服务（阿里云/腾讯云）
	log.Printf("📱 短信验证码 [%s]: %s", req.Phone, code)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "验证码已发送（开发模式见日志）",
	})
}

// smsLogin 短信验证码登录（无账号则自动注册）
func (h *AuthHandler) smsLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	// 验证码校验
	savedCode, ok := smsCodes.Load(req.Phone)
	if !ok || savedCode.(string) != strings.TrimSpace(req.Code) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "验证码错误或已过期"})
		return
	}

	// 检查过期
	if exp, ok := smsExpires.Load(req.Phone); ok {
		if time.Now().After(exp.(time.Time)) {
			smsCodes.Delete(req.Phone)
			smsExpires.Delete(req.Phone)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "验证码已过期"})
			return
		}
	}

	// 删除已使用的验证码
	smsCodes.Delete(req.Phone)
	smsExpires.Delete(req.Phone)

	// 查找或创建用户
	user, err := h.store.GetUserByPhone(req.Phone)
	if err != nil {
		user, err = h.store.CreateUser(req.Phone, "", "易友"+req.Phone[len(req.Phone)-4:])
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建用户失败"})
			return
		}
		for i := 0; i < 3; i++ {
			h.store.AddQuota(user.ID, "free")
		}
	}

	if user.IsActive == 0 {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "账号已被禁用"})
		return
	}

	token, _ := h.generateToken(user.ID, user.Role)
	user.Password = ""
	writeJSON(w, http.StatusOK, authResp{User: user, Token: token})
}

func (h *AuthHandler) generateToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secret))
}

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func validPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
