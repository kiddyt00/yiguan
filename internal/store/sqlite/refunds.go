package sqlite

import (
	"database/sql"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// CreateRefund 创建退款申请
func (s *Store) CreateRefund(userID, orderID int64, reason string) (*store.Refund, error) {
	// 获取订单信息
	order, err := s.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, store.ErrNotFound
	}
	// 检查是否已有退款申请
	var c int
	s.db.QueryRow("SELECT COUNT(*) FROM refunds WHERE order_id = ?", orderID).Scan(&c)
	if c > 0 {
		return nil, store.ErrQuotaExhausted
	}

	result, err := s.db.Exec(
		`INSERT INTO refunds (user_id, order_id, amount, reason, status, paid_at)
		 VALUES (?, ?, ?, ?, 'pending', ?)`,
		userID, orderID, order.Amount, reason, order.PaidAt,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return s.GetRefund(id)
}

// GetRefund 按ID获取退款记录
func (s *Store) GetRefund(id int64) (*store.Refund, error) {
	r := &store.Refund{}
	var paidAt, refundedAt sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, user_id, order_id, amount, COALESCE(reason,''), status, paid_at, refunded_at, created_at
		 FROM refunds WHERE id = ?`, id,
	).Scan(&r.ID, &r.UserID, &r.OrderID, &r.Amount, &r.Reason, &r.Status, &paidAt, &refundedAt, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if paidAt.Valid {
		r.PaidAt = &paidAt.Time
	}
	if refundedAt.Valid {
		r.RefundedAt = &refundedAt.Time
	}
	return r, nil
}

// GetRefundByOrder 按订单ID查退款
func (s *Store) GetRefundByOrder(orderID int64) (*store.Refund, error) {
	var id int64
	err := s.db.QueryRow("SELECT id FROM refunds WHERE order_id = ?", orderID).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetRefund(id)
}

// ListRefunds 列出用户退款记录
func (s *Store) ListRefunds(userID int64) ([]store.Refund, error) {
	return s.queryRefunds("WHERE user_id = ? ORDER BY created_at DESC", userID)
}

// ListAllRefunds 管理员列出全部退款
func (s *Store) ListAllRefunds() ([]store.Refund, error) {
	return s.queryRefunds("ORDER BY created_at DESC")
}

// ApproveRefund 批准退款
func (s *Store) ApproveRefund(id int64) error {
	_, err := s.db.Exec(
		"UPDATE refunds SET status = 'approved' WHERE id = ? AND status = 'pending'",
		id,
	)
	return err
}

// CompleteRefund 标记退款完成（已从支付渠道退回）
func (s *Store) CompleteRefund(id int64) error {
	now := time.Now()
	_, err := s.db.Exec(
		"UPDATE refunds SET status = 'completed', refunded_at = ? WHERE id = ? AND status = 'approved'",
		now, id,
	)
	return err
}

// RejectRefund 驳回退款
func (s *Store) RejectRefund(id int64, reason string) error {
	_, err := s.db.Exec(
		"UPDATE refunds SET status = 'rejected', reason = ? WHERE id = ? AND status = 'pending'",
		reason, id,
	)
	return err
}

func (s *Store) queryRefunds(where string, args ...interface{}) ([]store.Refund, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, order_id, amount, COALESCE(reason,''), status, paid_at, refunded_at, created_at
		 FROM refunds `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []store.Refund
	for rows.Next() {
		var r store.Refund
		var paidAt, refundedAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.UserID, &r.OrderID, &r.Amount, &r.Reason, &r.Status, &paidAt, &refundedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			r.PaidAt = &paidAt.Time
		}
		if refundedAt.Valid {
			r.RefundedAt = &refundedAt.Time
		}
		list = append(list, r)
	}
	return list, rows.Err()
}
