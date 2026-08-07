package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

// coinPriceFen 代币「金币」单价（分/个）。
// 与虚拟支付后台发布值必须一致；当前套餐金额（990/2990/4990/9900 分）
// 均可被 10 整除：99/299/499/999 金币。若后台单价不同需同步修改。
const coinPriceFen = 10

// VirtualPayHandler 微信小程序虚拟支付（米大师）处理器
//
// 体系说明：wx.requestVirtualPayment 是「米大师虚拟支付」，与微信支付 APIv2/APIv3 完全独立。
// 官方文档：https://developers.weixin.qq.com/miniprogram/dev/platform-capabilities/business-capabilities/virtual-payment.html
//   - 前端参数：signData(JSON字符串) + paySig + signature + mode
//   - signData 字段：offerId、buyQuantity、env(0现网/1沙箱)、currencyType(CNY)、outTradeNo、attach、mode
//   - mode: short_series_goods(道具直购) / short_series_coin(代币充值)
//   - paySig = hex(hmac_sha256(appKey, "requestVirtualPayment&" + signData))
//   - signature = hex(hmac_sha256(sessionKey, signData))
//   - 发货通知：消息推送 XML 事件（Event=xpay_coin_pay_notify / xpay_goods_deliver_notify），
//     在虚拟支付后台「基础配置-发货订阅」配置接收地址
type VirtualPayHandler struct {
	store         store.Store
	offerID       string // 支付应用 ID（mp 虚拟支付基础配置）
	appKey        string // 现网 AppKey（env=0）
	sandboxAppKey string // 沙箱 AppKey（env=1，联调用）
	sandbox       bool   // 沙箱模式开关（测试用，上线置 false）
	appID         string // 小程序 AppID（code2session 用）
	appSecret     string // 小程序 AppSecret（code2session 用）
	notifyURL     string // 发货订阅回调地址
}

// NewVirtualPayHandler 创建虚拟支付处理器
func NewVirtualPayHandler(st store.Store, offerID, appKey, sandboxAppKey string, sandbox bool, appID, appSecret, notifyURL string) *VirtualPayHandler {
	return &VirtualPayHandler{
		store:         st,
		offerID:       offerID,
		appKey:        appKey,
		sandboxAppKey: sandboxAppKey,
		sandbox:       sandbox,
		appID:         appID,
		appSecret:     appSecret,
		notifyURL:     notifyURL,
	}
}

// virtualCreateOrderReq 虚拟支付下单请求
type virtualCreateOrderReq struct {
	ProductID string `json:"product_id"`
	// LoginCode 为 wx.login() 返回的临时 code，后端用它换取 session_key 生成用户态签名。
	// 不落库：code2session 后微信侧 session_key 即更新，signature 与微信端校验一致。
	LoginCode string `json:"code"`
}

// CreateVirtualOrder 虚拟支付下单（POST /api/orders/virtual-create）
func (h *VirtualPayHandler) CreateVirtualOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var req virtualCreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	product, ok := products[req.ProductID]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "商品不存在"})
		return
	}
	if h.offerID == "" || h.appKey == "" {
		log.Printf("虚拟支付未配置: offerID=%q appKeySet=%v", h.offerID, h.appKey != "")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "虚拟支付未配置"})
		return
	}
	if req.LoginCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 wx.login code"})
		return
	}

	// 用 wx.login code 换取 session_key（生成用户态签名 signature 必需）
	sessionKey, err := h.exchangeSessionKey(req.LoginCode)
	if err != nil {
		log.Printf("虚拟支付 code2session 失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "会话校验失败，请重试"})
		return
	}

	outTradeNo := generateOutTradeNo()

	// 沙箱/现网切换：沙箱模式 env=1 + 沙箱 AppKey，现网 env=0 + 现网 AppKey
	env := 0
	signKey := h.appKey
	if h.sandbox {
		env = 1
		signKey = h.sandboxAppKey
	}

	// 代币充值模式（short_series_coin）：虚拟支付后台发布代币「金币」，
	// buyQuantity = 购买数量（金币个数），支付金额 = buyQuantity × 金币单价(分)。
	// 金币单价以虚拟支付后台发布值为准，默认 10 分（0.1元/金币）：
	// single=99 / monthly=299 / quarterly=499 / yearly=999 金币（均为整数）。
	// ⚠️ 若后台发布的金币单价不同，需同步修改 coinPriceFen。
	signData := map[string]interface{}{
		"offerId":      h.offerID,
		"buyQuantity":  product.Amount / coinPriceFen,
		"env":          env,
		"currencyType": "CNY",
		"outTradeNo":   outTradeNo,
		"attach":       product.ID,
	}

	// signData 以 JSON 字符串参与签名（官方要求：参与签名的串与传给 wx.requestVirtualPayment 的 signData 完全一致）
	signDataBytes, _ := json.Marshal(signData)
	signDataStr := string(signDataBytes)

	paySig := calcVirtualPaySig(signKey, "requestVirtualPayment", signDataStr)
	signature := calcVirtualUserSignature(sessionKey, signDataStr)

	// 落库（channel="virtual"，区分虚拟支付渠道）
	order, err := h.store.CreateOrder(userID, product, outTradeNo, "", "virtual")
	if err != nil {
		log.Printf("创建订单失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建订单失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           order.ID,
		"amount":       order.Amount,
		"out_trade_no": order.OutTradeNo,
		"virtual": map[string]string{
			"signData":  signDataStr,
			"paySig":    paySig,
			"signature": signature,
			"mode":      "short_series_coin",
		},
	})
}

