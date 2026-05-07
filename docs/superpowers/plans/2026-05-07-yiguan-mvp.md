# 易观·AI 算卦 MVP 实现计划

> **面向 AI 代理的工作者：** 使用 subagent-driven-development 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 构建 Web 端在线算卦 MVP — 用户输入问题 → 铜钱起卦 → 千问 AI 解卦 → 展示结果 + 大师二维码 + 广告位

**架构：** Go 标准库 HTTP 服务端 + HTML/HTMX/Tailwind CSS 前端。算卦引擎独立模块，千问 API 独立 client。无数据库，纯无状态服务。

**技术栈：** Go 1.21+, net/http, html/template, HTMX 2.0, Tailwind CSS CDN, 通义千问 API

---

## 文件结构

```
yiguan/
├── main.go                    # 入口，路由注册，启动服务
├── go.mod / go.sum
├── internal/
│   ├── engine/
│   │   ├── coins.go           # 铜钱抛掷，起爻算法
│   │   ├── hexagram.go        # 本卦/变卦构建，变爻计算
│   │   └── gua_data.go        # 64卦常量数据（卦名、卦辞、象征）
│   ├── qianwen/
│   │   └── client.go          # 通义千问 API 客户端
│   └── handler/
│       ├── home.go            # 首页处理器
│       └── divine.go          # 算卦处理器（接收问题，返回结果）
├── templates/
│   ├── layout.html            # 基础布局（header/footer/HTMX+Tailwind CDN）
│   ├── home.html              # 首页：问题输入框 + 开始提问按钮
│   ├── result.html            # 结果页：卦象展示 + AI解卦 + 二维码 + 广告
│   └── partials/
│       └── hexagram.html      # 卦象可视化组件（六爻图）
├── static/
│   └── style.css              # 自定义样式（动画、卦象样式）
├── config.yaml                # 配置文件（千问 API key、服务端口）
└── docs/
    └── PRD.md                 # 产品需求文档（已存在）
```

---

### 任务 1：项目骨架搭建

**文件：**
- 创建：`go.mod`、`main.go`、`config.yaml`
- 创建：`internal/`、`templates/`、`static/` 目录

- [ ] **步骤 1：初始化 Go module**

```bash
cd /home/kiddyt00/claude-projects/yiguan
go mod init github.com/kiddyt00/yiguan
```

- [ ] **步骤 2：创建 config.yaml**

```yaml
server:
  port: 8080
qianwen:
  api_key: ""  # 通过环境变量 DASHSCOPE_API_KEY 覆盖
  model: "qwen-plus"
  endpoint: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
```

- [ ] **步骤 3：创建 main.go 骨架 — 仅启动空服务**

```go
package main

import (
    "log"
    "net/http"
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Server struct {
        Port string `yaml:"port"`
    } `yaml:"server"`
    Qianwen struct {
        APIKey   string `yaml:"api_key"`
        Model    string `yaml:"model"`
        Endpoint string `yaml:"endpoint"`
    } `yaml:"qianwen"`
}

func main() {
    cfg := loadConfig("config.yaml")
    if key := os.Getenv("DASHSCOPE_API_KEY"); key != "" {
        cfg.Qianwen.APIKey = key
    }

    mux := http.NewServeMux()
    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("易观 API OK"))
    })

    log.Printf("易观服务启动 http://localhost:%s", cfg.Server.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}

func loadConfig(path string) *Config {
    // ... 读取并解析 config.yaml
}
```

- [ ] **步骤 4：创建目录结构**

```bash
mkdir -p internal/engine internal/qianwen internal/handler
mkdir -p templates/partials static
```

- [ ] **步骤 5：运行服务验证可启动**

```bash
go run main.go
# 预期: 易观服务启动 http://localhost:8080
# curl localhost:8080 → "易观 API OK"
```

- [ ] **步骤 6：Commit**

```bash
git add -A
git commit -m "chore: 初始化项目骨架 — Go module, config, main.go 空服务"
```

---

