package sqlite

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/kiddyt00/yiguan/internal/store"
)

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateInviteCode 生成/获取用户邀请码
func (s *Store) GenerateInviteCode(userID int64) (string, error) {
	// 先查是否已有
	var code string
	err := s.db.QueryRow("SELECT invite_code FROM users WHERE id = ? AND invite_code != ''", userID).Scan(&code)
	if err == nil && code != "" {
		return code, nil
	}

	// 生成8位邀请码
	for i := 0; i < 10; i++ {
		code = randInviteCode(8)
		// 检查唯一性
		var c int
		s.db.QueryRow("SELECT COUNT(*) FROM users WHERE invite_code = ?", code).Scan(&c)
		if c == 0 {
			_, err := s.db.Exec("UPDATE users SET invite_code = ? WHERE id = ?", code, userID)
			if err != nil {
				return "", err
			}
			return code, nil
		}
	}
	return "", fmt.Errorf("无法生成唯一邀请码")
}

// BindInvitation 注册时绑定邀请关系
func (s *Store) BindInvitation(inviteeID int64, inviteCode string) error {
	// 查找邀请码对应的用户
	var inviterID int64
	err := s.db.QueryRow("SELECT id FROM users WHERE invite_code = ?", inviteCode).Scan(&inviterID)
	if err == sql.ErrNoRows {
		return nil // 邀请码无效，静默忽略
	}
	if err != nil {
		return err
	}

	// 不能邀请自己
	if inviterID == inviteeID {
		return nil
	}

	// 记录邀请关系（UNIQUE 约束保证每人只能被邀请一次）
	_, err = s.db.Exec(
		`INSERT OR IGNORE INTO invitations (inviter_id, invitee_id, invite_code, status)
		 VALUES (?, ?, ?, 'registered')`,
		inviterID, inviteeID, inviteCode,
	)
	return err
}

// RewardInviterIfEligible 被邀请人完成测算后，检查邀请人是否达到奖励条件
func (s *Store) RewardInviterIfEligible(inviteeID int64) error {
	// 先查邀请人
	var inviterID int64
	err := s.db.QueryRow(
		"SELECT inviter_id FROM invitations WHERE invitee_id = ? AND status = 'registered'",
		inviteeID,
	).Scan(&inviterID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	// 标记为已测算
	s.db.Exec("UPDATE invitations SET status = 'divined' WHERE invitee_id = ?", inviteeID)

	// 检查邀请人是否达到奖励条件
	s.CheckAndReward(inviterID)
	return nil
}

// MarkInviteeDivined 被邀请人完成测算后标记
func (s *Store) MarkInviteeDivined(userID int64) error {
	_, err := s.db.Exec(
		"UPDATE invitations SET status = 'divined' WHERE invitee_id = ? AND status = 'registered'",
		userID,
	)
	return err
}

// GetInviteProgress 获取邀请人拉新进度
func (s *Store) GetInviteProgress(userID int64) (*store.InviteProgress, error) {
	p := &store.InviteProgress{}

	// 统计已注册的被邀请人数
	s.db.QueryRow(
		"SELECT COUNT(*) FROM invitations WHERE inviter_id = ? AND status IN ('registered','divined')",
		userID,
	).Scan(&p.RegisteredCount)

	// 统计已完成测算的被邀请人数
	s.db.QueryRow(
		"SELECT COUNT(*) FROM invitations WHERE inviter_id = ? AND status = 'divined'",
		userID,
	).Scan(&p.DivinedCount)

	// 计算已发放的奖励轮数（以 invitation 中被标记为 rewarded 的数量 ÷ 3 来计算）
	var rewardedCount int
	s.db.QueryRow(
		"SELECT COUNT(*) FROM invitations WHERE inviter_id = ? AND status = 'rewarded'",
		userID,
	).Scan(&rewardedCount)
	p.RewardRound = rewardedCount / 3

	// 判断是否有可领取的奖励
	p.PendingReward = p.RegisteredCount >= 3 && p.DivinedCount >= 1 &&
		(p.RegisteredCount/3) > p.RewardRound

	return p, nil
}

// CheckAndReward 检查并发放奖励
func (s *Store) CheckAndReward(userID int64) (bool, error) {
	p, err := s.GetInviteProgress(userID)
	if err != nil {
		return false, err
	}

	if !p.PendingReward {
		return false, nil
	}

	// 找到本轮已达到条件的3个邀请（最新注册的3个）
	rows, err := s.db.Query(
		`SELECT id FROM invitations
		 WHERE inviter_id = ? AND status IN ('registered','divined')
		 ORDER BY created_at ASC LIMIT 3`,
		userID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) < 3 {
		return false, nil
	}

	// 将这3条标记为 rewarded
	for _, id := range ids {
		s.db.Exec("UPDATE invitations SET status = 'rewarded' WHERE id = ?", id)
	}

	// 发放奖励：赠送1次免费 quota
	if err := s.AddQuota(userID, "referral"); err != nil {
		return false, err
	}

	return true, nil
}

func randInviteCode(length int) string {
	buf := make([]byte, length)
	rand.Read(buf)
	code := make([]byte, length)
	for i, b := range buf {
		code[i] = inviteCodeChars[int(b)%len(inviteCodeChars)]
	}
	return string(code)
}

func init() {
	// 确保 rand 能正确导入
	_ = hex.EncodeToString
}
