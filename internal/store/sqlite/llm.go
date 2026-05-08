package sqlite

import (
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// ListLLMProviders 列出所有 LLM 提供商
func (s *Store) ListLLMProviders() ([]store.LLMProvider, error) {
	rows, err := s.db.Query(
		`SELECT id, name, provider, api_key, endpoint, model, is_default, created_at, updated_at
		 FROM llm_providers ORDER BY is_default DESC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.LLMProvider
	for rows.Next() {
		var p store.LLMProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.Provider, &p.APIKey, &p.Endpoint,
			&p.Model, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// GetLLMProvider 获取单个 LLM 提供商
func (s *Store) GetLLMProvider(id int64) (*store.LLMProvider, error) {
	var p store.LLMProvider
	err := s.db.QueryRow(
		`SELECT id, name, provider, api_key, endpoint, model, is_default, created_at, updated_at
		 FROM llm_providers WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Provider, &p.APIKey, &p.Endpoint,
		&p.Model, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateLLMProvider 创建 LLM 提供商
func (s *Store) CreateLLMProvider(p *store.LLMProvider) error {
	now := time.Now()
	result, err := s.db.Exec(
		`INSERT INTO llm_providers (name, provider, api_key, endpoint, model, is_default, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.Provider, p.APIKey, p.Endpoint, p.Model, p.IsDefault, now, now,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// UpdateLLMProvider 更新 LLM 提供商
func (s *Store) UpdateLLMProvider(p *store.LLMProvider) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE llm_providers SET name=?, api_key=?, endpoint=?, model=?, is_default=?, updated_at=?
		 WHERE id=?`,
		p.Name, p.APIKey, p.Endpoint, p.Model, p.IsDefault, now, p.ID,
	)
	if err != nil {
		return err
	}
	p.UpdatedAt = now
	return nil
}

// DeleteLLMProvider 删除 LLM 提供商
func (s *Store) DeleteLLMProvider(id int64) error {
	_, err := s.db.Exec(`DELETE FROM llm_providers WHERE id = ?`, id)
	return err
}

// GetDefaultLLMProvider 获取默认 LLM 提供商
func (s *Store) GetDefaultLLMProvider() (*store.LLMProvider, error) {
	var p store.LLMProvider
	err := s.db.QueryRow(
		`SELECT id, name, provider, api_key, endpoint, model, is_default, created_at, updated_at
		 FROM llm_providers WHERE is_default = 1 LIMIT 1`,
	).Scan(&p.ID, &p.Name, &p.Provider, &p.APIKey, &p.Endpoint,
		&p.Model, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SetDefaultLLMProvider 设置默认 LLM 提供商
func (s *Store) SetDefaultLLMProvider(id int64) error {
	// 先取消所有默认
	if _, err := s.db.Exec(`UPDATE llm_providers SET is_default = 0`); err != nil {
		return err
	}
	// 设置指定 provider 为默认
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE llm_providers SET is_default = 1, updated_at = ? WHERE id = ?`,
		now, id,
	)
	return err
}