### 任务 2：64卦数据常量

**文件：**
- 创建：`internal/engine/gua_data.go`
- 测试：`internal/engine/gua_data_test.go`

- [ ] **步骤 1：定义数据结构**

每个卦的数据结构：
```go
type GuaInfo struct {
    ID      int    // 1-64
    Name    string // 乾为天
    Symbol  string // ☰☰
    ShangGua string // 上卦名（乾/坤/震/巽/坎/离/艮/兑）
    XiaGua  string // 下卦名
    YaoDesc string // 六爻阴阳描述，如 "111111" 表示六阳爻（1=阳，0=阴）
    GuaCi   string // 卦辞
    XiangCi string // 象辞
}
```

- [ ] **步骤 2：编写测试 — 验证 64 卦数据完整性**

```go
func TestGuaDataCount(t *testing.T) {
    if len(AllGua) != 64 {
        t.Errorf("expected 64 gua, got %d", len(AllGua))
    }
}

func TestGuaDataUniqueID(t *testing.T) {
    seen := make(map[int]bool)
    for _, g := range AllGua {
        if seen[g.ID] {
            t.Errorf("duplicate ID: %d", g.ID)
        }
        seen[g.ID] = true
    }
}

func TestFindGuaByPattern(t *testing.T) {
    g := FindGuaByYaoPattern("111111")
    if g.Name != "乾为天" {
        t.Errorf("expected 乾为天, got %s", g.Name)
    }
}
```

- [ ] **步骤 3：运行测试验证失败**

```bash
go test ./internal/engine/ -run TestGua -v
# 预期: FAIL
```

- [ ] **步骤 4：实现 64 卦数据常量 AllGua 切片和 FindGuaByYaoPattern 函数**

```go
var AllGua = []GuaInfo{
    {ID: 1, Name: "乾为天", Symbol: "☰☰", ShangGua: "乾", XiaGua: "乾", YaoDesc: "111111",
     GuaCi: "元亨利贞", XiangCi: "天行健，君子以自强不息"},
    {ID: 2, Name: "坤为地", Symbol: "☷☷", ShangGua: "坤", XiaGua: "坤", YaoDesc: "000000",
     GuaCi: "元亨，利牝马之贞", XiangCi: "地势坤，君子以厚德载物"},
    // ... 共64卦
}

func FindGuaByYaoPattern(pattern string) *GuaInfo {
    for i := range AllGua {
        if AllGua[i].YaoDesc == pattern {
            return &AllGua[i]
        }
    }
    return nil
}
```

> **注意：** 需要在 `YaoDesc` 中区分本卦和变卦的查找。六爻的阴阳模式确定后，YaoDesc 从上爻到下爻（传统排序）或初爻到上爻（与代码一致）。此处约定：YaoDesc 索引 0 = 初爻，索引 5 = 上爻。

- [ ] **步骤 5：运行测试验证通过**

```bash
go test ./internal/engine/ -run TestGua -v
# 预期: PASS
```

- [ ] **步骤 6：Commit**

```bash
git add internal/engine/gua_data.go internal/engine/gua_data_test.go
git commit -m "feat: 添加64卦数据常量和查找函数"
```

---

### 任务 3：算卦引擎 — 铜钱起爻

**文件：**
- 创建：`internal/engine/coins.go`
- 测试：`internal/engine/coins_test.go`

- [ ] **步骤 1：编写测试 — 起爻结果范围验证**

```go
func TestTossCoinsRange(t *testing.T) {
    for i := 0; i < 1000; i++ {
        result := tossCoins()
        if result < 6 || result > 9 {
            t.Errorf("tossCoins() = %d, want 6-9", result)
        }
    }
}

func TestTossCoinsDistribution(t *testing.T) {
    counts := map[int]int{6: 0, 7: 0, 8: 0, 9: 0}
    for i := 0; i < 10000; i++ {
        counts[tossCoins()]++
    }
    // 概率分布: 6(1/8) 7(3/8) 8(3/8) 9(1/8)
    // 不严格断言，但每种都应出现
    for _, v := range []int{6, 7, 8, 9} {
        if counts[v] == 0 {
            t.Errorf("value %d never appeared in 10000 tosses", v)
        }
    }
}

func TestCastSixLines(t *testing.T) {
    lines := CastSixLines()
    if len(lines) != 6 {
        t.Errorf("CastSixLines() returned %d lines, want 6", len(lines))
    }
    for i, v := range lines {
        if v < 6 || v > 9 {
            t.Errorf("line %d = %d, want 6-9", i, v)
        }
    }
}
```

