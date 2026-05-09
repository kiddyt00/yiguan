<template>
  <div class="glass-card p-6">
    <h3 class="text-sm font-medium text-center mb-1 text-green-400">结果展示</h3>
    <h2 class="text-xl font-bold mb-4 text-center text-stone-100">占卜结果</h2>

    <div v-if="loading" class="text-center py-8 text-stone-400">起卦中，请稍候...</div>
    <div v-else-if="error" class="text-center py-8 text-red-400">{{ error }}</div>
    <template v-else-if="data">
      <div class="grid grid-cols-2 gap-6 mb-6">
        <div class="rounded-lg p-4 text-center bg-slate-800/50">
          <span class="text-sm text-stone-400">本卦</span>
          <div class="text-2xl font-bold mt-1 text-amber-400">{{ data.primary.name }}</div>
          <div class="text-3xl">{{ data.primary.symbol }}</div>
          <p class="text-xs text-stone-500 mt-1">{{ data.primary.gua_ci }}</p>
        </div>
        <div class="rounded-lg p-4 text-center bg-slate-700/40">
          <span class="text-sm text-stone-400">变卦</span>
          <div class="text-2xl font-bold mt-1 text-amber-400">{{ data.changing.name }}</div>
          <div class="text-3xl">{{ data.changing.symbol }}</div>
          <p class="text-xs text-stone-500 mt-1">{{ data.changing.gua_ci }}</p>
        </div>
      </div>

      <div class="text-center mb-6">
        <span v-for="y in data.yao_positions" :key="y.position"
          class="inline-block px-2 py-0.5 rounded text-sm mx-0.5 bg-amber-500/20 text-amber-300">
          {{ y.label }}{{ y.is_master ? ' ★主' : '' }}
        </span>
      </div>

      <hexagram :lines="hexagramLines" :is-dark="true" />

      <div class="border-t border-stone-700 pt-6 mt-6">
        <h4 class="text-lg font-medium mb-3 text-stone-200">🤖 AI 解读</h4>
        <div class="markdown-body leading-relaxed" v-html="renderedMarkdown"></div>
      </div>

      <div class="mt-6 text-center">
        <span class="text-xs text-stone-500">剩余 {{ data.remaining_quota }} 次</span>
      </div>

      <!-- 大师入口 -->
      <div class="mt-4 pt-4 border-t border-stone-700">
        <button @click="showMaster = !showMaster"
          class="w-full py-3 rounded-lg text-center font-medium transition bg-amber-500/10 text-amber-400 hover:bg-amber-500/20">
          🎓 周易大师一对一详解
        </button>
        <div v-if="showMaster" class="mt-3 text-center">
          <img src="/qr-master.png" alt="大师二维码" class="w-48 h-auto mx-auto rounded-lg border border-stone-600" />
          <p class="text-xs text-stone-500 mt-2">扫码添加大师微信，获取深度解读</p>
        </div>
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
const loading = ref(true)
const error = ref('')
const showMaster = ref(false)
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
.markdown-body :deep(h2) { font-size: 1.1rem; font-weight: 600; margin-top: 1.25rem; margin-bottom: 0.5rem; padding-bottom: 0.25rem; border-bottom: 1px solid rgba(255,255,255,0.1); color: #e8e4d8; }
.markdown-body :deep(h3) { font-size: 1rem; font-weight: 600; margin-top: 1rem; margin-bottom: 0.4rem; color: #e8e4d8; }
.markdown-body :deep(p) { margin-bottom: 0.6rem; color: #d0ccc4; }
.markdown-body :deep(strong) { font-weight: 700; color: #e8c97a; }
.markdown-body :deep(blockquote) { border-left: 3px solid rgba(212,168,83,0.3); padding-left: 0.75rem; color: #9ca3af; margin: 0.5rem 0; }
.markdown-body :deep(ul), .markdown-body :deep(ol) { padding-left: 1.25rem; margin-bottom: 0.6rem; color: #d0ccc4; }
</style>
