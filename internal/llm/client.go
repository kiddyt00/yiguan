package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Config 单个 LLM provider 配置
type Config struct {
	APIKey   string
	Endpoint string
	Model    string
}

// Client 通用 LLM 客户端（OpenAI 兼容接口）
type Client struct {
	cfg        Config
	client     *http.Client
	mockDivine func(prompt string) (string, error) // 测试注入
}

// New 创建 LLM 客户端
func New(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// ModelName 返回当前模型名
func (c *Client) ModelName() string {
	return c.cfg.Model
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// chatURL 返回 chat/completions 端点 URL
func (c *Client) chatURL() string {
	base := strings.TrimRight(c.cfg.Endpoint, "/")
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

// Divine 调用 LLM 解卦
func (c *Client) Divine(prompt string) (string, error) {
	if c.mockDivine != nil {
		return c.mockDivine(prompt)
	}

	body, _ := json.Marshal(chatReq{
		Model: c.cfg.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	})

	req, err := http.NewRequest("POST", c.chatURL(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM返回错误: %d", resp.StatusCode)
	}

	var cr chatResp
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("LLM响应解析失败: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("LLM无有效响应")
	}
	return cr.Choices[0].Message.Content, nil
}

// BuildPrompt 构建解卦 prompt（要求返回 markdown 格式）
func BuildPrompt(question, primary, changing, yaoPositions string) string {
	return fmt.Sprintf(
		"请以专业易经解卦角度，结合用户问题「%s」和卦象进行分析：\n"+
			"本卦：%s，变卦：%s，变爻：%s\n\n"+
			"请按以下结构用 Markdown 格式给出解读：\n"+
			"## 本卦解义\n"+
			"（对本卦的解读，可使用**加粗**突出关键概念）\n\n"+
			"## 变爻启示\n"+
			"（变爻带来的启示和深层含义）\n\n"+
			"## 变卦趋势\n"+
			"（变卦所示的发展趋势）\n\n"+
			"## 综合建议\n"+
			"（结合问题给出的实用性建议）\n\n"+
			"> 请用流畅易懂的中文，避免过于玄奥的术语堆砌。",
		question, primary, changing, yaoPositions,
	)
}

// DivineWithRetry 调用 LLM 解卦，失败自动重试
func (c *Client) DivineWithRetry(prompt string, maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		result, err := c.Divine(prompt)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
	}
	return "", fmt.Errorf("LLM 调用失败(已重试%d次): %w", maxRetries, lastErr)
}

// DivineStreamWithRetry 流式调用 LLM，失败自动重试
func (c *Client) DivineStreamWithRetry(prompt string, onChunk func(string) error, maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		err := c.DivineStream(prompt, onChunk)
		if err == nil {
			return nil
		}
		lastErr = err
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
	}
	return fmt.Errorf("LLM流式调用失败(已重试%d次): %w", maxRetries, lastErr)
}
