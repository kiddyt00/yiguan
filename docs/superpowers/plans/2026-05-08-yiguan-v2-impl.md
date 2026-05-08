# 易观 v2.0 全栈增强实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现角色权限、SSE 流式起卦、模型/广告/用户管理、广告激励系统，改造后台前端为完整管理面板

**Architecture:** 后端 Go net/http + SQLite，前端 Vue 3（后台 Element Plus，前台 Tailwind）。SSE 通过 http.Flusher 逐事件推送，前端 EventSource/fetch+ReadableStream 接收。模型热切换通过内存原子指针替换。

**Tech Stack:** Go 1.24, SQLite (modernc.org/sqlite), Vue 3, Element Plus, Tailwind CSS, EventSource API, ReadableStream API

---

## 文件结构总览

### 新建文件

| 文件 | 职责 |
|------|------|
| `internal/store/models.go` | LLM 模型 Store 接口 + 实现 |
| `internal/store/ads.go` | 广告 Store 接口 + 实现 |
| `internal/store/hexagrams.go` | 管理员卦象查询接口 |
| `internal/handler/model_handler.go` | 模型管理 CRUD handler |
| `internal/handler/ad_handler.go` | 广告管理 CRUD + 用户端广告接口 |
| `internal/handler/hexagram_handler.go` | 卦象记录管理（列表/详情/删除） |
| `internal/handler/divine_stream.go` | SSE 流式起卦 handler |
| `internal/llm/stream.go` | LLM SSE 流式调用方法 |
| `internal/llm/router.go` | 模型热切换路由器 |
| `web/admin/src/layout/AdminLayout.vue` | 管理面板主布局（侧边栏+顶栏） |
| `web/admin/src/views/Login.vue` | 后台登录页 |
| `web/admin/src/views/Hexagrams.vue` | 卦象记录管理页 |
| `web/admin/src/views/Models.vue` | 模型管理页 |
| `web/admin/src/views/Ads.vue` | 广告管理页 |
| `web/admin/src/stores/auth.js` | 认证状态管理 |
| `web/admin/src/api/index.js` | API 请求封装 |
| `web/front/src/views/StreamDivine.vue` | 流式起卦页 |
| `web/front/src/views/AdCenter.vue` | 广告中心页 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `internal/store/store.go` | User 加 Role/IsActive 字段，Store 接口扩展 |
| `internal/store/sqlite/sqlite.go` | 迁移增加新表/字段 |
| `internal/store/sqlite/users.go` | SQL 增加 role/is_active |
| `internal/store/sqlite/history.go` | 增加管理员查询方法 |
| `internal/handler/auth.go` | JWT 加入 role 字段 |
| `internal/handler/admin.go` | 增强 Dashboard + 新增用户管理方法 |
| `internal/handler/divine.go` | 增加 LLM Router 参数 |
| `cmd/server/main.go` | 路由注册 + 管理员初始化 + SSE 路由 |
| `internal/middleware/auth.go` | 新增 AdminOnly 中间件 |
| `deploy/nginx.conf` | SSE 代理配置 + admin 路由 |
| `config.yaml` | 增加 admin 配置段 |
| `.env.example` | 增加 ADMIN_PHONE/PASSWORD |
| `Dockerfile.backend` | 复制新文件 |
| `web/admin/src/router/index.js` | 路由守卫 + 新页面路由 |
| `web/admin/src/App.vue` | 改为使用 AdminLayout |
| `web/admin/src/main.js` | Element Plus 引入 |
| `web/admin/package.json` | Element Plus 依赖 |
| `web/front/src/views/Home.vue` | 增加"看广告领次数"入口 |

---

## Task 1: 数据库迁移 + Store 接口扩展

**Files:**
- Modify: `internal/store/store.go`
- Modify: `internal/store/sqlite/sqlite.go`
- Create: `internal/store/models.go`
- Create: `internal/store/ads.go`
- Create: `internal/store/hexagrams.go`

### 步骤 1.1: 扩展 Store 接口和 User 模型

修改 `internal/store/store.go`：

```go
package store

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("record not found")
var ErrQuotaExhausted = errors.New("quota exhausted")

type User struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Address   string    `json:"address,omitempty"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`        // 新增
	IsActive  int       `json:"is_active"`   // 新增
	CreatedAt time.Time `json:"created_at"`
}

type Quota struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	QuotaType string     `json:"quota_type"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type History struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Question       string    `json:"question"`
	PrimaryGua     string    `json:"primary_gua"`
	ChangingGua    string    `json:"changing_gua"`
	YaoPositions   string    `json:"yao_positions"`
	Interpretation string    `json:"interpretation"`
	CreatedAt      time.Time `json:"created_at"`
}

// 新增: LLM 模型
type LLMModel struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Provider   string    `json:"provider"`
	Endpoint   string    `json:"endpoint"`
	APIKey     string    `json:"api_key"`
	IsDefault  int       `json:"is_default"`
	IsEnabled  int       `json:"is_enabled"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

// 新增: 广告
type Ad struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	AdType        string    `json:"ad_type"`
	ContentURL    string    `json:"content_url"`
	WatchDuration int       `json:"watch_duration"`
	RewardQuota   int       `json:"reward_quota"`
	IsEnabled     int       `json:"is_enabled"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
}

// 新增: 广告观看记录
type AdRecord struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	AdID          int64     `json:"ad_id"`
	WatchDuration int       `json:"watch_duration"`
	Status        string    `json:"status"`
	Rewarded      int       `json:"rewarded"`
	CreatedAt     time.Time `json:"created_at"`
}

// 新增: 广告统计
type AdStat struct {
	AdID       int64 `json:"ad_id"`
	AdName     string `json:"ad_name"`
	Total      int64  `json:"total"`
	Completed  int64  `json:"completed"`
	RewardTotal int64 `json:"reward_total"`
}

