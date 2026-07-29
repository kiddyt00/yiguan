package store

import (
	"errors"
	"time"
)

// ErrNotFound 资源不存在
var ErrNotFound = errors.New("record not found")

// ErrQuotaExhausted quota 次数已用完
var ErrQuotaExhausted = errors.New("quota exhausted")

// ========== 数据模型 ==========

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	OpenID    string    `json:"openid,omitempty"`
	UnionID   string    `json:"unionid,omitempty"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	WxAvatar  string    `json:"-"`       // 微信头像 URL，不直接暴露，后端计算回退
	Address   string    `json:"address,omitempty"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  int       `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Quota 次数配额
type Quota struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	QuotaType string     `json:"quota_type"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// History 算卦历史记录
type History struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Nickname       string    `json:"nickname,omitempty"`
	Question       string    `json:"question"`
	PrimaryGua     string    `json:"primary_gua"`
	ChangingGua    string    `json:"changing_gua"`
	YaoPositions   string    `json:"yao_positions"`
	PrimaryYao     string    `json:"primary_yao,omitempty"`
	ChangingYao    string    `json:"changing_yao,omitempty"`
	TossData       string    `json:"toss_data,omitempty"`
	MasterYao      int       `json:"master_yao,omitempty"`
	Interpretation string    `json:"interpretation"`
	Lang           string    `json:"lang"`
	CreatedAt      time.Time `json:"created_at"`
}

// Translation AI 解读翻译缓存
type Translation struct {
	ID        int64     `json:"id"`
	HistoryID int64     `json:"history_id"`
	Lang      string    `json:"lang"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// LLMModel LLM 模型配置
type LLMModel struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Provider    string    `json:"provider"`
	Endpoint    string    `json:"endpoint"`
	APIKey      string    `json:"api_key"`
	IsDefault   int       `json:"is_default"`
	IsEnabled   int       `json:"is_enabled"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
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

// Membership 会员记录
type Membership struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	OrderID   int64     `json:"order_id"`
	ProductID string    `json:"product_id"` // monthly|quarterly|yearly
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // active|terminated|refunded
	CreatedAt time.Time `json:"created_at"`
}

