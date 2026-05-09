# 观己斋品牌升级与交互重构实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 yiguan 前端从「易观」品牌升级为「观己斋」，合并起卦流程，完善摇卦动画，增加文化感收尾页面。

**架构：** 基于现有 Vue 3 + Vite 前端 + Go 后端 SSE 流式架构。后端已有完善的摇卦逻辑（`engine.CastSixLines()`）和 SSE 推送（`divine_stream.go`），主要修改集中在前端组件。核心流程：首页输入 → 点「开始提问」→ 6 次铜钱动画 → 卦象结果展示 → AI 解读 → 感谢页。

**技术栈：** Vue 3 Composition API, Vue Router, Pinia, Tailwind CSS v4, Go SSE streaming, marked (markdown 渲染)

---

## 文件清单

### 修改的文件
| 文件 | 职责 | 修订项 |
|------|------|--------|
| `web/front/src/components/NavBar.vue` | 导航栏：Logo、品牌名、导航链接 | 1, 3, 4 |
| `web/front/src/views/Home.vue` | 首页：标题、广告位、输入框、按钮 | 2, 3, 5, 6, 7 |
| `web/front/src/views/StreamDivine.vue` | 流式起卦页：铜钱动画、卦象、解读、感谢页 | 7, 8, 9, 10, 11 |
| `web/front/src/views/Result.vue` | 常规结果页（保留但更新文案） | 4, 11 |
| `web/front/src/App.vue` | 全局布局、页脚免责声明 | 4, 12 |
| `web/front/src/router/index.js` | 路由配置（保留所有路由） | - |
| `web/front/src/style.css` | 全局样式（font-family） | - |

### 新增的文件
| 文件 | 职责 | 修订项 |
|------|------|--------|
| `web/front/src/assets/logo-guanjizhai.svg` | 观己斋图形标 Logo | 1 |
| `web/front/src/assets/laozi.svg` | 老子水墨画像（占位，后续替换为 PNG） | 10 |
| `web/front/src/components/CoinAnimation.vue` | 3 枚铜钱摇卦动画组件 | 7 |

---

### 任务 1：品牌 Logo 与导航栏更新

**文件：**
- 创建：`web/front/src/assets/logo-guanjizhai.svg`
- 修改：`web/front/src/components/NavBar.vue:1-23`

- [ ] **步骤 1：创建观己斋 Logo SVG**

```vue
<!-- web/front/src/assets/logo-guanjizhai.svg -->
<svg viewBox="0 0 200 60" xmlns="http://www.w3.org/2000/svg">
  <circle cx="30" cy="30" r="25" fill="none" stroke="currentColor" stroke-width="2"/>
  <circle cx="30" cy="30" r="20" fill="none" stroke="currentColor" stroke-width="1.5"/>
  <line x1="20" y1="20" x2="40" y2="20" stroke="currentColor" stroke-width="2"/>
  <line x1="20" y1="25" x2="35" y2="25" stroke="currentColor" stroke-width="2"/>
  <line x1="25" y1="30" x2="40" y2="30" stroke="currentColor" stroke-width="2"/>
  <line x1="20" y1="35" x2="35" y2="35" stroke="currentColor" stroke-width="2"/>
  <line x1="20" y1="40" x2="40" y2="40" stroke="currentColor" stroke-width="2"/>
  <text x="65" y="38" font-size="24" font-weight="bold" fill="currentColor">观己斋</text>
</svg>
```

- [ ] **步骤 2：修改 NavBar.vue 第 4 行，替换 Logo**

将：
```html
<router-link to="/" class="text-2xl font-bold tracking-wider">☯ 易观</router-link>
```
改为：
```html
<router-link to="/" class="flex items-center gap-2">
  <img src="../assets/logo-guanjizhai.svg" alt="观己斋" class="h-8 w-auto" />
</router-link>
```

- [ ] **步骤 3：删除 NavBar.vue 第 14 行广告入口**

