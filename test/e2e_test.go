package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiddyt00/yiguan/internal/handler"
	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store/sqlite"
)

func newE2EServer(t *testing.T) (*httptest.Server, *sqlite.Store, string) {
	t.Helper()
	secret := "e2e-test-secret"
	st, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	llmClient := llm.New(llm.Config{APIKey: "", Endpoint: "https://example.com", Model: "test"})
	authMW := middleware.AuthRequired(secret)

	mux := http.NewServeMux()

	// CORS
	wcors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}

	ah := handler.NewAuthHandler(st, secret)
	mux.Handle("/api/auth/", wcors(ah.ServeMux()))

	uh := handler.NewUserHandler(st)
	hh := handler.NewHistoryHandler(st)
	dh := handler.NewDivineHandler(st, llmClient)
	admin := handler.NewAdminHandler(st, nil) // e2e test 不需要 LLM 管理器

	mux.Handle("GET /api/user", authMW(wcors(http.HandlerFunc(uh.GetUser))))
	mux.Handle("PUT /api/user", authMW(wcors(http.HandlerFunc(uh.UpdateUser))))
	mux.Handle("GET /api/history", authMW(wcors(http.HandlerFunc(hh.GetHistory))))
	mux.Handle("POST /api/divine", authMW(wcors(dh)))
	mux.Handle("GET /api/admin/dashboard", authMW(wcors(http.HandlerFunc(admin.Dashboard))))
	mux.Handle("GET /api/admin/users", authMW(wcors(http.HandlerFunc(admin.ListUsers))))

	return httptest.NewServer(mux), st, secret
}

func postJSON(url string, body map[string]interface{}, token string) (*http.Response, error) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func getWithAuth(url, token string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func TestFullUserFlow(t *testing.T) {
	ts, st, _ := newE2EServer(t)
	defer ts.Close()

	// 1. 注册
	resp, _ := postJSON(ts.URL+"/api/auth/register", map[string]interface{}{
		"phone": "13800138000", "password": "test1234", "nickname": "测试",
	}, "")
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("register: %d %s", resp.StatusCode, body)
	}
	var auth struct {
		Token string `json:"token"`
		User  struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	json.NewDecoder(resp.Body).Decode(&auth)
	resp.Body.Close()

	token := auth.Token
	userID := auth.User.ID

	// 2. 检查 quota = 3
	rem, _ := st.GetRemainingQuota(userID)
	if rem != 3 {
		t.Errorf("initial quota = %d, want 3", rem)
	}

	// 3. 算卦 2 次
	for i := 0; i < 2; i++ {
		resp, _ = postJSON(ts.URL+"/api/divine", map[string]interface{}{
			"question": "测试问题",
		}, token)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("divine %d: %d %s", i+1, resp.StatusCode, body)
		}
		resp.Body.Close()
	}

	rem, _ = st.GetRemainingQuota(userID)
	if rem != 1 {
		t.Errorf("quota after 2 = %d, want 1", rem)
	}

	// 4. 第 3 次
	resp, _ = postJSON(ts.URL+"/api/divine", map[string]interface{}{"question": "第三次"}, token)
	resp.Body.Close()

	// 5. 第 4 次 → 402
	resp, _ = postJSON(ts.URL+"/api/divine", map[string]interface{}{"question": "第四次"}, token)
	if resp.StatusCode != http.StatusPaymentRequired {
		t.Errorf("4th divine = %d, want 402", resp.StatusCode)
	}
	resp.Body.Close()

	// 6. 历史记录 = 3
	resp, _ = getWithAuth(ts.URL+"/api/history?limit=10", token)
	var hist struct {
		Items []interface{} `json:"items"`
	}
	json.NewDecoder(resp.Body).Decode(&hist)
	resp.Body.Close()
	if len(hist.Items) != 3 {
		t.Errorf("history count = %d, want 3", len(hist.Items))
	}

	// 7. Admin
	resp, _ = getWithAuth(ts.URL+"/api/admin/dashboard", token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("admin dashboard = %d", resp.StatusCode)
	}
	resp.Body.Close()

	t.Log("✅ Full E2E flow passed")
}
