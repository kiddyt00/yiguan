package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiddyt00/yiguan/internal/store"
	"github.com/kiddyt00/yiguan/internal/store/sqlite"
)

func setupAdTest(t *testing.T) (*AdHandler, *sqlite.Store) {
	t.Helper()
	st, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	h := NewAdHandler(st)
	return h, st
}

func TestAdHandler_CreateAndList(t *testing.T) {
	h, _ := setupAdTest(t)

	body, _ := json.Marshal(store.Ad{
		Name:          "测试广告",
		ContentURL:    "https://example.com/ad",
		AdType:        "iframe",
		WatchDuration: 30,
		RewardQuota:   1,
	})
	req := httptest.NewRequest("POST", "/api/admin/ads", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CreateAd(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest("GET", "/api/admin/ads", nil)
	rec2 := httptest.NewRecorder()
	h.ListAds(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	var resp struct {
		Items []store.Ad `json:"items"`
		Total int        `json:"total"`
	}
	json.NewDecoder(rec2.Body).Decode(&resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 ad, got %d", len(resp.Items))
	}
}

func TestAdHandler_CreateValidation(t *testing.T) {
	h, _ := setupAdTest(t)

	body, _ := json.Marshal(store.Ad{
		Name: "test",
		// ContentURL intentionally missing
	})
	req := httptest.NewRequest("POST", "/api/admin/ads", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CreateAd(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing content_url, got %d", rec.Code)
	}
}

func TestAdHandler_ListActive(t *testing.T) {
	h, st := setupAdTest(t)

	st.CreateAd(&store.Ad{
		Name: "active", ContentURL: "https://x.com", AdType: "iframe",
		WatchDuration: 30, RewardQuota: 1, IsEnabled: 1,
	})
	st.CreateAd(&store.Ad{
		Name: "disabled", ContentURL: "https://x.com", AdType: "iframe",
		WatchDuration: 30, RewardQuota: 1, IsEnabled: 0,
	})

	req := httptest.NewRequest("GET", "/api/ads/active", nil)
	rec := httptest.NewRecorder()
	h.ListActiveAds(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Items []store.Ad `json:"items"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 active ad, got %d", len(resp.Items))
	}
}

func TestAdHandler_Delete(t *testing.T) {
	h, st := setupAdTest(t)

	st.CreateAd(&store.Ad{
		Name: "to-delete", ContentURL: "https://x.com", AdType: "iframe",
		WatchDuration: 30, RewardQuota: 1,
	})
	ads, _ := st.ListAds()
	id := fmt.Sprint(ads[0].ID)

	req := httptest.NewRequest("DELETE", "/api/admin/ads/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	h.DeleteAd(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
