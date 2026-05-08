<template>
  <div>
    <h3 class="text-xl font-bold mb-4">历史记录</h3>
    <div v-if="items.length === 0" class="text-center opacity-50 py-12">暂无记录</div>
    <div v-for="h in items" :key="h.id"
      class="bg-white/80 backdrop-blur rounded-xl shadow-sm p-4 mb-3 cursor-pointer hover:shadow-md transition"
      :class="{ '!bg-slate-800/80': isDark }" @click="selected = selected?.id === h.id ? null : h">
      <div class="flex justify-between items-start">
        <div>
          <span class="font-medium">{{ h.primary_gua }}</span>
          <span class="mx-2 opacity-30">→</span>
          <span class="opacity-70">{{ h.changing_gua }}</span>
          <span v-if="h.nickname" class="ml-2 text-xs opacity-40">— {{ h.nickname }}</span>
        </div>
        <span class="text-xs opacity-40">{{ formatDate(h.created_at) }}</span>
      </div>
      <p class="text-sm opacity-60 mt-1">问：{{ h.question }}</p>
      <div v-if="selected?.id === h.id" class="mt-3 pt-3 border-t" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
        <div class="markdown-body text-sm opacity-70" v-html="renderMD(h.interpretation)"></div>
      </div>
    </div>
    <div v-if="total > items.length" class="text-center mt-4">
      <button @click="loadMore" :disabled="loading"
        class="px-6 py-2 rounded-lg text-sm underline opacity-50 hover:opacity-80">
        {{ loading ? '加载中...' : '加载更多' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { marked } from 'marked'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const items = ref([])
const total = ref(0)
const offset = ref(0)
const loading = ref(false)
const selected = ref(null)
const isDark = computed(() => document.documentElement.classList.contains('dark'))

marked.setOptions({ breaks: true, gfm: true })
function renderMD(text) {
  return marked.parse(text || '')
}

async function loadMore() {
  loading.value = true
  const res = await fetch(`/api/history?limit=10&offset=${offset.value}`, {
    headers: { Authorization: `Bearer ${auth.token}` },
  })
  const data = await res.json()
  if (res.ok) {
    items.value.push(...data.items)
    total.value = data.total
    offset.value += data.items.length
  } else if (res.status === 401) {
    auth.logout()
    router.push('/login')
    return
  }
  loading.value = false
}

function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-CN')
}

onMounted(loadMore)
</script>

<style scoped>
.markdown-body :deep(h2) {
  font-size: 1.05rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 0.4rem;
}
.markdown-body :deep(h3) {
  font-size: 0.95rem;
  font-weight: 600;
  margin-top: 0.8rem;
  margin-bottom: 0.3rem;
}
.markdown-body :deep(p) {
  margin-bottom: 0.5rem;
}
.markdown-body :deep(strong) {
  font-weight: 700;
}
.markdown-body :deep(blockquote) {
  border-left: 3px solid rgba(0,0,0,0.15);
  padding-left: 0.75rem;
  opacity: 0.7;
  margin: 0.4rem 0;
}
</style>
