package engine

import "testing"

func TestGuaDataCount(t *testing.T) {
	if len(AllGua) != 64 {
		t.Errorf("expected 64 gua, got %d", len(AllGua))
	}
}

func TestGuaDataUniqueID(t *testing.T) {
	seen := make(map[int]bool)
	for _, g := range AllGua {
		if seen[g.ID] {
			t.Errorf("duplicate ID: %d (%s)", g.ID, g.Name)
		}
		seen[g.ID] = true
	}
}

func TestFindGuaByYaoPattern(t *testing.T) {
	// 乾为天: 六阳爻 "111111"
	g := FindGuaByYaoPattern("111111")
	if g == nil {
		t.Fatal("FindGuaByYaoPattern('111111') returned nil")
	}
	if g.Name != "乾为天" {
		t.Errorf("expected 乾为天, got %s", g.Name)
	}

	// 坤为地: 六阴爻 "000000"
	g = FindGuaByYaoPattern("000000")
	if g == nil || g.Name != "坤为地" {
		t.Errorf("expected 坤为地, got %v", g)
	}
}

func TestYaoDescConvention(t *testing.T) {
	// 验证 YaoDesc 约定: 索引0=初爻(下), 索引5=上爻(上)
	// 地天泰: 下乾(111) + 上坤(000) → "111000"
	g := FindGuaByYaoPattern("111000")
	if g == nil || g.Name != "地天泰" {
		t.Errorf("expected 地天泰 for '111000', got %v", g)
	}
}
