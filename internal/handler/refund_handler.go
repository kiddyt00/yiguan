package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// RefundHandler 退款处理器
type RefundHandler struct {
	store store.Store
}

func NewRefundHandler(st store.Store) *RefundHandler {
	return &RefundHandler{store: st}
}

// RequestRefund 用户申请退款（POST /api/orders/{id}/refund）
func (h *RefundHandler) RequestRefund(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	var id int64
	if _, err := scanPathInt(r, "id", &id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数错误"})
		return
	}

	order, err := h.store.GetOrder(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "订单不存在"})
		return
	}
	if order.UserID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权操作"})
		return
	}
	if order.Status != "paid" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "订单未支付或已退款"})
		return
	}

	// 资格校验：24小时内
	if order.PaidAt == nil || time.Since(*order.PaidAt) > 24*time.Hour {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "已超过24小时退款期限"})
		return
	}

	// 资格校验：使用次数 ≤ 8
	product := store.FindProduct(order.ProductID)
	if product == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "商品不存在"})
		return
	}

	usageCount := 0
	if product.IsMembership() {
		// 会员卡：统计会员有效期内测算次数
		m, err := h.store.GetMembershipByOrderID(order.ID)
		if err == nil {
			history, _ := h.store.GetUserHistory(userID, 999, 0)
			for _, hh := range history {
				if !hh.CreatedAt.Before(m.StartTime) && !hh.CreatedAt.After(m.EndTime) {
					usageCount++
				}
			}
		}
	} else {
		// 单次：检查该订单的quota是否已使用
		usageCount = order.Quota
		remaining, _ := h.store.GetRemainingQuota(userID)
		if remaining < order.Quota {
			usageCount = order.Quota - remaining
		} else {
			usageCount = 0
		}
	}

	if usageCount > 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "该订单使用次数超过8次，不支持退款"})
		return
	}

	// 创建退款申请
	refund, err := h.store.CreateRefund(userID, id, "用户申请退款")
	if err != nil {
		log.Printf("创建退款失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建退款失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"refund": refund})
}

// ListRefunds 用户查看自己的退款记录
func (h *RefundHandler) ListRefunds(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	refunds, err := h.store.ListRefunds(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	if refunds == nil {
		refunds = []store.Refund{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": refunds})
}

// scanPathInt 从路径参数读取整数
func scanPathInt(r *http.Request, name string, val *int64) (bool, error) {
	_, err := scanInt(r.PathValue(name), val)
	return err == nil, err
}

func scanInt(s string, val *int64) (bool, error) {
	if s == "" {
		return false, nil
	}
	_, err := fmt.Sscanf(s, "%d", val)
	return err == nil, err
}
