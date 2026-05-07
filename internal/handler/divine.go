package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/qianwen"
)

// LineDisplay 前端六爻展示数据
type LineDisplay struct {
	Label    string // 初爻～上爻
	IsYang   bool
	Changing bool // 是否为变爻
	IsMaster bool // 是否为主变爻
}

// ResultData 结果页模板数据
type ResultData struct {
	Primary           *engine.GuaInfo
	Changing          *engine.GuaInfo
	ChangingPositions []int
	MasterYao         int
	Interpretation    string
	LineDisplays      []LineDisplay
}

// DivineHandler 算卦处理器
type DivineHandler struct {
	Tmpl    *template.Template
	Qianwen *qianwen.Client
}

func (h *DivineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "表单解析失败", http.StatusBadRequest)
		return
	}
	question := r.FormValue("question")
	if question == "" {
		http.Error(w, "请输入问题", http.StatusBadRequest)
		return
	}

	// 1. 起卦
	lines := engine.CastSixLines()

	// 2. 构建本卦变卦
	primary, changing, positions, master := engine.BuildHexagrams(lines)

	// 3. 生成变爻描述
	yaoDesc := formatYaoPositions(positions, master)

	// 4. 调用千问解卦
	prompt := qianwen.BuildPrompt(question, primary.Name, changing.Name, yaoDesc)
	interpretation, err := h.Qianwen.Divine(prompt)
	if err != nil {
		interpretation = fmt.Sprintf("解卦服务暂不可用：%v", err)
	}

	// 5. 构建六爻展示数据
	displays := buildLineDisplays(lines, positions, master)

	data := ResultData{
		Primary:           primary,
		Changing:          changing,
		ChangingPositions: positions,
		MasterYao:         master,
		Interpretation:    interpretation,
		LineDisplays:      displays,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "result", data); err != nil {
		http.Error(w, fmt.Sprintf("模板渲染失败: %v", err), http.StatusInternalServerError)
	}
}

// formatYaoPositions 格式化变爻描述文本
func formatYaoPositions(positions []int, master int) string {
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	var result string
	for i, p := range positions {
		if i > 0 {
			result += "、"
		}
		result += names[p]
	}
	if master >= 0 {
		result += fmt.Sprintf("（主变爻：%s）", names[master])
	}
	if result == "" {
		result = "无变爻"
	}
	return result
}

// buildLineDisplays 构建前端六爻展示数据
func buildLineDisplays(lines [6]int, changing []int, master int) []LineDisplay {
	names := []string{"上爻", "五爻", "四爻", "三爻", "二爻", "初爻"}
	displays := make([]LineDisplay, 6)

	// 从上到下展示 (index 5→0)
	for i := 0; i < 6; i++ {
		pos := 5 - i // 实际数组索引 (5=上爻, 0=初爻)
		displays[i] = LineDisplay{
			Label:    names[pos],
			IsYang:   (lines[pos] == 7 || lines[pos] == 9),
			Changing: containsInt(changing, pos),
			IsMaster: pos == master,
		}
	}
	return displays
}

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// YaoLabel 模板函数：0-based索引转中文爻名
func YaoLabel(pos int) string {
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	if pos < 0 || pos >= 6 {
		return "？爻"
	}
	return names[pos]
}

// YaoLabelFunc 暴露给模板的 yaoLabel 函数
func YaoLabelFunc() func(int) string {
	return YaoLabel
}

// sortPositions 排序变爻位置
func sortPositions(positions []int) {
	sort.Ints(positions)
}
