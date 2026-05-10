<template>
  <div>
    <h3 class="text-xl font-bold mb-4" :class="isDark ? 'text-white' : 'text-stone-800'">{{ t('nav.history') }}</h3>
    <div v-if="items.length === 0" class="text-center opacity-50 py-12">暂无记录</div>
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
    <div v-if="total > items.length" class="text-center mt-4">
      <button @click="loadMore" :disabled="loading"
        class="px-6 py-2 rounded-lg text-sm underline opacity-50 hover:opacity-80">
        {{ loading ? '加载中...' : '加载更多' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, reactive } from 'vue'
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
const offset = ref(0)
const loading = ref(false)
const selected = ref(null)
const isDark = computed(() => !document.documentElement.classList.contains('light'))
const translatingMap = reactive({})
const transCache = reactive({}) // { [historyId]: text }

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
    // 尝试加载已有翻译（静默，不提示）
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
    const target = locale.value === 'zh' ? 'en' : 'zh'
    // 当前用户想看的语言和 UI 一致
    const text = await generateTranslation(h.id, locale.value)
    transCache[h.id] = text
  } catch (e) {
    console.error('翻译失败:', e)
  } finally {
    translatingMap[h.id] = false
  }
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
