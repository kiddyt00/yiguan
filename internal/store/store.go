package store

import (
	"errors"
	"time"
)

// ErrNotFound 资源不存在
var ErrNotFound = errors.New("record not found")

// ErrQuotaExhausted quota 次数已用完
var ErrQuotaExhausted = errors.New("quota exhausted")

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Address   string    `json:"address,omitempty"`
	Password  string    `json:"-"` // bcrypt hash
	Role      string    `json:"role"`
	IsActive  int       `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Quota 次数配额
type Quota struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	QuotaType string     `json:"quota_type"` // "free", "paid", "share", "ad"
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// History 算卦历史记录
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

// LLMModel LLM 模型配置
type LLMModel struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	Endpoint  string    `json:"endpoint"`
	APIKey    string    `json:"api_key"`
	IsDefault int       `json:"is_default"`
	IsEnabled int       `json:"is_enabled"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// Ad 广告配置
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

// AdRecord 广告观看记录
type AdRecord struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	AdID          int64     `json:"ad_id"`
	WatchDuration int       `json:"watch_duration"`
	Status        string    `json:"status"`
	Rewarded      int       `json:"rewarded"`
	CreatedAt     time.Time `json:"created_at"`
}

// AdStat 广告统计
type AdStat struct {
	AdID        int64  `json:"ad_id"`
	AdName      string `json:"ad_name"`
	Total       int64  `json:"total"`
	Completed   int64  `json:"completed"`
	RewardTotal int64  `json:"reward_total"`
}

// Store 数据库抽象接口
type Store interface {
	// 用户
	CreateUser(phone, password, nickname string) (*User, error)
	GetUserByPhone(phone string) (*User, error)
	GetUserByID(id int64) (*User, error)
	UpdateUser(id int64, nickname, address string) error

	// 用户管理
	ToggleUser(id int64, active bool) error
	UpdateUserRole(id int64, role string) error
	UpdateUserQuota(userID int64, delta int) error
	GetUserQuota(userID int64) (int, error)

	// Quota
	GetRemainingQuota(userID int64) (int, error)
	AddQuota(userID int64, quotaType string) error
	ConsumeQuota(userID int64) error

	// 历史记录
	SaveHistory(h *History) error
	GetHistory(userID int64, limit, offset int) ([]History, error)
	GetHistoryCount(userID int64) (int64, error)

	// 卦象管理（管理员全量查询）
	ListAllHistory(limit, offset int, userID int64) ([]History, error)
	GetHistoryByID(id int64) (*History, error)
	DeleteHistory(id int64) error
	GetUserHistory(userID int64, limit, offset int) ([]History, error)

	// 管理统计
	GetTotalUsers() (int64, error)
	ListUsers(limit, offset int) ([]User, error)
	GetTodayDivineCount() (int64, error)
	GetActiveUserCount() (int64, error)
	GetTotalDivineCount() (int64, error)
	GetTodayAdWatchCount() (int64, error)
	GetTotalAdWatchCount() (int64, error)

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
