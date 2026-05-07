package test

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kiddyt00/yiguan/internal/handler"
	"github.com/kiddyt00/yiguan/internal/qianwen"
)

// newTestServer 创建测试用 HTTP 服务
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	funcMap := template.FuncMap{
		"yaoLabel": handler.YaoLabelFunc(),
	}

	tmpl := template.Must(
		template.New("").Funcs(funcMap).ParseFiles(
			"../templates/layout.html",
			"../templates/home.html",
			"../templates/result.html",
		),
	)

	qw := qianwen.NewClient("", "qwen-plus", "https://example.com/api")

	mux := http.NewServeMux()
	mux.Handle("GET /", &handler.HomeHandler{Tmpl: tmpl})
	mux.Handle("POST /divine", &handler.DivineHandler{Tmpl: tmpl, Qianwen: qw})

	return httptest.NewServer(mux)
}

func TestHomePage(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("home page returned %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	checks := []string{"易观", "开始提问", "请输入你想问的问题"}
	for _, c := range checks {
		if !strings.Contains(bodyStr, c) {
			t.Errorf("home page missing: %q", c)
		}
	}
}

func TestDivineNoQuestion(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.PostForm(ts.URL+"/divine", url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty question, got %d", resp.StatusCode)
	}
}

func TestDivineWithQuestion(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.PostForm(ts.URL+"/divine", url.Values{
		"question": {"我想知道今天的运气如何？"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("divine returned %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	required := []string{
		"卦象结果",
		"本卦",
		"变卦",
		"变爻",
		"周易大师一对一详解",
	}
	for _, r := range required {
		if !strings.Contains(bodyStr, r) {
			t.Errorf("result missing: %q", r)
		}
	}
}

func TestHomePageContentType(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/")
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Error("no Content-Type header")
	}
	// Go's html/template auto-sets Content-Type: text/html; charset=utf-8
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestDivineRandomness(t *testing.T) {
	// 多次算卦应产生不同结果
	ts := newTestServer(t)
	defer ts.Close()

	results := make(map[string]bool)
	for i := 0; i < 5; i++ {
		resp, _ := http.PostForm(ts.URL+"/divine", url.Values{
			"question": {"测试"},
		})
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		results[string(body)] = true
	}
	if len(results) < 2 {
		t.Error("5 divinations produced only 1 unique result — randomness broken")
	}
}
