package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	cfg    Config
	client *http.Client
}

// New 创建 LLM 客户端
func New(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
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

// Divine 调用 LLM 解卦
func (c *Client) Divine(prompt string) (string, error) {
	body, _ := json.Marshal(chatReq{
		Model: c.cfg.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	})

	req, err := http.NewRequest("POST", c.cfg.Endpoint, bytes.NewReader(body))
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
