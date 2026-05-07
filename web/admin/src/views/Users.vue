<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">用户管理</h2>
    <div class="bg-white rounded-lg shadow">
      <table class="w-full text-sm">
        <thead class="bg-gray-100">
          <tr><th class="p-3 text-left">ID</th><th class="p-3 text-left">手机号</th><th class="p-3 text-left">昵称</th><th class="p-3 text-left">注册时间</th></tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-t">
            <td class="p-3">{{ u.id }}</td>
            <td class="p-3">{{ u.phone }}</td>
            <td class="p-3">{{ u.nickname }}</td>
            <td class="p-3">{{ formatDate(u.created_at) }}</td>
          </tr>
        </tbody>
      </table>
      <div v-if="users.length === 0" class="p-6 text-center text-gray-400">暂无用户</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
const users = ref([])
onMounted(async () => {
  const res = await fetch('/api/admin/users?limit=50&offset=0')
  if (res.ok) {
    const data = await res.json()
    users.value = data.items
  }
})
function formatDate(d) { return new Date(d).toLocaleDateString('zh-CN') }
</script>
