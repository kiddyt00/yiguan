package sqlite

import (
	"database/sql"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// CreateMembership 创建会员记录（自动顺延叠加）
func (s *Store) CreateMembership(userID, orderID int64, productID string, endTime time.Time) (*store.Membership, error) {
	// 顺延逻辑：取当前有效会员最晚到期时间作为 start_time
	active, err := s.GetActiveMembership(userID)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}

	startTime := time.Now()
	if active != nil && active.EndTime.After(startTime) {
		startTime = active.EndTime
	}

	result, err := s.db.Exec(
		`INSERT INTO memberships (user_id, order_id, product_id, start_time, end_time, status)
		 VALUES (?, ?, ?, ?, ?, 'active')`,
		userID, orderID, productID, startTime, endTime,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	// 记录日志
	_ = s.LogMembership(userID, id, "created", "会员开通: "+productID)

	return s.getMembershipByID(id)
}

// GetActiveMembership 获取用户当前有效的会员（取最晚到期的那条）
func (s *Store) GetActiveMembership(userID int64) (*store.Membership, error) {
	m := &store.Membership{}
	err := s.db.QueryRow(
		`SELECT id, user_id, order_id, product_id, start_time, end_time, status, created_at
		 FROM memberships
		 WHERE user_id = ? AND status = 'active' AND end_time > datetime('now')
		 ORDER BY end_time DESC LIMIT 1`,
		userID,
	).Scan(&m.ID, &m.UserID, &m.OrderID, &m.ProductID, &m.StartTime, &m.EndTime, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// HasActiveMembership 检查用户是否有有效会员
func (s *Store) HasActiveMembership(userID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM memberships WHERE user_id = ? AND status = 'active' AND end_time > datetime('now')",
		userID,
	).Scan(&count)
	return count > 0, err
}

// TerminateMembership 终止会员（退款时用）
func (s *Store) TerminateMembership(membershipID int64) error {
	var userID int64
	err := s.db.QueryRow("SELECT user_id FROM memberships WHERE id = ?", membershipID).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"UPDATE memberships SET status = 'terminated' WHERE id = ? AND status = 'active'",
		membershipID,
	)
	if err != nil {
		return err
	}

	_ = s.LogMembership(userID, membershipID, "terminated", "会员终止（退款）")
	return nil
}

// ListMemberships 列出用户会员记录
func (s *Store) ListMemberships(userID int64) ([]store.Membership, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, order_id, product_id, start_time, end_time, status, created_at
		 FROM memberships WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Membership
	for rows.Next() {
		var m store.Membership
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrderID, &m.ProductID, &m.StartTime, &m.EndTime, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// LogMembership 记录会员权益变动流水
func (s *Store) LogMembership(userID int64, membershipID int64, action, detail string) error {
	_, err := s.db.Exec(
		`INSERT INTO membership_logs (user_id, membership_id, action, detail)
		 VALUES (?, ?, ?, ?)`,
		userID, membershipID, action, detail,
	)
	return err
}

// getMembershipByID 内部方法：按 ID 查会员
func (s *Store) getMembershipByID(id int64) (*store.Membership, error) {
	m := &store.Membership{}
	err := s.db.QueryRow(
		`SELECT id, user_id, order_id, product_id, start_time, end_time, status, created_at
		 FROM memberships WHERE id = ?`, id,
	).Scan(&m.ID, &m.UserID, &m.OrderID, &m.ProductID, &m.StartTime, &m.EndTime, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ListAllMemberships 管理员：列出全部会员记录
func (s *Store) ListAllMemberships(limit, offset int) ([]store.Membership, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, order_id, product_id, start_time, end_time, status, created_at
		 FROM memberships ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Membership
	for rows.Next() {
		var m store.Membership
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrderID, &m.ProductID, &m.StartTime, &m.EndTime, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// CountActiveMemberships 统计当前有效会员数
func (s *Store) CountActiveMemberships() (int, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM memberships WHERE status = 'active' AND end_time > datetime('now')",
	).Scan(&count)
	return count, err
}

// GetMembershipByOrderID 按 order_id 查会员（退款时用）
func (s *Store) GetMembershipByOrderID(orderID int64) (*store.Membership, error) {
	m := &store.Membership{}
	err := s.db.QueryRow(
		`SELECT id, user_id, order_id, product_id, start_time, end_time, status, created_at
		 FROM memberships WHERE order_id = ?`, orderID,
	).Scan(&m.ID, &m.UserID, &m.OrderID, &m.ProductID, &m.StartTime, &m.EndTime, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}
