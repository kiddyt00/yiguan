package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	visitors map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

var globalLimiter = &rateLimiter{
	visitors: make(map[string][]time.Time),
	limit:    60,
	window:   time.Minute,
}

// allLimiters 所有限流器实例，init 定期统一清理
var (
	allLimitersMu sync.Mutex
	allLimiters   = []*rateLimiter{globalLimiter}
)

// clientIP 提取真实客户端 IP：
// 优先 X-Forwarded-For 首值（部署于可信 Nginx 反代之后），
// 再回退 X-Real-IP / RemoteAddr（去掉端口），避免按连接端口绕过限流。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// RateLimit 基于真实 IP 的滑动窗口限流（默认 60次/分钟）
func RateLimit(next http.Handler) http.Handler {
	return RateLimitWithLimit(next, globalLimiter.limit)
}

// RateLimitWithLimit 指定阈值的滑动窗口限流（如认证接口单独收紧）
func RateLimitWithLimit(next http.Handler, limit int) http.Handler {
	lim := &rateLimiter{
		visitors: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Minute,
	}
	allLimitersMu.Lock()
	allLimiters = append(allLimiters, lim)
	allLimitersMu.Unlock()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		lim.mu.Lock()
		now := time.Now()
		times := lim.visitors[ip]

		// 清理超过窗口期的记录
		var recent []time.Time
		for _, t := range times {
			if now.Sub(t) < lim.window {
				recent = append(recent, t)
			}
		}

		if len(recent) >= lim.limit {
			lim.mu.Unlock()
			log.Printf("[rate-limit] %s 请求超限", ip)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"请求过于频繁，请60秒后重试"}`))
			return
		}

		lim.visitors[ip] = append(recent, now)
		lim.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// init 定期清理过期记录，防止内存泄漏
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			allLimitersMu.Lock()
			lims := make([]*rateLimiter, len(allLimiters))
			copy(lims, allLimiters)
			allLimitersMu.Unlock()
			for _, lim := range lims {
				cleanupLimiter(lim)
			}
		}
	}()
}

func cleanupLimiter(lim *rateLimiter) {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	now := time.Now()
	for ip, times := range lim.visitors {
		var recent []time.Time
		for _, t := range times {
			if now.Sub(t) < lim.window {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(lim.visitors, ip)
		} else {
			lim.visitors[ip] = recent
		}
	}
}
