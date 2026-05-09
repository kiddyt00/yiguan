<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
    <h3 class="text-xl font-bold mb-4 text-center">🔮 {{ statusText }}</h3>

    <!-- 铜钱动画 -->
    <div v-if="phase === 'coins'" class="text-center py-8">
      <div class="text-4xl mb-4" :class="isDark ? 'text-cyan-400' : 'text-amber-600'">
        🪙 {{ coinsLabel }}
      </div>
      <div class="animate-bounce text-3xl">🎲</div>
    </div>

    <!-- 卦象展示 -->
    <div v-if="showHexagram" class="grid grid-cols-2 gap-6 mb-6">
      <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-amber-50'">
        <span class="text-sm opacity-60">本卦</span>
        <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ hexagram.primary_gua }}</div>
      </div>
      <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-stone-50'">
        <span class="text-sm opacity-60">变卦</span>
        <div class="text-2xl font-bold mt-1">{{ hexagram.changing_gua }}</div>
      </div>
    </div>

    <!-- 等待 AI 响应的加载动画 -->
    <div v-if="showLoading" class="text-center py-6">
      <div class="inline-flex items-center gap-3 px-5 py-3 rounded-lg"
        :class="isDark ? 'bg-slate-700 text-cyan-400' : 'bg-amber-50 text-amber-700'">
        <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span class="text-sm font-medium">AI 正在思考中{{ loadingDots }}</span>
      </div>
      <p v-if="statusMsg" class="text-xs mt-2 opacity-50">{{ statusMsg }}</p>
    </div>

    <!-- AI 解卦流式渲染 -->
    <div v-if="showAI" class="border-t pt-6 mt-6" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <h4 class="text-lg font-medium mb-3">🤖 AI 解卦</h4>
      <div class="markdown-body leading-relaxed" v-html="renderedAI"></div>
    </div>

    <!-- 错误 -->
    <div v-if="error" class="text-center text-red-500 py-4">{{ error }}</div>

    <!-- 完成按钮 -->
    <div v-if="phase === 'done'" class="mt-6 text-center">
      <router-link to="/"
        class="px-6 py-2 rounded-lg font-medium inline-block"
        :class="isDark ? 'bg-cyan-600 text-white' : 'bg-red-800 text-amber-100'">
        返回
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { marked } from 'marked'

const router = useRouter()
const auth = useAuthStore()
const props = defineProps(['isDark'])
const emit = defineEmits(['complete', 'error'])

const question = history.state?.question || ''
const token = auth.token

if (!question) {
  router.push('/')
}

const phase = ref('coins') // coins, hexagram, ai, done
const coinsLabel = ref('')
const hexagram = ref({ primary_gua: '', changing_gua: '' })
const aiText = ref('')
const error = ref('')
const recordData = ref(null)
const statusMsg = ref('')

// 加载动画的点（... → .... → ..... → ...）
const loadingDots = ref('...')
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

const showHexagram = computed(() => ['hexagram', 'ai', 'done'].includes(phase.value))
const showAI = computed(() => ['ai', 'done'].includes(phase.value))
const showLoading = computed(() => phase.value === 'hexagram')
const statusText = computed(() => {
  const map = { coins: '起卦中...', hexagram: '卦象已现', ai: 'AI 解卦中...', done: '解卦完成', error: '出错了' }
  return map[phase.value] || ''
})

marked.setOptions({ breaks: true, gfm: true })
const renderedAI = computed(() => {
  let html = marked.parse(aiText.value || '')
  if (phase.value === 'ai') html += '<span class="animate-pulse">▊</span>'
  return html
})

async function startStream() {
  try {
    const resp = await fetch('/api/divine/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({ question: question }),
    })

    if (!resp.ok) {
      if (resp.status === 401) {
        auth.logout()
        router.push('/login')
        return
      }
      const err = await resp.json()
      error.value = err.error || '起卦失败'
      phase.value = 'error'
      emit('error', err.error)
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      let currentEvent = ''
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
                phase.value = 'coins'
                coinsLabel.value = `${data.data.label} — ${data.data.result}`
              } else if (data.phase === 'hexagram') {
                phase.value = 'hexagram'
                hexagram.value = { primary_gua: data.data.primary_gua, changing_gua: data.data.changing_gua }
                startDots()
              }
            } else if (currentEvent === 'ai') {
              stopDots()
              if (phase.value === 'hexagram') phase.value = 'ai'
              aiText.value += data.chunk
            } else if (currentEvent === 'status') {
              statusMsg.value = data.msg || ''
            } else if (currentEvent === 'done') {
              stopDots()
              phase.value = 'done'
              recordData.value = data
              emit('complete', data)
            } else if (currentEvent === 'error') {
              stopDots()
              error.value = data.error
              phase.value = 'done'
              emit('error', data.error)
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
    phase.value = 'error'
    emit('error', e.message)
  }
}

onMounted(startStream)
onUnmounted(stopDots)
</script>

<style scoped>
.markdown-body :deep(h2) {
  font-size: 1.1rem;
  font-weight: 600;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  padding-bottom: 0.25rem;
  border-bottom: 1px solid rgba(0,0,0,0.08);
}
.markdown-body :deep(h3) {
  font-size: 1rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 0.4rem;
}
.markdown-body :deep(p) {
  margin-bottom: 0.6rem;
}
.markdown-body :deep(strong) {
  font-weight: 700;
}
.markdown-body :deep(blockquote) {
  border-left: 3px solid rgba(0,0,0,0.15);
  padding-left: 0.75rem;
  opacity: 0.7;
  margin: 0.5rem 0;
}
.markdown-body :deep(ul), .markdown-body :deep(ol) {
  padding-left: 1.25rem;
  margin-bottom: 0.6rem;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
.animate-spin {
  animation: spin 1s linear infinite;
}
</style>
