package engine

// linesToPattern 将六爻阴阳数组转为查找模式字符串
// index 0 = 初爻, index 5 = 上爻
func linesToPattern(yaos []bool) string {
	b := make([]byte, 6)
	for i := 0; i < 6; i++ {
		if yaos[i] {
			b[i] = '1'
		} else {
			b[i] = '0'
		}
	}
	return string(b)
}

// BuildPrimary 从六爻值构建本卦
// lines[0]=初爻, lines[5]=上爻
func BuildPrimary(lines [6]int) *GuaInfo {
	yaos := make([]bool, 6)
	for i, v := range lines {
		yaos[i] = (v == 7 || v == 9) // 7/9 为阳
	}
	return FindGuaByYaoPattern(linesToPattern(yaos))
}

// BuildHexagrams 从六爻构建本卦、变卦、变爻位置、主变爻索引
// 返回: 本卦, 变卦, 变爻索引列表(0-based), 主变爻索引
func BuildHexagrams(lines [6]int) (*GuaInfo, *GuaInfo, []int, int) {
	// 本卦
	yaos := make([]bool, 6)
	changingPositions := []int{}
	for i, v := range lines {
		yaos[i] = (v == 7 || v == 9)
		if v == 6 || v == 9 {
			changingPositions = append(changingPositions, i)
		}
	}
	primary := FindGuaByYaoPattern(linesToPattern(yaos))

	// 变卦：先应用 6/9 自动变爻
	changed := make([]bool, 6)
	copy(changed, yaos)
	for i, v := range lines {
		if v == 6 || v == 9 {
			changed[i] = !changed[i] // 翻转阴阳
		}
	}

	// 55 法定主变爻
	sum := 0
	for _, v := range lines {
		sum += v
	}
	masterPos := calcMasterYao(sum)

	// 如果主变爻位置不在自动变爻列表中，强制变化
	found := false
	for _, p := range changingPositions {
		if p == masterPos {
			found = true
			break
		}
	}
	if !found {
		changed[masterPos] = !changed[masterPos]
		changingPositions = append(changingPositions, masterPos)
	}

	changing := FindGuaByYaoPattern(linesToPattern(changed))
	return primary, changing, changingPositions, masterPos
}

// calcMasterYao 55减总数定主变爻，返回 0-based 索引
func calcMasterYao(total int) int {
	remainder := (55 - total) % 6
	if remainder == 0 {
		remainder = 6
	}
	// 路径: 1,2,3,4,5,6,6,5,4,3,2,1,...
	pos := walkPath(remainder)
	return pos - 1 // 转 0-indexed
}

// walkPath 按循环路径走 steps 步，返回 1-indexed 爻位
// 路径序列: 1→2→3→4→5→6→6→5→4→3→2→1→1→2→...
func walkPath(steps int) int {
	cycle := []int{1, 2, 3, 4, 5, 6, 6, 5, 4, 3, 2, 1}
	idx := (steps - 1) % len(cycle)
	return cycle[idx]
}

// FormatYaoPositions 格式化变爻描述文本（用于 prompt）
func FormatYaoPositions(positions []int, master int) string {
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	var result string
	for i, p := range positions {
		if i > 0 {
			result += "、"
		}
		result += names[p]
	}
	if master >= 0 && master < 6 {
		result += "（主变爻：" + names[master] + "）"
	}
	if result == "" {
		result = "无变爻"
	}
	return result
}
