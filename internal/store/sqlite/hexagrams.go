package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListAllHistory(limit, offset int, userID int64) ([]store.History, error) {
	var rows *sql.Rows
	var err error
	query := `SELECT h.id, h.user_id, u.nickname, h.question, h.primary_gua, h.changing_gua,
	           h.yao_positions, h.primary_yao, h.changing_yao, h.interpretation, h.lang, h.created_at
	          FROM history h LEFT JOIN users u ON h.user_id = u.id`
	if userID > 0 {
		rows, err = s.db.Query(query+` WHERE h.user_id = ? ORDER BY h.id DESC LIMIT ? OFFSET ?`, userID, limit, offset)
	} else {
		rows, err = s.db.Query(query+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.History
	for rows.Next() {
		var h store.History
		if err := rows.Scan(&h.ID, &h.UserID, &h.Nickname, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.PrimaryYao, &h.ChangingYao, &h.Interpretation, &h.Lang, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, h)
	}
	return list, rows.Err()
}

func (s *Store) GetHistoryByID(id int64) (*store.History, error) {
	h := &store.History{}
	err := s.db.QueryRow(
		`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, primary_yao, changing_yao, interpretation, lang, created_at
		 FROM history WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.PrimaryYao, &h.ChangingYao, &h.Interpretation, &h.Lang, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return h, err
}

func (s *Store) DeleteHistory(id int64) error {
	_, err := s.db.Exec("DELETE FROM history WHERE id = ?", id)
	return err
}

func (s *Store) GetUserHistory(userID int64, limit, offset int) ([]store.History, error) {
	return s.GetHistory(userID, limit, offset)
}

// SearchHistory 关键词搜索历史记录
func (s *Store) SearchHistory(userID int64, keyword string, limit, offset int) ([]store.History, error) {
	pattern := "%" + keyword + "%"
	rows, err := s.db.Query(
		`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, lang, created_at
		 FROM history WHERE user_id = ? AND (question LIKE ? OR primary_gua LIKE ? OR changing_gua LIKE ?)
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, pattern, pattern, pattern, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.History
	for rows.Next() {
		var h store.History
		if err := rows.Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua,
			&h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.Lang, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, h)
	}
	return list, rows.Err()
}

// SearchHistoryCount 搜索匹配的记录总数
func (s *Store) SearchHistoryCount(userID int64, keyword string) (int64, error) {
	pattern := "%" + keyword + "%"
	var count int64
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM history WHERE user_id = ? AND (question LIKE ? OR primary_gua LIKE ? OR changing_gua LIKE ?)`,
		userID, pattern, pattern, pattern,
	).Scan(&count)
	return count, err
}

// GetLatestHistory 获取用户最新一条历史记录
func (s *Store) GetLatestHistory(userID int64) (*store.History, error) {
	h := &store.History{}
	err := s.db.QueryRow(
		`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, lang, created_at
		 FROM history WHERE user_id = ? ORDER BY id DESC LIMIT 1`,
		userID,
	).Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.Lang, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return h, err
}

func (s *Store) GetActiveUserCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = 1").Scan(&count)
	return count, err
}

func (s *Store) GetTotalDivineCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&count)
	return count, err
}

func (s *Store) GetTodayAdWatchCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM ad_records WHERE date(created_at) = date('now')").Scan(&count)
	return count, err
}

func (s *Store) GetTodayAdWatchCountByUser(userID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM ad_records WHERE user_id = ? AND date(created_at) = date('now') AND rewarded = 1",
		userID,
	).Scan(&count)
	return count, err
}

func (s *Store) GetTotalAdWatchCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM ad_records").Scan(&count)
	return count, err
}

func (s *Store) ToggleUser(id int64, active bool) error {
	v := 0
	if active {
		v = 1
	}
	_, err := s.db.Exec("UPDATE users SET is_active = ? WHERE id = ?", v, id)
	return err
}

func (s *Store) UpdateUserRole(id int64, role string) error {
	_, err := s.db.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	return err
}

func (s *Store) UpdateUserQuota(userID int64, delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		for i := 0; i < delta; i++ {
			if err := s.AddQuota(userID, "admin_grant"); err != nil {
				return err
			}
		}
	} else {
		count := -delta
		_, err := s.db.Exec(`DELETE FROM quotas WHERE id IN (
			SELECT id FROM quotas WHERE user_id = ? AND used_at IS NULL ORDER BY id LIMIT ?
		)`, userID, count)
		return err
	}
	return nil
}

func (s *Store) GetUserQuota(userID int64) (int, error) {
	return s.GetRemainingQuota(userID)
}
