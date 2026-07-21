package handler

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// OrderHandler 微信支付订单处理器
type OrderHandler struct {
	store     store.Store
	mchID     string // 商户号
	apiKey    string // API密钥
	appID     string // 公众号AppID
	notifyURL string // 回调地址
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(st store.Store, mchID, apiKey, appID, notifyURL string) *OrderHandler {
	return &OrderHandler{
		store:     st,
		mchID:     mchID,
		apiKey:    apiKey,
		appID:     appID,
		notifyURL: notifyURL,
	}
}

// products 商品定义（与前端 Recharge.vue 一致）
var products = map[string]*store.OrderProduct{
	"test": {
		ID:     "test",
		Name:  "测试包",
		Amount: 1,
		Quota:  1,
	},
	"trial": {
		ID:     "trial",
		Name:   "尝鲜包",
		Amount: 500,
		Quota:  10,
	},
	"standard": {
		ID:     "standard",
		Name:   "标准包",
		Amount: 2000,
		Quota:  50,
	},
	"unlimited": {
		ID:     "unlimited",
		Name:   "畅享包",
		Amount: 6000,
		Quota:  200,
	},
}

// createOrderReq 下单请求
type createOrderReq struct {
	ProductID string `json:"product_id"`
}

// CreateOrder 创建订单（POST /api/orders/create）
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	product, ok := products[req.ProductID]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "商品不存在"})
		return
	}

	outTradeNo := generateOutTradeNo()

	codeURL, err := h.wechatPayNative(product, outTradeNo)
	if err != nil {
		log.Printf("微信支付下单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "支付下单失败"})
		return
	}

	order, err := h.store.CreateOrder(userID, product, outTradeNo, codeURL)
	if err != nil {
		log.Printf("创建订单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建订单失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           order.ID,
		"amount":       order.Amount,
		"quota":        order.Quota,
		"code_url":     order.CodeURL,
		"out_trade_no": order.OutTradeNo,
	})
}

// jsapiCreateOrderReq JSAPI 下单请求
type jsapiCreateOrderReq struct {
	ProductID string `json:"product_id"`
	OpenID    string `json:"openid"`
}

// CreateJSAPIOrder 小程序下单（POST /api/orders/jsapi-create）
func (h *OrderHandler) CreateJSAPIOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var req jsapiCreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.OpenID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 openid"})
		return
	}

	product, ok := products[req.ProductID]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "商品不存在"})
		return
	}

	outTradeNo := generateOutTradeNo()

	prepayID, err := h.wechatPayJSAPI(product, outTradeNo, req.OpenID)
	if err != nil {
		log.Printf("微信支付JSAPI下单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "支付下单失败"})
		return
	}

	order, err := h.store.CreateOrder(userID, product, outTradeNo, "")
	if err != nil {
		log.Printf("创建订单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建订单失败"})
		return
	}

	// 生成 JSAPI 调起支付参数
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := randStr(16)
	packageStr := "prepay_id=" + prepayID
	payParams := map[string]string{
		"appId":     h.appID,
		"timeStamp": timeStamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "MD5",
	}
	paySign := wechatSign(payParams, h.apiKey)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           order.ID,
		"amount":       order.Amount,
		"quota":        order.Quota,
		"out_trade_no": order.OutTradeNo,
		"payment": map[string]string{
			"appId":     h.appID,
			"timeStamp": timeStamp,
			"nonceStr":  nonceStr,
			"package":   packageStr,
			"signType":  "MD5",
			"paySign":   paySign,
		},
	})
}

// GetOrder 获取订单详情（GET /api/orders/{id}）
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var id int64
	if _, err := fmt.Sscanf(r.PathValue("id"), "%d", &id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数错误"})
		return
	}

	order, err := h.store.GetOrder(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权访问"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"order": order})
}

// ListOrders 列出用户订单（GET /api/orders）
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	limit := 20
	offset := 0

	orders, err := h.store.ListOrders(userID, limit, offset)
	if err != nil {
		log.Printf("查询订单列表失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	if orders == nil {
		orders = []store.Order{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": orders,
	})
}

// ========== XML types for WeChat Pay ==========

