package llm

import (
	"fmt"
	"sync"

	"github.com/kiddyt00/yiguan/internal/store"
)

// Router 模型热切换路由器
type Router struct {
	mu      sync.RWMutex
	st      store.Store
	current *Client
}

// NewRouter 创建路由器并从数据库加载默认模型
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
		// 数据库没有模型，用 config 兜底
		r.current = New(cfg)
		return r, nil
	}
	client := New(Config{
		APIKey:   m.APIKey,
		Endpoint: m.Endpoint,
		Model:    m.Name,
	})
	r.current = client
	return r, nil
}

// Get 返回当前默认模型客户端
func (r *Router) Get() *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.current
}

// Reload 从数据库重新加载默认模型
func (r *Router) Reload() error {
	m, err := r.st.GetDefaultModel()
	if err != nil {
		return fmt.Errorf("获取默认模型失败: %w", err)
	}
	client := New(Config{
		APIKey:   m.APIKey,
		Endpoint: m.Endpoint,
		Model:    m.Name,
	})
	r.mu.Lock()
	r.current = client
	r.mu.Unlock()
	return nil
}
