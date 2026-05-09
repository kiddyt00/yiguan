<template>
  <header class="shadow-lg bg-transparent">
    <div class="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
      <router-link to="/" class="flex items-center gap-2">
        <img src="../assets/logo-guanjizhai.png" alt="观己斋" class="h-9 w-auto brightness-0 invert" />
      </router-link>
      <div class="flex items-center gap-4 text-sm">
        <template v-if="auth.isLoggedIn()">
          <span v-if="quota !== null" class="text-xs px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300">
            {{ quota }} 次
          </span>
          <router-link to="/history" class="text-gray-300 hover:text-amber-300 transition">历史</router-link>
          <router-link to="/profile" class="text-gray-300 hover:text-amber-300 transition">{{ auth.user?.nickname || '我' }}</router-link>
          <button @click="doLogout" class="text-gray-400 hover:text-amber-300 transition">退出</button>
        </template>
        <template v-else>
          <router-link to="/login" class="text-gray-300 hover:text-amber-300 transition">登录</router-link>
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
