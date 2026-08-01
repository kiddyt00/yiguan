package handler

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// AlipayHandler 支付宝支付处理器
type AlipayHandler struct {
	store        store.Store
	appID        string // 支付宝应用ID
	merchantID   string // 商户号
	privateKey   *rsa.PrivateKey
	alipayPubKey *rsa.PublicKey
	notifyURL    string // 异步通知地址
	returnURL    string // 同步回调地址
}

// NewAlipayHandler 创建支付宝处理器
// appID: 支付宝应用ID, merchantID: 商户号, privKeyPath: 应用私钥路径,
// aliPubKeyPath: 支付宝公钥路径, notifyURL: 异步通知, returnURL: 同步回调
func NewAlipayHandler(st store.Store, appID, merchantID, privKeyPath, aliPubKeyPath, notifyURL, returnURL string) (*AlipayHandler, error) {
	privKey, err := loadPrivateKey(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载应用私钥失败: %w", err)
	}

	aliPubKey, err := loadPublicKey(aliPubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝公钥失败: %w", err)
	}

	return &AlipayHandler{
		store:        st,
		appID:        appID,
		merchantID:   merchantID,
		privateKey:   privKey,
		alipayPubKey: aliPubKey,
		notifyURL:    notifyURL,
		returnURL:    returnURL,
	}, nil
}

// loadPrivateKey 加载 RSA 私钥（PKCS8 PEM）
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("PEM 解码失败")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥不是 RSA 密钥")
	}
	return rsaKey, nil
}

// loadPublicKey 加载 RSA 公钥（PKCS1/X509 PEM）
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("PEM 解码失败")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS1 格式
		rsaKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析公钥失败: %w", err)
		}
		return rsaKey, nil
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥不是 RSA 公钥")
	}
	return rsaKey, nil
}

// rsaSign RSA2 (SHA256WithRSA) 签名
func rsaSign(data string, privKey *rsa.PrivateKey) (string, error) {
	hash := sha256.Sum256([]byte(data))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// rsaVerify RSA2 (SHA256WithRSA) 验签
func rsaVerify(data, sign string, pubKey *rsa.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("base64 解码签名失败: %w", err)
	}
	hash := sha256.Sum256([]byte(data))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sig)
}

// buildAlipaySignStr 按支付宝规则构建待签名字符串
// 参数按 key 排序，格式: key1=value1&key2=value2
// includeSignType=true 用于下单签名（网关要求包含 sign_type）
// includeSignType=false 用于回调验签（支付宝回调签名规则排除 sign_type）
func buildAlipaySignStr(params map[string]string, includeSignType bool) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		if !includeSignType && k == "sign_type" {
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
	return buf.String()
}

// isMobile 通过 User-Agent 检测是否移动端浏览器
func isMobile(r *http.Request) bool {
	ua := strings.ToLower(r.UserAgent())
	keywords := []string{
		"mobile", "android", "iphone", "ipad", "ipod",
		"windows phone", "blackberry", "opera mini",
		"iemobile", "webos",
	}
	for _, kw := range keywords {
		if strings.Contains(ua, kw) {
			return true
		}
	}
	return false
}

// CreateAlipayOrder 支付宝下单（POST /api/orders/alipay-create）
func (h *AlipayHandler) CreateAlipayOrder(w http.ResponseWriter, r *http.Request) {
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

	outTradeNo := generateOutTradeNo() // YG + timestamp + random

	// 根据设备类型选择支付方式
	var payURL string
	var err error
	mobile := isMobile(r)
	if mobile {
		payURL, err = h.buildAlipayWapPayURL(product, outTradeNo)
	} else {
		payURL, err = h.buildAlipayPayURL(product, outTradeNo)
	}
	log.Printf("支付宝下单: order=%s is_mobile=%v ua=%s", outTradeNo, mobile, r.UserAgent())
	if err != nil {
		log.Printf("支付宝下单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "支付下单失败"})
		return
	}

	order, err := h.store.CreateOrder(userID, product, outTradeNo, payURL, "alipay")
	if err != nil {
		log.Printf("创建订单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建订单失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           order.ID,
		"amount":       order.Amount,
		"quota":        order.Quota,
		"pay_url":      payURL,
		"out_trade_no": order.OutTradeNo,
	})
}

// buildAlipayPayURL 构造支付宝电脑网站支付 URL
func (h *AlipayHandler) buildAlipayPayURL(product *store.OrderProduct, outTradeNo string) (string, error) {
	// biz_content
	bizContent := map[string]interface{}{
		"subject":      "易观-占卜次数",
		"out_trade_no": outTradeNo,
		"total_amount": fmt.Sprintf("%.2f", float64(product.Amount)/100),
		"product_code": "FAST_INSTANT_TRADE_PAY",
		"body":         fmt.Sprintf("易观占卜 %d 次", product.Quota),
	}
	bizJSON, _ := json.Marshal(bizContent)

	// 公共参数
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	params := map[string]string{
		"app_id":      h.appID,
		"method":      "alipay.trade.page.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   timestamp,
		"version":     "1.0",
		"notify_url":  h.notifyURL,
		"return_url":  h.returnURL,
		"biz_content": string(bizJSON),
	}

	// 签名
	signStr := buildAlipaySignStr(params, true)
	sign, err := rsaSign(signStr, h.privateKey)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}
	params["sign"] = sign

	// 构造 URL（alipay.trade.page.pay 是表单跳转，返回 URL 让前端打开）
	gateway := "https://openapi.alipay.com/gateway.do"
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return gateway + "?" + query.Encode(), nil
}

