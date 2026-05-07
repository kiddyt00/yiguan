package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kiddyt00/yiguan/internal/store"
)

// AuthHandler 认证相关处理器
type AuthHandler struct {
	mux    *http.ServeMux
	store  store.Store
	secret string
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(st store.Store, jwtSecret string) *AuthHandler {
	h := &AuthHandler{
		mux:    http.NewServeMux(),
		store:  st,
		secret: jwtSecret,
	}
	h.mux.HandleFunc("POST /api/auth/register", h.register)
	h.mux.HandleFunc("POST /api/auth/login", h.login)
	return h
}

// ServeMux 返回内部 mux（用于测试和挂载）
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

	// 新用户赠送 3 个免费 quota
	for i := 0; i < 3; i++ {
		h.store.AddQuota(user.ID, "free")
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成token失败"})
		return
	}

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

	token, err := h.generateToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成token失败"})
		return
	}
	user.Password = "" // 不暴露

	writeJSON(w, http.StatusOK, authResp{User: user, Token: token})
}

func (h *AuthHandler) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secret))
}

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func validPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