type Store interface {
	// 用户（现有）
	CreateUser(phone, password, nickname string) (*User, error)
	GetUserByPhone(phone string) (*User, error)
	GetUserByID(id int64) (*User, error)
	UpdateUser(id int64, nickname, address string) error

	// 用户（新增）
	ToggleUser(id int64, active bool) error
	UpdateUserRole(id int64, role string) error
	UpdateUserQuota(userID int64, delta int) error
	GetUserQuota(userID int64) (int, error)

	// Quota
	GetRemainingQuota(userID int64) (int, error)
	AddQuota(userID int64, quotaType string) error
	ConsumeQuota(userID int64) error

	// 历史记录（现有）
	SaveHistory(h *History) error
	GetHistory(userID int64, limit, offset int) ([]History, error)
	GetHistoryCount(userID int64) (int64, error)

	// 历史记录（新增：管理员全量查询）
	ListAllHistory(limit, offset int, userID int64) ([]History, error)
	GetHistoryByID(id int64) (*History, error)
	DeleteHistory(id int64) error

	// 管理（现有）
	GetTotalUsers() (int64, error)
	ListUsers(limit, offset int) ([]User, error)
	GetTodayDivineCount() (int64, error)

	// 管理（新增）
	GetActiveUserCount() (int64, error)
	GetTotalDivineCount() (int64, error)
	GetTodayAdWatchCount() (int64, error)
	GetTotalAdWatchCount() (int64, error)
	GetUserHistory(userID int64, limit, offset int) ([]History, error)

	// LLM 模型
	ListModels() ([]LLMModel, error)
	GetModelByID(id int64) (*LLMModel, error)
	GetDefaultModel() (*LLMModel, error)
	CreateModel(m *LLMModel) error
	UpdateModel(m *LLMModel) error
	DeleteModel(id int64) error
	SetDefaultModel(id int64) error
	ToggleModel(id int64, enabled bool) error

	// 广告
	ListAds() ([]Ad, error)
	ListActiveAds() ([]Ad, error)
	GetAdByID(id int64) (*Ad, error)
	CreateAd(ad *Ad) error
	UpdateAd(ad *Ad) error
	DeleteAd(id int64) error
	ToggleAd(id int64, enabled bool) error
	CreateAdRecord(rec *AdRecord) error
	UpdateAdRecord(rec *AdRecord) error
	GetAdRecord(userID, adID int64) (*AdRecord, error)
	GetAdStats() ([]AdStat, error)

	Close() error
}
```

### 步骤 1.2: 数据库迁移

修改 `internal/store/sqlite/sqlite.go`，在 `migrate` 函数末尾的 schema 列表中追加：

```go
func migrate(db *sql.DB) error {
	schemas := []string{
		// ... 现有建表语句不变 ...

		// users 表新增字段（SQLite 3.35+ 支持 ALTER TABLE ADD COLUMN）
		`CREATE TABLE IF NOT EXISTS users_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone TEXT UNIQUE NOT NULL,
			nickname TEXT NOT NULL,
			avatar TEXT DEFAULT '',
			address TEXT DEFAULT '',
			password TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT OR IGNORE INTO users_new (id, phone, nickname, avatar, address, password, created_at)
		 SELECT id, phone, nickname, avatar, address, password, created_at FROM users`,
		`DROP TABLE IF EXISTS users`,
		`ALTER TABLE users_new RENAME TO users`,

		// LLM 模型表
		`CREATE TABLE IF NOT EXISTS llm_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			provider TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			api_key TEXT NOT NULL,
			is_default INTEGER DEFAULT 0,
			is_enabled INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 广告表
		`CREATE TABLE IF NOT EXISTS ads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			ad_type TEXT DEFAULT 'iframe',
			content_url TEXT NOT NULL,
			watch_duration INTEGER NOT NULL DEFAULT 30,
			reward_quota INTEGER NOT NULL DEFAULT 1,
			is_enabled INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 广告观看记录表
		`CREATE TABLE IF NOT EXISTS ad_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			ad_id INTEGER NOT NULL,
			watch_duration INTEGER NOT NULL DEFAULT 0,
			status TEXT DEFAULT 'watching',
			rewarded INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (ad_id) REFERENCES ads(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ad_records_user_ad ON ad_records(user_id, ad_id)`,

		// ... 现有 index 语句不变 ...
	}
	// ... 后续循环执行不变 ...
}
```

### 步骤 1.3: 创建 models.go

创建 `internal/store/models.go`：

```go
package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListModels() ([]store.LLMModel, error) {
	rows, err := s.db.Query("SELECT id, name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.LLMModel
	for rows.Next() {
		var m store.LLMModel
		if err := rows.Scan(&m.ID, &m.Name, &m.Provider, &m.Endpoint, &m.APIKey, &m.IsDefault, &m.IsEnabled, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (s *Store) GetModelByID(id int64) (*store.LLMModel, error) {
	m := &store.LLMModel{}
	err := s.db.QueryRow(
		"SELECT id, name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models WHERE id = ?",
		id,
	).Scan(&m.ID, &m.Name, &m.Provider, &m.Endpoint, &m.APIKey, &m.IsDefault, &m.IsEnabled, &m.SortOrder, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return m, err
}

func (s *Store) GetDefaultModel() (*store.LLMModel, error) {
	m := &store.LLMModel{}
	err := s.db.QueryRow(
		"SELECT id, name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models WHERE is_default = 1 AND is_enabled = 1 ORDER BY id LIMIT 1",
	).Scan(&m.ID, &m.Name, &m.Provider, &m.Endpoint, &m.APIKey, &m.IsDefault, &m.IsEnabled, &m.SortOrder, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return m, err
}

func (s *Store) CreateModel(m *store.LLMModel) error {
	result, err := s.db.Exec(
		"INSERT INTO llm_models (name, provider, endpoint, api_key, is_default, is_enabled, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)",
		m.Name, m.Provider, m.Endpoint, m.APIKey, m.IsDefault, m.IsEnabled, m.SortOrder,
	)
	if err != nil {
		return err
	}
	m.ID, _ = result.LastInsertId()
	return nil
}

func (s *Store) UpdateModel(m *store.LLMModel) error {
	_, err := s.db.Exec(
		"UPDATE llm_models SET name = ?, provider = ?, endpoint = ?, api_key = ?, is_enabled = ?, sort_order = ? WHERE id = ?",
		m.Name, m.Provider, m.Endpoint, m.APIKey, m.IsEnabled, m.SortOrder, m.ID,
	)
	return err
}

func (s *Store) DeleteModel(id int64) error {
	_, err := s.db.Exec("DELETE FROM llm_models WHERE id = ?", id)
	return err
}

func (s *Store) SetDefaultModel(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("UPDATE llm_models SET is_default = 0"); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE llm_models SET is_default = 1 WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ToggleModel(id int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec("UPDATE llm_models SET is_enabled = ? WHERE id = ?", v, id)
	return err
}
```

### 步骤 1.4: 创建 ads.go

创建 `internal/store/ads.go`：

```go
package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListAds() ([]store.Ad, error) {
	rows, err := s.db.Query("SELECT id, name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order, created_at FROM ads ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Ad
	for rows.Next() {
		var a store.Ad
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.AdType, &a.ContentURL, &a.WatchDuration, &a.RewardQuota, &a.IsEnabled, &a.SortOrder, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) ListActiveAds() ([]store.Ad, error) {
	rows, err := s.db.Query("SELECT id, name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order, created_at FROM ads WHERE is_enabled = 1 ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Ad
	for rows.Next() {
		var a store.Ad
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.AdType, &a.ContentURL, &a.WatchDuration, &a.RewardQuota, &a.IsEnabled, &a.SortOrder, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) GetAdByID(id int64) (*store.Ad, error) {
	a := &store.Ad{}
	err := s.db.QueryRow(
		"SELECT id, name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order, created_at FROM ads WHERE id = ?",
		id,
	).Scan(&a.ID, &a.Name, &a.Description, &a.AdType, &a.ContentURL, &a.WatchDuration, &a.RewardQuota, &a.IsEnabled, &a.SortOrder, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return a, err
}

func (s *Store) CreateAd(ad *store.Ad) error {
	result, err := s.db.Exec(
		"INSERT INTO ads (name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		ad.Name, ad.Description, ad.AdType, ad.ContentURL, ad.WatchDuration, ad.RewardQuota, ad.IsEnabled, ad.SortOrder,
	)
	if err != nil {
		return err
	}
	ad.ID, _ = result.LastInsertId()
	return nil
}

func (s *Store) UpdateAd(ad *store.Ad) error {
	_, err := s.db.Exec(
		"UPDATE ads SET name = ?, description = ?, ad_type = ?, content_url = ?, watch_duration = ?, reward_quota = ?, is_enabled = ?, sort_order = ? WHERE id = ?",
		ad.Name, ad.Description, ad.AdType, ad.ContentURL, ad.WatchDuration, ad.RewardQuota, ad.IsEnabled, ad.SortOrder, ad.ID,
	)
	return err
}

func (s *Store) DeleteAd(id int64) error {
	_, err := s.db.Exec("DELETE FROM ads WHERE id = ?", id)
	return err
}

func (s *Store) ToggleAd(id int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec("UPDATE ads SET is_enabled = ? WHERE id = ?", v, id)
	return err
}

func (s *Store) CreateAdRecord(rec *store.AdRecord) error {
	result, err := s.db.Exec(
		"INSERT INTO ad_records (user_id, ad_id, watch_duration, status, rewarded) VALUES (?, ?, ?, ?, ?)",
		rec.UserID, rec.AdID, rec.WatchDuration, rec.Status, rec.Rewarded,
	)
	if err != nil {
		return err
	}
	rec.ID, _ = result.LastInsertId()
	return nil
}

func (s *Store) UpdateAdRecord(rec *store.AdRecord) error {
	_, err := s.db.Exec(
		"UPDATE ad_records SET watch_duration = ?, status = ?, rewarded = ? WHERE id = ?",
		rec.WatchDuration, rec.Status, rec.Rewarded, rec.ID,
	)
	return err
}

func (s *Store) GetAdRecord(userID, adID int64) (*store.AdRecord, error) {
	r := &store.AdRecord{}
	err := s.db.QueryRow(
		"SELECT id, user_id, ad_id, watch_duration, status, rewarded, created_at FROM ad_records WHERE user_id = ? AND ad_id = ? ORDER BY id DESC LIMIT 1",
		userID, adID,
	).Scan(&r.ID, &r.UserID, &r.AdID, &r.WatchDuration, &r.Status, &r.Rewarded, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return r, err
}

func (s *Store) GetAdStats() ([]store.AdStat, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.name,
			   COUNT(ar.id),
			   SUM(CASE WHEN ar.status = 'completed' THEN 1 ELSE 0 END),
			   SUM(CASE WHEN ar.rewarded = 1 THEN 1 ELSE 0 END)
		FROM ads a LEFT JOIN ad_records ar ON a.id = ar.ad_id
		GROUP BY a.id ORDER BY a.sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []store.AdStat
	for rows.Next() {
		var s store.AdStat
		if err := rows.Scan(&s.AdID, &s.AdName, &s.Total, &s.Completed, &s.RewardTotal); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
```

### 步骤 1.5: 创建 hexagrams.go

创建 `internal/store/hexagrams.go`：

```go
package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListAllHistory(limit, offset int, userID int64) ([]store.History, error) {
	var rows *sql.Rows
	var err error
	if userID > 0 {
		rows, err = s.db.Query(
			`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
			 FROM history WHERE user_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
			userID, limit, offset,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
			 FROM history ORDER BY id DESC LIMIT ? OFFSET ?`,
			limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.History
	for rows.Next() {
		var h store.History
		if err := rows.Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, h)
	}
	return list, rows.Err()
}

func (s *Store) GetHistoryByID(id int64) (*store.History, error) {
	h := &store.History{}
	err := s.db.QueryRow(
		`SELECT id, user_id, question, primary_gua, changing_gua, yao_positions, interpretation, created_at
		 FROM history WHERE id = ?`,
		id,
	).Scan(&h.ID, &h.UserID, &h.Question, &h.PrimaryGua, &h.ChangingGua, &h.YaoPositions, &h.Interpretation, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return h, err
}

func (s *Store) DeleteHistory(id int64) error {
	_, err := s.db.Exec("DELETE FROM history WHERE id = ?", id)
	return err
}

func (s *Store) GetUserHistory(userID int64, limit, offset int) ([]store.History, error) {
	return s.GetHistory(userID, limit, offset)
}

func (s *Store) GetActiveUserCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = 1").Scan(&count)
	return count, err
}

func (s *Store) GetTotalDivineCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&count)
	return count, err
}

func (s *Store) GetTodayAdWatchCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM ad_records WHERE date(created_at) = date('now')").Scan(&count)
	return count, err
}

func (s *Store) GetTotalAdWatchCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM ad_records").Scan(&count)
	return count, err
}

func (s *Store) ToggleUser(id int64, active bool) error {
	v := 0
	if active {
		v = 1
	}
	_, err := s.db.Exec("UPDATE users SET is_active = ? WHERE id = ?", v, id)
	return err
}

func (s *Store) UpdateUserRole(id int64, role string) error {
	_, err := s.db.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	return err
}

func (s *Store) UpdateUserQuota(userID int64, delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		for i := 0; i < delta; i++ {
			if err := s.AddQuota(userID, "admin_grant"); err != nil {
				return err
			}
		}
	} else {
		// delta < 0: 删除未使用的 quota 记录
		count := -delta
		_, err := s.db.Exec(`DELETE FROM quotas WHERE id IN (
			SELECT id FROM quotas WHERE user_id = ? AND used_at IS NULL ORDER BY id LIMIT ?
		)`, userID, count)
		return err
	}
	return nil
}

func (s *Store) GetUserQuota(userID int64) (int, error) {
	return s.GetRemainingQuota(userID)
}
```

### 步骤 1.6: 更新 users.go SQL

修改 `internal/store/sqlite/users.go`，所有 SELECT 语句中增加 `role, is_active` 字段：

```go
// GetUserByPhone — 第 37-40 行改为：
err := s.db.QueryRow(
	"SELECT id, phone, nickname, avatar, address, password, role, is_active, created_at FROM users WHERE phone = ?",
	phone,
).Scan(&u.ID, &u.Phone, &u.Nickname, &u.Avatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

// GetUserByID — 第 53-56 行改为：
err := s.db.QueryRow(
	"SELECT id, phone, nickname, avatar, address, password, role, is_active, created_at FROM users WHERE id = ?",
	id,
).Scan(&u.ID, &u.Phone, &u.Nickname, &u.Avatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

// ListUsers — 第 59-61 行改为：
rows, err := s.db.Query(
	"SELECT id, phone, nickname, avatar, address, role, is_active, created_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?",
	limit, offset,
)
// Scan 增加: &u.Role, &u.IsActive
```

### 步骤 1.7: 编译验证

```bash
cd /tmp/yiguan
GOPROXY=https://goproxy.cn,direct go build ./...
```

- [ ] **Commit**
```bash
git add internal/store/
git commit -m "feat: 扩展 Store 接口，新增模型/广告/管理员查询存储层"
```

---

## Task 2: 管理员初始化 + 认证增强 + SSE 中间件

**Files:**
- Create: `internal/middleware/admin.go`
- Modify: `internal/handler/auth.go`
- Modify: `cmd/server/main.go`

### 步骤 2.1: 创建 AdminOnly 中间件

创建 `internal/middleware/admin.go`：

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

const RoleKey contextKey = "role"

// AdminOnly 管理员权限中间件
func AdminOnly(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearer(r)
			if tokenStr == "" {
				writeAuthError(w, "需要登录")
				return
			}

			claims, err := parseJWT(tokenStr, secret)
			if err != nil {
				writeAuthError(w, "token无效")
				return
			}

			role, _ := claims["role"].(string)
			if role != "admin" {
				writeAuthError(w, "需要管理员权限")
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				writeAuthError(w, "token无效")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

### 步骤 2.2: auth.go 加入 role 到 JWT

修改 `internal/handler/auth.go` 的 `generateToken` 方法和注册/登录逻辑：

```go
// generateToken 改为接受 role
func (h *AuthHandler) generateToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 延长到 24 小时
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secret))
}
```

修改 `register` 方法（第 85 行）：
```go
token, err := h.generateToken(user.ID, "user")
```

修改 `login` 方法（第 112 行后，获取用户后）：
```go
// 检查用户是否被禁用
if user.IsActive == 0 {
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "账号已被禁用"})
	return
}
token, err := h.generateToken(user.ID, user.Role)
```

### 步骤 2.3: 管理员初始化

修改 `cmd/server/main.go`，在 main 函数中，数据库初始化后、路由注册前：

```go
// 管理员初始化
if phone := cfg.Admin.Phone; phone != "" {
	existing, _ := st.GetUserByPhone(phone)
	if existing != nil {
		if existing.Role != "admin" {
			st.UpdateUserRole(existing.ID, "admin")
			log.Printf("已将用户 %s 升级为管理员", phone)
		}
	} else {
		_, err := st.CreateUser(phone, cfg.Admin.Password, "管理员")
		if err != nil {
			log.Printf("创建管理员失败: %v", err)
		} else {
			st.UpdateUserRole(1, "admin") // 确保第一个用户是 admin
			log.Printf("已创建管理员账号: %s", phone)
		}
	}
}
```

Config struct 增加 Admin 字段：
```go
type Config struct {
	Server    struct { Port string `yaml:"port"` } `yaml:"server"`
	LLM       struct {
		Default   string `yaml:"default"`
		Providers map[string]struct {
			APIKey   string `yaml:"api_key"`
			Endpoint string `yaml:"endpoint"`
			Model    string `yaml:"model"`
		} `yaml:"providers"`
	} `yaml:"llm"`
	JWTSecret string `yaml:"jwt_secret"`
	DBPath    string `yaml:"db_path"`
	Admin     struct { Phone string `yaml:"phone"`; Password string `yaml:"password"` } `yaml:"admin"`
}
```

环境变量覆盖：
```go
if phone := os.Getenv("ADMIN_PHONE"); phone != "" {
	cfg.Admin.Phone = phone
}
if pwd := os.Getenv("ADMIN_PASSWORD"); pwd != "" {
	cfg.Admin.Password = pwd
}
```

### 步骤 2.4: 编译验证

```bash
cd /tmp/yiguan
GOPROXY=https://goproxy.cn,direct go build ./...
```

- [ ] **Commit**
```bash
git add internal/middleware/admin.go internal/handler/auth.go cmd/server/main.go
git commit -m "feat: 管理员初始化 + JWT 角色 + AdminOnly 中间件"
```

---

## Task 3: 后端 Handlers — 模型/广告/卦象管理

**Files:**
- Create: `internal/handler/model_handler.go`
- Create: `internal/handler/ad_handler.go`
- Create: `internal/handler/hexagram_handler.go`
- Modify: `internal/handler/admin.go`

### 步骤 3.1: 模型管理 Handler

创建 `internal/handler/model_handler.go`：

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/store"
)

type ModelHandler struct {
	store   store.Store
	onReload func() // 模型变更后回调，用于热切换
}

func NewModelHandler(st store.Store, onReload func()) *ModelHandler {
	return &ModelHandler{store: st, onReload: onReload}
}

func (h *ModelHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.store.ListModels()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取模型列表失败"})
		return
	}
	if models == nil {
		models = []store.LLMModel{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": models, "total": len(models)})
}

func (h *ModelHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var m store.LLMModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if m.Name == "" || m.Endpoint == "" || m.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "名称、endpoint、api_key 不能为空"})
		return
	}
	if err := h.store.CreateModel(&m); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *ModelHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var m store.LLMModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	m.ID = id
	if err := h.store.UpdateModel(&m); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *ModelHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.DeleteModel(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}

func (h *ModelHandler) SetDefaultModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.SetDefaultModel(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "设置默认失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已设为默认"})
}

func (h *ModelHandler) ToggleModel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	enabled := r.URL.Query().Get("enabled") == "true"
	if err := h.store.ToggleModel(id, enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "操作失败"})
		return
	}
	if h.onReload != nil {
		h.onReload()
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
}
```

### 步骤 3.2: 广告管理 Handler

创建 `internal/handler/ad_handler.go`：

```go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

type AdHandler struct {
	store store.Store
}

func NewAdHandler(st store.Store) *AdHandler {
	return &AdHandler{store: st}
}

// === Admin API ===

func (h *AdHandler) ListAds(w http.ResponseWriter, r *http.Request) {
	ads, err := h.store.ListAds()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取广告列表失败"})
		return
	}
	if ads == nil {
		ads = []store.Ad{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": ads, "total": len(ads)})
}

func (h *AdHandler) CreateAd(w http.ResponseWriter, r *http.Request) {
	var ad store.Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if ad.Name == "" || ad.ContentURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "名称和 content_url 不能为空"})
		return
	}
	if ad.WatchDuration <= 0 {
		ad.WatchDuration = 30
	}
	if ad.RewardQuota <= 0 {
		ad.RewardQuota = 1
	}
	if err := h.store.CreateAd(&ad); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	writeJSON(w, http.StatusCreated, ad)
}

func (h *AdHandler) UpdateAd(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var ad store.Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	ad.ID = id
	if err := h.store.UpdateAd(&ad); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	writeJSON(w, http.StatusOK, ad)
}

func (h *AdHandler) DeleteAd(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.DeleteAd(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}

func (h *AdHandler) ToggleAd(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	enabled := r.URL.Query().Get("enabled") == "true"
	if err := h.store.ToggleAd(id, enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "操作失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
}

func (h *AdHandler) GetAdStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetAdStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取统计失败"})
		return
	}
	if stats == nil {
		stats = []store.AdStat{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": stats})
}

// === 用户端 API ===

func (h *AdHandler) ListActiveAds(w http.ResponseWriter, r *http.Request) {
	ads, err := h.store.ListActiveAds()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取广告失败"})
		return
	}
	if ads == nil {
		ads = []store.Ad{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": ads})
}

func (h *AdHandler) StartWatch(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	adID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	// 检查广告是否存在且启用
	ad, err := h.store.GetAdByID(adID)
	if err != nil || ad.IsEnabled == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "广告不存在"})
		return
	}

	rec := &store.AdRecord{
		UserID:        userID,
		AdID:          adID,
		WatchDuration: 0,
		Status:        "watching",
		Rewarded:      0,
	}
	if err := h.store.CreateAdRecord(rec); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建记录失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"record_id": rec.ID, "watch_duration": ad.WatchDuration})
}

func (h *AdHandler) CompleteWatch(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	adID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	var req struct {
		Duration int `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}

	ad, err := h.store.GetAdByID(adID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "广告不存在"})
		return
	}

	if req.Duration < ad.WatchDuration {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   fmt.Sprintf("观看时长不足，还需 %d 秒", ad.WatchDuration-req.Duration),
		})
		return
	}

	// 查找最近一条 watching 状态的记录
	rec, err := h.store.GetAdRecord(userID, adID)
	if err != nil || rec.Status != "watching" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无进行中的观看记录"})
		return
	}

	// 更新记录
	rec.WatchDuration = req.Duration
	rec.Status = "completed"
	rec.Rewarded = 1
	if err := h.store.UpdateAdRecord(rec); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新记录失败"})
		return
	}

	// 发放配额
	for i := 0; i < ad.RewardQuota; i++ {
		h.store.AddQuota(userID, "ad")
	}

	remaining, _ := h.store.GetRemainingQuota(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rewarded":        ad.RewardQuota,
		"remaining_quota": remaining,
	})
}
```

注意：import 需要 `"fmt"`。

### 步骤 3.3: 卦象记录管理 Handler

创建 `internal/handler/hexagram_handler.go`：

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/kiddyt00/yiguan/internal/store"
)

type HexagramHandler struct {
	store store.Store
}

func NewHexagramHandler(st store.Store) *HexagramHandler {
	return &HexagramHandler{store: st}
}

func (h *HexagramHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	items, err := h.store.ListAllHistory(limit, offset, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取记录失败"})
		return
	}
	if items == nil {
		items = []store.History{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "total": len(items)})
}

func (h *HexagramHandler) GetHistoryDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	history, err := h.store.GetHistoryByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "记录不存在"})
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (h *HexagramHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.store.DeleteHistory(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已删除"})
}
```