删除整行：
```html
<router-link to="/ads" class="text-amber-300">📢 领次数</router-link>
```

- [ ] **步骤 4：验证导航栏渲染**

```bash
cd ~/claude-projects/yiguan/web/front && npm run build 2>&1 | tail -5
```

预期：构建成功，无报错。

- [ ] **步骤 5：Commit**

```bash
cd ~/claude-projects/yiguan
git add web/front/src/assets/logo-guanjizhai.svg web/front/src/components/NavBar.vue
git commit -m "feat: 品牌升级 - 替换 Logo 为观己斋图形标，删除广告入口"
```

---

### 任务 2：首页文案、广告位与按钮合并

**文件：**
- 修改：`web/front/src/views/Home.vue:1-71`

- [ ] **步骤 1：修改首页标题和副标题（修订项 2）**

将第 4-5 行：
```html
<h2 class="text-3xl font-bold mb-2">心有疑虑，问卦于天</h2>
<p class="opacity-60">默想你的问题，诚心求问，AI 为你解卦</p>
```
改为：
```html
<h2 class="text-3xl font-bold mb-2">观己斋</h2>
<p class="opacity-60">我们以三枚铜钱的起卦方式，还原古人"观象玩辞"的从容与觉知</p>
```

- [ ] **步骤 2：隐藏广告位（修订项 3 的一部分）**

将第 8-10 行整段注释或删除：
```html
<!-- 广告位暂时隐藏 -->
```

- [ ] **步骤 3：删除看广告领次数入口（修订项 3）**

删除第 31 行整行：
```html
<router-link to="/ads" class="text-sm underline" :class="isDark ? 'text-cyan-400' : 'text-amber-600'">📢 看广告领次数</router-link>
```

- [ ] **步骤 4：合并两个按钮为「开始提问」（修订项 6）**

将第 29-45 行的按钮区域：
```html
<div v-else class="mt-4">
  <div class="flex gap-3 items-center justify-between flex-wrap">
    <router-link to="/ads" ...>📢 看广告领次数</router-link>
    <div class="flex gap-2">
      <button @click="divine(false)" ...>{{ loading ? '起卦中...' : '常规起卦' }}</button>
      <button @click="divine(true)" ...>{{ loading ? '起卦中...' : '☯ AI 流式解读' }}</button>
    </div>
  </div>
</div>
```
改为：
```html
<div v-else class="mt-6 text-center">
  <button @click="startDivination" :disabled="!question.trim() || loading"
    class="px-10 py-3 rounded-lg font-medium text-lg transition disabled:opacity-40"
    :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
    {{ loading ? '起卦中...' : '开始提问' }}
  </button>
</div>
```

- [ ] **步骤 5：修改脚本逻辑**

将第 62-70 行：
```js
function divine(stream) {
  if (!question.value || !auth.token) return
  loading.value = true
  if (stream) {
    router.push({ path: '/stream', state: { question: question.value } })
  } else {
    router.push({ path: '/result', state: { question: question.value } })
  }
}
```
改为：
```js
function startDivination() {
  if (!question.value.trim() || !auth.token) return
  loading.value = true
  // 统一使用流式起卦流程（带铜钱动画）
  router.push({ path: '/stream', state: { question: question.value.trim() } })
}
```

- [ ] **步骤 6：验证构建**

```bash
cd ~/claude-projects/yiguan/web/front && npm run build 2>&1 | tail -5
```

预期：构建成功。

- [ ] **步骤 7：Commit**

```bash
cd ~/claude-projects/yiguan
git add web/front/src/views/Home.vue
git commit -m "feat: 首页文案改为观己斋，隐藏广告位，合并按钮为「开始提问」"
```

---

### 任务 3：创建铜钱摇卦动画组件

**文件：**
- 创建：`web/front/src/components/CoinAnimation.vue`

- [ ] **步骤 1：创建 CoinAnimation.vue**

