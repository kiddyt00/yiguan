package llm

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// Manager 管理多个 LLM provider，从数据库读取配置
type Manager struct {
	store       store.Store
	mu          sync.RWMutex
	cache       map[string]*Client // provider key -> client
	cacheExpiry time.Time
	cacheTTL    time.Duration
}

// NewManager 创建 LLM 管理器
func NewManager(st store.Store) *Manager {
	return &Manager{
		store:    st,
		cache:    make(map[string]*Client),
		cacheTTL: 30 * time.Second,
	}
}

// GetDefault 获取默认 provider 的客户端
func (m *Manager) GetDefault() (*Client, error) {
	m.mu.RLock()
	if time.Now().Before(m.cacheExpiry) {
		if c, ok := m.cache["__default__"]; ok {
			m.mu.RUnlock()
			return c, nil
		}
	}
	m.mu.RUnlock()

	// 从 DB 加载默认 provider
	p, err := m.store.GetDefaultLLMProvider()
	if err != nil {
		// 如果没有默认 provider，尝试取第一个
		providers, err2 := m.store.ListLLMProviders()
		if err2 != nil || len(providers) == 0 {
			return nil, fmt.Errorf("没有可用的 LLM 提供商: %w", err)
		}
		p = &providers[0]
	}

	client := New(Config{
		APIKey:   p.APIKey,
		Endpoint: p.Endpoint,
		Model:    p.Model,
	})

	m.mu.Lock()
	m.cache["__default__"] = client
	m.cacheExpiry = time.Now().Add(m.cacheTTL)
	m.mu.Unlock()

	log.Printf("[LLM] 使用默认提供商: %s (%s)", p.Name, p.Model)
	return client, nil
}

// GetByProvider 按 provider key 获取客户端
func (m *Manager) GetByProvider(providerKey string) (*Client, error) {
	m.mu.RLock()
	if time.Now().Before(m.cacheExpiry) {
		if c, ok := m.cache[providerKey]; ok {
			m.mu.RUnlock()
			return c, nil
		}
	}
	m.mu.RUnlock()

	providers, err := m.store.ListLLMProviders()
	if err != nil {
		return nil, fmt.Errorf("查询 LLM 提供商失败: %w", err)
	}

	for _, p := range providers {
		if p.Provider == providerKey {
			client := New(Config{
				APIKey:   p.APIKey,
				Endpoint: p.Endpoint,
				Model:    p.Model,
			})

			m.mu.Lock()
			m.cache[providerKey] = client
			m.cacheExpiry = time.Now().Add(m.cacheTTL)
			m.mu.Unlock()

			return client, nil
		}
	}

	return nil, fmt.Errorf("LLM 提供商 '%s' 未找到", providerKey)
}

// InvalidateCache 使缓存失效（provider 配置变更时调用）
func (m *Manager) InvalidateCache() {
	m.mu.Lock()
	m.cache = make(map[string]*Client)
	m.cacheExpiry = time.Time{}
	m.mu.Unlock()
}
