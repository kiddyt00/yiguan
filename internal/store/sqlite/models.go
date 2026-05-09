package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListModels() ([]store.LLMModel, error) {
	rows, err := s.db.Query("SELECT id, name, display_name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.LLMModel
	for rows.Next() {
		var m store.LLMModel
		if err := rows.Scan(&m.ID, &m.Name, &m.DisplayName, &m.Provider, &m.Endpoint, &m.APIKey, &m.IsDefault, &m.IsEnabled, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (s *Store) GetModelByID(id int64) (*store.LLMModel, error) {
	m := &store.LLMModel{}
	err := s.db.QueryRow(
		"SELECT id, name, display_name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models WHERE id = ?",
		id,
	).Scan(&m.ID, &m.Name, &m.DisplayName, &m.Provider, &m.Endpoint, &m.APIKey, &m.IsDefault, &m.IsEnabled, &m.SortOrder, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return m, err
}

func (s *Store) GetDefaultModel() (*store.LLMModel, error) {
	m := &store.LLMModel{}
	err := s.db.QueryRow(
		"SELECT id, name, display_name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models WHERE is_default = 1 AND is_enabled = 1 ORDER BY id LIMIT 1",
	).Scan(&m.ID, &m.Name, &m.DisplayName, &m.Provider, &m.Endpoint, &m.APIKey, &m.IsDefault, &m.IsEnabled, &m.SortOrder, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return m, err
}

func (s *Store) CreateModel(m *store.LLMModel) error {
	result, err := s.db.Exec(
		"INSERT INTO llm_models (name, display_name, provider, endpoint, api_key, is_default, is_enabled, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		m.Name, m.DisplayName, m.Provider, m.Endpoint, m.APIKey, m.IsDefault, m.IsEnabled, m.SortOrder,
	)
	if err != nil {
		return err
	}
	m.ID, _ = result.LastInsertId()
	return nil
}

func (s *Store) UpdateModel(m *store.LLMModel) error {
	_, err := s.db.Exec(
		"UPDATE llm_models SET name = ?, display_name = ?, provider = ?, endpoint = ?, api_key = ?, is_enabled = ?, sort_order = ? WHERE id = ?",
		m.Name, m.DisplayName, m.Provider, m.Endpoint, m.APIKey, m.IsEnabled, m.SortOrder, m.ID,
	)
	return err
}

func (s *Store) DeleteModel(id int64) error {
	_, err := s.db.Exec("DELETE FROM llm_models WHERE id = ?", id)
	return err
}

func (s *Store) SetDefaultModel(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("UPDATE llm_models SET is_default = 0"); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE llm_models SET is_default = 1 WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ToggleModel(id int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec("UPDATE llm_models SET is_enabled = ? WHERE id = ?", v, id)
	return err
}
