package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

// GetTranslation 获取一条翻译缓存
func (s *Store) GetTranslation(historyID int64, lang string) (*store.Translation, error) {
	t := &store.Translation{}
	err := s.db.QueryRow(
		`SELECT id, history_id, lang, content, created_at
		 FROM translations WHERE history_id = ? AND lang = ?`,
		historyID, lang,
	).Scan(&t.ID, &t.HistoryID, &t.Lang, &t.Content, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return t, err
}

// SaveTranslation 保存一条翻译缓存（由于 UNIQUE 约束，重复写入会失败但不影响已有数据）
func (s *Store) SaveTranslation(t *store.Translation) error {
	result, err := s.db.Exec(
		`INSERT OR IGNORE INTO translations (history_id, lang, content)
		 VALUES (?, ?, ?)`,
		t.HistoryID, t.Lang, t.Content,
	)
	if err != nil {
		return err
	}
	t.ID, _ = result.LastInsertId()
	return nil
}
