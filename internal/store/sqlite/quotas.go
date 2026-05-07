package sqlite

import (
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// GetRemainingQuota 返回用户剩余可用次数
func (s *Store) GetRemainingQuota(userID int64) (int, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM quotas WHERE user_id = ? AND used_at IS NULL",
		userID,
	).Scan(&count)
	return count, err
}

// AddQuota 为用户增加一次配额
func (s *Store) AddQuota(userID int64, quotaType string) error {
	_, err := s.db.Exec(
		"INSERT INTO quotas (user_id, quota_type) VALUES (?, ?)",
		userID, quotaType,
	)
	return err
}

// ConsumeQuota 扣减一次配额（FIFO：最早的未使用配额）
func (s *Store) ConsumeQuota(userID int64) error {
	// 找出最早的一条未使用配额
	var id int64
	err := s.db.QueryRow(
		"SELECT id FROM quotas WHERE user_id = ? AND used_at IS NULL ORDER BY created_at ASC LIMIT 1",
		userID,
	).Scan(&id)
	if err != nil {
		return store.ErrQuotaExhausted
	}

	now := time.Now()
	_, err = s.db.Exec("UPDATE quotas SET used_at = ? WHERE id = ?", now, id)
	return err
}