// MembershipStatus 当前会员状态（API 响应用）
type MembershipStatus struct {
	IsActive  bool      `json:"is_active"`
	ProductID string    `json:"product_id,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	DaysLeft  int       `json:"days_left"`
}

// Invitation 邀请关系
type Invitation struct {
	ID         int64     `json:"id"`
	InviterID  int64     `json:"inviter_id"`
	InviteeID  int64     `json:"invitee_id"`
	InviteCode string    `json:"invite_code"`
	Status     string    `json:"status"` // registered|divined|rewarded
	CreatedAt  time.Time `json:"created_at"`
}

// InviteProgress 拉新进度
type InviteProgress struct {
	RegisteredCount int `json:"registered_count"`
	DivinedCount    int `json:"divined_count"`
	RewardRound     int `json:"reward_round"`
	PendingReward   bool `json:"pending_reward"`
}

// MembershipLog 会员权益变动流水
type MembershipLog struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	MembershipID int64     `json:"membership_id,omitempty"`
	Action       string    `json:"action"`  // created|terminated|refunded|expired
	Detail       string    `json:"detail"`  // 关联业务说明
	CreatedAt    time.Time `json:"created_at"`
}

// AdStat 广告统计
type AdStat struct {
	AdID        int64  `json:"ad_id"`
	AdName      string `json:"ad_name"`
	Total       int64  `json:"total"`
	Completed   int64  `json:"completed"`
	RewardTotal int64  `json:"reward_total"`
}

// ========== 子接口 ==========

// UserStore 用户与配额操作
type UserStore interface {
	CreateUser(phone, password, nickname string) (*User, error)
	CreateUserByOpenID(openid, nickname, wxAvatar string) (*User, error)
	UpdateUserWechatInfo(id int64, nickname, wxAvatar string) error
	UpdateUserGender(id int64, sex int) error
	GetUserByPhone(phone string) (*User, error)
	GetUserByOpenID(openid string) (*User, error)
	GetUserByUnionID(unionid string) (*User, error)
	GetUserByID(id int64) (*User, error)
	UpdateUser(id int64, nickname, address string) error
	UpdateUserOpenID(id int64, openid string) error
	UpdateUserUnionID(id int64, unionid string) error

	ToggleUser(id int64, active bool) error
	UpdateUserRole(id int64, role string) error
	UpdateUserQuota(userID int64, delta int) error
	GetUserQuota(userID int64) (int, error)

	GetRemainingQuota(userID int64) (int, error)
	AddQuota(userID int64, quotaType string) error
	ConsumeQuota(userID int64) error

	GetTotalUsers() (int64, error)
	ListUsers(limit, offset int) ([]User, error)
	GetTodayDivineCount() (int64, error)
	GetActiveUserCount() (int64, error)
	GetTotalDivineCount() (int64, error)
	SearchUsers(keyword string, limit, offset int) ([]User, error)
	SearchUsersCount(keyword string) (int64, error)
}

// HistoryStore 历史记录与卦象管理
type HistoryStore interface {
	SaveHistory(h *History) error
	GetHistory(userID int64, limit, offset int) ([]History, error)
	GetHistoryCount(userID int64) (int64, error)

	ListAllHistory(limit, offset int, userID int64, keyword, dateFrom, dateTo string) ([]History, error)
	CountAllHistory(userID int64, keyword, dateFrom, dateTo string) (int64, error)
	GetHistoryByID(id int64) (*History, error)
	DeleteHistory(id int64) error
	GetUserHistory(userID int64, limit, offset int) ([]History, error)

	// SearchHistory 关键词搜索历史记录（支持 question/primary_gua/changing_gua）
	SearchHistory(userID int64, keyword string, limit, offset int) ([]History, error)
	SearchHistoryCount(userID int64, keyword string) (int64, error)
	// GetLatestHistory 获取用户最新一条历史记录
	GetLatestHistory(userID int64) (*History, error)
	// GetDailyDivineTrend 近7天每日起卦趋势
	GetDailyDivineTrend() (map[string]int64, error)
}

// TranslationStore 翻译缓存
type TranslationStore interface {
	GetTranslation(historyID int64, lang string) (*Translation, error)
	SaveTranslation(t *Translation) error
}

// ModelStore LLM 模型管理
type ModelStore interface {
	ListModels() ([]LLMModel, error)
	GetModelByID(id int64) (*LLMModel, error)
	GetDefaultModel() (*LLMModel, error)
	CreateModel(m *LLMModel) error
	UpdateModel(m *LLMModel) error
	DeleteModel(id int64) error
	SetDefaultModel(id int64) error
	ToggleModel(id int64, enabled bool) error
}

// AdStore 广告管理
type AdStore interface {
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
	GetTodayAdWatchCount() (int64, error)
	GetTodayAdWatchCountByUser(userID int64) (int64, error)
	GetTotalAdWatchCount() (int64, error)
}

// InviteStore 拉新裂变
type InviteStore interface {
	// GenerateInviteCode 生成/获取用户邀请码
	GenerateInviteCode(userID int64) (string, error)
	// BindInvitation 注册时绑定邀请关系（通过邀请码）
	BindInvitation(inviteeID int64, inviteCode string) error
	// MarkInviteeDivined 被邀请人完成测算后标记
	MarkInviteeDivined(userID int64) error
	// RewardInviterIfEligible 被邀请人测算后,检查邀请人奖励条件
	RewardInviterIfEligible(inviteeID int64) error
	// GetInviteProgress 获取邀请人拉新进度
	GetInviteProgress(userID int64) (*InviteProgress, error)
	// CheckAndReward 检查并发放奖励（返回是否发放了奖励）
	CheckAndReward(userID int64) (bool, error)
}

// MembershipStore 会员管理
type MembershipStore interface {
	// CreateMembership 创建会员记录（自动处理顺延）
	CreateMembership(userID, orderID int64, productID string, endTime time.Time) (*Membership, error)
	// GetActiveMembership 获取用户当前有效会员（取最晚到期的）
	GetActiveMembership(userID int64) (*Membership, error)
	// HasActiveMembership 检查用户是否有有效会员
	HasActiveMembership(userID int64) (bool, error)
	// TerminateMembership 终止会员（退款时用）
	TerminateMembership(membershipID int64) error
	// ListMemberships 列出用户会员记录
	ListMemberships(userID int64) ([]Membership, error)
	// ListAllMemberships 管理员：列出全部会员记录
	ListAllMemberships(limit, offset int) ([]Membership, error)
	// CountActiveMemberships 统计当前有效会员数
	CountActiveMemberships() (int, error)
	// LogMembership 记录权益变动
	LogMembership(userID int64, membershipID int64, action, detail string) error
}

// Store 组合接口（向后兼容）
type Store interface {
	UserStore
	HistoryStore
	TranslationStore
	ModelStore
	AdStore
	AnalyticsStore
	OrderStore
	MembershipStore
	InviteStore
	Close() error
}
