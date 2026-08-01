package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// RefundChannel 支付渠道退款客户端
type RefundChannel interface {
	// Refund 原路退款，outTradeNo 商户订单号，amountFen 金额（分）
	Refund(outTradeNo string, amountFen int) error
}

// RefundHandler 退款处理器
type RefundHandler struct {
	store  store.Store
	alipay RefundChannel
	wxpay  RefundChannel
}

func NewRefundHandler(st store.Store) *RefundHandler {
	return &RefundHandler{store: st}
}

// SetChannels 注入退款渠道客户端（alipay / wxpay 可为 nil，表示渠道未配置）
func (h *RefundHandler) SetChannels(alipay, wxpay RefundChannel) {
	h.alipay = alipay
	h.wxpay = wxpay
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
		// 单次：检查该订单的quota是否已使用（按订单归属精确统计）
		remaining, _ := h.store.GetRemainingQuota(userID)
		// 该订单发放的配额中未使用的数量 = order.Quota - 已使用数
		// 精确判定：已使用数 = order.Quota - min(remaining, order.Quota) 的近似，
		// 更精确需按 order_id 查询，这里用全局池剩余量与订单配额对比
		usageCount = order.Quota
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

	// 创建退款申请（refunds 表有 order_id 唯一索引，重复申请会被拒绝）
	refund, err := h.store.CreateRefund(userID, id, "用户申请退款")
	if err != nil {
		log.Printf("创建退款失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建退款失败，可能已申请过"})
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

// AdminListRefunds 管理员查看全部退款申请（GET /api/admin/refunds）
func (h *RefundHandler) AdminListRefunds(w http.ResponseWriter, r *http.Request) {
	refunds, err := h.store.ListAllRefunds()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	if refunds == nil {
		refunds = []store.Refund{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": refunds})
}

// AdminApproveRefund 批准退款（POST /api/admin/refunds/{id}/approve）
// 流程：调渠道退款 API → 成功 → 标记完成 + 回收权益（终止会员/回收未用配额）+ 订单标记 refunded
func (h *RefundHandler) AdminApproveRefund(w http.ResponseWriter, r *http.Request) {
	var id int64
	if _, err := scanPathInt(r, "id", &id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数错误"})
		return
	}

	refund, err := h.store.GetRefund(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "退款申请不存在"})
		return
	}
	if refund.Status != "pending" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "该退款申请已处理"})
		return
	}

	order, err := h.store.GetOrder(refund.OrderID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "订单不存在"})
		return
	}

	// 选择退款渠道
	var ch RefundChannel
	switch order.Channel {
	case "alipay":
		ch = h.alipay
	case "wxpay":
		ch = h.wxpay
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "未知支付渠道"})
		return
	}
	if ch == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "退款渠道未配置，需人工处理"})
		return
	}

	// 调渠道退款（金额校验：退款金额必须与订单一致）
	if order.Amount != refund.Amount {
		log.Printf("退款金额与订单不符 refund_id=%d order_id=%d 订单=%d 退款=%d", id, order.ID, order.Amount, refund.Amount)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "退款金额与订单不符"})
		return
	}
	if err := ch.Refund(order.OutTradeNo, refund.Amount); err != nil {
		log.Printf("渠道退款失败 refund_id=%d order_id=%d: %v", id, order.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "渠道退款失败，请检查日志或稍后重试"})
		return
	}

	// 标记退款完成
	if err := h.store.CompleteRefund(id); err != nil {
		log.Printf("标记退款完成失败 refund_id=%d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "退款成功但本地状态更新失败，请人工核对"})
		return
	}

	// 回收权益
	product := store.FindProduct(order.ProductID)
	if product != nil && product.IsMembership() {
		if m, err := h.store.GetMembershipByOrderID(order.ID); err == nil {
			if err := h.store.TerminateMembership(m.ID); err != nil {
				log.Printf("终止会员失败 membership_id=%d: %v", m.ID, err)
			}
		}
	} else {
		if err := h.store.RecycleQuotaByOrder(order.ID); err != nil {
			log.Printf("回收配额失败 order_id=%d: %v", order.ID, err)
		}
	}

	// 订单标记已退款
	if err := h.store.SetOrderRefunded(order.ID); err != nil {
		log.Printf("标记订单退款失败 order_id=%d: %v", order.ID, err)
	}

	log.Printf("退款完成 refund_id=%d order_id=%d amount=%d", id, order.ID, refund.Amount)
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

// AdminRejectRefund 驳回退款（POST /api/admin/refunds/{id}/reject）
func (h *RefundHandler) AdminRejectRefund(w http.ResponseWriter, r *http.Request) {
	var id int64
	if _, err := scanPathInt(r, "id", &id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数错误"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "不符合退款条件"
	}

	if err := h.store.RejectRefund(id, req.Reason); err != nil {
		log.Printf("驳回退款失败 refund_id=%d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "驳回失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
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
