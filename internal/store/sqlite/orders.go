package sqlite

import (
	"database/sql"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// CreateOrder 创建订单
func (s *Store) CreateOrder(userID int64, product *store.OrderProduct, outTradeNo, codeURL string) (*store.Order, error) {
	result, err := s.db.Exec(
		`INSERT INTO orders (user_id, amount, quota, product_id, out_trade_no, code_url)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, product.Amount, product.Quota, product.ID, outTradeNo, codeURL,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return s.GetOrder(id)
}

// GetOrder 按 ID 获取订单
func (s *Store) GetOrder(id int64) (*store.Order, error) {
	o := &store.Order{}
	var paidAt sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), status,
		        COALESCE(out_trade_no,''), COALESCE(prepay_id,''), COALESCE(code_url,''),
		        paid_at, created_at
		 FROM orders WHERE id = ?`, id,
	).Scan(&o.ID, &o.UserID, &o.Amount, &o.Quota, &o.ProductID,
		&o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
		&paidAt, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}
	return o, nil
}

// GetOrderByOutTradeNo 按商户订单号获取订单
func (s *Store) GetOrderByOutTradeNo(outTradeNo string) (*store.Order, error) {
	o := &store.Order{}
	var paidAt sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), status,
		        COALESCE(out_trade_no,''), COALESCE(prepay_id,''), COALESCE(code_url,''),
		        paid_at, created_at
		 FROM orders WHERE out_trade_no = ?`, outTradeNo,
	).Scan(&o.ID, &o.UserID, &o.Amount, &o.Quota, &o.ProductID,
		&o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
		&paidAt, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}
	return o, nil
}

// ListOrders 列出用户订单
func (s *Store) ListOrders(userID int64, limit, offset int) ([]store.Order, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), status,
		        COALESCE(out_trade_no,''), COALESCE(prepay_id,''), COALESCE(code_url,''),
		        paid_at, created_at
		 FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Order
	for rows.Next() {
		var o store.Order
		var paidAt sql.NullTime
		if err := rows.Scan(&o.ID, &o.UserID, &o.Amount, &o.Quota, &o.ProductID,
			&o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
			&paidAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			o.PaidAt = &paidAt.Time
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

// MarkOrderPaid 标记订单已支付
func (s *Store) MarkOrderPaid(id int64, prepayID string) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE orders SET status = 'paid', prepay_id = ?, paid_at = ? WHERE id = ?`,
		prepayID, now, id,
	)
	return err
}

// UpdateCodeURL 更新订单付款二维码 URL
func (s *Store) UpdateCodeURL(id int64, codeURL string) error {
	_, err := s.db.Exec(
		`UPDATE orders SET code_url = ? WHERE id = ?`,
		codeURL, id,
	)
	return err
}
