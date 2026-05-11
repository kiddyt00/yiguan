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
	// 乾为天: 上爻为老阳(9)
	// sum=44, 55-44=11, walkPath(11)=2(三爻, index 1)
	// 6/9变[5](上爻) + 55法强制变[1](三爻)
	// 本卦: 111111(乾) → 变卦: 泽火革 "101110"
	lines := [6]int{7, 7, 7, 7, 7, 9}
	primary, changing, positions, _ := BuildHexagrams(lines)
	if primary == nil || primary.Name != "乾为天" {
		t.Fatalf("primary = %v, want 乾为天", primary)
	}
	if changing == nil || changing.Name != "泽火革" {
		t.Errorf("changing = %v, want 泽火革", changing)
	}
	if len(positions) != 2 {
		t.Errorf("positions = %v, want 2 positions", positions)
	}

	// 坤为地: 初爻为老阴(6)
	// sum=46, 55-46=9, walkPath(9)=4(四爻, index 3)
	// 6/9变[0](初爻) + 55法强制变[3](四爻)
	// 本卦: 000000(坤) → 变卦: 震为雷 "100100"
	lines2 := [6]int{6, 8, 8, 8, 8, 8}
	_, changing2, _, _ := BuildHexagrams(lines2)
	if changing2 == nil || changing2.Name != "震为雷" {
		t.Errorf("changing = %v, want 震为雷", changing2)
	}
}

func TestChangingLinesMultiple(t *testing.T) {
	// 老阴老阳各一处: lines=[6,7,7,9,8,8], sum=45
	// steps=10, walkPath(10)=3(三爻, index 2) → 55法强制变
	// 6/9 auto: [0, 3] + 55: [2] → positions=[0,3,2]
	lines := [6]int{6, 7, 7, 9, 8, 8}
	_, _, positions, _ := BuildHexagrams(lines)
	if len(positions) != 3 {
		t.Errorf("expected 3 changing lines, got %d: %v", len(positions), positions)
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
		t.Errorf("expected positions to contain 0(初爻) and 3(四爻), got %v", positions)
	}
}

func TestFiftyFiveMethod(t *testing.T) {
	// sum=45: steps=55-45=10, walkPath(10)=3(三爻), 0-indexed=2
	pos := calcMasterYao(45)
	if pos != 2 {
		t.Errorf("calcMasterYao(45) = %d, want 2 (三爻)", pos)
	}

	// sum=42: steps=55-42=13, walkPath(13)=1(初爻), 0-indexed=0
	pos = calcMasterYao(42)
	if pos != 0 {
		t.Errorf("calcMasterYao(42) = %d, want 0 (初爻)", pos)
	}

	// sum=49: steps=55-49=6, walkPath(6)=6(上爻), 0-indexed=5
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
	// sum=45, steps=55-45=10, walkPath(10)=3(三爻), 0-indexed=2
	// 第3爻(7=少阳)变阴 → "010001" → 火风鼎
	lines := [6]int{8, 7, 8, 7, 8, 7}
	primary, changing, positions, master := BuildHexagrams(lines)
	if primary == nil || primary.Name != "火水未济" {
		t.Fatalf("primary = %v, want 火水未济", primary)
	}
	if changing == nil || changing.Name != "火风鼎" {
		t.Errorf("changing = %v, want 火风鼎", changing)
	}
	if master != 2 {
		t.Errorf("master yao = %d, want 2 (三爻)", master)
	}
	if len(positions) != 1 || positions[0] != 2 {
		t.Errorf("positions = %v, want [2]", positions)
	}
}