- [ ] **步骤 2：运行测试验证失败**

```bash
go test ./internal/engine/ -run TestToss -v
# 预期: FAIL
```

- [ ] **步骤 3：实现铜钱起爻**

```go
package engine

import "math/rand"

// tossCoins 模拟抛3枚铜钱，返回 6/7/8/9
// 正面=3, 反面=2
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

// CastSixLines 起六爻，返回 [初爻...上爻] 6个值 ∈ {6,7,8,9}
func CastSixLines() [6]int {
    var lines [6]int
    for i := 0; i < 6; i++ {
        lines[i] = tossCoins()
    }
    return lines
}
```

- [ ] **步骤 4：运行测试验证通过**

```bash
go test ./internal/engine/ -run TestToss -v
# 预期: PASS
```

- [ ] **步骤 5：Commit**

```bash
git add internal/engine/coins.go internal/engine/coins_test.go
git commit -m "feat: 铜钱起爻引擎 — 3枚铜钱6次抛掷"
```

---

### 任务 4：算卦引擎 — 本卦/变卦构建

**文件：**
- 创建：`internal/engine/hexagram.go`
- 测试：`internal/engine/hexagram_test.go`

- [ ] **步骤 1：编写测试 — 本卦查找**

```go
func TestBuildPrimary(t *testing.T) {
    // 全部少阴(8) → 坤为地 "000000"
    lines := [6]int{8, 8, 8, 8, 8, 8}
    primary := BuildPrimary(lines)
    if primary == nil || primary.Name != "坤为地" {
        t.Errorf("BuildPrimary(all 8) = %v, want 坤为地", primary)
    }

    // 全部少阳(7) → 乾为天 "111111"
    lines2 := [6]int{7, 7, 7, 7, 7, 7}
    primary2 := BuildPrimary(lines2)
    if primary2 == nil || primary2.Name != "乾为天" {
        t.Errorf("BuildPrimary(all 7) = %v, want 乾为天", primary2)
    }
}

func TestBuildChanging(t *testing.T) {
    // 老阳(9)变阴 → 乾变其他
    lines := [6]int{7, 7, 7, 7, 7, 9} // 上爻为老阳
    lines[5] = 9
    primary, changing, positions := BuildHexagrams(lines)
    if primary == nil || changing == nil {
        t.Fatal("BuildHexagrams returned nil")
    }
    if primary.Name != "乾为天" {
        t.Errorf("primary = %s, want 乾为天", primary.Name)
    }
    // 第6爻(上爻)9变阴 → 泽天夬 "011111"
    if changing.Name != "泽天夬" {
        t.Errorf("changing = %s, want 泽天夬", changing.Name)
    }
}

func TestChangingLines(t *testing.T) {
    // 老阴老阳各一处
    lines := [6]int{6, 7, 7, 9, 8, 8}
    _, _, positions := BuildHexagrams(lines)
    // positions 应包含 初爻(6) 和 第四爻(9)
    if len(positions) != 2 {
        t.Errorf("expected 2 changing lines, got %d", len(positions))
    }
}

func TestFiftyFiveMethod(t *testing.T) {
    lines := [6]int{7, 8, 7, 8, 7, 8} // 无 6/9 可变的爻
    sum := 0
    for _, v := range lines {
        sum += v
    }
    // 总数=45, 余数=(55-45)%6=10%6=4
    // 路径: 1→2→3→4, 落第4爻
    pos := calcMasterYao(sum)
    if pos != 3 { // 0-indexed
        t.Errorf("calcMasterYao(45) = %d, want 3 (第4爻)", pos)
    }
}
```