### 步骤 3.4: 增强 Admin Handler

修改 `internal/handler/admin.go`：

```go
// Dashboard 增强
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	totalUsers, _ := h.store.GetTotalUsers()
	todayDivines, _ := h.store.GetTodayDivineCount()
	totalDivines, _ := h.store.GetTotalDivineCount()
	activeUsers, _ := h.store.GetActiveUserCount()
	adWatchesToday, _ := h.store.GetTodayAdWatchCount()
	totalAdWatches, _ := h.store.GetTotalAdWatchCount()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_users":      totalUsers,
		"active_users":     activeUsers,
		"today_divines":    todayDivines,
		"total_divines":    totalDivines,
		"ad_watches_today": adWatchesToday,
		"total_ads_watched": totalAdWatches,
	})
}

// 增强 ListUsers 支持 role/is_active 筛选
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// 保持现有实现，store 层已返回 role/is_active
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	users, _ := h.store.ListUsers(limit, offset)
	total, _ := h.store.GetTotalUsers()
	if users == nil {
		users = []store.User{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": users,
		"total": total,
	})
}

// 新增：用户操作
func (h *AdminHandler) ToggleUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.store.GetUserByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	if err := h.store.ToggleUser(id, user.IsActive == 0); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "操作失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "ok"})
}

func (h *AdminHandler) AdjustUserQuota(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Delta int `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if err := h.store.UpdateUserQuota(id, req.Delta); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "调整配额失败"})
		return
	}
	remaining, _ := h.store.GetUserQuota(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{"remaining_quota": remaining})
}

