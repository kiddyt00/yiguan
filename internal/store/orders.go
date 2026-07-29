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
	ID       string
	Name     string // 显示名
	Quota    int    // 次数（-1 = 不限次/会员卡）
	Amount   int    // 金额（分）
	Duration int    // 有效天数（0 = 长期有效/单次）
}

// 预定义商品（二期定价）
var Products = []OrderProduct{
	{ID: "single", Name: "单次测算", Quota: 1, Amount: 990, Duration: 0},      // ¥9.9 / 1次
	{ID: "monthly", Name: "月卡", Quota: -1, Amount: 2990, Duration: 30},     // ¥29.9 / 30天不限次
	{ID: "quarterly", Name: "季卡", Quota: -1, Amount: 4990, Duration: 90},   // ¥49.9 / 90天不限次
	{ID: "yearly", Name: "年卡", Quota: -1, Amount: 9900, Duration: 365},     // ¥99 / 365天不限次
}

// IsMembership 判断是否是会员卡商品
func (p *OrderProduct) IsMembership() bool {
	return p.Duration > 0
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
