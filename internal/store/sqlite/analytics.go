package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) SaveLoginLog(log *store.LoginLog) error {
	_, err := s.db.Exec(
		`INSERT INTO login_logs (user_id, ip, city, device, os, browser) VALUES (?, ?, ?, ?, ?, ?)`,
		log.UserID, log.IP, log.City, log.Device, log.OS, log.Browser,
	)
	return err
}

func (s *Store) GetLoginAnalytics() (*store.LoginAnalytics, error) {
	a := &store.LoginAnalytics{}

	// 总登录次数
	s.db.QueryRow("SELECT COUNT(*) FROM login_logs").Scan(&a.TotalLogins)

	// 今日登录
	s.db.QueryRow("SELECT COUNT(*) FROM login_logs WHERE date(created_at) = date('now')").Scan(&a.TodayLogins)

	// 按小时分布
	rows, err := s.db.Query("SELECT CAST(strftime('%H', created_at) AS INTEGER) AS h, COUNT(*) FROM login_logs GROUP BY h ORDER BY h")
	if err == nil {
		defer rows.Close()
		a.LoginByHour = make(map[int]int64)
		for rows.Next() {
			var h int
			var c int64
			rows.Scan(&h, &c)
			a.LoginByHour[h] = c
		}
	}

	// 按城市
	rows2, err := s.db.Query("SELECT COALESCE(NULLIF(city,''),'未知') AS c, COUNT(*) FROM login_logs GROUP BY c ORDER BY COUNT(*) DESC")
	if err == nil {
		defer rows2.Close()
		a.LoginByCity = make(map[string]int64)
		for rows2.Next() {
			var c string
			var n int64
			rows2.Scan(&c, &n)
			a.LoginByCity[c] = n
		}
	}

	// 按设备
	rows3, err := s.db.Query("SELECT COALESCE(NULLIF(device,''),'未知') AS d, COUNT(*) FROM login_logs GROUP BY d ORDER BY COUNT(*) DESC")
	if err == nil {
		defer rows3.Close()
		a.LoginByDevice = make(map[string]int64)
		for rows3.Next() {
			var d string
			var n int64
			rows3.Scan(&d, &n)
			a.LoginByDevice[d] = n
		}
	}

	// 按操作系统
	rows4, err := s.db.Query("SELECT COALESCE(NULLIF(os,''),'未知') AS o, COUNT(*) FROM login_logs GROUP BY o ORDER BY COUNT(*) DESC")
	if err == nil {
		defer rows4.Close()
		a.LoginByOS = make(map[string]int64)
		for rows4.Next() {
			var o string
			var n int64
			rows4.Scan(&o, &n)
			a.LoginByOS[o] = n
		}
	}

	// 性别统计
	rows5, err := s.db.Query(`SELECT CASE WHEN wx_sex=1 THEN '男' WHEN wx_sex=2 THEN '女' ELSE '未知' END, COUNT(*) FROM users GROUP BY wx_sex`)
	if err == nil {
		defer rows5.Close()
		a.GenderStats = make(map[string]int64)
		for rows5.Next() {
			var g string
			var n int64
			rows5.Scan(&g, &n)
			a.GenderStats[g] = n
		}
	}

	// 本周每日登录趋势
	rows6, err := s.db.Query(`
		SELECT date(created_at) as d, COUNT(*) FROM login_logs
		WHERE created_at >= date('now', '-6 days')
		GROUP BY d ORDER BY d`)
	if err == nil {
		defer rows6.Close()
		a.DailyTrend = make(map[string]int64)
		for rows6.Next() {
			var d string
			var n int64
			rows6.Scan(&d, &n)
			a.DailyTrend[d] = n
		}
	}

	return a, nil
}

func (s *Store) GetRecentLogins(limit int) ([]store.LoginLog, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`
		SELECT l.id, l.user_id, l.ip, l.city, l.device, l.os, l.browser, l.created_at,
			COALESCE(u.nickname, '')
		FROM login_logs l LEFT JOIN users u ON l.user_id = u.id
		ORDER BY l.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.LoginLog
	for rows.Next() {
		var l store.LoginLog
		rows.Scan(&l.ID, &l.UserID, &l.IP, &l.City, &l.Device, &l.OS, &l.Browser, &l.CreatedAt, &l.Nickname)
		list = append(list, l)
	}
	return list, rows.Err()
}
