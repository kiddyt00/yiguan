package engine

import "math/rand"

// tossCoins 模拟抛3枚铜钱，返回总和 6/7/8/9
// 正面=3, 反面=2，每枚正反概率相同
func tossCoins() int {
	sum := 0
	for i := 0; i < 3; i++ {
		if rand.Intn(2) == 0 {
			sum += 3 // 正面
		} else {
			sum += 2 // 反面
		}
	}
	return sum
}

// CastSixLines 起六爻，返回 [初爻, 二爻, 三爻, 四爻, 五爻, 上爻] 6个值 ∈ {6,7,8,9}
// index 0 = 初爻 (最下), index 5 = 上爻 (最上)
func CastSixLines() [6]int {
	var lines [6]int
	for i := 0; i < 6; i++ {
		lines[i] = tossCoins()
	}
	return lines
}
