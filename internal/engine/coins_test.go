package engine

import "testing"

func TestTossCoinsRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := tossCoins()
		if result < 6 || result > 9 {
			t.Errorf("tossCoins() = %d, want 6-9 (iteration %d)", result, i)
		}
	}
}

func TestTossCoinsDistribution(t *testing.T) {
	counts := map[int]int{6: 0, 7: 0, 8: 0, 9: 0}
	for i := 0; i < 10000; i++ {
		counts[tossCoins()]++
	}
	// 概率分布: 6(1/8) 7(3/8) 8(3/8) 9(1/8)
	// 每种都应至少出现一次
	for _, v := range []int{6, 7, 8, 9} {
		if counts[v] == 0 {
			t.Errorf("value %d never appeared in 10000 tosses", v)
		}
	}
	// 7和8应该比6和9多
	if counts[7]+counts[8] <= counts[6]+counts[9] {
		t.Errorf("unexpected distribution: 6=%d 7=%d 8=%d 9=%d", counts[6], counts[7], counts[8], counts[9])
	}
}

func TestCastSixLines(t *testing.T) {
	lines := CastSixLines()
	if len(lines) != 6 {
		t.Fatalf("CastSixLines() returned %d lines, want 6", len(lines))
	}
	for i, v := range lines {
		if v < 6 || v > 9 {
			t.Errorf("line %d (0=初爻) = %d, want 6-9", i, v)
		}
	}
}

func TestCastSixLinesReproducibility(t *testing.T) {
	// 多次调用应该产生不同结果（概率上）
	results := make(map[[6]int]bool)
	for i := 0; i < 100; i++ {
		results[CastSixLines()] = true
	}
	if len(results) < 10 {
		t.Errorf("only %d unique results in 100 calls, expected more randomness", len(results))
	}
}