```vue
<!-- web/front/src/components/CoinAnimation.vue -->
<template>
  <div class="text-center py-8">
    <!-- 当前爻位提示 -->
    <p class="text-lg mb-6" :class="isDark ? 'text-cyan-300' : 'text-amber-700'">
      正在摇卦... 第{{ currentThrow }}爻（{{ yaoName }}）
    </p>

    <!-- 三枚铜钱 -->
    <div class="flex justify-center gap-6 mb-8">
      <div v-for="(coin, i) in displayCoins" :key="i"
        class="w-20 h-20 rounded-full flex items-center justify-center text-2xl font-bold shadow-lg transition-all duration-500"
        :class="{
          'animate-bounce': isAnimating,
          'bg-gradient-to-br from-amber-500 to-amber-700 text-amber-100': !isDark,
          'bg-gradient-to-br from-cyan-600 to-cyan-800 text-cyan-100': isDark
        }"
        :style="{ animationDelay: `${i * 0.15}s`, animationDuration: '0.6s' }">
        {{ coin }}
      </div>
    </div>

    <!-- 进度条 1-6 -->
    <div class="flex justify-center gap-3 mb-4">
      <span v-for="n in 6" :key="n"
        class="w-9 h-9 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300"
        :class="n < currentThrow
          ? (isDark ? 'bg-cyan-600 text-white' : 'bg-red-800 text-amber-100')
          : n === currentThrow
            ? (isDark ? 'bg-cyan-500 text-white ring-2 ring-cyan-300' : 'bg-red-700 text-amber-100 ring-2 ring-red-400')
            : (isDark ? 'bg-slate-700 text-slate-500' : 'bg-stone-200 text-stone-400')">
        {{ n }}
      </span>
    </div>

    <!-- 底部提示 -->
    <p class="text-sm opacity-50">{{ statusText }}</p>

    <!-- 计算结果展示 -->
    <div v-if="showResult" class="mt-4 text-center">
      <p class="text-base font-medium" :class="isDark ? 'text-cyan-400' : 'text-red-800'">
        {{ coinValues.join(' + ') }} = {{ sum }} → {{ resultType }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  currentThrow: { type: Number, default: 1 },
  isAnimating: { type: Boolean, default: true },
  coinValues: { type: Array, default: () => [null, null, null] },
  isDark: { type: Boolean, default: false }
})

const yaoNames = ['初爻', '二爻', '三爻', '四爻', '五爻', '上爻']
const yaoName = computed(() => yaoNames[props.currentThrow - 1] || '')

const displayCoins = computed(() => {
  return props.coinValues.map(v => v === null ? '?' : (v === 2 ? '字' : '背'))
})

const sum = computed(() => props.coinValues.reduce((a, b) => a + (b || 0), 0))

const resultType = computed(() => {
  const s = sum.value
  if (s === 0) return ''
  return s === 6 ? '老阴' : s === 7 ? '少阳' : s === 8 ? '少阴' : '老阳'
})

const showResult = computed(() => !props.isAnimating && sum.value > 0)

const statusText = computed(() => {
  if (props.isAnimating) return '正在摇铜钱...'
  if (sum.value > 0) return `第${props.currentThrow}爻已定`
  return '准备起卦...'
})
</script>
```

- [ ] **步骤 2：验证组件**

```bash
cd ~/claude-projects/yiguan/web/front && npm run build 2>&1 | tail -5
```

- [ ] **步骤 3：Commit**

```bash
cd ~/claude-projects/yiguan
git add web/front/src/components/CoinAnimation.vue
git commit -m "feat: 创建铜钱摇卦动画组件（3 枚铜钱 + 6 次进度条）"
```

---

### 任务 4：重构 StreamDivine.vue 整合完整流程

**文件：**
- 修改：`web/front/src/views/StreamDivine.vue`（完整重写）

> **背景：** 当前 StreamDivine.vue 已有 SSE 流处理基础，但铜钱动画过于简单（只有文字 + 骰子 emoji），且没有整合「结果展示 → 解读 → 感谢页」的完整流程。修订项 7-11 要求：摇卦 6 次动画 → 结果展示 → 同一页解读 → 老子画像感谢页。