func (h *AdminHandler) GetUserHistory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.store.GetUserHistory(id, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取记录失败"})
		return
	}
	if items == nil {
		items = []store.History{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}
```

新增 import `"encoding/json"`。

### 步骤 3.5: 编译验证

```bash
cd /tmp/yiguan
GOPROXY=https://goproxy.cn,direct go build ./...
```

- [ ] **Commit**
```bash
git add internal/handler/
git commit -m "feat: 模型/广告/卦象管理 handler + admin 增强"
```

---

## Task 4: SSE 流式起卦 + LLM SSE + 模型路由器

**Files:**
- Create: `internal/llm/stream.go`
- Create: `internal/llm/router.go`
- Create: `internal/handler/divine_stream.go`
- Modify: `internal/handler/divine.go`（增加 Router 参数）
- Modify: `cmd/server/main.go`（注册 SSE 路由）

### 步骤 4.1: LLM SSE 流式调用

创建 `internal/llm/stream.go`：

```go
package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type streamChatReq struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// DivineStream SSE 流式调用 LLM，对每个 chunk 调用 onChunk
func (c *Client) DivineStream(prompt string, onChunk func(chunk string) error) error {
	body, _ := json.Marshal(streamChatReq{
		Model: c.cfg.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	})

	req, err := http.NewRequest("POST", c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("LLM请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM返回错误: %d %s", resp.StatusCode, string(respBody))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取SSE行失败: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				if err := onChunk(content); err != nil {
					return err
				}
			}
			if chunk.Choices[0].FinishReason == "stop" {
				break
			}
		}
	}
	return nil
}
```

### 步骤 4.2: 模型路由器

创建 `internal/llm/router.go`：

```go
package llm

import (
	"sync"

	"github.com/kiddyt00/yiguan/internal/store"
)

type Router struct {
	mu      sync.RWMutex
	st      store.Store
	current *Client
}

func NewRouter(st store.Store) (*Router, error) {
	r := &Router{st: st}
	if err := r.Reload(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Router) Get() *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.current
}

func (r *Router) Reload() error {
	m, err := r.st.GetDefaultModel()
	if err != nil {
		return fmt.Errorf("获取默认模型失败: %w", err)
	}
	client := New(Config{
		APIKey:   m.APIKey,
		Endpoint: m.Endpoint,
		Model:    m.Model,
	})
	r.mu.Lock()
	r.current = client
	r.mu.Unlock()
	return nil
}
```

需要 import `"fmt"`。

### 步骤 4.3: SSE 流式起卦 Handler

创建 `internal/handler/divine_stream.go`：

```go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

type DivineStreamHandler struct {
	store  store.Store
	router *llm.Router
}

func NewDivineStreamHandler(st store.Store, router *llm.Router) *DivineStreamHandler {
	return &DivineStreamHandler{store: st, router: router}
}

type sseEvent struct {
	Event string      `json:"-"`
	Data  interface{} `json:"data"`
}

func (h *DivineStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	// 检查配额
	remaining, err := h.store.GetRemainingQuota(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询配额失败"})
		return
	}
	if remaining <= 0 {
		writeJSON(w, http.StatusPaymentRequired, map[string]interface{}{
			"error":           "次数不足",
			"remaining_quota": 0,
		})
		return
	}

	var req divineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入问题"})
		return
	}

	// 扣减配额
	if err := h.store.ConsumeQuota(userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "扣减配额失败"})
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "不支持流式响应"})
		return
	}

	writeSSE := func(event string, data interface{}) {
		jsonData, _ := json.Marshal(map[string]interface{}{"event": event, "data": data})
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
		flusher.Flush()
	}

	// 阶段 1: 铜钱抛掷
	lines := engine.CastSixLines()
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	for i, line := range lines {
		writeSSE("phase", map[string]interface{}{
			"phase": "coins",
			"data":  map[string]interface{}{"throw": i + 1, "label": names[i], "result": lineType(line), "yang": line%2 != 0},
		})
		time.Sleep(200 * time.Millisecond) // 给前端动画时间
	}

	// 阶段 2: 构建卦象
	primary, changing, positions, master := engine.BuildHexagrams(lines)
	yaoPositions := buildYaoPositions(positions, master)
	yaoDesc := engine.FormatYaoPositions(positions, master)

	writeSSE("phase", map[string]interface{}{
		"phase": "hexagram",
		"data": map[string]interface{}{
			"primary_gua":   primary.Name,
			"changing_gua":  changing.Name,
			"yao_positions": yaoPositions,
		},
	})

	// 阶段 3: AI 解卦流式
	llmClient := h.router.Get()
	prompt := llm.BuildPrompt(req.Question, primary.Name, changing.Name, yaoDesc)

	var interpretation strings.Builder
	llmErr := llmClient.DivineStream(prompt, func(chunk string) error {
		interpretation.WriteString(chunk)
		writeSSE("ai", map[string]interface{}{"chunk": chunk})
		return nil
	})

	if llmErr != nil {
		writeSSE("error", map[string]interface{}{"error": "解卦服务暂不可用: " + llmErr.Error()})
	}

	// 保存历史记录
	history := &store.History{
		UserID:         userID,
		Question:       req.Question,
		PrimaryGua:     primary.Name,
		ChangingGua:    changing.Name,
		YaoPositions:   yaoDesc,
		Interpretation: interpretation.String(),
	}
	if err := h.store.SaveHistory(history); err != nil {
		writeSSE("error", map[string]interface{}{"error": "保存记录失败"})
	}

	// 完成
	remaining, _ = h.store.GetRemainingQuota(userID)
	writeSSE("done", map[string]interface{}{
		"id":              history.ID,
		"interpretation":  interpretation.String(),
		"remaining_quota": remaining,
		"created_at":      time.Now().Format(time.RFC3339),
	})
}

