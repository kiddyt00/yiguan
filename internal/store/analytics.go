package store

import "time"

// LoginLog 登录日志
type LoginLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	IP        string    `json:"ip"`
	City      string    `json:"city"`
	Device    string    `json:"device"`
	OS        string    `json:"os"`
	Browser   string    `json:"browser"`
	Nickname  string    `json:"nickname,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginAnalytics 登录分析汇总
type LoginAnalytics struct {
	TotalLogins int64            `json:"total_logins"`
	TodayLogins int64            `json:"today_logins"`
	LoginByHour map[int]int64   `json:"login_by_hour"`
	LoginByCity map[string]int64 `json:"login_by_city"`
	LoginByDevice map[string]int64 `json:"login_by_device"`
	LoginByOS   map[string]int64 `json:"login_by_os"`
	GenderStats map[string]int64 `json:"gender_stats"`
	DailyTrend  map[string]int64 `json:"daily_trend"`
}

// AnalyticsStore 分析数据接口
type AnalyticsStore interface {
	SaveLoginLog(log *LoginLog) error
	GetLoginAnalytics() (*LoginAnalytics, error)
	GetRecentLogins(limit int) ([]LoginLog, error)
}
