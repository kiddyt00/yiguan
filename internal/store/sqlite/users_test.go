package sqlite

import (
	"testing"

	"github.com/kiddyt00/yiguan/internal/store"
)

func openTestDB(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateUser(t *testing.T) {
	s := openTestDB(t)
	u, err := s.CreateUser("13800138000", "pass123", "测试用户")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == 0 {
		t.Error("user ID should not be 0")
	}
	if u.Phone != "13800138000" {
		t.Errorf("phone = %s, want 13800138000", u.Phone)
	}
	if u.Password != "" {
		t.Error("returned user should not expose password hash")
	}
}

func TestCreateUserDuplicatePhone(t *testing.T) {
	s := openTestDB(t)
	s.CreateUser("13800138000", "pass1", "user1")
	_, err := s.CreateUser("13800138000", "pass2", "user2")
	if err == nil {
		t.Error("expected error for duplicate phone")
	}
}

func TestGetUserByPhone(t *testing.T) {
	s := openTestDB(t)
	s.CreateUser("13800138000", "testpass", "张三")
	u, err := s.GetUserByPhone("13800138000")
	if err != nil {
		t.Fatal(err)
	}
	if u.Nickname != "张三" {
		t.Errorf("nickname = %s, want 张三", u.Nickname)
	}
	if u.Password == "testpass" {
		t.Error("password should be hashed, not plaintext")
	}
}

func TestGetUserByPhoneNotFound(t *testing.T) {
	s := openTestDB(t)
	_, err := s.GetUserByPhone("99999999999")
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	s := openTestDB(t)
	u, _ := s.CreateUser("13800138000", "pass", "旧名")
	err := s.UpdateUser(u.ID, "新名", "北京市")
	if err != nil {
		t.Fatal(err)
	}
	u2, _ := s.GetUserByID(u.ID)
	if u2.Nickname != "新名" {
		t.Errorf("nickname = %s, want 新名", u2.Nickname)
	}
	if u2.Address != "北京市" {
		t.Errorf("address = %s, want 北京市", u2.Address)
	}
}
