package sqlite

import (
	"testing"

	"github.com/kiddyt00/yiguan/internal/store"
)

// TestOrderStatusChannelRoundtrip 回归测试：所有订单查询方法的
// Status / Channel 字段必须正确映射（曾因 Scan 顺序错位导致
// Status 读到 channel 值、Channel 读到 status 值）。
func TestOrderStatusChannelRoundtrip(t *testing.T) {
	s := openTestDB(t)

	product := store.FindProduct("single")
	if product == nil {
		t.Fatal("product 'single' not found")
	}
	o, err := s.CreateOrder(1, product, "TEST-SCAN-001", "http://qr", "wxpay")
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	check := func(name string, o *store.Order) {
		t.Helper()
		if o.Status != "pending" {
			t.Errorf("%s status = %q, want pending", name, o.Status)
		}
		if o.Channel != "wxpay" {
			t.Errorf("%s channel = %q, want wxpay", name, o.Channel)
		}
	}

	got, err := s.GetOrder(o.ID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	check("GetOrder", got)

	got2, err := s.GetOrderByOutTradeNo("TEST-SCAN-001")
	if err != nil {
		t.Fatalf("GetOrderByOutTradeNo: %v", err)
	}
	check("GetOrderByOutTradeNo", got2)

	list, err := s.ListOrders(1, 10, 0)
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ListOrders len = %d, want 1", len(list))
	}
	check("ListOrders", &list[0])

	all, err := s.ListAllOrders(10, 0)
	if err != nil {
		t.Fatalf("ListAllOrders: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("ListAllOrders len = %d, want 1", len(all))
	}
	check("ListAllOrders", &all[0])

	// MarkOrderPaid 后 status 应变 paid、channel 保持
	ok, err := s.MarkOrderPaid(o.ID, "txn001")
	if err != nil || !ok {
		t.Fatalf("MarkOrderPaid: ok=%v err=%v", ok, err)
	}
	paid, _ := s.GetOrder(o.ID)
	if paid.Status != "paid" {
		t.Errorf("paid status = %q, want paid", paid.Status)
	}
	if paid.Channel != "wxpay" {
		t.Errorf("paid channel = %q, want wxpay", paid.Channel)
	}
}

// TestOrderChannelAlipay 验证支付宝渠道订单的 channel 字段
func TestOrderChannelAlipay(t *testing.T) {
	s := openTestDB(t)
	product := store.FindProduct("monthly")
	o, err := s.CreateOrder(2, product, "TEST-SCAN-002", "http://pay", "alipay")
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	got, err := s.GetOrder(o.ID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if got.Channel != "alipay" {
		t.Errorf("channel = %q, want alipay", got.Channel)
	}
	if got.Status != "pending" {
		t.Errorf("status = %q, want pending", got.Status)
	}
}