- [ ] **步骤 2：运行测试验证失败**

```bash
go test ./internal/engine/ -run "TestBuild|TestChanging|TestFifty" -v
# 预期: FAIL
```

- [ ] **步骤 3：实现本卦/变卦构建逻辑**

```go
package engine

// Line 表示一个爻
type Line struct {
    Value    int  // 6,7,8,9
    IsYang   bool // 7或9为阳
    Changing bool // 6或9为可变
}

// linesToPattern 将六爻转为查找模式字符串
// index 0 = 初爻, index 5 = 上爻
func linesToPattern(yaos []bool) string {
    // 从上爻到下爻排列，与 YaoDesc 约定一致
    // YaoDesc 约定: index 0=初爻, index 5=上爻
    // 对于传统表示，我们直接用布尔数组的字符串表示
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

// BuildPrimary 从六爻构建本卦
func BuildPrimary(lines [6]int) *GuaInfo {
    yaos := make([]bool, 6)
    for i, v := range lines {
        yaos[i] = (v == 7 || v == 9) // 7/9 为阳
    }
    return FindGuaByYaoPattern(linesToPattern(yaos))
}

// BuildHexagrams 从六爻构建本卦、变卦、变爻位置
// 返回: 本卦, 变卦, 变爻索引列表(0-based), 主变爻索引
func BuildHexagrams(lines [6]int) (*GuaInfo, *GuaInfo, []int, int) {
    // 本卦
    yaos := make([]bool, 6)
    changingYaos := make([]bool, 6)
    changingPositions := []int{}
    for i, v := range lines {
        yaos[i] = (v == 7 || v == 9)
        if v == 6 || v == 9 {
            changingYaos[i] = true
            changingPositions = append(changingPositions, i)
        }
    }
    primary := FindGuaByYaoPattern(linesToPattern(yaos))

    // 变卦：先应用 6/9 变爻
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
    // 如果主变爻位置不在 changingPositions 中，强制变化
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
    // 路径: 1,2,3,4,5,6,6,5,4,3,2,1,1,2,...
    // 第K步的爻位 (1-indexed)
    pos := walkPath(remainder)
    return pos - 1 // 转 0-indexed
}

// walkPath 按循环路径走 remainder 步，返回 1-indexed 爻位
func walkPath(steps int) int {
    // 路径序列: 1,2,3,4,5,6,6,5,4,3,2,1
    cycle := []int{1, 2, 3, 4, 5, 6, 6, 5, 4, 3, 2, 1}
    idx := (steps - 1) % len(cycle)
    return cycle[idx]
}
```

- [ ] **步骤 4：运行测试验证通过**

```bash
go test ./internal/engine/ -run "TestBuild|TestChanging|TestFifty" -v
# 预期: PASS
```

- [ ] **步骤 5：Commit**

```bash
git add internal/engine/hexagram.go internal/engine/hexagram_test.go
git commit -m "feat: 本卦/变卦构建引擎 — 6/9变爻+55法定主变爻"
```

---

### 任务 5：千问 API 客户端

**文件：**
- 创建：`internal/qianwen/client.go`
- 测试：`internal/qianwen/client_test.go`

- [ ] **步骤 1：编写测试 — API 请求/响应结构**

```go
func TestBuildPrompt(t *testing.T) {
    prompt := BuildPrompt("我该换工作吗？", "乾为天", "天风姤", "上爻(第6爻)")
    if !contains(prompt, "我该换工作吗") {
        t.Error("prompt missing question")
    }
    if !contains(prompt, "乾为天") {
        t.Error("prompt missing primary")
    }
    if !contains(prompt, "天风姤") {
        t.Error("prompt missing changing")
    }
}

func TestClientConfig(t *testing.T) {
    c := NewClient("test-key", "qwen-plus", "https://example.com")
    if c.apiKey != "test-key" {
        t.Error("api key not set")
    }
}
```

