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

    <!-- 阶段 2: 结果展示 + 阶段 3: 解读 + 阶段 4: 感谢页 -->
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
          <div class="text-2xl my-1">{{ guaResult.primary?.symbol }}</div>
          <p v-if="guaResult.primary?.gua_ci" class="text-xs opacity-40 mt-1">{{ guaResult.primary.gua_ci }}</p>
        </div>
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-stone-50'">
          <span class="text-sm opacity-60">变卦</span>
          <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ guaResult.changing?.name }}</div>
          <div class="text-2xl my-1">{{ guaResult.changing?.symbol }}</div>
          <p v-if="guaResult.changing?.gua_ci" class="text-xs opacity-40 mt-1">{{ guaResult.changing.gua_ci }}</p>
        </div>
      </div>

      <!-- 变爻提示 -->
      <div v-if="guaResult.yaoPositions?.length" class="rounded-lg p-3 mb-6 text-center text-sm"
        :class="isDark ? 'bg-red-900/50 text-amber-100' : 'bg-red-50 text-red-800'">
        变爻：{{ yaoLabels }}
        <template v-if="guaResult.masterYao !== null"> | 主变爻：第{{ guaResult.masterYao + 1 }}爻（最重要）</template>
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

if (!question) { router.push('/') }

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
                  primary: {
                    name: data.data.primary_gua,
                    gua_ci: data.data.primary_gua_ci || '',
                    symbol: data.data.primary_symbol || '',
                    yao_desc: data.data.primary_yao_desc || ''
                  },
                  changing: {
                    name: data.data.changing_gua,
                    gua_ci: data.data.changing_gua_ci || '',
                    symbol: data.data.changing_symbol || ''
                  },
                  yaoPositions: data.data.yao_positions || [],
                  masterYao: (() => {
                    const mp = (data.data.yao_positions || []).find(y => y.is_master)
                    return mp ? mp.position : null
                  })()
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

<style scoped>
.markdown-body :deep(h2) { font-size: 1.1rem; font-weight: 600; margin-top: 1.25rem; margin-bottom: 0.5rem; padding-bottom: 0.25rem; border-bottom: 1px solid rgba(0,0,0,0.08); }
.markdown-body :deep(h3) { font-size: 1rem; font-weight: 600; margin-top: 1rem; margin-bottom: 0.4rem; }
.markdown-body :deep(p) { margin-bottom: 0.6rem; }
.markdown-body :deep(strong) { font-weight: 700; }
.markdown-body :deep(blockquote) { border-left: 3px solid rgba(0,0,0,0.15); padding-left: 0.75rem; opacity: 0.7; margin: 0.5rem 0; }
.markdown-body :deep(ul), .markdown-body :deep(ol) { padding-left: 1.25rem; margin-bottom: 0.6rem; }
</style>
