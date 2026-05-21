package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

// AnalyticsHandler 用户分析
type AnalyticsHandler struct {
	store store.Store
}

func NewAnalyticsHandler(st store.Store) *AnalyticsHandler {
	return &AnalyticsHandler{store: st}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetLoginAnalytics()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取分析数据失败"})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetRecentLogins(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.GetRecentLogins(30)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取登录记录失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": list})
}

// ========== 登录日志记录 ==========

// parseUserAgent 解析 User-Agent 提取 OS、浏览器、设备类型
func parseUserAgent(ua string) (os, browser, device string) {
	ua = strings.ToLower(ua)

	// 微信内置浏览器
	if strings.Contains(ua, "micromessenger") {
		browser = "微信"
		// 进一步区分 iOS/Android 微信
	}

	// 设备
	switch {
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod"):
		device = "iPhone/iPad"
	case strings.Contains(ua, "android") && strings.Contains(ua, "mobile"):
		device = "Android手机"
	case strings.Contains(ua, "android"):
		device = "Android设备"
	default:
		device = "PC/桌面"
	}

	// 操作系统
	switch {
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		os = "iOS"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "mac os") || strings.Contains(ua, "macintosh"):
		os = "macOS"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	default:
		os = "其他"
	}

	// 浏览器
	if browser == "" {
		switch {
		case strings.Contains(ua, "edg"):
			browser = "Edge"
		case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
			browser = "Chrome"
		case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
			browser = "Safari"
		case strings.Contains(ua, "firefox"):
			browser = "Firefox"
		default:
			browser = "其他"
		}
	}

	return
}

// geoLookup 通过 ip-api.com 免费接口查询 IP 归属地
// 容错：查询失败时返回空字符串，不影响主流程
func geoLookup(ip string) string {
	// 内网/本地 IP 不查
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "172.") {
		return ""
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		City string `json:"city"`
	}
	json.Unmarshal(body, &result)
	return result.City
}

// logLogin 记录登录日志，从请求中提取 IP/UA/设备信息
func logLogin(st store.Store, userID int64, r *http.Request) {
	// 获取真实 IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		// 从 RemoteAddr 提取
		parts := strings.Split(r.RemoteAddr, ":")
		if len(parts) > 0 {
			ip = parts[0]
		}
	}
	// X-Forwarded-For 可能是逗号分隔列表
	if idx := strings.Index(ip, ","); idx > 0 {
		ip = strings.TrimSpace(ip[:idx])
	}

	ua := r.Header.Get("User-Agent")
	os, browser, device := parseUserAgent(ua)

	// 异步查询 IP 归属地（不影响登录响应）
	go func() {
		city := geoLookup(ip)
		if err := st.SaveLoginLog(&store.LoginLog{
			UserID:  userID,
			IP:      ip,
			City:    city,
			Device:  device,
			OS:      os,
			Browser: browser,
		}); err != nil {
			log.Printf("[analytics] 保存登录日志失败: %v", err)
		}
	}()
}
