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
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone TEXT UNIQUE NOT NULL,
			nickname TEXT NOT NULL,
			avatar TEXT DEFAULT '',
			address TEXT DEFAULT '',
			password TEXT NOT NULL,
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
		`CREATE TABLE IF NOT EXISTS llm_providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			provider TEXT NOT NULL UNIQUE,
			api_key TEXT NOT NULL DEFAULT '',
			endpoint TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			is_default INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`PRAGMA foreign_keys = ON`,
	}

	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("执行建表失败: %w\nSQL: %s", err, s)
		}
	}
	return nil
}
