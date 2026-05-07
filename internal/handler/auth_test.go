package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiddyt00/yiguan/internal/store/sqlite"
)

func setupAuthTest(t *testing.T) (*AuthHandler, *sqlite.Store) {
	t.Helper()
	st, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	h := NewAuthHandler(st, "test-secret")
	return h, st
}

func TestRegisterSuccess(t *testing.T) {
	h, _ := setupAuthTest(t)

	body := `{"phone":"13800138000","password":"123456","nickname":"张三"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeMux().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("response missing token")
	}
	if resp["user"] == nil {
		t.Error("response missing user")
	}
}

func TestRegisterDuplicatePhone(t *testing.T) {
	h, _ := setupAuthTest(t)

	body := `{"phone":"13800138000","password":"123456","nickname":"张三"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeMux().ServeHTTP(w, req)

	// 第二次
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte(body)))
	req2.Header.Set("Content-Type", "application/json")
	h.ServeMux().ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w2.Code)
	}
}

func TestRegisterShortPassword(t *testing.T) {
	h, _ := setupAuthTest(t)

	body := `{"phone":"13800138000","password":"123","nickname":"张三"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeMux().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	h, st := setupAuthTest(t)

	st.CreateUser("13800138000", "mypassword", "张三")

	body := `{"phone":"13800138000","password":"mypassword"}`
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeMux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == "" {
		t.Error("response missing token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h, st := setupAuthTest(t)

	st.CreateUser("13800138000", "mypassword", "张三")

	body := `{"phone":"13800138000","password":"wrongpass"}`
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeMux().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRegisterAutoGrantsQuota(t *testing.T) {
	h, st := setupAuthTest(t)

	body := `{"phone":"13900139000","password":"123456","nickname":"李四"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeMux().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %s", w.Body.String())
	}

	// 解析返回的 user ID
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	userMap := resp["user"].(map[string]interface{})
	userID := int64(userMap["id"].(float64))

	remaining, _ := st.GetRemainingQuota(userID)
	if remaining != 3 {
		t.Errorf("remaining quota = %d, want 3", remaining)
	}
}
