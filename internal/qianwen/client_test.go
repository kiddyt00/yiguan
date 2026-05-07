package qianwen

import "testing"

func TestBuildPrompt(t *testing.T) {
	prompt := BuildPrompt("我该换工作吗？", "乾为天", "天风姤", "上爻(第6爻)")

	checks := []string{
		"我该换工作吗？",
		"乾为天",
		"天风姤",
		"上爻",
		"本卦解义",
		"变爻启示",
		"变卦趋势",
		"综合建议",
	}
	for _, c := range checks {
		if !contains(prompt, c) {
			t.Errorf("prompt missing expected content: %q", c)
		}
	}
}

func TestBuildPromptEmpty(t *testing.T) {
	prompt := BuildPrompt("", "", "", "")
	if len(prompt) == 0 {
		t.Error("prompt should not be empty even with empty inputs")
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("test-api-key", "qwen-plus", "https://example.com/api")
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.apiKey != "test-api-key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "test-api-key")
	}
	if c.model != "qwen-plus" {
		t.Errorf("model = %q, want %q", c.model, "qwen-plus")
	}
}

func TestClientTimeout(t *testing.T) {
	c := NewClient("key", "model", "https://example.com")
	if c.client == nil {
		t.Error("http client not initialized")
	}
	if c.client.Timeout == 0 {
		t.Error("http client should have a timeout")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
