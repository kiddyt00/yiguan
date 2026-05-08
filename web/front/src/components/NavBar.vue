<template>
  <header class="shadow-lg" :class="isDark ? 'bg-slate-800 text-cyan-300' : 'bg-red-900 text-amber-100'">
    <div class="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
      <router-link to="/" class="text-2xl font-bold tracking-wider">☯ 易观</router-link>
      <div class="flex items-center gap-4 text-sm">
        <button @click="$emit('toggleTheme')" class="opacity-75 hover:opacity-100">
          {{ isDark ? '☀' : '🌙' }}
        </button>
        <template v-if="auth.isLoggedIn()">
          <span v-if="quota !== null" class="text-xs px-2 py-0.5 rounded-full" :class="isDark ? 'bg-slate-700' : 'bg-red-800'">
            {{ quota }} 次
          </span>
          <router-link to="/history">历史</router-link>
          <router-link to="/ads" class="text-amber-300">📢 领次数</router-link>
          <router-link to="/profile">{{ auth.user?.nickname || '我' }}</router-link>
          <button @click="doLogout" class="opacity-75">退出</button>
        </template>
        <template v-else>
          <router-link to="/login">登录</router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
defineProps(['isDark'])
defineEmits(['toggleTheme'])
const auth = useAuthStore()
const $router = useRouter()
const quota = ref(null)

onMounted(async () => {
  if (auth.isLoggedIn()) {
    try {
      const res = await fetch('/api/user', {
        headers: { Authorization: `Bearer ${auth.token}` }
      })
      if (res.ok) {
        const data = await res.json()
        quota.value = data.remaining_quota
      } else if (res.status === 401) {
        auth.logout()
        $router.push('/login')
      }
    } catch {}
  }
})

function doLogout() {
  auth.logout()
  $router.push('/')
}
</script>
