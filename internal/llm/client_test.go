package llm

import (
	"errors"
	"testing"
)

func TestDivineWithRetry_SuccessFirstAttempt(t *testing.T) {
	callCount := 0
	c := &Client{
		cfg: Config{Model: "test", Endpoint: "http://localhost:1", APIKey: "k"},
		client: nil,
		// 注入 mock divine
		mockDivine: func(prompt string) (string, error) {
			callCount++
			return "解读内容", nil
		},
	}

	result, err := c.DivineWithRetry("测试问题", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "解读内容" {
		t.Fatalf("expected 解读内容, got %s", result)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
}

func TestDivineWithRetry_RetriesOnFailure(t *testing.T) {
	callCount := 0
	c := &Client{
		cfg: Config{Model: "test", Endpoint: "http://localhost:1", APIKey: "k"},
		mockDivine: func(prompt string) (string, error) {
			callCount++
			if callCount < 3 {
				return "", errors.New("network error")
			}
			return "重试后成功", nil
		},
	}

	result, err := c.DivineWithRetry("测试问题", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "重试后成功" {
		t.Fatalf("expected 重试后成功, got %s", result)
	}
	if callCount != 3 {
		t.Fatalf("expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
}

func TestDivineWithRetry_ExhaustsRetries(t *testing.T) {
	callCount := 0
	c := &Client{
		cfg: Config{Model: "test", Endpoint: "http://localhost:1", APIKey: "k"},
		mockDivine: func(prompt string) (string, error) {
			callCount++
			return "", errors.New("persistent failure")
		},
	}

	_, err := c.DivineWithRetry("测试问题", 2)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if callCount != 3 {
		t.Fatalf("expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}