// wechatPayRequest 微信支付统一下单请求
type wechatPayRequest struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       string   `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	OpenID         string   `xml:"openid,omitempty"`
}

// wechatPayResponse 微信支付统一下单响应
type wechatPayResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	AppID      string   `xml:"appid,omitempty"`
	MchID      string   `xml:"mch_id,omitempty"`
	NonceStr   string   `xml:"nonce_str,omitempty"`
	Sign       string   `xml:"sign,omitempty"`
	ResultCode string   `xml:"result_code,omitempty"`
	ErrCode    string   `xml:"err_code,omitempty"`
	ErrCodeDes string   `xml:"err_code_des,omitempty"`
	PrepayID   string   `xml:"prepay_id,omitempty"`
	TradeType  string   `xml:"trade_type,omitempty"`
	CodeURL    string   `xml:"code_url,omitempty"`
}

// wechatNotifyResponse 微信支付回调响应
type wechatNotifyResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
}

// wechatNotifyParams 微信支付回调通知参数（map 形式）
type wechatNotifyParams map[string]string

// WechatNotify 微信支付回调通知（POST /api/orders/notify）
func (h *OrderHandler) WechatNotify(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeWechatResponse(w, "FAIL", "读取通知失败")
		return
	}

	params, err := parseWechatXML(body)
	if err != nil {
		writeWechatResponse(w, "FAIL", "解析XML失败")
		return
	}

	// 验签
	sign := params["sign"]
	if sign == "" {
		writeWechatResponse(w, "FAIL", "缺少签名")
		return
	}

	expectedSign := wechatSign(params, h.apiKey)
	if !strings.EqualFold(sign, expectedSign) {
		log.Printf("微信回调签名验证失败: got %s, expected %s", sign, expectedSign)
		writeWechatResponse(w, "FAIL", "签名验证失败")
		return
	}

	// 检查业务结果
	if params["return_code"] != "SUCCESS" || params["result_code"] != "SUCCESS" {
		writeWechatResponse(w, "SUCCESS", "OK")
		return
	}

	outTradeNo := params["out_trade_no"]
	prepayID := params["prepay_id"]

	order, err := h.store.GetOrderByOutTradeNo(outTradeNo)
	if err != nil {
		log.Printf("订单查询失败 out_trade_no=%s: %v", outTradeNo, err)
		writeWechatResponse(w, "SUCCESS", "OK") // 幂等：已处理当作成功
		return
	}

	if order.Status == "paid" {
		// 已支付，避免重复处理
		writeWechatResponse(w, "SUCCESS", "OK")
		return
	}

	if err := h.store.MarkOrderPaid(order.ID, prepayID); err != nil {
		log.Printf("标记订单支付失败 id=%d: %v", order.ID, err)
		writeWechatResponse(w, "FAIL", "更新订单失败")
		return
	}

	// 增加配额
	for i := 0; i < order.Quota; i++ {
		if err := h.store.AddQuota(order.UserID, "purchase"); err != nil {
			log.Printf("添加配额失败 user_id=%d: %v", order.UserID, err)
		}
	}

	writeWechatResponse(w, "SUCCESS", "OK")
}

// wechatPayNative 调用微信支付 Native 下单
func (h *OrderHandler) wechatPayNative(product *store.OrderProduct, outTradeNo string) (string, error) {
	if h.mchID == "" {
		return "wechat://pay/test?order=" + outTradeNo, nil
	}
	return h.unifiedOrder(product, outTradeNo, "NATIVE", "")
}

// wechatPayJSAPI 调用微信支付 JSAPI 下单（小程序）
func (h *OrderHandler) wechatPayJSAPI(product *store.OrderProduct, outTradeNo, openid string) (string, error) {
	if h.mchID == "" {
		return "", fmt.Errorf("微信支付未配置")
	}
	return h.unifiedOrder(product, outTradeNo, "JSAPI", openid)
}

// unifiedOrder 统一下单（共用逻辑）
func (h *OrderHandler) unifiedOrder(product *store.OrderProduct, outTradeNo, tradeType, openid string) (string, error) {
	nonceStr := randStr(16)
	params := map[string]string{
		"appid":            h.appID,
		"mch_id":           h.mchID,
		"nonce_str":        nonceStr,
		"body":             "易观-占卜次数",
		"out_trade_no":     outTradeNo,
		"total_fee":        fmt.Sprintf("%d", product.Amount),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       h.notifyURL,
		"trade_type":       tradeType,
	}
	if openid != "" {
		params["openid"] = openid
	}

	sign := wechatSign(params, h.apiKey)
	params["sign"] = sign

	req := wechatPayRequest{
		AppID:          h.appID,
		MchID:          h.mchID,
		NonceStr:       nonceStr,
		Sign:           sign,
		Body:           "易观-占卜次数",
		OutTradeNo:     outTradeNo,
		TotalFee:       fmt.Sprintf("%d", product.Amount),
		SpbillCreateIP: "127.0.0.1",
		NotifyURL:      h.notifyURL,
		TradeType:      tradeType,
		OpenID:         openid,
	}

	xmlBody, err := xml.MarshalIndent(req, "", "")
	if err != nil {
		return "", fmt.Errorf("构建XML请求失败: %w", err)
	}

	log.Printf("微信支付下单请求: %s", string(xmlBody))

	resp, err := http.Post("https://api.mch.weixin.qq.com/pay/unifiedorder", "application/xml", bytes.NewReader(xmlBody))
	if err != nil {
		return "", fmt.Errorf("请求微信支付失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("微信支付响应: %s", string(bodyBytes))

	var payResp wechatPayResponse
	if err := xml.Unmarshal(bodyBytes, &payResp); err != nil {
		return "", fmt.Errorf("解析微信支付响应失败: %w", err)
	}

	if payResp.ReturnCode != "SUCCESS" || payResp.ResultCode != "SUCCESS" {
		return "", fmt.Errorf("微信支付下单失败: return_code=%s, result_code=%s, err_code=%s, err_msg=%s",
			payResp.ReturnCode, payResp.ResultCode, payResp.ErrCode, payResp.ErrCodeDes)
	}

	if tradeType == "JSAPI" {
		return payResp.PrepayID, nil
	}
	return payResp.CodeURL, nil
}

// generateOutTradeNo 生成商户订单号
func generateOutTradeNo() string {
	return "YG" + time.Now().Format("20060102150405") + randStr(4)
}

// randStr 生成随机字符串
func randStr(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

// wechatSign 微信支付签名（MD5）
func wechatSign(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for i, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
	}
	buf.WriteString("&key=")
	buf.WriteString(apiKey)

	hash := md5.Sum([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// parseWechatXML 将微信支付 XML 解析为 map（用于验签）
func parseWechatXML(data []byte) (wechatNotifyParams, error) {
	params := make(wechatNotifyParams)

	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "xml" {
				continue
			}
			var content string
			if err := decoder.DecodeElement(&content, &se); err != nil {
				return nil, err
			}
			params[se.Name.Local] = content
		}
	}

	return params, nil
}

// writeWechatResponse 写入微信回调通知响应（XML）
func writeWechatResponse(w http.ResponseWriter, returnCode, returnMsg string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	resp := wechatNotifyResponse{
		ReturnCode: returnCode,
		ReturnMsg:  returnMsg,
	}
	xml.NewEncoder(w).Encode(resp)
}