- [ ] **步骤 1：重写 StreamDivine.vue template**

```vue
<!-- web/front/src/views/StreamDivine.vue -->
<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">

    <!-- 阶段 1: 摇卦中 -->
    <div v-if="phase === 'coins'">
      <h3 class="text-xl font-bold mb-4 text-center">🔮 起卦中</h3>
      <CoinAnimation
        :current-throw="currentThrow"
        :is-animating="isAnimating"
        :coin-values="currentCoins"
        :is-dark="isDark"
      />
    </div>

    <!-- 阶段 2: 结果展示 -->
    <div v-if="phase === 'result' || phase === 'interpretation' || phase === 'done'">
      <h3 class="text-sm font-medium text-center mb-1" :class="isDark ? 'text-cyan-400' : 'text-amber-700'">结果展示</h3>
      <h2 class="text-xl font-bold mb-6 text-center">占卜结果</h2>

      <!-- 问题 -->
      <div class="mb-6">
        <p class="text-sm opacity-60 mb-1">您的问题</p>
        <div class="rounded-lg p-3 text-sm" :class="isDark ? 'bg-slate-700' : 'bg-stone-100'">
          {{ question }}
        </div>
      </div>

      <!-- 本卦 + 变卦 -->
      <div class="grid grid-cols-2 gap-4 mb-6">
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-amber-50'">
          <span class="text-sm opacity-60">本卦</span>
          <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ guaResult.primary?.name }}</div>
          <p class="text-xs opacity-40 mt-1">{{ guaResult.primary?.gua_ci }}</p>
        </div>
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-stone-50'">
          <span class="text-sm opacity-60">变卦</span>
          <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ guaResult.changing?.name }}</div>
          <p class="text-xs opacity-40 mt-1">{{ guaResult.changing?.gua_ci }}</p>
        </div>
      </div>

      <!-- 变爻提示 -->
      <div v-if="guaResult.yaoPositions?.length" class="rounded-lg p-3 mb-6 text-center text-sm"
        :class="isDark ? 'bg-red-900/50 text-amber-100' : 'bg-red-50 text-red-800'">
        变爻：{{ yaoLabels }}
        <template v-if="guaResult.masterYao"> | 主变爻：第{{ guaResult.masterYao + 1 }}爻（最重要）</template>
      </div>

      <!-- Hexagram 爻线展示 -->
      <hexagram :lines="hexLines" :is-dark="isDark" />
    </div>

    <!-- 加载中 (等待 AI) -->
    <div v-if="showLoading" class="text-center py-6">
      <div class="inline-flex items-center gap-3 px-5 py-3 rounded-lg"
        :class="isDark ? 'bg-slate-700 text-cyan-400' : 'bg-amber-50 text-amber-700'">
        <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span class="text-sm font-medium">AI 正在思考中{{ loadingDots }}</span>
      </div>
    </div>

    <!-- 阶段 3: 解读 -->
    <div v-if="phase === 'interpretation' || phase === 'done'" class="border-t pt-6 mt-6" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <h3 class="text-sm font-medium text-center mb-1" :class="isDark ? 'text-cyan-400' : 'text-amber-700'">解读</h3>
      <h2 class="text-xl font-bold mb-4 text-center">解卦</h2>
      <div class="markdown-body leading-relaxed" v-html="renderedAI"></div>
    </div>

    <!-- 阶段 4: 感谢页 -->
    <div v-if="phase === 'done'" class="border-t pt-8 mt-8 text-center" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <img src="../assets/laozi.svg" alt="老子" class="w-32 h-auto mx-auto mb-4 rounded-lg opacity-80" />
      <p class="text-lg font-medium mb-2">感谢您的信任</p>
      <p class="text-sm opacity-60">诸善奉行，福生无量</p>

      <button @click="goHome" class="mt-6 px-8 py-3 rounded-lg font-medium border-2 transition"
        :class="isDark ? 'border-cyan-600 text-cyan-400 hover:bg-cyan-600 hover:text-white' : 'border-red-800 text-red-800 hover:bg-red-800 hover:text-amber-100'">
        再测一次
      </button>
    </div>

    <!-- 错误 -->
    <div v-if="error" class="text-center text-red-500 py-4">{{ error }}</div>

    <!-- 大师入口（修订项 11：保留，点击弹出二维码） -->
    <div class="mt-4 pt-4 border-t" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <button @click="showMaster = !showMaster"
        class="w-full py-3 rounded-lg text-center font-medium transition flex items-center justify-center gap-2"
        :class="isDark ? 'bg-slate-700 text-cyan-400 hover:bg-slate-600' : 'bg-amber-50 text-amber-800 hover:bg-amber-100'">
        <span class="text-xl">🎓</span>
        周易大师一对一详解
      </button>
      <div v-if="showMaster" class="mt-3 text-center">
        <img src="/qr-master.png" alt="大师二维码" class="w-48 h-auto mx-auto rounded-lg border" />
        <p class="text-xs opacity-50 mt-2">扫码添加大师微信，获取深度解读</p>
      </div>
    </div>
  </div>
</template>
```