// buildAlipayWapPayURL 构造支付宝手机网站支付 URL（alipay.trade.wap.pay）
func (h *AlipayHandler) buildAlipayWapPayURL(product *store.OrderProduct, outTradeNo string) (string, error) {
	bizContent := map[string]interface{}{
		"subject":      "易观-占卜次数",
		"out_trade_no": outTradeNo,
		"total_amount": fmt.Sprintf("%.2f", float64(product.Amount)/100),
		"product_code": "QUICK_WAP_PAY",
		"body":         fmt.Sprintf("易观占卜 %d 次", product.Quota),
	}
	bizJSON, _ := json.Marshal(bizContent)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	params := map[string]string{
		"app_id":      h.appID,
		"method":      "alipay.trade.wap.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   timestamp,
		"version":     "1.0",
		"notify_url":  h.notifyURL,
		"return_url":  h.returnURL,
		"biz_content": string(bizJSON),
	}

	signStr := buildAlipaySignStr(params, true)
	sign, err := rsaSign(signStr, h.privateKey)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}
	params["sign"] = sign

	gateway := "https://openapi.alipay.com/gateway.do"
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return gateway + "?" + query.Encode(), nil
}

// AlipayReturn 支付宝同步回调（GET /api/orders/alipay-return）
// 支付成功后支付宝重定向用户回到此 URL
func (h *AlipayHandler) AlipayReturn(w http.ResponseWriter, r *http.Request) {
	// 从 query 中获取支付宝返回的参数
	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	sign := params["sign"]
	if sign == "" {
		http.Error(w, "缺少签名参数", http.StatusBadRequest)
		return
	}
	// Go 的 url.Query() 会把字面 + 解码成空格，但 base64 签名里 + 是有效字符，需恢复
	sign = strings.ReplaceAll(sign, " ", "+")

	signStr := buildAlipaySignStr(params, false)
	if err := rsaVerify(signStr, sign, h.alipayPubKey); err != nil {
		log.Printf("支付宝同步回调验签失败: %v", err)
		http.Error(w, "验签失败", http.StatusBadRequest)
		return
	}

	// 同步回调仅做页面跳转，支付确认在异步通知中完成
	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"] // 支付宝交易号

	log.Printf("支付宝同步回调: out_trade_no=%s, trade_no=%s", outTradeNo, tradeNo)

	// 重定向到前端充值页，由前端展示支付成功提示
	http.Redirect(w, r, "/recharge?alipay_success="+outTradeNo, http.StatusFound)
}

