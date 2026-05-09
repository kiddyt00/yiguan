package sqlite

import "github.com/kiddyt00/yiguan/internal/store"

// SaveHistory 保存算卦历史记录
func (s *Store) SaveHistory(h *store.History) error {
	result, err := s.db.Exec(
		`INSERT INTO history (user_id, question, primary_gua, changing_gua, yao_positions, interpretation)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		h.UserID, h.Question, h.PrimaryGua, h.ChangingGua, h.YaoPositions, h.Interpretation,
	)
	if err != nil {
		return err
	}
	h.ID, _ = result.LastInsertId()
	return nil
}

// GetHistory 分页获取算卦历史
func (s *Store) GetHistory(userID int64, limit, offset int) ([]store.History, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
		 FROM history WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.History
	for rows.Next() {
		var h store.History
		if err := rows.Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua,
			&h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, h)
	}
	return list, rows.Err()
}

// GetHistoryCount 获取用户历史记录总数
func (s *Store) GetHistoryCount(userID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM history WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// GetTotalUsers 用户总数
func (s *Store) GetTotalUsers() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// ListUsers 分页列出用户
func (s *Store) ListUsers(limit, offset int) ([]store.User, error) {
	rows, err := s.db.Query(
		"SELECT id, phone, COALESCE(openid,''), nickname, avatar, address, role, is_active, created_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.User
	for rows.Next() {
		var u store.User
		if err := rows.Scan(&u.ID, &u.Phone, &u.OpenID, &u.Nickname, &u.Avatar, &u.Address, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

// GetTodayDivineCount 今日算卦次数
func (s *Store) GetTodayDivineCount() (int64, error) {
	var count int64
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM history WHERE date(created_at) = date('now')",
	).Scan(&count)
	return count, err
}
