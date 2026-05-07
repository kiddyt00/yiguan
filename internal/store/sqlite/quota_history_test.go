package sqlite

import (
	"testing"

	"github.com/kiddyt00/yiguan/internal/store"
)

func TestQuotaFlow(t *testing.T) {
	s := openTestDB(t)
	u, _ := s.CreateUser("13800138000", "pass", "test")

	// 新用户无 quota
	remaining, _ := s.GetRemainingQuota(u.ID)
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}

	// 赠送 3 个免费 quota
	s.AddQuota(u.ID, "free")
	s.AddQuota(u.ID, "free")
	s.AddQuota(u.ID, "free")

	remaining, _ = s.GetRemainingQuota(u.ID)
	if remaining != 3 {
		t.Errorf("remaining = %d, want 3", remaining)
	}

	// 扣减 1 次
	err := s.ConsumeQuota(u.ID)
	if err != nil {
		t.Fatal(err)
	}

	remaining, _ = s.GetRemainingQuota(u.ID)
	if remaining != 2 {
		t.Errorf("after 1 consume, remaining = %d, want 2", remaining)
	}
}

func TestConsumeQuotaEmpty(t *testing.T) {
	s := openTestDB(t)
	u, _ := s.CreateUser("13800138000", "pass", "test")

	err := s.ConsumeQuota(u.ID)
	if err != store.ErrQuotaExhausted {
		t.Errorf("expected ErrQuotaExhausted, got %v", err)
	}
}

func TestConsumeQuotaFIFO(t *testing.T) {
	s := openTestDB(t)
	u, _ := s.CreateUser("13800138000", "pass", "test")

	// 混合类型: free + paid
	s.AddQuota(u.ID, "free")
	s.AddQuota(u.ID, "paid")

	// 先消耗最早创建的
	s.ConsumeQuota(u.ID)
	remaining, _ := s.GetRemainingQuota(u.ID)
	if remaining != 1 {
		t.Errorf("remaining = %d, want 1", remaining)
	}
}

func TestHistorySaveAndList(t *testing.T) {
	s := openTestDB(t)
	u, _ := s.CreateUser("13800138000", "pass", "test")

	h := &store.History{
		UserID: u.ID, Question: "今天运气如何？",
		PrimaryGua: "乾为天", ChangingGua: "天风姤",
		YaoPositions: "上爻", Interpretation: "大吉大利",
	}
	if err := s.SaveHistory(h); err != nil {
		t.Fatal(err)
	}

	list, err := s.GetHistory(u.ID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("count = %d, want 1", len(list))
	}
	if list[0].PrimaryGua != "乾为天" {
		t.Errorf("primary = %s", list[0].PrimaryGua)
	}
}

func TestHistoryPagination(t *testing.T) {
	s := openTestDB(t)
	u, _ := s.CreateUser("13800138000", "pass", "test")

	for i := 0; i < 5; i++ {
		s.SaveHistory(&store.History{
			UserID: u.ID, Question: "test",
			PrimaryGua: "乾", ChangingGua: "坤",
			YaoPositions: "初", Interpretation: "ok",
		})
	}

	// 取前 3
	list, _ := s.GetHistory(u.ID, 3, 0)
	if len(list) != 3 {
		t.Fatalf("page1 = %d, want 3", len(list))
	}

	// 取后 2
	list, _ = s.GetHistory(u.ID, 3, 3)
	if len(list) != 2 {
		t.Fatalf("page2 = %d, want 2", len(list))
	}
}
