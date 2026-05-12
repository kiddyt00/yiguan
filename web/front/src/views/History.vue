<template>
  <div class="glass-card p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-xl font-bold" :class="isDark ? 'text-white' : 'text-stone-800'">{{ t('nav.history') }}</h3>
          <div class="flex items-center gap-3">
            <!-- 搜索框 -->
            <div class="relative">
              <input v-model="searchQuery" @input="onSearch" type="text"
                :placeholder="t('history.search') || '搜索问题/卦象...'"
                class="text-sm rounded-lg px-3 py-1.5 outline-none transition focus:ring-1 focus:ring-amber-500/30"
                :class="isDark ? 'bg-slate-700/60 text-stone-200 placeholder-stone-500 border border-stone-600 w-48' : 'bg-white text-stone-700 placeholder-stone-400 border border-stone-300 w-48'" />
              <button v-if="searchQuery" @click="clearSearch"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-xs opacity-50 hover:opacity-80">
                ✕
              </button>
            </div>
            <!-- 每页数量 -->
            <div class="flex items-center gap-1.5 text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
              <span>每页</span>
              <select v-model.number="pageSize" @change="onPageSizeChange"
                class="rounded px-1.5 py-0.5 text-xs outline-none"
                :class="isDark ? 'bg-slate-700/60 border border-stone-600 text-stone-200' : 'bg-white border border-stone-300 text-stone-700'">
                <option :value="5">5</option>
                <option :value="10">10</option>
                <option :value="20">20</option>
                <option :value="50">50</option>
              </select>
              <span>条</span>
            </div>
          </div>
        </div>

        <div v-if="items.length === 0 && !loading" class="text-center opacity-50 py-12">
          {{ searchQuery ? '未找到匹配的记录' : '暂无记录' }}
        </div>
        <div v-if="loading && items.length === 0" class="text-center opacity-50 py-12">加载中...</div>

        <div v-for="h in items" :key="h.id"
          class="bg-white/80 backdrop-blur rounded-xl shadow-sm p-4 mb-3 cursor-pointer hover:shadow-md transition"
          :class="{ '!bg-slate-800/80': isDark }" @click="toggleSelect(h)">
          <div class="flex justify-between items-start">
            <div>
              <span class="font-medium">{{ h.primary_gua }}</span>
              <span class="mx-2 opacity-30">→</span>
              <span class="opacity-70">{{ h.changing_gua }}</span>
              <span v-if="h.nickname" class="ml-2 text-xs" :class="isDark ? 'opacity-40' : 'opacity-60'">— {{ h.nickname }}</span>
            </div>
            <span class="text-xs opacity-40">{{ formatDate(h.created_at) }}</span>
          </div>
          <p class="text-sm opacity-60 mt-1">问：{{ h.question }}</p>
          <div v-if="selected?.id === h.id" class="mt-3 pt-3 border-t" :class="isDark ? 'border-slate-600' : 'border-stone-200'">

            <!-- 翻译提示条 -->
            <div v-if="needsTranslation(h) && !translatingMap[h.id]" class="mb-3 px-3 py-2 rounded-lg flex items-center justify-between gap-2 bg-amber-500/10 border border-amber-500/30">
              <span class="text-xs text-amber-400">
                {{ h.lang === 'zh' ? '⚠ 此解读以中文生成' : '⚠ This reading was generated in English' }}
              </span>
              <button @click.stop="doTranslateHistory(h)" :disabled="translatingMap[h.id]"
                class="text-xs px-2 py-0.5 rounded-full font-medium transition border border-amber-500/50 text-amber-400 hover:bg-amber-500/20 disabled:opacity-50">
                {{ translatingMap[h.id] ? '...' : (h.lang === 'zh' ? '翻译为 EN →' : '翻译为 中文 →') }}
              </button>
            </div>

            <div class="markdown-body text-sm opacity-70" v-html="renderMD(getDisplayText(h))"></div>
          </div>
        </div>

        <!-- 分页控件 -->
        <div v-if="totalPages > 1" class="flex items-center justify-between mt-6 pt-4 border-t"
          :class="isDark ? 'border-stone-700' : 'border-stone-200'">
          <span class="text-xs opacity-50">共 {{ total }} 条, 第 {{ page }}/{{ totalPages }} 页</span>
          <div class="flex items-center gap-2">
            <button @click="goToPage(1)" :disabled="page <= 1"
              class="px-2 py-1 rounded text-xs transition disabled:opacity-30 hover:opacity-80">
              ‹‹
            </button>
            <button @click="prevPage" :disabled="page <= 1"
              class="px-3 py-1 rounded text-xs transition disabled:opacity-30 hover:opacity-80">
              ‹ 上一页
            </button>
            <span class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ page }} / {{ totalPages }}</span>
            <button @click="nextPage" :disabled="page >= totalPages"
              class="px-3 py-1 rounded text-xs transition disabled:opacity-30 hover:opacity-80">
              下一页 ›
            </button>
            <button @click="goToPage(totalPages)" :disabled="page >= totalPages"
              class="px-2 py-1 rounded text-xs transition disabled:opacity-30 hover:opacity-80">
              ››
            </button>
          </div>
        </div>
      </div>