- [ ] **步骤 2：实现千问客户端**

```go
package qianwen

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    apiKey   string
    model    string
    endpoint string
    client   *http.Client
}

func NewClient(apiKey, model, endpoint string) *Client {
    return &Client{
        apiKey:   apiKey,
        model:    model,
        endpoint: endpoint,
        client:   &http.Client{Timeout: 30 * time.Second},
    }
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

// BuildPrompt 构建解卦 prompt
func BuildPrompt(question, primary, changing string, yaoPositions string) string {
    return fmt.Sprintf(
        "请以专业易经解卦角度，结合用户问题「%s」和卦象进行分析：\n"+
        "本卦：%s，变卦：%s，变爻：%s\n\n"+
        "请按以下结构给出解读：\n"+
        "1. 本卦解义\n"+
        "2. 变爻启示\n"+
        "3. 变卦趋势\n"+
        "4. 综合建议\n\n"+
        "请用流畅易懂的中文，避免过于玄奥的术语堆砌。",
        question, primary, changing, yaoPositions,
    )
}

// Divine 调用千问 API 解卦
func (c *Client) Divine(prompt string) (string, error) {
    reqBody := ChatRequest{
        Model: c.model,
        Messages: []Message{
            {Role: "user", Content: prompt},
        },
    }
    body, _ := json.Marshal(reqBody)

    req, _ := http.NewRequest("POST", c.endpoint, bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("千问API请求失败: %w", err)
    }
    defer resp.Body.Close()

    var cr ChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
        return "", fmt.Errorf("千问API响应解析失败: %w", err)
    }
    if len(cr.Choices) == 0 {
        return "", fmt.Errorf("千问API无有效响应")
    }
    return cr.Choices[0].Message.Content, nil
}
```

- [ ] **步骤 3：运行测试**

```bash
go test ./internal/qianwen/ -v
# 预期: 单元测试 PASS（不测试真实 API 调用）
```

- [ ] **步骤 4：Commit**

```bash
git add internal/qianwen/
git commit -m "feat: 千问API客户端 — prompt构建+解卦调用"
```

---

### 任务 6：HTTP 处理器 + HTML 模板

**文件：**
- 创建：`internal/handler/home.go`、`internal/handler/divine.go`
- 创建：`templates/layout.html`、`templates/home.html`、`templates/result.html`、`templates/partials/hexagram.html`
- 修改：`main.go`

- [ ] **步骤 1：创建基础布局模板 layout.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>易观 · AI 算卦</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@2.0.0"></script>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body class="bg-stone-50 min-h-screen">
    <header class="bg-red-900 text-amber-100 py-4 shadow-lg">
        <div class="max-w-2xl mx-auto px-4 flex items-center justify-between">
            <a href="/" class="text-2xl font-bold tracking-wider">☯ 易观</a>
            <span class="text-sm opacity-75">AI 在线算卦</span>
        </div>
    </header>
    <main class="max-w-2xl mx-auto px-4 py-8">
        {{block "content" .}}{{end}}
    </main>
    <footer class="text-center text-stone-400 text-sm py-8">
        <p>易观 · AI 算卦 | 仅供娱乐参考</p>
    </footer>
</body>
</html>
```

- [ ] **步骤 2：创建首页模板 home.html**

```html
{{define "content"}}
<div class="text-center mb-8">
    <h2 class="text-3xl font-bold text-stone-800 mb-2">心有疑虑，问卦于天</h2>
    <p class="text-stone-500">默想你的问题，诚心求问，AI 为你解卦</p>
</div>

<!-- 广告位 -->
<div class="bg-stone-200 rounded-lg h-16 mb-8 flex items-center justify-center text-stone-400 text-sm">
    广告位 (Banner)
</div>

