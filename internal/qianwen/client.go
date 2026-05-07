package qianwen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client 千问 API 客户端
type Client struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
}

// NewClient 创建千问 API 客户端
func NewClient(apiKey, model, endpoint string) *Client {
	return &Client{
		apiKey:   apiKey,
		model:    model,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ChatRequest OpenAI 兼容的请求体
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse OpenAI 兼容的响应体
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// BuildPrompt 构建解卦 prompt 模板
func BuildPrompt(question, primary, changing string, yaoPositions string) string {
	return fmt.Sprintf(
		"请以专业易经解卦角度，结合用户问题「%s」和卦象进行分析：\n"+
			"本卦：%s，变卦：%s，变爻：%s\n\n"+
			"请按以下结构给出解读：\n"+
			"1. 本卦解义\n"+
			"2. 变爻启示\n"+
			"3. 变卦趋势\n"+
			"4. 综合建议\n\n"+
			"请用流畅易懂的中文，避免过于玄奥的术语堆砌。",
		question, primary, changing, yaoPositions,
	)
}

// Divine 调用千问 API 解卦，返回解读文本
func (c *Client) Divine(prompt string) (string, error) {
	reqBody := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("千问API请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("千问API返回错误状态码: %d", resp.StatusCode)
	}

	var cr ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("千问API响应解析失败: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("千问API无有效响应")
	}
	return cr.Choices[0].Message.Content, nil
}