// exchangeSessionKey 用 wx.login code 换取 session_key（复用公共 code2session）
func (h *VirtualPayHandler) exchangeSessionKey(code string) (string, error) {
	ws, err := wechatCode2Session(h.appID, h.appSecret, code)
	if err != nil {
		return "", err
	}
	if ws.SessionKey == "" {
		return "", fmt.Errorf("session_key 为空")
	}
	return ws.SessionKey, nil
}

// ========== 签名（官方《签名详解》，含可复现验证值，见 virtual_pay_test.go） ==========

// calcVirtualPaySig 支付签名：
//
//	paySig = to_hex(hmac_sha256(appKey, uri + "&" + signData))
//
// uri：wx.requestVirtualPayment 固定填 "requestVirtualPayment"；服务器 API 用接口路径（如 /xpay/query_user_balance）。
// appKey 按 env 区分：env=0 用现网 AppKey，env=1 用沙箱 AppKey。
func calcVirtualPaySig(appKey, uri, signData string) string {
	mac := hmac.New(sha256.New, []byte(appKey))
	mac.Write([]byte(uri + "&" + signData))
	return hex.EncodeToString(mac.Sum(nil))
}

// calcVirtualUserSignature 用户态签名：
//
//	signature = to_hex(hmac_sha256(sessionKey, signData))
func calcVirtualUserSignature(sessionKey, signData string) string {
	mac := hmac.New(sha256.New, []byte(sessionKey))
	mac.Write([]byte(signData))
	return hex.EncodeToString(mac.Sum(nil))
}

// ========== 发货通知（消息推送 XML 事件） ==========

// virtualNotifyXML 虚拟支付消息推送事件
// 推送内容为 XML 格式时，响应也必须是 XML 格式；响应格式不对微信会重推（最多 15 次）。
type virtualNotifyXML struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Event        string   `xml:"Event"`
	OpenID       string   `xml:"OpenId"`
	OutTradeNo   string   `xml:"OutTradeNo"`
	Env          int      `xml:"Env"`
	WeChatPayInfo struct {
		MchOrderNo    string `xml:"MchOrderNo"`
		TransactionID string `xml:"TransactionId"`
		PaidTime      int64  `xml:"PaidTime"`
	} `xml:"WeChatPayInfo"`
	CoinInfo struct {
		Quantity    int    `xml:"Quantity"`
		OrigPrice   int    `xml:"OrigPrice"`
		ActualPrice int    `xml:"ActualPrice"`
		Attach      string `xml:"Attach"`
	} `xml:"CoinInfo"`
}

// VirtualNotify 虚拟支付发货通知（POST /api/orders/virtual-notify）
// 在虚拟支付后台「基础配置 - 发货订阅」配置该地址为接收推送 URL。
func (h *VirtualPayHandler) VirtualNotify(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeVirtualNotifyXML(w, 1, "读取通知失败")
		return
	}
	log.Printf("虚拟支付发货通知: %s", string(body))

	var ev virtualNotifyXML
	if err := xml.Unmarshal(body, &ev); err != nil {
		log.Printf("虚拟支付通知 XML 解析失败: %v", err)
		writeVirtualNotifyXML(w, 1, "解析失败")
		return
	}

	// 只处理发货事件；退款/投诉/风控事件记录后返回成功避免重推
	if ev.Event != "xpay_coin_pay_notify" && ev.Event != "xpay_goods_deliver_notify" {
		log.Printf("虚拟支付忽略非发货事件: %s", ev.Event)
		writeVirtualNotifyXML(w, 0, "success")
		return
	}

	if ev.OutTradeNo == "" {
		log.Printf("虚拟支付发货通知缺少 OutTradeNo")
		writeVirtualNotifyXML(w, 1, "缺少订单号")
		return
	}

	order, err := h.store.GetOrderByOutTradeNo(ev.OutTradeNo)
	if err != nil {
		log.Printf("虚拟支付订单查询失败 out_trade_no=%s: %v", ev.OutTradeNo, err)
		writeVirtualNotifyXML(w, 0, "success") // 幂等：查不到视为已处理
		return
	}

	// 金额校验（CoinInfo.ActualPrice 实际支付金额，单位分）
	if ev.CoinInfo.ActualPrice > 0 && ev.CoinInfo.ActualPrice != order.Amount {
		log.Printf("虚拟支付回调金额不匹配 out_trade_no=%s 期望=%d 实际=%d", ev.OutTradeNo, order.Amount, ev.CoinInfo.ActualPrice)
		writeVirtualNotifyXML(w, 1, "金额不匹配")
		return
	}

	// 原子完成：标记支付 + 发放权益（GrantOrderBenefits 幂等，重复推送不会重复发放）
	txID := ev.WeChatPayInfo.TransactionID
	if txID == "" {
		txID = ev.OutTradeNo
	}
	if err := h.store.GrantOrderBenefits(order.ID, txID); err != nil {
		log.Printf("处理虚拟支付回调失败 order_id=%d: %v", order.ID, err)
		writeVirtualNotifyXML(w, 1, "处理失败")
		return
	}

	writeVirtualNotifyXML(w, 0, "success")
}

// writeVirtualNotifyXML 发货推送响应（XML 格式，ErrCode=0 表示成功）
func writeVirtualNotifyXML(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<xml><ErrCode>%d</ErrCode><ErrMsg><![CDATA[%s]]></ErrMsg></xml>", code, msg)
}
