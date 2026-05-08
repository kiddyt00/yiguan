package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListAllHistory(limit, offset int, userID int64) ([]store.History, error) {
	var rows *sql.Rows
	var err error
	if userID > 0 {
		rows, err = s.db.Query(
			`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
			 FROM history WHERE user_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
			userID, limit, offset,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
			 FROM history ORDER BY id DESC LIMIT ? OFFSET ?`,
			limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.History
	for rows.Next() {
		var h store.History
		if err := rows.Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, h)
	}
	return list, rows.Err()
}

func (s *Store) GetHistoryByID(id int64) (*store.History, error) {
	h := &store.History{}
	err := s.db.QueryRow(
		`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
		 FROM history WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.CreatedAt)
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
