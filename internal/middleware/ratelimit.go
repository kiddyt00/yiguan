package middleware

import (
	"log"
	"net/http"
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

// RateLimit 基于 IP 的滑动窗口限流（默认 60次/分钟）
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		globalLimiter.mu.Lock()
		now := time.Now()
		times := globalLimiter.visitors[ip]

		// 清理超过窗口期的记录
		var recent []time.Time
		for _, t := range times {
			if now.Sub(t) < globalLimiter.window {
				recent = append(recent, t)
			}
		}

		if len(recent) >= globalLimiter.limit {
			globalLimiter.mu.Unlock()
			log.Printf("[rate-limit] %s 请求超限", ip)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"请求过于频繁，请60秒后重试"}`))
			return
		}

		globalLimiter.visitors[ip] = append(recent, now)
		globalLimiter.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// init 定期清理过期记录，防止内存泄漏
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			globalLimiter.mu.Lock()
			now := time.Now()
			for ip, times := range globalLimiter.visitors {
				var recent []time.Time
				for _, t := range times {
					if now.Sub(t) < globalLimiter.window {
						recent = append(recent, t)
					}
				}
				if len(recent) == 0 {
					delete(globalLimiter.visitors, ip)
				} else {
					globalLimiter.visitors[ip] = recent
				}
			}
			globalLimiter.mu.Unlock()
		}
	}()
}
