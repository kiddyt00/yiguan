package llm

import (
	"fmt"
	"sync"

	"github.com/kiddyt00/yiguan/internal/store"
)

// Router 模型热切换路由器 + 容错链
type Router struct {
	mu      sync.RWMutex
	st      store.Store
	current *Client   // 默认模型
	all     []*Client // 所有已启用模型（按 sort_order）
}

// NewRouter 创建路由器并从数据库加载
func NewRouter(st store.Store) (*Router, error) {
	r := &Router{st: st}
	if err := r.Reload(); err != nil {
		return nil, err
	}
	return r, nil
}

// NewRouterWithFallback 创建路由器，数据库无模型时用 config 兜底
func NewRouterWithFallback(st store.Store, cfg Config) (*Router, error) {
	r := &Router{st: st}
	m, err := st.GetDefaultModel()
	if err != nil {
		r.current = New(cfg)
		r.all = []*Client{r.current}
		return r, nil
	}
	client := New(Config{
		APIKey:   m.APIKey,
		Endpoint: m.Endpoint,
		Model:    m.Name,
	})
	r.current = client
	r.all = []*Client{client}
	return r, nil
}

// Get 返回当前默认模型客户端
func (r *Router) Get() *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.current
}

// GetAllEnabled 返回所有已启用模型客户端（容错链用）
func (r *Router) GetAllEnabled() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clients := make([]*Client, len(r.all))
	copy(clients, r.all)
	return clients
}

// Reload 从数据库重新加载所有已启用模型
func (r *Router) Reload() error {
	models, err := r.st.ListModels()
	if err != nil {
		return fmt.Errorf("获取模型列表失败: %w", err)
	}

	var enabled []store.LLMModel
	var def *store.LLMModel
	for i := range models {
		m := &models[i]
		if m.IsEnabled != 1 {
			continue
		}
		enabled = append(enabled, *m)
		if m.IsDefault == 1 {
			def = m
		}
	}

	if len(enabled) == 0 {
		return fmt.Errorf("没有已启用的模型")
	}

	// 默认模型不存在时取第一个
	if def == nil {
		def = &enabled[0]
	}

	clients := make([]*Client, len(enabled))
	for i, m := range enabled {
		clients[i] = New(Config{
			APIKey:   m.APIKey,
			Endpoint: m.Endpoint,
			Model:    m.Name,
		})
	}

	r.mu.Lock()
	r.current = New(Config{
		APIKey:   def.APIKey,
		Endpoint: def.Endpoint,
		Model:    def.Name,
	})
	r.all = clients
	r.mu.Unlock()
	return nil
}
