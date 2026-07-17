package store

import "time"

// Order 支付订单
type Order struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Amount      int        `json:"amount"`       // 金额（分）
	Quota       int        `json:"quota"`        // 购买次数
	ProductID   string     `json:"product_id"`   // 套餐标识
	Status      string     `json:"status"`       // pending/paid/failed/refunded
	OutTradeNo  string     `json:"out_trade_no"` // 商户订单号
	PrepayID    string     `json:"-"`
	CodeURL     string     `json:"code_url"`     // 支付二维码URL
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// OrderProduct 套餐定义
type OrderProduct struct {
	ID     string
	Name   string // 显示名
	Quota  int    // 次数
	Amount int    // 金额（分）
}

// 预定义套餐
var Products = []OrderProduct{
	{ID: "test", Name: "测试包", Quota: 1, Amount: 1},          // ¥0.01（仅测试）
	{ID: "trial", Name: "尝鲜包", Quota: 10, Amount: 500},    // ¥5
	{ID: "standard", Name: "标准包", Quota: 50, Amount: 2000}, // ¥20
	{ID: "unlimited", Name: "畅享包", Quota: 200, Amount: 6000}, // ¥60
}

// FindProduct 按 ID 查找套餐
func FindProduct(id string) *OrderProduct {
	for _, p := range Products {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// OrderStore 订单操作
type OrderStore interface {
	CreateOrder(userID int64, product *OrderProduct, outTradeNo, codeURL string) (*Order, error)
	GetOrder(id int64) (*Order, error)
	GetOrderByOutTradeNo(outTradeNo string) (*Order, error)
	ListOrders(userID int64, limit, offset int) ([]Order, error)
	MarkOrderPaid(id int64, prepayID string) error
	UpdateCodeURL(id int64, codeURL string) error
}
