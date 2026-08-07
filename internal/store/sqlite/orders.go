package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// CreateOrder 创建订单
func (s *Store) CreateOrder(userID int64, product *store.OrderProduct, outTradeNo, codeURL, channel string) (*store.Order, error) {
	result, err := s.db.Exec(
		`INSERT INTO orders (user_id, amount, quota, product_id, out_trade_no, code_url, channel)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, product.Amount, product.Quota, product.ID, outTradeNo, codeURL, channel,
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
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), COALESCE(channel,''), status,
		        COALESCE(out_trade_no,''), COALESCE(prepay_id,''), COALESCE(code_url,''),
		        paid_at, created_at
		 FROM orders WHERE id = ?`, id,
	).Scan(&o.ID, &o.UserID, &o.Amount, &o.Quota, &o.ProductID,
		&o.Channel, &o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
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
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), COALESCE(channel,''), status,
		        COALESCE(out_trade_no,''), COALESCE(prepay_id,''), COALESCE(code_url,''),
		        paid_at, created_at
		 FROM orders WHERE out_trade_no = ?`, outTradeNo,
	).Scan(&o.ID, &o.UserID, &o.Amount, &o.Quota, &o.ProductID,
		&o.Channel, &o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
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
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), COALESCE(channel,''), status,
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
			&o.Channel, &o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
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

// MarkOrderPaid 原子标记订单已支付（仅 pending → paid，返回是否真正更新，
// 防止并发回调重复进入发放逻辑）
func (s *Store) MarkOrderPaid(id int64, prepayID string) (bool, error) {
	now := time.Now()
	res, err := s.db.Exec(
		`UPDATE orders SET status = 'paid', prepay_id = ?, paid_at = ? WHERE id = ? AND status = 'pending'`,
		prepayID, now, id,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// GrantOrderBenefits 在单事务内完成：标记订单已支付 + 发放对应权益（幂等）。
// - 会员商品：创建会员记录（自动顺延叠加，endTime 基于顺延后的 startTime）
// - 次数商品：写入购买配额（带 order_id 归属）
// 任一步失败则整体回滚；重复调用不会重复发放。
func (s *Store) GrantOrderBenefits(orderID int64, prepayID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID, quota, amount int64
	var productID string
	err = tx.QueryRow(
		"SELECT user_id, quota, COALESCE(product_id,''), amount FROM orders WHERE id = ?", orderID,
	).Scan(&userID, &quota, &productID, &amount)
	if err != nil {
		return err
	}

	now := time.Now()
	if _, err := tx.Exec(
		"UPDATE orders SET status = 'paid', prepay_id = ?, paid_at = ? WHERE id = ? AND status = 'pending'",
		prepayID, now, orderID,
	); err != nil {
		return err
	}

	product := store.FindProduct(productID)
	if product != nil && product.IsMembership() {
		// 幂等：该订单已开过会员则跳过
		var c int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM memberships WHERE order_id = ?", orderID,
		).Scan(&c); err != nil {
			return err
		}
		if c == 0 {
			// 顺延：已有有效会员则从最晚到期时间起算
			var maxEnd sql.NullString
			_ = tx.QueryRow(
				`SELECT MAX(end_time) FROM memberships
				 WHERE user_id = ? AND status = 'active' AND end_time > datetime('now')`,
				userID,
			).Scan(&maxEnd)
			startTime := now
			if maxEnd.Valid && maxEnd.String != "" {
				if t, err := time.Parse("2006-01-02 15:04:05", maxEnd.String); err == nil && t.After(startTime) {
					startTime = t
				}
			}
			endTime := startTime.AddDate(0, 0, product.Duration)
			res, err := tx.Exec(
				`INSERT INTO memberships (user_id, order_id, product_id, start_time, end_time, status)
				 VALUES (?, ?, ?, ?, ?, 'active')`,
				userID, orderID, productID, startTime, endTime,
			)
			if err != nil {
				return err
			}
			mid, _ := res.LastInsertId()
			_, err = tx.Exec(
				`INSERT INTO membership_logs (user_id, membership_id, action, detail)
				 VALUES (?, ?, 'created', ?)`,
				userID, mid, "会员开通: "+productID,
			)
			if err != nil {
				return err
			}
		}
	} else {
		// 幂等：该订单已发过配额则跳过
		var c int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM quotas WHERE order_id = ? AND quota_type = 'purchase'", orderID,
		).Scan(&c); err != nil {
			return err
		}
		if c == 0 {
			quotaCount := quota
			if quotaCount <= 0 {
				quotaCount = 1
			}
			for i := int64(0); i < quotaCount; i++ {
				if _, err := tx.Exec(
					"INSERT INTO quotas (user_id, quota_type, order_id) VALUES (?, 'purchase', ?)",
					userID, orderID,
				); err != nil {
					return err
				}
			}
		}
	}

	// 金币累加（1元=10金币，amount 单位分）：购买自动记录累计金币，用于等级
	if coinAdd := amount / 10; coinAdd > 0 {
		if _, err := tx.Exec(
			"UPDATE users SET coin_total = coin_total + ? WHERE id = ?",
			coinAdd, userID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SetOrderRefunded 标记订单已退款（仅 paid → refunded）
func (s *Store) SetOrderRefunded(orderID int64) error {
	res, err := s.db.Exec(
		"UPDATE orders SET status = 'refunded' WHERE id = ? AND status = 'paid'",
		orderID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("订单状态不是 paid，无法标记退款")
	}
	return nil
}

// RecycleQuotaByOrder 回收某订单未使用的购买配额（退款时用）
func (s *Store) RecycleQuotaByOrder(orderID int64) error {
	_, err := s.db.Exec(
		"DELETE FROM quotas WHERE order_id = ? AND used_at IS NULL",
		orderID,
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

// ListAllOrders 管理员：列出全部订单
func (s *Store) ListAllOrders(limit, offset int) ([]store.Order, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, amount, quota, COALESCE(product_id,''), COALESCE(channel,''), status,
		        COALESCE(out_trade_no,''), COALESCE(prepay_id,''), COALESCE(code_url,''),
		        paid_at, created_at
		 FROM orders ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
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
			&o.Channel, &o.Status, &o.OutTradeNo, &o.PrepayID, &o.CodeURL,
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

// CountAllOrders 统计全部订单数
func (s *Store) CountAllOrders() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	return count, err
}
