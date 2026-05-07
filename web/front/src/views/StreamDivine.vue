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

    <!-- AI 解卦流式渲染 -->
    <div v-if="showAI" class="border-t pt-6 mt-6" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <h4 class="text-lg font-medium mb-3">🤖 AI 解卦</h4>
      <div class="leading-relaxed whitespace-pre-wrap opacity-80">
        {{ aiText }}<span v-if="phase === 'ai'" class="animate-pulse">▊</span>
      </div>
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

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

const showHexagram = computed(() => ['hexagram', 'ai', 'done'].includes(phase.value))
const showAI = computed(() => ['ai', 'done'].includes(phase.value))
const statusText = computed(() => {
  const map = { coins: '起卦中...', hexagram: '卦象已现', ai: 'AI 解卦中...', done: '解卦完成', error: '出错了' }
  return map[phase.value] || ''
})

async function startStream() {
  try {
    const resp = await fetch('/api/divine/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${props.token}`,
      },
      body: JSON.stringify({ question: props.question }),
    })

    if (!resp.ok) {
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

      for (const line of lines) {
        if (line.startsWith('event:')) {
          const eventType = line.replace('event:', '').trim()
        } else if (line.startsWith('data:')) {
          const dataStr = line.replace('data:', '').trim()
          if (!dataStr) continue

          try {
            const parsed = JSON.parse(dataStr)
            const evt = parsed.event
            const data = parsed.data

            if (evt === 'phase') {
              if (data.phase === 'coins') {
                phase.value = 'coins'
                coinsLabel.value = `${data.label} — ${data.result}`
              } else if (data.phase === 'hexagram') {
                phase.value = 'hexagram'
                hexagram.value = { primary_gua: data.primary_gua, changing_gua: data.changing_gua }
              }
            } else if (evt === 'ai') {
              if (phase.value === 'hexagram') phase.value = 'ai'
              aiText.value += data.chunk
            } else if (evt === 'done') {
              phase.value = 'done'
              recordData.value = data
              emit('complete', data)
            } else if (evt === 'error') {
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
    error.value = '网络连接失败: ' + e.message
    phase.value = 'error'
    emit('error', e.message)
  }
}

onMounted(startStream)
</script>