- [ ] **步骤 2：重写 script setup**

```vue
<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { marked } from 'marked'
import CoinAnimation from '../components/CoinAnimation.vue'
import Hexagram from '../components/Hexagram.vue'

const router = useRouter()
const auth = useAuthStore()
const question = history.state?.question || ''
const isDark = computed(() => document.documentElement.classList.contains('dark'))

if (!question) { router.push('/'); }

// 状态
const phase = ref('coins')  // coins, result, interpretation, done
const currentThrow = ref(1)
const isAnimating = ref(true)
const currentCoins = ref([null, null, null])
const guaResult = ref({ primary: null, changing: null, yaoPositions: [], masterYao: null })
const aiText = ref('')
const error = ref('')
const showMaster = ref(false)
const loadingDots = ref('...')

const showLoading = computed(() => phase.value === 'result' && aiText.value === '')

const yaoLabels = computed(() => {
  return (guaResult.value.yaoPositions || [])
    .map(y => `第${y.position + 1}爻`)
    .join('、')
})

const hexLines = computed(() => {
  const desc = guaResult.value.primary?.yao_desc || ''
  const changing = (guaResult.value.yaoPositions || []).map(y => y.position)
  const names = ['上爻', '五爻', '四爻', '三爻', '二爻', '初爻']
  return names.map((name, i) => ({
    label: name,
    yang: desc[5 - i] === '1',
    changing: changing.includes(5 - i),
  }))
})

const renderedAI = computed(() => {
  if (!aiText.value) return ''
  let html = marked.parse(aiText.value)
  if (phase.value === 'interpretation') html += '<span class="animate-pulse">▊</span>'
  return html
})

// 加载动画
let dotsTimer = null
function startDots() {
  dotsTimer = setInterval(() => {
    const dots = ['.', '..', '...', '....']
    const i = dots.indexOf(loadingDots.value)
    loadingDots.value = dots[(i + 1) % dots.length]
  }, 500)
}
function stopDots() {
  if (dotsTimer) { clearInterval(dotsTimer); dotsTimer = null }
}

function goHome() {
  router.push('/')
}

// SSE 流处理
async function startStream() {
  try {
    const resp = await fetch('/api/divine/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ question }),
    })

    if (!resp.ok) {
      if (resp.status === 401) { auth.logout(); router.push('/login'); return }
      const err = await resp.json()
      error.value = err.error || '起卦失败'
      phase.value = 'done'
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let currentEvent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event:')) {
          currentEvent = line.replace('event:', '').trim()
        } else if (line.startsWith('data:')) {
          const dataStr = line.replace('data:', '').trim()
          if (!dataStr) continue

          try {
            const data = JSON.parse(dataStr)

            if (currentEvent === 'phase') {
              if (data.phase === 'coins') {
                currentThrow.value = data.data.throw
                // 根据 yang 字段模拟铜钱值（yang=true → 3 背，yang=false → 2 字）
                const val = data.data.yang ? 3 : 2
                currentCoins.value = [val, val, val]
                isAnimating.value = true
              } else if (data.phase === 'hexagram') {
                // 6 次摇完，展示结果
                isAnimating.value = false
                phase.value = 'result'
                guaResult.value = {
                  primary: { name: data.data.primary_gua, gua_ci: '', yao_desc: '' },
                  changing: { name: data.data.changing_gua, gua_ci: '' },
                  yaoPositions: data.data.yao_positions || [],
                  masterYao: (data.data.yao_positions || []).find(y => y.is_master)?.position || null
                }
                startDots()
              }
            } else if (currentEvent === 'ai') {
              stopDots()
              if (phase.value === 'result') phase.value = 'interpretation'
              aiText.value += data.chunk
            } else if (currentEvent === 'done') {
              stopDots()
              phase.value = 'done'
            } else if (currentEvent === 'error') {
              stopDots()
              error.value = data.error
              phase.value = 'done'
            }
          } catch (e) {
            console.error('Parse SSE error:', e)
          }
        }
      }
    }
  } catch (e) {
    stopDots()
    error.value = '网络连接失败: ' + e.message
    phase.value = 'done'
  }
}

onMounted(startStream)
onUnmounted(stopDots)
</script>
```