<form hx-post="/divine" hx-target="#result" hx-swap="innerHTML"
      class="bg-white rounded-xl shadow-md p-6">
    <label class="block text-stone-700 font-medium mb-2">请输入你想问的问题：</label>
    <textarea name="question" rows="3" required
              placeholder="例如：我该不该换工作？这段感情能长久吗？..."
              class="w-full border border-stone-300 rounded-lg p-3 focus:ring-2 focus:ring-red-700 focus:border-transparent resize-none"
              hx-indicator="#loading"></textarea>
    <div class="mt-4 flex items-center justify-between">
        <span id="loading" class="htmx-indicator text-stone-400 text-sm">
            ⏳ 诚心起卦中...
        </span>
        <button type="submit"
                class="bg-red-800 text-amber-100 px-8 py-3 rounded-lg font-medium hover:bg-red-700 transition">
            ☯ 开始提问
        </button>
    </div>
</form>

<div id="result" class="mt-8"></div>
{{end}}
```

- [ ] **步骤 3：创建卦象组件 hexagram.html**

```html
{{define "hexagram"}}
<div class="flex flex-col items-center py-4">
    {{range $i, $v := .Lines}}
    <div class="flex items-center gap-4 my-1">
        <span class="text-xs text-stone-400 w-8 text-right">{{$i | yaoName}}</span>
        <div class="flex gap-1">
            {{if $v.IsYang}}
            <span class="block w-16 h-1.5 bg-red-800 rounded"></span>
            {{else}}
            <span class="flex gap-1">
                <span class="block w-7 h-1.5 bg-stone-700 rounded"></span>
                <span class="block w-7 h-1.5 bg-stone-700 rounded"></span>
            </span>
            {{end}}
        </div>
        {{if $v.Changing}}
        <span class="text-red-600 text-xs font-bold">○ 变</span>
        {{end}}
    </div>
    {{end}}
    <div class="mt-2 text-stone-500 text-sm">{{.GuaName}} {{.GuaSymbol}}</div>
</div>
{{end}}
```

- [ ] **步骤 4：创建结果页模板 result.html**

```html
{{define "result"}}
<div class="bg-white rounded-xl shadow-md p-6">
    <h3 class="text-xl font-bold text-stone-800 mb-4 text-center">📜 卦象结果</h3>

    <div class="grid grid-cols-2 gap-6 mb-6">
        <div class="bg-amber-50 rounded-lg p-4 text-center">
            <span class="text-sm text-stone-500">本卦</span>
            <div class="text-2xl font-bold text-red-900 mt-1">{{.Primary.Name}}</div>
            <div class="text-3xl">{{.Primary.Symbol}}</div>
            <p class="text-xs text-stone-400 mt-1">{{.Primary.GuaCi}}</p>
        </div>
        <div class="bg-stone-50 rounded-lg p-4 text-center">
            <span class="text-sm text-stone-500">变卦</span>
            <div class="text-2xl font-bold text-stone-800 mt-1">{{.Changing.Name}}</div>
            <div class="text-3xl">{{.Changing.Symbol}}</div>
            <p class="text-xs text-stone-400 mt-1">{{.Changing.GuaCi}}</p>
        </div>
    </div>

    <!-- 变爻标注 -->
    <div class="text-center mb-6">
        <span class="text-sm text-stone-500">变爻：</span>
        {{range .ChangingPositions}}
        <span class="inline-block bg-red-100 text-red-700 px-2 py-0.5 rounded text-sm mx-0.5">
            {{yaoLabel .}}
        </span>
        {{end}}
        {{if ne .MasterYao -1}}
        <span class="inline-block bg-red-700 text-amber-100 px-2 py-0.5 rounded text-sm mx-0.5" title="主变爻">
            {{yaoLabel .MasterYao}} ★主
        </span>
        {{end}}
    </div>

    <!-- AI解卦 -->
    <div class="border-t border-stone-200 pt-6">
        <h4 class="text-lg font-medium text-stone-700 mb-3 flex items-center gap-2">
            <span>🤖</span> AI 解卦
        </h4>
        <div class="prose prose-stone max-w-none text-stone-700 leading-relaxed whitespace-pre-wrap">
            {{.Interpretation}}
        </div>
    </div>

    <!-- 广告位 -->
    <div class="bg-stone-200 rounded-lg h-16 my-6 flex items-center justify-center text-stone-400 text-sm">
        广告位 (结果页)
    </div>

    <!-- 大师咨询 -->
    <div class="border-t border-stone-200 pt-6 text-center">
        <p class="text-stone-500 mb-3">想获得更深入的解读吗？</p>
        <button onclick="toggleQR()"
                class="bg-amber-600 text-white px-8 py-3 rounded-lg font-medium hover:bg-amber-500 transition">
            🔮 周易大师一对一详解
        </button>
        <div id="qr-code" class="hidden mt-4">
            <img src="/static/qr-placeholder.png" alt="大师微信" class="mx-auto w-48 h-48">
            <p class="text-stone-400 text-sm mt-2">扫码添加周易大师微信</p>
        </div>
    </div>

    <!-- 分享按钮 -->
    <div class="text-center mt-4">
        <button onclick="share()"
                class="text-stone-400 hover:text-stone-600 text-sm underline">
            📤 分享结果给朋友
        </button>
    </div>
