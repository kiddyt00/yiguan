package engine

import "testing"

func TestBuildPrimary(t *testing.T) {
	// 全部少阴(8) → 坤为地 "000000"
	lines := [6]int{8, 8, 8, 8, 8, 8}
	primary := BuildPrimary(lines)
	if primary == nil {
		t.Fatal("BuildPrimary(all 8) returned nil")
	}
	if primary.Name != "坤为地" {
		t.Errorf("BuildPrimary(all 8) = %s, want 坤为地", primary.Name)
	}

	// 全部少阳(7) → 乾为天 "111111"
	lines2 := [6]int{7, 7, 7, 7, 7, 7}
	primary2 := BuildPrimary(lines2)
	if primary2 == nil || primary2.Name != "乾为天" {
		t.Errorf("BuildPrimary(all 7) = %v, want 乾为天", primary2)
	}

	// 混合: 7,8,7,8,7,8 → 阳阴阳阴阳阴 → "010101" → 水火既济
	lines3 := [6]int{7, 8, 7, 8, 7, 8}
	primary3 := BuildPrimary(lines3)
	if primary3 == nil || primary3.Name != "水火既济" {
		t.Errorf("BuildPrimary(7,8,7,8,7,8) = %v, want 水火既济", primary3)
	}
}

func TestBuildChangingAuto(t *testing.T) {
	// 乾为天: 上爻为老阳(9) → 6/9自动变 + 55法合并
	// sum=44, 55-44=11, 11%6=5, walkPath(5)=5(五爻, index 4)
	// 6/9变[5](上爻) + 55法强制变[4](五爻)
	// 本卦: 111111(乾) → 变卦: 111000(天地否)
	lines := [6]int{7, 7, 7, 7, 7, 9}
	primary, changing, positions, _ := BuildHexagrams(lines)
	if primary == nil || primary.Name != "乾为天" {
		t.Fatalf("primary = %v, want 乾为天", primary)
	}
	if changing == nil || changing.Name != "雷天大壮" {
		t.Errorf("changing = %v, want 雷天大壮", changing)
	}
	if len(positions) != 2 {
		t.Errorf("positions = %v, want [4, 5]", positions)
	}

	// 坤为地: 初爻为老阴(6) + 55法合并
	// sum=46, 55-46=9, 9%6=3, walkPath(3)=3(三爻, index 2)
	// 6/9变[0](初爻) + 55法强制变[2](三爻)
	// 本卦: 000000(坤) → 变卦: 101000(火地晋)
	lines2 := [6]int{6, 8, 8, 8, 8, 8}
	_, changing2, _, _ := BuildHexagrams(lines2)
	if changing2 == nil || changing2.Name != "地火明夷" {
		t.Errorf("changing = %v, want 地火明夷", changing2)
	}
}

func TestChangingLinesMultiple(t *testing.T) {
	// 老阴老阳各一处
	lines := [6]int{6, 7, 7, 9, 8, 8}
	_, _, positions, _ := BuildHexagrams(lines)
	if len(positions) != 2 {
		t.Errorf("expected 2 changing lines, got %d: %v", len(positions), positions)
	}
	has0 := false
	has3 := false
	for _, p := range positions {
		if p == 0 {
			has0 = true
		}
		if p == 3 {
			has3 = true
		}
	}
	if !has0 || !has3 {
		t.Errorf("expected positions [0, 3] (初爻+四爻), got %v", positions)
	}
}

func TestFiftyFiveMethod(t *testing.T) {
	// 无 6/9: lines = {7,8,7,8,7,8}, sum=45
	// 余数 = (55-45)%6 = 10%6 = 4
	// 路径: 第1步→1, 第2步→2, 第3步→3, 第4步→4
	// 落第4爻 (0-indexed=3)
	pos := calcMasterYao(45)
	if pos != 3 {
		t.Errorf("calcMasterYao(45) = %d, want 3 (第4爻)", pos)
	}

	// sum=42: 余数 = (55-42)%6 = 13%6 = 1 → 第1步→1, 落初爻
	pos = calcMasterYao(42)
	if pos != 0 {
		t.Errorf("calcMasterYao(42) = %d, want 0 (初爻)", pos)
	}

	// sum=49: 余数 = (55-49)%6 = 6%6 = 0 → 取6 → 第6步→6, 落上爻
	pos = calcMasterYao(49)
	if pos != 5 {
		t.Errorf("calcMasterYao(49) = %d, want 5 (上爻)", pos)
	}
}

func TestWalkPath(t *testing.T) {
	tests := []struct {
		steps int
		want  int // 1-indexed 爻位
	}{
		{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6},
		{7, 6}, {8, 5}, {9, 4}, {10, 3}, {11, 2}, {12, 1},
		{13, 1}, {14, 2}, // 新循环
	}
	for _, tc := range tests {
		got := walkPath(tc.steps)
		if got != tc.want {
			t.Errorf("walkPath(%d) = %d, want %d", tc.steps, got, tc.want)
		}
	}
}

func TestBuildHexagrams55ForceChange(t *testing.T) {
	// 所有爻为少阴少阳(无6/9): {8,7,8,7,8,7}
	// 本卦=阴阳交替 → 火水未济 "010101"
	// 变卦: 55法强制变爻 (sum=45, 主变爻=第4爻, 0-indexed=3)
	// 第4爻(7=少阳)变阴 → "010001" → 山水蒙
	lines := [6]int{8, 7, 8, 7, 8, 7}
	primary, changing, positions, master := BuildHexagrams(lines)
	if primary == nil || primary.Name != "火水未济" {
		t.Fatalf("primary = %v, want 火水未济", primary)
	}
	if changing == nil || changing.Name != "山水蒙" {
		t.Errorf("changing = %v, want 山水蒙", changing)
	}
	if master != 3 {
		t.Errorf("master yao = %d, want 3 (第4爻)", master)
	}
	if len(positions) != 1 || positions[0] != 3 {
		t.Errorf("positions = %v, want [3]", positions)
	}
}
