<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
    <h3 class="text-xl font-bold mb-4 text-center">📜 卦象结果</h3>

    <div v-if="loading" class="text-center py-8 opacity-50">起卦中，请稍候...</div>
    <div v-else-if="error" class="text-center py-8 text-red-500">{{ error }}</div>
    <template v-else-if="data">
      <div class="grid grid-cols-2 gap-6 mb-6">
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-amber-50'">
          <span class="text-sm opacity-60">本卦</span>
          <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ data.primary.name }}</div>
          <div class="text-3xl">{{ data.primary.symbol }}</div>
          <p class="text-xs opacity-40 mt-1">{{ data.primary.gua_ci }}</p>
        </div>
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-stone-50'">
          <span class="text-sm opacity-60">变卦</span>
          <div class="text-2xl font-bold mt-1">{{ data.changing.name }}</div>
          <div class="text-3xl">{{ data.changing.symbol }}</div>
          <p class="text-xs opacity-40 mt-1">{{ data.changing.gua_ci }}</p>
        </div>
      </div>

      <div class="text-center mb-6">
        <span v-for="y in data.yao_positions" :key="y.position"
          class="inline-block px-2 py-0.5 rounded text-sm mx-0.5"
          :class="y.is_master ? 'bg-red-700 text-amber-100' : 'bg-red-100 text-red-700'">
          {{ y.label }}{{ y.is_master ? ' ★主' : '' }}
        </span>
      </div>

      <hexagram :lines="hexagramLines" :is-dark="isDark" />

      <div class="border-t pt-6 mt-6" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
        <h4 class="text-lg font-medium mb-3">🤖 AI 解卦</h4>
        <div class="markdown-body leading-relaxed" v-html="renderedMarkdown"></div>
      </div>

      <div class="mt-6 text-center">
        <span class="text-xs opacity-40">剩余 {{ data.remaining_quota }} 次</span>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { marked } from 'marked'
import Hexagram from '../components/Hexagram.vue'

const router = useRouter()
const auth = useAuthStore()
const question = history.state?.question || ''
const isDark = computed(() => document.documentElement.classList.contains('dark'))
const loading = ref(true)
const error = ref('')
const data = ref(null)

marked.setOptions({ breaks: true, gfm: true })
const renderedMarkdown = computed(() => marked.parse(data.value?.interpretation || ''))

const hexagramLines = computed(() => {
  if (!data.value) return []
  const names = ['上爻', '五爻', '四爻', '三爻', '二爻', '初爻']
  const desc = data.value.primary.yao_desc || ''
  const changing = (data.value.yao_positions || []).map(y => y.position)
  return names.map((name, i) => ({
    label: name,
    yang: desc[5 - i] === '1',
    changing: changing.includes(5 - i),
  }))
})

onMounted(async () => {
  if (!question) {
    router.push('/')
    return
  }
  try {
    const res = await fetch('/api/divine', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ question }),
    })
    if (res.status === 401) { auth.logout(); router.push('/login'); return }
    const json = await res.json()
    if (res.ok) {
      data.value = json
    } else {
      error.value = json.error || '起卦失败'
    }
  } catch (e) {
    error.value = '网络错误: ' + e.message
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.markdown-body :deep(h2) { font-size: 1.1rem; font-weight: 600; margin-top: 1.25rem; margin-bottom: 0.5rem; padding-bottom: 0.25rem; border-bottom: 1px solid rgba(0,0,0,0.08); }
.markdown-body :deep(h3) { font-size: 1rem; font-weight: 600; margin-top: 1rem; margin-bottom: 0.4rem; }
.markdown-body :deep(p) { margin-bottom: 0.6rem; }
.markdown-body :deep(strong) { font-weight: 700; }
.markdown-body :deep(blockquote) { border-left: 3px solid rgba(0,0,0,0.15); padding-left: 0.75rem; opacity: 0.7; margin: 0.5rem 0; }
.markdown-body :deep(ul), .markdown-body :deep(ol) { padding-left: 1.25rem; margin-bottom: 0.6rem; }
</style>