func lineType(line int) string {
	switch line {
	case 6:
		return "old_yin"
	case 7:
		return "young_yang"
	case 8:
		return "young_yin"
	case 9:
		return "old_yang"
	default:
		return "unknown"
	}
}
```

### 步骤 4.4: 路由注册

修改 `cmd/server/main.go`，在路由部分增加：

```go
import "github.com/kiddyt00/yiguan/internal/llm"

// ... 在 main 函数中 ...

// LLM Router（替代原来的固定 client）
llmRouter, err := llm.NewRouter(st)
if err != nil {
	log.Fatalf("LLM 路由器初始化失败: %v", err)
}

// 现有 divine handler 也改用 router
dh := handler.NewDivineHandler(st, llmRouter.Get())

// SSE 流式起卦
streamHandler := handler.NewDivineStreamHandler(st, llmRouter)
mux.Handle("POST /api/divine/stream", authMW(corsWrap(streamHandler)))

// 模型管理
mh := handler.NewModelHandler(st, llmRouter.Reload)
mux.Handle("GET /api/admin/models", adminMW(corsWrap(http.HandlerFunc(mh.ListModels))))
mux.Handle("POST /api/admin/models", adminMW(corsWrap(http.HandlerFunc(mh.CreateModel))))
mux.Handle("PUT /api/admin/models/{id}", adminMW(corsWrap(http.HandlerFunc(mh.UpdateModel))))
mux.Handle("DELETE /api/admin/models/{id}", adminMW(corsWrap(http.HandlerFunc(mh.DeleteModel))))
mux.Handle("POST /api/admin/models/{id}/set-default", adminMW(corsWrap(http.HandlerFunc(mh.SetDefaultModel))))
mux.Handle("POST /api/admin/models/{id}/toggle", adminMW(corsWrap(http.HandlerFunc(mh.ToggleModel))))

// 广告管理
ah := handler.NewAdHandler(st)
mux.Handle("GET /api/admin/ads", adminMW(corsWrap(http.HandlerFunc(ah.ListAds))))
mux.Handle("POST /api/admin/ads", adminMW(corsWrap(http.HandlerFunc(ah.CreateAd))))
mux.Handle("PUT /api/admin/ads/{id}", adminMW(corsWrap(http.HandlerFunc(ah.UpdateAd))))
mux.Handle("DELETE /api/admin/ads/{id}", adminMW(corsWrap(http.HandlerFunc(ah.DeleteAd))))
mux.Handle("POST /api/admin/ads/{id}/toggle", adminMW(corsWrap(http.HandlerFunc(ah.ToggleAd))))
mux.Handle("GET /api/admin/ads/stats", adminMW(corsWrap(http.HandlerFunc(ah.GetAdStats))))
mux.Handle("GET /api/ads/active", corsWrap(http.HandlerFunc(ah.ListActiveAds)))
mux.Handle("POST /api/ads/{id}/watch", authMW(corsWrap(http.HandlerFunc(ah.StartWatch))))
mux.Handle("POST /api/ads/{id}/complete", authMW(corsWrap(http.HandlerFunc(ah.CompleteWatch))))

// 卦象记录管理
hh2 := handler.NewHexagramHandler(st)
mux.Handle("GET /api/admin/hexagrams", adminMW(corsWrap(http.HandlerFunc(hh2.ListHistory))))
mux.Handle("GET /api/admin/hexagrams/{id}", adminMW(corsWrap(http.HandlerFunc(hh2.GetHistoryDetail))))
mux.Handle("DELETE /api/admin/hexagrams/{id}", adminMW(corsWrap(http.HandlerFunc(hh2.DeleteHistory))))

// 用户管理增强
mux.Handle("POST /api/admin/users/{id}/toggle", adminMW(corsWrap(http.HandlerFunc(adminHandler.ToggleUser))))
mux.Handle("POST /api/admin/users/{id}/quota", adminMW(corsWrap(http.HandlerFunc(adminHandler.AdjustUserQuota))))
mux.Handle("GET /api/admin/users/{id}/history", adminMW(corsWrap(http.HandlerFunc(adminHandler.GetUserHistory))))

// admin 中间件
adminMW := middleware.AdminOnly(cfg.JWTSecret)
```

需要 import `"fmt"`。

### 步骤 4.5: 编译验证

```bash
cd /tmp/yiguan
GOPROXY=https://goproxy.cn,direct go build ./...
```

- [ ] **Commit**
```bash
git add internal/llm/stream.go internal/llm/router.go internal/handler/divine_stream.go cmd/server/main.go internal/handler/divine.go
git commit -m "feat: SSE 流式起卦 + LLM SSE + 模型热切换路由器"
```

---

## Task 5: 配置文件 + Docker + Nginx

**Files:**
- Modify: `config.yaml`
- Modify: `.env.example`
- Modify: `deploy/nginx.conf`
- Modify: `docker-compose.yml`

### 步骤 5.1: config.yaml

修改 `config.yaml`：

```yaml
server:
  port: 8080
jwt_secret: "yiguan-dev-secret-change-in-production"
db_path: "yiguan.db"

admin:
  phone: "13800000000"
  password: "admin123"

llm:
  default: "qwen"
  providers:
    qwen:
      api_key: "sk-b8bd0f8077764c59b948c503cf1ee5f7"
      endpoint: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
      model: "qwen-plus"
    deepseek:
      api_key: ""
      endpoint: "https://api.deepseek.com/v1/chat/completions"
      model: "deepseek-chat"
```

### 步骤 5.2: .env.example

```env
LLM_API_KEY=sk-...
JWT_SECRET=your-secret-change-me
ADMIN_PHONE=13800000000
ADMIN_PASSWORD=admin123
HTTP_PORT=80
```

### 步骤 5.3: docker-compose.yml

增加环境变量：

```yaml
services:
  backend:
    environment:
      - DB_PATH=/data/yiguan.db
      - JWT_SECRET=${JWT_SECRET}
      - LLM_API_KEY=${LLM_API_KEY}
      - SERVER_PORT=8080
      - ADMIN_PHONE=${ADMIN_PHONE}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
```

### 步骤 5.4: deploy/nginx.conf

增加 SSE 代理配置：

```nginx
    # SSE 流式起卦（不缓冲）
    location /api/divine/stream {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 120s;
        add_header X-Accel-Buffering no;
    }
```

- [ ] **Commit**
```bash
git add config.yaml .env.example docker-compose.yml deploy/nginx.conf
git commit -m "chore: 配置管理员账号 + SSE nginx 代理 + 环境变量"
```

---

## Task 6: 后台前端重构 — 布局 + 登录 + 路由守卫

**Files:**
- Modify: `web/admin/package.json`
- Modify: `web/admin/src/main.js`
- Create: `web/admin/src/api/index.js`
- Create: `web/admin/src/stores/auth.js`
- Create: `web/admin/src/layout/AdminLayout.vue`
- Create: `web/admin/src/views/Login.vue`
- Modify: `web/admin/src/router/index.js`
- Modify: `web/admin/src/App.vue`

### 步骤 6.1: 安装 Element Plus

修改 `web/admin/package.json` dependencies 增加：
```json
{
  "dependencies": {
    "element-plus": "^2.9.0",
    "@element-plus/icons-vue": "^2.3.0"
  }
}
```

### 步骤 6.2: main.js

修改 `web/admin/src/main.js`：

```js
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'