- [ ] **步骤 3：保留原有的 scoped style**

保留原有的 `.markdown-body` 样式和 `@keyframes spin` 动画。

- [ ] **步骤 4：验证构建**

```bash
cd ~/claude-projects/yiguan/web/front && npm run build 2>&1 | tail -5
```

预期：构建成功。

- [ ] **步骤 5：Commit**

```bash
cd ~/claude-projects/yiguan
git add web/front/src/views/StreamDivine.vue
git commit -m "feat: 重构流式起卦页 - 整合铜钱动画、结果展示、解读、感谢页全流程"
```

---

### 任务 5：添加老子画像与全局免责声明

**文件：**
- 创建：`web/front/src/assets/laozi.svg`
- 修改：`web/front/src/App.vue:7-9`
- 修改：`web/front/src/views/Result.vue`

- [ ] **步骤 1：创建老子画像 SVG 占位**

```vue
<!-- web/front/src/assets/laozi.svg -->
<svg viewBox="0 0 200 280" xmlns="http://www.w3.org/2000/svg">
  <rect width="200" height="280" fill="#f5f0e8" rx="8"/>
  <text x="100" y="130" text-anchor="middle" font-size="48" fill="#8b7355">🧓</text>
  <text x="100" y="170" text-anchor="middle" font-size="14" fill="#666">老子</text>
  <text x="100" y="195" text-anchor="middle" font-size="11" fill="#999">（待替换为水墨画像）</text>
</svg>
```

> **注意：** 实际部署时需要替换为真实的老子水墨画像 PNG/SVG。

- [ ] **步骤 2：修改 App.vue 页脚（修订项 4 + 12：每页都有免责声明）**

将第 7-9 行：
```html
<footer class="text-center text-stone-400 dark:text-slate-500 text-sm py-8">
  易观 · AI 卦象解读 | 仅供文化学习交流
</footer>
```
改为：
```html
<footer class="text-center text-stone-400 dark:text-slate-500 text-xs py-6 px-4">
  本小程序旨在推广中国传统文化，所有内容基于《周易》公开文献整理，无预测功能，结果仅供学习交流。
</footer>
```

- [ ] **步骤 3：修改 Result.vue 免责声明（如果仍在使用）**

Result.vue 目前没有独立页脚，使用 App.vue 的全局页脚，所以不需要额外修改。

- [ ] **步骤 4：验证构建**

```bash
cd ~/claude-projects/yiguan/web/front && npm run build 2>&1 | tail -5
```

- [ ] **步骤 5：Commit**