</div>

<script>
function toggleQR() {
    document.getElementById('qr-code').classList.toggle('hidden');
}
function share() {
    if (navigator.share) {
        navigator.share({title: '易观AI算卦', url: window.location.href});
    } else {
        alert('请复制链接分享给朋友');
    }
}
</script>
{{end}}
```

- [ ] **步骤 5：创建首页处理器 home.go**

```go
package handler

import (
    "html/template"
    "net/http"
)

type HomeHandler struct {
    Tmpl *template.Template
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.Tmpl.ExecuteTemplate(w, "layout.html", nil)
}
```

- [ ] **步骤 6：创建算卦处理器 divine.go**

```go
package handler

import (
    "html/template"
    "net/http"

    "github.com/kiddyt00/yiguan/internal/engine"
    "github.com/kiddyt00/yiguan/internal/qianwen"
)

type DivineHandler struct {
    Tmpl    *template.Template
    Qianwen *qianwen.Client
}

func (h *DivineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
    yaoDesc := formatYaoPositions(positions)
    if master >= 0 {
        yaoDesc += fmt.Sprintf("（主变爻：%s）", yaoLabel(master))
    }

    // 4. 调用千问解卦
    prompt := qianwen.BuildPrompt(question, primary.Name, changing.Name, yaoDesc)
    interpretation, err := h.Qianwen.Divine(prompt)
    if err != nil {
        interpretation = fmt.Sprintf("解卦服务暂不可用：%v", err)
    }

    // 5. 渲染结果
    data := map[string]interface{}{
        "Primary":           primary,
        "Changing":          changing,
        "ChangingPositions": positions,
        "MasterYao":         master,
        "Interpretation":    interpretation,
        "Lines":             buildLineDisplay(lines, positions, master),
    }
    h.Tmpl.ExecuteTemplate(w, "result", data)
}

func formatYaoPositions(positions []int) string {
    // 将0-based转为中文爻名
}

func yaoLabel(pos int) string {
    names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
    return names[pos]
}

func buildLineDisplay(lines [6]int, changing []int, master int) []LineDisplay {
    // 构建前端展示用的六爻数据
}
```

- [ ] **步骤 7：更新 main.go 整合路由和模板**

```go
func main() {
    cfg := loadConfig("config.yaml")
    if key := os.Getenv("DASHSCOPE_API_KEY"); key != "" {
        cfg.Qianwen.APIKey = key
    }

    // 模板函数映射
    funcMap := template.FuncMap{
        "yaoName":  yaoNameFunc,
        "yaoLabel": yaoLabelFunc,
    }

    // 解析模板
    tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/**/*.html"))

    // 千问客户端
    qw := qianwen.NewClient(cfg.Qianwen.APIKey, cfg.Qianwen.Model, cfg.Qianwen.Endpoint)

    mux := http.NewServeMux()

    // 静态文件
    mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // 页面路由
    home := &handler.HomeHandler{Tmpl: tmpl}
    divine := &handler.DivineHandler{Tmpl: tmpl, Qianwen: qw}

    mux.Handle("GET /", home)
    mux.Handle("POST /divine", divine)

    log.Printf("☯ 易观服务启动 http://localhost:%s", cfg.Server.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}
```

- [ ] **步骤 8：Commit**

```bash
git add -A
git commit -m "feat: HTTP处理器+HTML模板 — 首页/算卦/结果全流程"
```

---

### 任务 7：样式美化与动画

**文件：**
- 创建：`static/style.css`

- [ ] **步骤 1：编写 CSS 样式**

```css
/* 卦象动画 */
@keyframes fadeInUp {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
}

#result {
    animation: fadeInUp 0.5s ease-out;
}