const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus)
app.use(router)
app.mount('#app')
```

### 步骤 6.3: API 封装

创建 `web/admin/src/api/index.js`：

```js
function getAuthHeader() {
  const token = localStorage.getItem('admin_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function api(path, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...getAuthHeader() }
  const res = await fetch(`/api${path}`, { ...options, headers })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: '请求失败' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export const adminApi = {
  login: (data) => fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  }).then(r => r.json()),

  dashboard: () => api('/admin/dashboard'),

  // 用户
  users: (params) => api(`/admin/users?limit=${params.limit || 20}&offset=${params.offset || 0}`),
  toggleUser: (id) => api(`/admin/users/${id}/toggle`, { method: 'POST' }),
  adjustQuota: (id, delta) => api(`/admin/users/${id}/quota`, { method: 'POST', body: JSON.stringify({ delta }) }),
  userHistory: (id, limit = 20, offset = 0) => api(`/admin/users/${id}/history?limit=${limit}&offset=${offset}`),

  // 卦象
  hexagrams: (params) => api(`/admin/hexagrams?limit=${params.limit || 20}&offset=${params.offset || 0}${params.userId ? '&user_id=' + params.userId : ''}`),
  hexagramDetail: (id) => api(`/admin/hexagrams/${id}`),
  deleteHexagram: (id) => api(`/admin/hexagrams/${id}`, { method: 'DELETE' }),

  // 模型
  models: () => api('/admin/models'),
  createModel: (data) => api('/admin/models', { method: 'POST', body: JSON.stringify(data) }),
  updateModel: (id, data) => api(`/admin/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteModel: (id) => api(`/admin/models/${id}`, { method: 'DELETE' }),
  setDefaultModel: (id) => api(`/admin/models/${id}/set-default`, { method: 'POST' }),
  toggleModel: (id, enabled) => api(`/admin/models/${id}/toggle?enabled=${enabled}`, { method: 'POST' }),

  // 广告
  ads: () => api('/admin/ads'),
  createAd: (data) => api('/admin/ads', { method: 'POST', body: JSON.stringify(data) }),
  updateAd: (id, data) => api(`/admin/ads/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteAd: (id) => api(`/admin/ads/${id}`, { method: 'DELETE' }),
  toggleAd: (id, enabled) => api(`/admin/ads/${id}/toggle?enabled=${enabled}`, { method: 'POST' }),
  adStats: () => api('/admin/ads/stats'),
}
```

### 步骤 6.4: 认证 Store

创建 `web/admin/src/stores/auth.js`：

```js
import { ref } from 'vue'
import { adminApi } from '../api'

export const user = ref(null)
export const token = ref(localStorage.getItem('admin_token') || '')

export async function login(phone, password) {
  const resp = await adminApi.login({ phone, password })
  if (resp.error) throw new Error(resp.error)
  if (resp.user?.role !== 'admin') throw new Error('需要管理员权限')
  token.value = resp.token
  user.value = resp.user
  localStorage.setItem('admin_token', resp.token)
  return resp
}

export function logout() {
  token.value = ''
  user.value = null
  localStorage.removeItem('admin_token')
}
```

### 步骤 6.5: 管理布局

创建 `web/admin/src/layout/AdminLayout.vue`：

```vue
<template>
  <el-container class="min-h-screen">
    <el-aside width="200px" class="bg-slate-800 text-white">
      <div class="text-xl font-bold p-4 text-center">☯ 易观后台</div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#1e293b"
        text-color="#cbd5e1"
        active-text-color="#38bdf8"
        class="border-0"
      >
        <el-menu-item index="/"><el-icon><DataBoard /></el-icon>仪表盘</el-menu-item>
        <el-menu-item index="/users"><el-icon><User /></el-icon>用户管理</el-menu-item>
        <el-menu-item index="/hexagrams"><el-icon><List /></el-icon>卦象任务</el-menu-item>
        <el-menu-item index="/models"><el-icon><Cpu /></el-icon>模型管理</el-menu-item>
        <el-menu-item index="/ads"><el-icon><Notification /></el-icon>广告管理</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="bg-white shadow-sm flex items-center justify-between px-4">
        <span class="text-lg font-medium">{{ pageTitle }}</span>
        <el-dropdown @command="handleCommand">
          <span class="cursor-pointer">{{ user?.nickname || '管理员' }} <el-icon><ArrowDown /></el-icon></span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="bg-gray-50 p-6">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { user } from '../stores/auth'
import { logout } from '../stores/auth'

const route = useRoute()
const router = useRouter()

const titles = { '/': '仪表盘', '/users': '用户管理', '/hexagrams': '卦象任务', '/models': '模型管理', '/ads': '广告管理' }
const pageTitle = computed(() => titles[route.path] || '')

function handleCommand(cmd) {
  if (cmd === 'logout') {
    logout()
    router.push('/login')
  }
}
</script>
```

### 步骤 6.6: 登录页

创建 `web/admin/src/views/Login.vue`：

```vue
<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-100">
    <el-card class="w-96">
      <template #header>
        <div class="text-center text-xl font-bold">☯ 易观后台登录</div>
      </template>
      <el-form @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="phone" placeholder="手机号" prefix-icon="Phone" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="w-full" :loading="loading" @click="handleLogin">登录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../stores/auth'
import { ElMessage } from 'element-plus'

const phone = ref('')
const password = ref('')
const loading = ref(false)
const router = useRouter()

async function handleLogin() {
  loading.value = true
  try {
    await login(phone.value, password.value)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    loading.value = false
  }
}
</script>
```

### 步骤 6.7: 路由

修改 `web/admin/src/router/index.js`：

```js
import { createRouter, createWebHistory } from 'vue-router'
import { token } from '../stores/auth'
import AdminLayout from '../layout/AdminLayout.vue'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import Users from '../views/Users.vue'

export default createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/login', component: Login },
    {
      path: '/',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', component: Dashboard },
        { path: 'users', component: Users },
        { path: 'hexagrams', component: () => import('../views/Hexagrams.vue') },
        { path: 'models', component: () => import('../views/Models.vue') },
        { path: 'ads', component: () => import('../views/Ads.vue') },
      ],
    },
  ],
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !token.value) {
    next('/login')
  } else if (to.path === '/login' && token.value) {
    next('/')
  } else {
    next()
  }
})
```

### 步骤 6.8: App.vue

修改 `web/admin/src/App.vue`：

```vue
<template>
  <router-view />