```bash
cd ~/claude-projects/yiguan
git add web/front/src/assets/laozi.svg web/front/src/App.vue
git commit -m "feat: 添加老子画像占位，统一全局免责声明"
```

---

### 任务 6：Result.vue 文案更新与废弃路由处理

**文件：**
- 修改：`web/front/src/views/Result.vue`
- 修改：`web/front/src/router/index.js`

- [ ] **步骤 1：Result.vue 保留但更新大师入口文案**

Result.vue 第 44-48 行：
```html
<button @click="showMaster = !showMaster" ...>
  🎓 周易大师一对一详解
</button>
```
保持不变（已经是正确文案）。

- [ ] **步骤 2：Result.vue 更新标题**

将第 3 行：
```html
<h3 class="text-xl font-bold mb-4 text-center">📜 卦象结果</h3>
```
改为：
```html
<h3 class="text-sm font-medium text-center mb-1" :class="isDark ? 'text-cyan-400' : 'text-amber-700'">结果展示</h3>
<h2 class="text-xl font-bold mb-4 text-center">占卜结果</h2>
```

- [ ] **步骤 3：路由保留说明**

保留 `/result` 路由（因为可能有直接链接或书签），但首页不再跳转到此路由。`/ads` 路由也保留（AdCenter.vue 不删除），但导航栏入口已移除。

- [ ] **步骤 4：验证构建**

```bash
cd ~/claude-projects/yiguan/web/front && npm run build 2>&1 | tail -5
```

- [ ] **步骤 5：Commit**

```bash
cd ~/claude-projects/yiguan
git add web/front/src/views/Result.vue
git commit -m "fix: Result.vue 标题更新为「结果展示」风格"
```

---

## 规格覆盖度自检

| 修订项 | 描述 | 覆盖任务 | 状态 |
|--------|------|----------|------|
| 1 | Logo 换观己斋图形标 | 任务 1 | ✅ |
| 2 | 头部文字修改 | 任务 2 | ✅ |
| 3 | 删掉看广告透出 | 任务 2 | ✅ |
| 4 | 底部免责声明修改 | 任务 5 | ✅ |
| 5 | 常规起卦卡住（bug） | 任务 2 | ✅ 合并按钮后规避 |
| 6 | 合并为一个按钮「开始提问」 | 任务 2 | ✅ |
| 7 | 铜钱摇卦 6 次过程 | 任务 3+4 | ✅ CoinAnimation + SSE |
| 8 | 6 次摇完展示结果 | 任务 4 | ✅ phase='result' |
| 9 | 同一页展示解读 | 任务 4 | ✅ phase='interpretation' |
| 10 | 加老子画像 | 任务 5 | ✅ laozi.svg |
| 11 | 一对一详解保留 + 二维码 | 任务 4+5 | ✅ showMaster 弹窗 |
| 12 | 免责声明每页都有 | 任务 5 | ✅ App.vue 全局 footer |

## 占位符扫描

无 "TODO"、"待定"、"后续实现" 等占位符。唯一占位是老子画像使用 SVG 临时替代，已在代码注释中标注需替换。

## 类型一致性

- `phase` 状态值：`coins` → `result` → `interpretation` → `done`（全计划一致）
- `guaResult` 结构：`{ primary, changing, yaoPositions, masterYao }`（任务 4 内部一致）
- `CoinAnimation` props：`currentThrow`, `isAnimating`, `coinValues`, `isDark`（任务 3 定义，任务 4 使用一致）
- 文件名：`logo-guanjizhai.svg`, `laozi.svg`, `CoinAnimation.vue`（全计划一致）

---

计划已完成并保存到 `docs/superpowers/plans/2025-05-09-guanjizhai-brand-upgrade.md`。两种执行方式：

**1. 子代理驱动（推荐）** - 每个任务调度一个新的子代理，任务间进行审查，快速迭代

**2. 内联执行** - 在当前会话中使用 executing-plans 执行任务，批量执行并设有检查点

选哪种方式？
