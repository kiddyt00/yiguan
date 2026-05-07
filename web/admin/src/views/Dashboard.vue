<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">仪表盘</h2>
    <div class="grid grid-cols-3 gap-4">
      <div class="bg-white rounded-lg shadow p-4">
        <div class="text-sm text-gray-500">注册用户</div>
        <div class="text-3xl font-bold">{{ stats.total_users }}</div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="text-sm text-gray-500">今日算卦</div>
        <div class="text-3xl font-bold">{{ stats.today_divines }}</div>
      </div>
      <div class="bg-white rounded-lg shadow p-4">
        <div class="text-sm text-gray-500">总计算卦</div>
        <div class="text-3xl font-bold">{{ stats.total_divines }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
const stats = ref({ total_users: 0, today_divines: 0, total_divines: 0 })
onMounted(async () => {
  const res = await fetch('/api/admin/dashboard')
  if (res.ok) stats.value = await res.json()
})
</script>
