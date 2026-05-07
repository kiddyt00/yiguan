package store

import (
	"errors"
	"time"
)

// ErrNotFound 资源不存在
var ErrNotFound = errors.New("record not found")

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Address   string    `json:"address,omitempty"`
	Password  string    `json:"-"` // bcrypt hash
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

// Store 数据库抽象接口
type Store interface {
	// 用户
	CreateUser(phone, password, nickname string) (*User, error)
	GetUserByPhone(phone string) (*User, error)
	GetUserByID(id int64) (*User, error)
	UpdateUser(id int64, nickname, address string) error

	// Quota
	GetRemainingQuota(userID int64) (int, error)
	AddQuota(userID int64, quotaType string) error
	ConsumeQuota(userID int64) error // 扣减 1 次

	// 历史记录
	SaveHistory(h *History) error
	GetHistory(userID int64, limit, offset int) ([]History, error)
	GetHistoryCount(userID int64) (int64, error)

	// 管理
	GetTotalUsers() (int64, error)
	ListUsers(limit, offset int) ([]User, error)
	GetTodayDivineCount() (int64, error)

	Close() error
}
