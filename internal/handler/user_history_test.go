package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
	"github.com/kiddyt00/yiguan/internal/store/sqlite"
)

func makeAuthHeader(userID int64, secret string) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return "Bearer " + s
}

func setupHandlerTest(t *testing.T) (*http.ServeMux, *sqlite.Store, string) {
	t.Helper()
	secret := "test-secret"
	st, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	mux := http.NewServeMux()
	authMW := middleware.AuthRequired(secret)

	// User routes
	uh := NewUserHandler(st)
	mux.Handle("GET /api/user", authMW(http.HandlerFunc(uh.GetUser)))
	mux.Handle("PUT /api/user", authMW(http.HandlerFunc(uh.UpdateUser)))

	// History routes
	hh := NewHistoryHandler(st)
	mux.Handle("GET /api/history", authMW(http.HandlerFunc(hh.GetHistory)))

	return mux, st, secret
}

func TestGetUserProfile(t *testing.T) {
	mux, st, secret := setupHandlerTest(t)
	u, _ := st.CreateUser("13800138000", "pass", "张三")

	req := httptest.NewRequest("GET", "/api/user", nil)
	req.Header.Set("Authorization", makeAuthHeader(u.ID, secret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	userObj := resp["user"].(map[string]interface{})
	if userObj["nickname"] != "张三" {
		t.Errorf("nickname = %v", userObj["nickname"])

	}
	if _, ok := resp["remaining_quota"]; !ok {
		t.Error("response missing remaining_quota")
	}
}

func TestGetUserProfileUnauthenticated(t *testing.T) {
	mux, _, _ := setupHandlerTest(t)

	req := httptest.NewRequest("GET", "/api/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestUpdateUserProfile(t *testing.T) {
	mux, st, secret := setupHandlerTest(t)
	u, _ := st.CreateUser("13800138000", "pass", "旧名")

	body := `{"nickname":"新名","address":"上海"}`
	req := httptest.NewRequest("PUT", "/api/user", bytes.NewReader([]byte(body)))
	req.Header.Set("Authorization", makeAuthHeader(u.ID, secret))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}

	updated, _ := st.GetUserByID(u.ID)
	if updated.Nickname != "新名" {
		t.Errorf("nickname = %s, want 新名", updated.Nickname)
	}
}

func TestGetHistory(t *testing.T) {
	mux, st, secret := setupHandlerTest(t)
	u, _ := st.CreateUser("13800138000", "pass", "张三")

	// 直接通过 store 存历史
	st.SaveHistory(&store.History{
		UserID: u.ID, Question: "测试问题",
		PrimaryGua: "乾为天", ChangingGua: "天风姤",
		YaoPositions: "上爻", Interpretation: "解读",
	})

	req := httptest.NewRequest("GET", "/api/history?limit=10&offset=0", nil)
	req.Header.Set("Authorization", makeAuthHeader(u.ID, secret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	items := resp["items"].([]interface{})
	if len(items) != 1 {
		t.Errorf("items count = %d, want 1", len(items))
	}
}