</template>
```

### 步骤 6.9: 前端本地安装 + 构建验证

```bash
cd /tmp/yiguan/web/admin
npm install
npx vite build  # 验证无报错
```

- [ ] **Commit**
```bash
git add web/admin/
git commit -m "feat: 后台前端重构 — Element Plus + 管理布局 + 登录 + 路由守卫"
```

---

## Task 7: 后台前端页面 — Dashboard/Users 增强 + 卦象/模型/广告页

**Files:**
- Modify: `web/admin/src/views/Dashboard.vue`
- Modify: `web/admin/src/views/Users.vue`
- Create: `web/admin/src/views/Hexagrams.vue`
- Create: `web/admin/src/views/Models.vue`
- Create: `web/admin/src/views/Ads.vue`

### 步骤 7.1: Dashboard 增强

修改 `web/admin/src/views/Dashboard.vue`：

```vue
<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">仪表盘</h2>
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <el-card>
        <div class="text-sm text-gray-500">注册用户</div>
        <div class="text-3xl font-bold mt-2">{{ stats.total_users }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">活跃用户</div>
        <div class="text-3xl font-bold mt-2 text-green-600">{{ stats.active_users }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">今日算卦</div>
        <div class="text-3xl font-bold mt-2">{{ stats.today_divines }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">总起卦数</div>
        <div class="text-3xl font-bold mt-2">{{ stats.total_divines }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">今日广告播放</div>
        <div class="text-3xl font-bold mt-2 text-blue-600">{{ stats.ad_watches_today }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">总广告播放</div>
        <div class="text-3xl font-bold mt-2">{{ stats.total_ads_watched }}</div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'

const stats = ref({})
onMounted(async () => {
  try {
    stats.value = await adminApi.dashboard()
  } catch (e) {
    console.error(e)
  }
})
</script>
```

### 步骤 7.2: Users 增强

修改 `web/admin/src/views/Users.vue`：

```vue
<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">用户管理</h2>
    <el-table :data="users" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column prop="nickname" label="昵称" width="120" />
      <el-table-column prop="role" label="角色" width="80">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">{{ row.role }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'danger'">{{ row.is_active ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="toggleUser(row)">
            {{ row.is_active ? '禁用' : '启用' }}
          </el-button>
          <el-button size="small" type="primary" @click="adjustQuota(row)">配额</el-button>
          <el-button size="small" type="info" @click="viewHistory(row)">记录</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="quotaVisible" title="调整配额" width="400px">
      <el-input-number v-model="quotaDelta" :min="-100" :max="100" />
      <template #footer>
        <el-button @click="quotaVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmQuota">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="historyVisible" :title="'用户起卦记录'" width="800px">
      <el-table :data="userHistory" size="small">
        <el-table-column prop="id" label="ID" width="50" />
        <el-table-column prop="question" label="问题" />
        <el-table-column prop="primary_gua" label="本卦" width="80" />
        <el-table-column prop="changing_gua" label="变卦" width="80" />
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'

const users = ref([])
const quotaVisible = ref(false)
const quotaDelta = ref(0)
const currentUser = ref(null)
const historyVisible = ref(false)
const userHistory = ref([])

onMounted(async () => {
  try {
    const data = await adminApi.users({ limit: 100 })
    users.value = data.items
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  }
})

async function toggleUser(row) {
  try {
    await adminApi.toggleUser(row.id)
    ElMessage.success('操作成功')
    row.is_active = row.is_active ? 0 : 1
  } catch (e) {
    ElMessage.error(e.message)
  }
}

async function adjustQuota(row) {
  currentUser.value = row
  quotaDelta.value = 0
  quotaVisible.value = true
}

async function confirmQuota() {
  try {
    await adminApi.adjustQuota(currentUser.value.id, quotaDelta.value)
    ElMessage.success('配额调整成功')
    quotaVisible.value = false
  } catch (e) {
    ElMessage.error(e.message)
  }
}

async function viewHistory(row) {
  currentUser.value = row
  try {
    const data = await adminApi.userHistory(row.id)
    userHistory.value = data.items
    historyVisible.value = true
  } catch (e) {
    ElMessage.error(e.message)
  }
}

function formatDate(d) { return new Date(d).toLocaleString('zh-CN') }
</script>
```

### 步骤 7.3: 卦象任务页

创建 `web/admin/src/views/Hexagrams.vue`：

```vue
<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">卦象任务管理</h2>
    <div class="mb-4 flex gap-2">
      <el-input v-model="userIdFilter" placeholder="用户ID筛选" style="width:150px" />
      <el-button @click="load">筛选</el-button>
    </div>
    <el-table :data="items" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="user_id" label="用户ID" width="80" />
      <el-table-column prop="question" label="问题" min-width="200" show-overflow-tooltip />
      <el-table-column prop="primary_gua" label="本卦" width="80" />
      <el-table-column prop="changing_gua" label="变卦" width="80" />
      <el-table-column prop="created_at" label="时间" width="170">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" @click="showDetail(row)">详情</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="detailVisible" title="卦象详情" width="700px">
      <div v-if="detail" class="space-y-3">
        <p><strong>问题：</strong>{{ detail.question }}</p>
        <p><strong>本卦：</strong>{{ detail.primary_gua }} <strong>变卦：</strong>{{ detail.changing_gua }}</p>
        <p><strong>变爻：</strong>{{ detail.yao_positions }}</p>
        <p><strong>AI解卦：</strong></p>
        <div class="bg-gray-50 p-3 rounded whitespace-pre-wrap">{{ detail.interpretation }}</div>
        <p class="text-sm text-gray-400">{{ formatDate(detail.created_at) }}</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const items = ref([])
const detail = ref(null)
const detailVisible = ref(false)
const userIdFilter = ref('')

onMounted(() => load())

async function load() {
  try {
    const params = { limit: 50 }
    if (userIdFilter.value) params.userId = userIdFilter.value
    const data = await adminApi.hexagrams(params)
    items.value = data.items
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  }
}

function showDetail(row) {
  detail.value = row
  detailVisible.value = true
}

async function remove(row) {
  try {
    await ElMessageBox.confirm('确定删除此记录？', '确认')
    await adminApi.deleteHexagram(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message)
  }
}

function formatDate(d) { return new Date(d).toLocaleString('zh-CN') }
</script>
```

### 步骤 7.4: 模型管理页

创建 `web/admin/src/views/Models.vue`：

```vue
<template>
  <div>
    <h2 class="text-2xl font-bold mb-4 flex justify-between items-center">
      模型管理
      <el-button type="primary" @click="showCreate = true">新增模型</el-button>
    </h2>
    <el-table :data="models" stripe>
      <el-table-column prop="id" label="ID" width="50" />
      <el-table-column prop="name" label="模型名" width="120" />
      <el-table-column prop="provider" label="提供商" width="100" />
      <el-table-column prop="endpoint" label="Endpoint" min-width="200" show-overflow-tooltip />
      <el-table-column prop="is_default" label="默认" width="60">
        <template #default="{ row }"><el-tag v-if="row.is_default" type="warning">默认</el-tag></template>
      </el-table-column>
      <el-table-column prop="is_enabled" label="状态" width="60">
        <template #default="{ row }">
          <el-switch v-model="row.is_enabled" :active-value="1" :inactive-value="0"
            @change="toggleModel(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" v-if="!row.is_default" @click="setDefault(row)">设为默认</el-button>
          <el-button size="small" type="primary" @click="editModel(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" :title="editTarget ? '编辑模型' : '新增模型'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="模型名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="提供商"><el-input v-model="form.provider" /></el-form-item>
        <el-form-item label="Endpoint"><el-input v-model="form.endpoint" /></el-form-item>
        <el-form-item label="API Key"><el-input v-model="form.api_key" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const models = ref([])
const showCreate = ref(false)
const editTarget = ref(null)
const form = ref({ name: '', provider: '', endpoint: '', api_key: '' })

onMounted(load)

async function load() {
  try {
    const data = await adminApi.models()
    models.value = data.items
  } catch (e) {
    ElMessage.error('加载失败')
  }
}

async function toggleModel(row) {
  try {
    await adminApi.toggleModel(row.id, row.is_enabled === 1)
    ElMessage.success('操作成功')
  } catch (e) {
    ElMessage.error(e.message)
    row.is_enabled = row.is_enabled ? 0 : 1
  }
}

async function setDefault(row) {
  try {
    await adminApi.setDefaultModel(row.id)
    ElMessage.success('已设为默认')
    load()
  } catch (e) {
    ElMessage.error(e.message)
  }
}

function editModel(row) {
  editTarget.value = row
  form.value = { name: row.name, provider: row.provider, endpoint: row.endpoint, api_key: row.api_key }
  showCreate.value = true
}

async function save() {
  try {
    if (editTarget.value) {
      await adminApi.updateModel(editTarget.value.id, form.value)
    } else {
      await adminApi.createModel(form.value)
    }
    ElMessage.success('保存成功')
    showCreate.value = false
    editTarget.value = null
    load()
  } catch (e) {
    ElMessage.error(e.message)
  }
}

async function remove(row) {
  try {
    await ElMessageBox.confirm('确定删除？', '确认')
    await adminApi.deleteModel(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message)
  }
}
</script>
```

### 步骤 7.5: 广告管理页

创建 `web/admin/src/views/Ads.vue`：

```vue
<template>
  <div>
    <h2 class="text-2xl font-bold mb-4 flex justify-between items-center">
      广告管理
      <el-button type="primary" @click="showCreate = true">新增广告</el-button>
    </h2>
    <el-table :data="ads" stripe>
      <el-table-column prop="id" label="ID" width="50" />
      <el-table-column prop="name" label="名称" width="120" />
      <el-table-column prop="content_url" label="URL" min-width="200" show-overflow-tooltip />
      <el-table-column prop="watch_duration" label="时长(秒)" width="90" />
      <el-table-column prop="reward_quota" label="奖励次数" width="90" />
      <el-table-column prop="is_enabled" label="状态" width="60">
        <template #default="{ row }">
          <el-switch v-model="row.is_enabled" :active-value="1" :inactive-value="0"
            @change="toggleAd(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="editAd(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-divider>广告播放统计</el-divider>
    <el-table :data="stats" stripe>
      <el-table-column prop="ad_name" label="广告" width="120" />
      <el-table-column prop="total" label="总播放" width="80" />
      <el-table-column prop="completed" label="完成" width="80" />
      <el-table-column prop="reward_total" label="发放奖励" width="80" />
    </el-table>

    <el-dialog v-model="showCreate" :title="editTarget ? '编辑广告' : '新增广告'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" /></el-form-item>
        <el-form-item label="广告URL"><el-input v-model="form.content_url" /></el-form-item>
        <el-form-item label="观看时长"><el-input-number v-model="form.watch_duration" :min="5" :max="300" /></el-form-item>
        <el-form-item label="奖励次数"><el-input-number v-model="form.reward_quota" :min="1" :max="10" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const ads = ref([])
const stats = ref([])
const showCreate = ref(false)
const editTarget = ref(null)
const form = ref({ name: '', description: '', content_url: '', watch_duration: 30, reward_quota: 1 })

onMounted(load)

async function load() {
  try {
    const data = await adminApi.ads()
    ads.value = data.items
  } catch (e) { ElMessage.error('加载失败') }
  try {
    const data = await adminApi.adStats()
    stats.value = data.items
  } catch (e) {}
}

async function toggleAd(row) {
  try {
    await adminApi.toggleAd(row.id, row.is_enabled === 1)
    ElMessage.success('操作成功')
  } catch (e) {
    row.is_enabled = row.is_enabled ? 0 : 1
  }
}

function editAd(row) {
  editTarget.value = row
  form.value = { name: row.name, description: row.description, content_url: row.content_url, watch_duration: row.watch_duration, reward_quota: row.reward_quota }
  showCreate.value = true
}

async function save() {
  try {
    if (editTarget.value) {
      await adminApi.updateAd(editTarget.value.id, form.value)
    } else {
      await adminApi.createAd(form.value)
    }
    ElMessage.success('保存成功')
    showCreate.value = false
    editTarget.value = null
    load()
  } catch (e) {
    ElMessage.error(e.message)
  }
}

async function remove(row) {
  try {
    await ElMessageBox.confirm('确定删除？', '确认')
    await adminApi.deleteAd(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message)
  }
}
</script>
```

- [ ] **Commit**
```bash
git add web/admin/src/views/ web/admin/src/layout/ web/admin/src/stores/ web/admin/src/api/
git commit -m "feat: 后台前端页面 — Dashboard/Users/Hexagrams/Models/Ads 完整实现"
```

---

## Task 8: 前台前端 — 流式起卦 + 广告中心

**Files:**
- Create: `web/front/src/views/StreamDivine.vue`
- Create: `web/front/src/views/AdCenter.vue`
- Modify: `web/front/src/views/Home.vue`

### 步骤 8.1: 流式起卦页

创建 `web/front/src/views/StreamDivine.vue`：

```vue
<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6">
    <h3 class="text-xl font-bold mb-4 text-center">🔮 起卦中...</h3>

    <!-- 铜钱动画 -->
    <div v-if="phase === 'coins'" class="text-center py-8">
      <div class="text-4xl mb-4" :class="isDark ? 'text-cyan-400' : 'text-amber-600'">
        🪙 {{ coinsLabel }}
      </div>
      <div class="animate-bounce text-2xl">🎲</div>
    </div>

    <!-- 卦象展示 -->
    <div v-if="phase === 'hexagram' || phase === 'ai' || phase === 'done'" class="grid grid-cols-2 gap-6 mb-6">
      <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-amber-50'">
        <span class="text-sm opacity-60">本卦</span>
        <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ hexagram.primary_gua }}</div>
      </div>
      <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-stone-50'">
        <span class="text-sm opacity-60">变卦</span>
        <div class="text-2xl font-bold mt-1">{{ hexagram.changing_gua }}</div>
      </div>
    </div>

    <!-- AI 解卦流式渲染 -->
    <div v-if="phase === 'ai' || phase === 'done'" class="border-t pt-6 mt-6" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <h4 class="text-lg font-medium mb-3">🤖 AI 解卦</h4>
      <div class="leading-relaxed whitespace-pre-wrap opacity-80">
        {{ aiText }}
        <span v-if="phase === 'ai'" class="animate-pulse">▊</span>
      </div>
    </div>

    <!-- 错误 -->
    <div v-if="error" class="text-center text-red-500 py-4">{{ error }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps(['question', 'isDark', 'token'])
const emit = defineEmits(['complete', 'error'])

const phase = ref('coins')
const coinsLabel = ref('')
const hexagram = ref({ primary_gua: '', changing_gua: '' })
const aiText = ref('')
const error = ref('')

async function startStream() {
  try {
    const resp = await fetch('/api/divine/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${props.token}`,
      },
      body: JSON.stringify({ question: props.question }),
    })

    if (!resp.ok) {
      const err = await resp.json()
      error.value = err.error || '起卦失败'
      emit('error', err.error)
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event:')) {
          const eventType = line.replace('event:', '').trim()
        } else if (line.startsWith('data:')) {
          const dataStr = line.replace('data:', '').trim()
          if (!dataStr) continue

          try {
            const parsed = JSON.parse(dataStr)
            const evt = parsed.event
            const data = parsed.data

            if (evt === 'phase') {
              if (data.phase === 'coins') {
                phase.value = 'coins'
                coinsLabel.value = `${data.label} — ${data.result}`
              } else if (data.phase === 'hexagram') {
                phase.value = 'hexagram'
                hexagram.value = { primary_gua: data.primary_gua, changing_gua: data.changing_gua }
              }
            } else if (evt === 'ai') {
              if (phase.value === 'hexagram') phase.value = 'ai'
              aiText.value += data.chunk
            } else if (evt === 'done') {
              phase.value = 'done'
              emit('complete', data)
            } else if (evt === 'error') {
              error.value = data.error
              phase.value = 'done'
              emit('error', data.error)
            }
          } catch (e) {
            console.error('Parse SSE error:', e)
          }
        }
      }
    }
  } catch (e) {
    error.value = '网络连接失败'
    emit('error', e.message)
  }
}

startStream()
</script>
```

### 步骤 8.2: 广告中心

创建 `web/front/src/views/AdCenter.vue`：

```vue
<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6">
    <h3 class="text-xl font-bold mb-4 text-center">📢 看广告领次数</h3>

    <div v-if="ads.length === 0" class="text-center text-gray-400 py-8">暂无可用广告</div>

    <div v-for="ad in ads" :key="ad.id" class="mb-4 p-4 rounded-lg border" :class="isDark ? 'border-slate-600 bg-slate-700' : 'border-stone-200'">
      <div class="flex justify-between items-center">
        <div>
          <h4 class="font-bold">{{ ad.name }}</h4>
          <p class="text-sm opacity-60">{{ ad.description }}</p>
          <p class="text-sm mt-1">观看 {{ ad.watch_duration }} 秒，奖励 <span class="text-amber-600 font-bold">{{ ad.reward_quota }}</span> 次起卦</p>
        </div>
        <button @click="watchAd(ad)" class="px-4 py-2 rounded-lg font-medium" :class="isDark ? 'bg-cyan-600 hover:bg-cyan-500' : 'bg-amber-600 hover:bg-amber-500'">
          观看
        </button>
      </div>
    </div>

    <!-- 广告弹窗 -->
    <div v-if="watchingAd" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-2xl">
        <div class="flex justify-between items-center mb-4">
          <h4 class="font-bold">{{ watchingAd.name }}</h4>
          <button @click="closeAd" class="text-gray-400 text-2xl">&times;</button>
        </div>
        <iframe :src="watchingAd.content_url" class="w-full h-64 border rounded" />
        <div class="mt-4 text-center">
          <p class="text-lg mb-2">
            <template v-if="countdown > 0">还需观看 {{ countdown }} 秒</template>
            <template v-else>✅ 观看完成！</template>
          </p>
          <button v-if="countdown <= 0" @click="claimReward"
            class="px-8 py-3 rounded-lg font-medium bg-green-600 text-white hover:bg-green-500">
            领取次数
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps(['token', 'isDark'])
const emit = defineEmits(['rewarded'])

const ads = ref([])
const watchingAd = ref(null)
const countdown = ref(0)
let timer = null

onMounted(async () => {
  try {
    const res = await fetch('/api/ads/active')
    const data = await res.json()
    ads.value = data.items || []
  } catch (e) {}
})

async function watchAd(ad) {
  try {
    await fetch(`/api/ads/${ad.id}/watch`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${props.token}` },
    })
    watchingAd.value = ad
    countdown.value = ad.watch_duration
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) {}
}

async function claimReward() {
  try {
    const res = await fetch(`/api/ads/${watchingAd.value.id}/complete`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${props.token}`,
      },
      body: JSON.stringify({ duration: watchingAd.value.watch_duration }),
    })
    const data = await res.json()
    if (data.rewarded) {
      emit('rewarded', data)
    }
  } catch (e) {}
  closeAd()
}

function closeAd() {
  watchingAd.value = null
  if (timer) clearInterval(timer)
  countdown.value = 0
}
</script>
```

### 步骤 8.3: Home.vue 增加入口

修改 `web/front/src/views/Home.vue`，在适当位置添加"看广告领次数"按钮或入口。

- [ ] **Commit**
```bash
git add web/front/src/views/StreamDivine.vue web/front/src/views/AdCenter.vue web/front/src/views/Home.vue
git commit -m "feat: 前台流式起卦 + 广告中心"
```

---

## Task 9: 最终集成 + 构建 + 验证

### 步骤 9.1: 复制项目到 /tmp 并构建

```bash
cp -r /mnt/d/projects/yiguan /tmp/yiguan-v2
cd /tmp/yiguan-v2

# 安装 admin 依赖
cd web/admin && npm install && cd ../..

# 编译前端
cd web/admin && npm run build && cd ../..
cd web/front && npm run build && cd ../..

# 编译后端
GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o deploy/bin/yiguan ./cmd/server/
```

### 步骤 9.2: 本地启动验证

```bash
./yiguan  # 启动后端
cd web/front && npx vite --host 0.0.0.0 --port 3000  # 前台
cd web/admin && npx vite --host 0.0.0.0 --port 3001  # 后台
```

### 步骤 9.3: 验证清单

```bash
# 1. 管理员初始化
curl -s http://localhost:8080/api/admin/dashboard -H "Authorization: Bearer $(curl -s -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' -d '{"phone":"13800000000","password":"admin123"}' | jq -r .token)"

# 2. 非管理员访问后台（应返回 403）

# 3. SSE 流式起卦
curl -N -X POST http://localhost:8080/api/divine/stream \
  -H "Authorization: Bearer <user_token>" \
  -H "Content-Type: application/json" \
  -d '{"question":"今天运势如何"}'

# 4. 广告 CRUD
# 5. 模型 CRUD
```

- [ ] **Final Commit**
```bash
cd /mnt/d/projects/yiguan
git add -A
git commit -m "feat: v2.0 全栈增强完成 — 角色权限/流式起卦/模型广告管理/广告激励系统"
git push
```
