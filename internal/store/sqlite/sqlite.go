package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store SQLite 实现的存储
type Store struct {
	db *sql.DB
}

// New 创建 SQLite 存储并执行建表
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite 单写

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return &Store{db: db}, nil
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	return s.db.Close()
}

// DB 返回底层 *sql.DB（测试用）
func (s *Store) DB() *sql.DB {
	return s.db
}

func migrate(db *sql.DB) error {
	schemas := []string{
		// users 表（完整定义，含新字段）
		`CREATE TABLE IF NOT EXISTS users (
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
		`CREATE TABLE IF NOT EXISTS quotas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			quota_type TEXT NOT NULL DEFAULT 'free',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			used_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			question TEXT NOT NULL,
			primary_gua TEXT NOT NULL,
			changing_gua TEXT NOT NULL,
			yao_positions TEXT NOT NULL,
			interpretation TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_quotas_user_id ON quotas(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_quotas_used_at ON quotas(used_at)`,
		`CREATE INDEX IF NOT EXISTS idx_history_user_id ON history(user_id)`,

		// 旧库升级：users 表加字段（SQLite 不支持直接加列到 UNIQUE 约束表，用重建方式）
		`CREATE TABLE IF NOT EXISTS __users_migrated (done INTEGER DEFAULT 1)`,

		// LLM 模型表
		`CREATE TABLE IF NOT EXISTS llm_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			display_name TEXT DEFAULT '',
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
	}

	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("执行建表失败: %w\nSQL: %s", err, s)
		}
	}

	// 运行时检查旧 users 表是否需要迁移
	var colExists int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name IN ('role', 'is_active')`).Scan(&colExists)
	if err != nil {
		return err
	}
	if colExists < 2 {
		// 旧表缺少字段，执行迁移
		migrationSQLs := []string{
			`ALTER TABLE users RENAME TO users_old`,
			`CREATE TABLE users (
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
			`INSERT INTO users (id, phone, nickname, avatar, address, password, created_at)
			 SELECT id, phone, nickname, avatar, address, password, created_at FROM users_old`,
			`DROP TABLE users_old`,
		}
		for _, s := range migrationSQLs {
			if _, err := db.Exec(s); err != nil {
				return fmt.Errorf("执行迁移失败: %w\nSQL: %s", err, s)
			}
		}
	}

	// v2.2: 存量 display_name 迁移
	var colCount int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('llm_models') WHERE name = 'display_name'").Scan(&colCount)
	if colCount == 0 {
		// 旧表缺少 display_name 列，先加列
		if _, err := db.Exec("ALTER TABLE llm_models ADD COLUMN display_name TEXT DEFAULT ''"); err != nil {
			return fmt.Errorf("添加 display_name 列失败: %w", err)
		}
	}
	// 填充存量数据
	var emptyCount int
	db.QueryRow("SELECT COUNT(*) FROM llm_models WHERE display_name = '' OR display_name IS NULL").Scan(&emptyCount)
	if emptyCount > 0 {
		db.Exec("UPDATE llm_models SET display_name = provider || ' ' || name WHERE display_name = '' OR display_name IS NULL")
	}

	return nil
}
