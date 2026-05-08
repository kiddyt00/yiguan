package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiddyt00/yiguan/internal/store"
	"github.com/kiddyt00/yiguan/internal/store/sqlite"
)

func setupHexagramTest(t *testing.T) (*HexagramHandler, *sqlite.Store) {
	t.Helper()
	st, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	st.CreateUser("13800000001", "test", "testuser")
	h := NewHexagramHandler(st)
	return h, st
}

func TestHexagramHandler_ListHistory(t *testing.T) {
	h, st := setupHexagramTest(t)

	st.SaveHistory(&store.History{
		UserID: 1, Question: "测试问题", PrimaryGua: "乾为天",
		ChangingGua: "天风姤", YaoPositions: "初九", Interpretation: "测试解读",
	})

	req := httptest.NewRequest("GET", "/api/admin/hexagrams", nil)
	rec := httptest.NewRecorder()
	h.ListHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHexagramHandler_DeleteHistory(t *testing.T) {
	h, st := setupHexagramTest(t)

	st.SaveHistory(&store.History{
		UserID: 1, Question: "will delete", PrimaryGua: "坤为地",
		ChangingGua: "坤为地", YaoPositions: "无变爻", Interpretation: "test",
	})

	req := httptest.NewRequest("DELETE", "/api/admin/hexagrams/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.DeleteHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHexagramHandler_GetDetail(t *testing.T) {
	h, st := setupHexagramTest(t)

	st.SaveHistory(&store.History{
		UserID: 1, Question: "detail test", PrimaryGua: "震为雷",
		ChangingGua: "雷地豫", YaoPositions: "六二", Interpretation: "detail",
	})

	req := httptest.NewRequest("GET", "/api/admin/hexagrams/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.GetHistoryDetail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
