<template>
  <div class="rounded-xl p-4 space-y-3" :class="isDark ? 'bg-slate-800/50' : 'bg-stone-100'">
    <!-- 标题 -->
    <h3 class="text-sm font-bold" :class="isDark ? 'text-stone-300' : 'text-stone-600'">
      {{ t('sidebar.recent') || '最近记录' }}
    </h3>

    <!-- 搜索框 -->
    <div class="relative">
      <input v-model="searchQuery" @input="onSearch" type="text"
        :placeholder="t('sidebar.search') || '搜索问题/卦象...'"
        class="w-full text-xs rounded-lg px-3 py-2 outline-none transition focus:ring-1 focus:ring-amber-500/30"
        :class="isDark ? 'bg-slate-700/60 text-stone-200 placeholder-stone-500 border border-stone-600' : 'bg-white text-stone-700 placeholder-stone-400 border border-stone-300'" />
      <button v-if="searchQuery" @click="clearSearch"
        class="absolute right-2 top-1/2 -translate-y-1/2 text-xs opacity-50 hover:opacity-80">
        ✕
      </button>
    </div>

    <!-- 每页数量 -->
    <div class="flex items-center gap-2 text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
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
      <span v-if="searchQuery" class="ml-auto opacity-60">共 {{ total }} 条匹配</span>
    </div>

    <!-- 记录列表 -->
    <div class="space-y-1.5 max-h-96 overflow-y-auto">
      <div v-if="loading" class="text-center text-xs py-4 opacity-50">加载中...</div>
      <div v-else-if="items.length === 0" class="text-center text-xs py-4 opacity-40">暂无记录</div>
      <div v-for="h in items" :key="h.id"
        @click="$emit('select', h)"
        class="rounded-lg p-2.5 cursor-pointer hover:shadow-md transition text-xs"
        :class="[
          selectedId === h.id
            ? (isDark ? 'bg-amber-500/20 border border-amber-500/40' : 'bg-amber-50 border border-amber-300')
            : (isDark ? 'bg-slate-700/30 hover:bg-slate-700/50 border border-transparent' : 'bg-white/80 hover:bg-white border border-transparent')
        ]">
        <div class="flex justify-between items-start">
          <span class="font-medium truncate max-w-[70%]" :class="isDark ? 'text-stone-200' : 'text-stone-700'">
            {{ h.primary_gua }} → {{ h.changing_gua }}
          </span>
          <span class="text-xs opacity-40 whitespace-nowrap ml-2">{{ formatDate(h.created_at) }}</span>
        </div>
        <p class="mt-0.5 truncate opacity-60" :class="isDark ? 'text-stone-300' : 'text-stone-600'">{{ h.question }}</p>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > pageSize" class="flex items-center justify-between text-xs pt-2 border-t"
      :class="isDark ? 'border-stone-700 text-stone-400' : 'border-stone-200 text-stone-500'">
      <button @click="prevPage" :disabled="page <= 1"
        class="px-2 py-1 rounded transition disabled:opacity-30 hover:opacity-80">
        ‹ 上一页
      </button>
      <span>第 {{ page }} / {{ totalPages }} 页</span>
      <button @click="nextPage" :disabled="page >= totalPages"
        class="px-2 py-1 rounded transition disabled:opacity-30 hover:opacity-80">
        下一页 ›
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  isDark: Boolean,
  selectedId: [Number, String],
})

const emit = defineEmits(['select'])

const auth = useAuthStore()
const router = useRouter()
const { t, locale } = useI18n()

const items = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')
const loading = ref(false)
const searchTimer = ref(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

async function fetchHistory() {
  loading.value = true
  const offset = (page.value - 1) * pageSize.value
  let url
  if (searchQuery.value.trim()) {
    url = `/api/history/search?q=${encodeURIComponent(searchQuery.value.trim())}&limit=${pageSize.value}&offset=${offset}`
  } else {
    url = `/api/history/recent?limit=${pageSize.value}`
  }
  try {
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    const data = await res.json()
    if (res.ok) {
      if (searchQuery.value.trim()) {
        items.value = data.items
        total.value = data.total
      } else {
        items.value = data
        total.value = data.length
      }
    } else if (res.status === 401) {
      auth.logout()
      router.push('/login')
    }
  } catch (e) {
    console.error('获取历史记录失败:', e)
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

function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-CN')
}

onMounted(fetchHistory)
</script>