/* HTMX 加载指示器 */
.htmx-indicator {
    opacity: 0;
    transition: opacity 0.2s ease-in;
}
.htmx-request .htmx-indicator,
.htmx-indicator.htmx-request {
    opacity: 1;
}

/* 二维码弹窗动画 */
#qr-code {
    transition: all 0.3s ease;
}
#qr-code:not(.hidden) {
    animation: fadeInUp 0.3s ease-out;
}

/* 卦象爻线悬浮效果 */
.hexagram-line:hover {
    transform: scaleX(1.02);
    transition: transform 0.2s ease;
}

/* 按钮波纹 */
button[type="submit"]:active {
    transform: scale(0.98);
}
```

- [ ] **步骤 2：Commit**

```bash
git add static/style.css
git commit -m "feat: CSS样式 — 卦象动画+HTMX加载指示器"
```

---

### 任务 8：端到端集成测试

**文件：**
- 创建：`test/e2e_test.go`

- [ ] **步骤 1：编写 E2E 测试**

```go
func TestHomePage(t *testing.T) {
    ts := newTestServer(t)
    defer ts.Close()

    resp, err := http.Get(ts.URL + "/")
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != 200 {
        t.Errorf("home page returned %d", resp.StatusCode)
    }
    body, _ := io.ReadAll(resp.Body)
    if !contains(string(body), "易观") {
        t.Error("home page missing 易观 title")
    }
}

func TestDivineWithoutQuestion(t *testing.T) {
    ts := newTestServer(t)
    defer ts.Close()

    resp, _ := http.PostForm(ts.URL+"/divine", url.Values{})
    if resp.StatusCode != 400 {
        t.Errorf("expected 400, got %d", resp.StatusCode)
    }
}

func TestDivineWithQuestion(t *testing.T) {
    if os.Getenv("DASHSCOPE_API_KEY") == "" {
        t.Skip("DASHSCOPE_API_KEY not set, skipping integration test")
    }
    ts := newTestServer(t)
    defer ts.Close()

    resp, _ := http.PostForm(ts.URL+"/divine", url.Values{
        "question": {"测试问题"},
    })
    if resp.StatusCode != 200 {
        t.Errorf("divine returned %d", resp.StatusCode)
    }
}
```

- [ ] **步骤 2：验证完整流程**

```bash
# 启动服务
DASHSCOPE_API_KEY=your_key go run main.go &

# 验证首页
curl -s localhost:8080 | grep "易观"

# 验证算卦（无千问key时会显示错误但页面正常）
curl -s -X POST localhost:8080/divine -d "question=测试" | grep "卦象结果"
```

- [ ] **步骤 3：Commit**

```bash
git add test/
git commit -m "test: E2E集成测试 — 首页+算卦全流程"
```

---

## 环境变量

| 变量 | 说明 | 必需 |
|------|------|------|
| `DASHSCOPE_API_KEY` | 阿里云 DashScope API Key | 是（用于千问解卦） |

## 启动方式

```bash
cd /home/kiddyt00/claude-projects/yiguan
export DASHSCOPE_API_KEY="your-key"
go run main.go
# 访问 http://localhost:8080
```