// AlipayNotify 支付宝异步通知（POST /api/orders/alipay-notify）
func (h *AlipayHandler) AlipayNotify(w http.ResponseWriter, r *http.Request) {
	// 读取支付宝 POST 的表单数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAlipayResponse(w, false, "读取通知失败")
		return
	}

	// 支付宝异步通知的 Content-Type 是 application/x-www-form-urlencoded
	values, err := url.ParseQuery(string(body))
	if err != nil {
		writeAlipayResponse(w, false, "解析参数失败")
		return
	}

	// 将表单参数转为 map
	params := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	sign := params["sign"]
	signType := params["sign_type"]
	if sign == "" {
		writeAlipayResponse(w, false, "缺少签名参数")
		return
	}

	signStr := buildAlipaySignStr(params, false)
	if err := rsaVerify(signStr, sign, h.alipayPubKey); err != nil {
		log.Printf("支付宝异步通知验签失败: %v", err)
		writeAlipayResponse(w, false, "验签失败")
		return
	}

	// 检查通知类型
	tradeStatus := params["trade_status"]
	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	_ = signType

	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		log.Printf("支付宝通知: out_trade_no=%s, trade_status=%s (无需处理)", outTradeNo, tradeStatus)
		writeAlipayResponse(w, true, "success")
		return
	}

	order, err := h.store.GetOrderByOutTradeNo(outTradeNo)
	if err != nil {
		log.Printf("订单查询失败 out_trade_no=%s: %v", outTradeNo, err)
		writeAlipayResponse(w, true, "success") // 幂等
		return
	}

	// 金额校验（纵深防御）
	if amt := params["total_amount"]; amt != "" {
		want := fmt.Sprintf("%.2f", float64(order.Amount)/100.0)
		if amt != want {
			log.Printf("支付宝回调金额不匹配 out_trade_no=%s 期望=%s 实际=%s", outTradeNo, want, amt)
			writeAlipayResponse(w, false, "金额不匹配")
			return
		}
	}

	// 原子完成：标记支付 + 发放权益（含会员商品，幂等，重复回调不会重复发放）
	if err := h.store.GrantOrderBenefits(order.ID, tradeNo); err != nil {
		log.Printf("处理支付宝回调失败 order_id=%d: %v", order.ID, err)
		writeAlipayResponse(w, false, "处理失败")
		return
	}

	log.Printf("支付宝支付成功: out_trade_no=%s, trade_no=%s, amount=%d", outTradeNo, tradeNo, order.Amount)
	writeAlipayResponse(w, true, "success")
}

// writeAlipayResponse 写入支付宝异步通知响应
func writeAlipayResponse(w http.ResponseWriter, success bool, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if success {
		w.Write([]byte("success"))
	} else {
		w.Write([]byte("failure"))
	}
}


// Refund 支付宝原路退款（alipay.trade.refund）
func (h *AlipayHandler) Refund(outTradeNo string, amountFen int) error {
	bizContent := map[string]interface{}{
		"out_trade_no":  outTradeNo,
		"refund_amount": fmt.Sprintf("%.2f", float64(amountFen)/100),
	}
	bizJSON, _ := json.Marshal(bizContent)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	params := map[string]string{
		"app_id":      h.appID,
		"method":      "alipay.trade.refund",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   timestamp,
		"version":     "1.0",
		"biz_content": string(bizJSON),
	}

	signStr := buildAlipaySignStr(params, true)
	sign, err := rsaSign(signStr, h.privateKey)
	if err != nil {
		return fmt.Errorf("退款签名失败: %w", err)
	}
	params["sign"] = sign

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://openapi.alipay.com/gateway.do", form)
	if err != nil {
		return fmt.Errorf("请求支付宝退款失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取支付宝退款响应失败: %w", err)
	}

	// 验签响应（防止中间人篡改）
	if err := verifyAlipayRefundResponse(body, h.alipayPubKey); err != nil {
		return err
	}

	var result struct {
		AlipayTradeRefundResponse struct {
			Code   string `json:"code"`
			Msg    string `json:"msg"`
			SubMsg string `json:"sub_msg"`
		} `json:"alipay_trade_refund_response"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析支付宝退款响应失败: %w", err)
	}
	if result.AlipayTradeRefundResponse.Code != "10000" {
		return fmt.Errorf("支付宝退款失败: code=%s msg=%s sub_msg=%s",
			result.AlipayTradeRefundResponse.Code,
			result.AlipayTradeRefundResponse.Msg,
			result.AlipayTradeRefundResponse.SubMsg)
	}
	log.Printf("支付宝退款成功: out_trade_no=%s amount=%d", outTradeNo, amountFen)
	return nil
}

// verifyAlipayRefundResponse 校验支付宝退款响应签名
func verifyAlipayRefundResponse(body []byte, pubKey *rsa.PublicKey) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	signBytes, ok := m["sign"]
	if !ok {
		return fmt.Errorf("响应缺少签名")
	}
	var sign string
	_ = json.Unmarshal(signBytes, &sign)

	respBytes, ok := m["alipay_trade_refund_response"]
	if !ok {
		return fmt.Errorf("响应缺少业务数据")
	}
	var respObj map[string]string
	if err := json.Unmarshal(respBytes, &respObj); err != nil {
		return fmt.Errorf("解析业务数据失败: %w", err)
	}

	keys := make([]string, 0, len(respObj))
	for k := range respObj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		v := respObj[k]
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
	return rsaVerify(buf.String(), sign, pubKey)
}