</template>

<script setup>
import { ref, computed, onMounted, reactive } from 'vue'
import { marked } from 'marked'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useTranslation } from '../composables/useTranslation'

const auth = useAuthStore()
const router = useRouter()
const { t, locale } = useI18n()
const { needsTranslation, getTranslation, generateTranslation } = useTranslation()
const items = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const selected = ref(null)
const isDark = computed(() => !document.documentElement.classList.contains('light'))
const translatingMap = reactive({})
const transCache = reactive({})
const searchQuery = ref('')
const searchTimer = ref(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

marked.setOptions({ breaks: true, gfm: true })

function renderMD(text) {
  return marked.parse(text || '')
}

function getDisplayText(h) {
  return transCache[h.id] || h.interpretation
}

function toggleSelect(h) {
  if (selected.value?.id === h.id) {
    selected.value = null
  } else {
    selected.value = h
    if (needsTranslation(h)) {
      getTranslation(h.id, locale.value).then(text => {
        if (text) transCache[h.id] = text
      }).catch(() => {})
    }
  }
}


async function doTranslateHistory(h) {
  if (translatingMap[h.id]) return
  translatingMap[h.id] = true
  try {
    const text = await generateTranslation(h.id, locale.value)
    transCache[h.id] = text
  } catch (e) {
    console.error('翻译失败:', e)
  } finally {
    translatingMap[h.id] = false
  }
}

async function fetchHistory() {
  loading.value = true
  const offset = (page.value - 1) * pageSize.value
  let url
  if (searchQuery.value.trim()) {
    url = `/api/history/search?q=${encodeURIComponent(searchQuery.value.trim())}&limit=${pageSize.value}&offset=${offset}`
  } else {
    url = `/api/history?limit=${pageSize.value}&offset=${offset}`
  }
  try {
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    const data = await res.json()
    if (res.ok) {
      items.value = data.items
      total.value = data.total
    } else if (res.status === 401) {
      auth.logout()
      router.push('/login')
      return
    }
  } catch (e) {
    console.error('加载历史失败:', e)
  } finally {
    loading.value = false
  }
}

function onSearch() {
  clearTimeout(searchTimer.value)
  searchTimer.value = setTimeout(() => {
    page.value = 1
    fetchHistory()
  }, 300)
}

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  fetchHistory()
}

function onPageSizeChange() {
  page.value = 1
  fetchHistory()
}

function prevPage() {
  if (page.value > 1) {
    page.value--
    fetchHistory()
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value++
    fetchHistory()
  }
}

function goToPage(p) {
  if (p >= 1 && p <= totalPages.value) {
    page.value = p
    fetchHistory()
  }
}

function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-CN')
}

onMounted(fetchHistory)
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
