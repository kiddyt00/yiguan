package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser 创建用户，密码 bcrypt 哈希，返回 User（不含密码）
func (s *Store) CreateUser(phone, password, nickname string) (*store.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(
		"INSERT INTO users (phone, nickname, password) VALUES (?, ?, ?)",
		phone, nickname, string(hash),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	u, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

// CreateUserByOpenID 微信用户创建（无密码，phone 用 wx: 前缀占位）
func (s *Store) CreateUserByOpenID(openid, nickname string) (*store.User, error) {
	phone := "wx:" + openid[:min(10, len(openid))]
	result, err := s.db.Exec(
		"INSERT INTO users (phone, openid, nickname, password) VALUES (?, ?, ?, ?)",
		phone, openid, nickname, "",
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	// 赠送 3 次免费 quota
	for i := 0; i < 3; i++ {
		s.AddQuota(id, "free")
	}

	u, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

// GetUserByPhone 按手机号查找用户
func (s *Store) GetUserByPhone(phone string) (*store.User, error) {
	u := &store.User{}
	err := s.db.QueryRow(
		"SELECT id, phone, COALESCE(openid,''), nickname, avatar, address, password, role, is_active, created_at FROM users WHERE phone = ?",
		phone,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.Nickname, &u.Avatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByOpenID 按微信 OpenID 查找用户
func (s *Store) GetUserByOpenID(openid string) (*store.User, error) {
	u := &store.User{}
	err := s.db.QueryRow(
		"SELECT id, phone, COALESCE(openid,''), nickname, avatar, address, password, role, is_active, created_at FROM users WHERE openid = ?",
		openid,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.Nickname, &u.Avatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByID 按 ID 查找用户
func (s *Store) GetUserByID(id int64) (*store.User, error) {
	u := &store.User{}
	err := s.db.QueryRow(
		"SELECT id, phone, COALESCE(openid,''), nickname, avatar, address, password, role, is_active, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.Nickname, &u.Avatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUser 更新昵称和地址
func (s *Store) UpdateUser(id int64, nickname, address string) error {
	_, err := s.db.Exec(
		"UPDATE users SET nickname = ?, address = ? WHERE id = ?",
		nickname, address, id,
	)
	return err
}

func min(a, b int) int {
	if a < b { return a }
	return b
}
