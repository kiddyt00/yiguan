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

func setupModelTest(t *testing.T) (*ModelHandler, *sqlite.Store) {
	t.Helper()
	st, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	h := NewModelHandler(st, nil)
	return h, st
}

func TestModelHandler_CreateAndList(t *testing.T) {
	h, _ := setupModelTest(t)

	body, _ := json.Marshal(store.LLMModel{
		Name:     "qwen-turbo",
		Provider: "dashscope",
		Endpoint: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		APIKey:   "sk-test",
	})
	req := httptest.NewRequest("POST", "/api/admin/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CreateModel(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest("GET", "/api/admin/models", nil)
	rec2 := httptest.NewRecorder()
	h.ListModels(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	var resp struct {
		Items []store.LLMModel `json:"items"`
		Total int              `json:"total"`
	}
	json.NewDecoder(rec2.Body).Decode(&resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 model, got %d", len(resp.Items))
	}
	if resp.Items[0].Name != "qwen-turbo" {
		t.Fatalf("expected qwen-turbo, got %s", resp.Items[0].Name)
	}
	if resp.Items[0].DisplayName == "" {
		t.Fatal("expected display_name to be set")
	}
}

func TestModelHandler_CreateValidation(t *testing.T) {
	h, _ := setupModelTest(t)

	body, _ := json.Marshal(store.LLMModel{
		Endpoint: "https://example.com",
		APIKey:   "sk-test",
	})
	req := httptest.NewRequest("POST", "/api/admin/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CreateModel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", rec.Code)
	}
}

func TestModelHandler_Delete(t *testing.T) {
	h, st := setupModelTest(t)
	st.CreateModel(&store.LLMModel{
		Name: "to-delete", Provider: "test", Endpoint: "https://x.com", APIKey: "k",
	})
	models, _ := st.ListModels()
	id := fmt.Sprint(models[0].ID)

	req := httptest.NewRequest("DELETE", "/api/admin/models/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	h.DeleteModel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestModelHandler_ToggleModel(t *testing.T) {
	h, st := setupModelTest(t)
	st.CreateModel(&store.LLMModel{
		Name: "t", Provider: "p", Endpoint: "https://x.com", APIKey: "k",
	})
	models, _ := st.ListModels()
	id := fmt.Sprint(models[0].ID)

	req := httptest.NewRequest("POST", "/api/admin/models/"+id+"/toggle?enabled=true", nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	h.ToggleModel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m, _ := st.GetModelByID(models[0].ID)
	if m.IsEnabled != 1 {
		t.Fatalf("expected enabled=1, got %d", m.IsEnabled)
	}
}
