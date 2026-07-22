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
func (s *Store) CreateUserByOpenID(openid, nickname, wxAvatar string) (*store.User, error) {
	phone := "wx:" + openid[:min(10, len(openid))]
	result, err := s.db.Exec(
		"INSERT INTO users (phone, openid, nickname, wx_avatar, password) VALUES (?, ?, ?, ?, ?)",
		phone, openid, nickname, wxAvatar, "",
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
		"SELECT id, phone, COALESCE(openid,''), COALESCE(unionid,''), nickname, avatar, COALESCE(wx_avatar,''), address, password, role, is_active, created_at FROM users WHERE phone = ?",
		phone,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.UnionID, &u.Nickname, &u.Avatar, &u.WxAvatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
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
		"SELECT id, phone, COALESCE(openid,''), COALESCE(unionid,''), nickname, avatar, COALESCE(wx_avatar,''), address, password, role, is_active, created_at FROM users WHERE openid = ?",
		openid,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.UnionID, &u.Nickname, &u.Avatar, &u.WxAvatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByUnionID 按微信 UnionID 查找用户
func (s *Store) GetUserByUnionID(unionid string) (*store.User, error) {
	u := &store.User{}
	err := s.db.QueryRow(
		"SELECT id, phone, COALESCE(openid,''), COALESCE(unionid,''), nickname, avatar, COALESCE(wx_avatar,''), address, password, role, is_active, created_at FROM users WHERE unionid = ?",
		unionid,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.UnionID, &u.Nickname, &u.Avatar, &u.WxAvatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserUnionID 更新用户的 unionid
func (s *Store) UpdateUserUnionID(id int64, unionid string) error {
	_, err := s.db.Exec("UPDATE users SET unionid = ? WHERE id = ?", unionid, id)
	return err
}

// GetUserByID 按 ID 查找用户
func (s *Store) GetUserByID(id int64) (*store.User, error) {
	u := &store.User{}
	err := s.db.QueryRow(
		"SELECT id, phone, COALESCE(openid,''), nickname, avatar, COALESCE(wx_avatar,''), address, password, role, is_active, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Phone, &u.OpenID, &u.Nickname, &u.Avatar, &u.WxAvatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserOpenID 绑定微信 openid（在已存在用户上设置）
func (s *Store) UpdateUserOpenID(id int64, openid string) error {
	_, err := s.db.Exec("UPDATE users SET openid = ? WHERE id = ?", openid, id)
	return err
}

// UpdateUser 更新昵称和地址
func (s *Store) UpdateUser(id int64, nickname, address string) error {
	_, err := s.db.Exec(
		"UPDATE users SET nickname = ?, address = ? WHERE id = ?",
		nickname, address, id,
	)
	return err
}

// UpdateUserWechatInfo 更新微信用户的昵称、头像和性别
func (s *Store) UpdateUserWechatInfo(id int64, nickname, wxAvatar string) error {
	_, err := s.db.Exec(
		"UPDATE users SET nickname = ?, wx_avatar = ? WHERE id = ?",
		nickname, wxAvatar, id,
	)
	return err
}

// UpdateUserGender 更新微信用户性别（1=男, 2=女）
func (s *Store) UpdateUserGender(id int64, sex int) error {
	_, err := s.db.Exec("UPDATE users SET wx_sex = ? WHERE id = ?", sex, id)
	return err
}

// SearchUsers 按昵称或手机号搜索用户
func (s *Store) SearchUsers(keyword string, limit, offset int) ([]store.User, error) {
	like := "%" + keyword + "%"
	rows, err := s.db.Query(
		"SELECT id, phone, COALESCE(openid,''), nickname, avatar, COALESCE(wx_avatar,''), address, password, role, is_active, created_at FROM users WHERE nickname LIKE ? OR phone LIKE ? ORDER BY id DESC LIMIT ? OFFSET ?",
		like, like, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []store.User
	for rows.Next() {
		var u store.User
		rows.Scan(&u.ID, &u.Phone, &u.OpenID, &u.Nickname, &u.Avatar, &u.WxAvatar, &u.Address, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
		u.Password = ""
		list = append(list, u)
	}
	return list, rows.Err()
}

// SearchUsersCount 搜索结果总数
func (s *Store) SearchUsersCount(keyword string) (int64, error) {
	var count int64
	like := "%" + keyword + "%"
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE nickname LIKE ? OR phone LIKE ?", like, like).Scan(&count)
	return count, err
}

func min(a, b int) int {
	if a < b { return a }
	return b
}
