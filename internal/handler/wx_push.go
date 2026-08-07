package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

// ========== 微信消息推送协议（虚拟支付发货订阅走此通道） ==========
//
// 配置位置：mp.weixin.qq.com → 开发 → 开发管理 → 消息推送
//   - URL：https://zgjz.insightj.cn/api/orders/virtual-notify
//   - Token / EncodingAESKey：见 SECRETS.md「微信消息推送」小节
//
// 协议：
//  1. 保存配置时微信 GET 验证：signature=sha1(排序[token,timestamp,nonce])，通过则回显 echostr
//  2. 事件推送 POST：明文模式 body 直接是 XML；安全模式 body 为
//     <xml><Encrypt>base64</Encrypt></xml>，query 带 msg_signature=sha1(排序[token,timestamp,nonce,encrypt])，
//     密文用 AES-256-CBC（key=base64(EncodingAESKey+"=")，iv=key前16字节）解密，
//     明文 = 16字节随机 + 4字节大端消息长度 + 消息XML + appid

// verifyPushSignature 消息推送地址验证签名：sha1(排序拼接[token, timestamp, nonce])
func (h *VirtualPayHandler) verifyPushSignature(signature, timestamp, nonce string) bool {
	if signature == "" || h.pushToken == "" {
		return false
	}
	arr := []string{h.pushToken, timestamp, nonce}
	sort.Strings(arr)
	hsh := sha1.Sum([]byte(strings.Join(arr, "")))
	return hex.EncodeToString(hsh[:]) == signature
}

// verifyPushMsgSignature 加密推送验签：sha1(排序拼接[token, timestamp, nonce, encrypt])
func (h *VirtualPayHandler) verifyPushMsgSignature(msgSignature, timestamp, nonce, encrypt string) bool {
	if msgSignature == "" || h.pushToken == "" {
		return false
	}
	arr := []string{h.pushToken, timestamp, nonce, encrypt}
	sort.Strings(arr)
	hsh := sha1.Sum([]byte(strings.Join(arr, "")))
	return hex.EncodeToString(hsh[:]) == msgSignature
}

// decryptWxMsg 解密微信消息推送密文（AES-256-CBC + PKCS7）
func decryptWxMsg(encodingAESKey, encryptedB64 string) (string, error) {
	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return "", fmt.Errorf("EncodingAESKey 解码失败: %w", err)
	}
	if len(aesKey) != 32 {
		return "", fmt.Errorf("EncodingAESKey 长度非法: %d", len(aesKey))
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", fmt.Errorf("密文解码失败: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("AES 初始化失败: %w", err)
	}
	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("密文长度非法: %d", len(ciphertext))
	}
	iv := aesKey[:16]
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)
	// PKCS7 去填充
	padding := int(plain[len(plain)-1])
	if padding < 1 || padding > aes.BlockSize {
		return "", fmt.Errorf("PKCS7 填充非法: %d", padding)
	}
	plain = plain[:len(plain)-padding]
	// 明文结构：16字节随机 + 4字节大端消息长度 + 消息XML + appid
	if len(plain) < 20 {
		return "", fmt.Errorf("解密明文过短: %d", len(plain))
	}
	msgLen := binary.BigEndian.Uint32(plain[16:20])
	if len(plain) < 20+int(msgLen) {
		return "", fmt.Errorf("解密消息长度非法: len=%d msgLen=%d", len(plain), msgLen)
	}
	return string(plain[20 : 20+msgLen]), nil
}

// extractEncrypt 从 <xml><Encrypt>xxx</Encrypt></xml> 提取密文
func extractEncrypt(body []byte) (string, error) {
	var wrap struct {
		XMLName xml.Name `xml:"xml"`
		Encrypt string   `xml:"Encrypt"`
	}
	if err := xml.Unmarshal(body, &wrap); err != nil {
		return "", err
	}
	if wrap.Encrypt == "" {
		return "", fmt.Errorf("缺少 Encrypt 字段")
	}
	return wrap.Encrypt, nil
}

// VirtualNotifyVerify 消息推送地址验证（GET /api/orders/virtual-notify）
// 微信保存消息推送配置时会立即调用；验证通过回显 echostr。
func (h *VirtualPayHandler) VirtualNotifyVerify(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	signature := q.Get("signature")
	timestamp := q.Get("timestamp")
	nonce := q.Get("nonce")
	echostr := q.Get("echostr")
	if h.verifyPushSignature(signature, timestamp, nonce) {
		fmt.Fprint(w, echostr)
		return
	}
	log.Printf("消息推送地址验证失败: sig=%s ts=%s nonce=%s", signature, timestamp, nonce)
	http.Error(w, "verify failed", http.StatusForbidden)
}

// decryptPushBody 推送 POST 体统一处理：安全模式解密为明文 XML，明文模式原样返回
func (h *VirtualPayHandler) decryptPushBody(r *http.Request, body []byte) ([]byte, error) {
	// 无 Encrypt 字段即明文模式
	if !strings.Contains(string(body), "<Encrypt>") {
		return body, nil
	}
	if h.pushAESKey == "" {
		return nil, fmt.Errorf("收到加密推送但未配置 EncodingAESKey")
	}
	encrypt, err := extractEncrypt(body)
	if err != nil {
		return nil, err
	}
	q := r.URL.Query()
	if !h.verifyPushMsgSignature(q.Get("msg_signature"), q.Get("timestamp"), q.Get("nonce"), encrypt) {
		return nil, fmt.Errorf("加密推送签名验证失败")
	}
	plain, err := decryptWxMsg(h.pushAESKey, encrypt)
	if err != nil {
		return nil, err
	}
	return []byte(plain), nil
}

var _ = io.ReadAll
