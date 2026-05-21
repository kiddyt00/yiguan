package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) DB() *sql.DB  { return s.db }

func migrate(db *sql.DB) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone TEXT UNIQUE NOT NULL,
			openid TEXT DEFAULT '',
			nickname TEXT NOT NULL,
			avatar TEXT DEFAULT '',
			wx_avatar TEXT DEFAULT '',
			wx_sex INTEGER DEFAULT 0,
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
		`CREATE TABLE IF NOT EXISTS __users_migrated (done INTEGER DEFAULT 1)`,
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
		`CREATE TABLE IF NOT EXISTS translations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			history_id INTEGER NOT NULL,
			lang TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (history_id) REFERENCES history(id),
			UNIQUE(history_id, lang)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_translations_history_id ON translations(history_id)`,
		// v2.8: login_logs
		`CREATE TABLE IF NOT EXISTS login_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			ip TEXT DEFAULT '',
			city TEXT DEFAULT '',
			device TEXT DEFAULT '',
			os TEXT DEFAULT '',
			browser TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_login_logs_user_id ON login_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_login_logs_created ON login_logs(created_at)`,
	}
	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("执行建表失败: %w\nSQL: %s", err, s)
		}
	}
	// 旧表迁移：运行时加列
	migrations := []struct {
		table string
		col   string
		typ   string
		def   string
	}{
		{"users", "role", "TEXT", "'user'"},
		{"users", "is_active", "INTEGER", "1"},
		{"users", "openid", "TEXT", "''"},
		{"users", "avatar", "TEXT", "''"},
		{"users", "wx_avatar", "TEXT", "''"},
		{"users", "wx_sex", "INTEGER", "0"},
		{"llm_models", "display_name", "TEXT", "''"},
		{"history", "lang", "TEXT", "'zh'"},
		{"history", "primary_yao", "TEXT", "''"},
		{"history", "changing_yao", "TEXT", "''"},
		{"history", "toss_data", "TEXT", "''"},
		{"history", "master_yao", "INTEGER", "0"},
	}
	for _, m := range migrations {
		var n int
		db.QueryRow("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", m.table, m.col).Scan(&n)
		if n == 0 {
			sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s DEFAULT %s", m.table, m.col, m.typ, m.def)
			if _, err := db.Exec(sql); err != nil {
				return fmt.Errorf("添加 %s.%s 列失败: %w", m.table, m.col, err)
			}
		}
	}
	return nil
}
