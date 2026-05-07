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
        </div>
        <span class="text-xs opacity-40">{{ formatDate(h.created_at) }}</span>
      </div>
      <p class="text-sm opacity-60 mt-1">问：{{ h.question }}</p>
      <div v-if="selected?.id === h.id" class="mt-3 pt-3 border-t" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
        <p class="text-sm opacity-70 whitespace-pre-wrap">{{ h.interpretation }}</p>
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
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const items = ref([])
const total = ref(0)
const offset = ref(0)
const loading = ref(false)
const selected = ref(null)
const isDark = computed(() => document.documentElement.classList.contains('dark'))

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
  }
  loading.value = false
}

function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-CN')
}

onMounted(loadMore)
</script>
