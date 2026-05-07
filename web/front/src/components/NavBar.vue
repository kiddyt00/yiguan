<template>
  <header class="shadow-lg" :class="isDark ? 'bg-slate-800 text-cyan-300' : 'bg-red-900 text-amber-100'">
    <div class="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
      <router-link to="/" class="text-2xl font-bold tracking-wider">☯ 易观</router-link>
      <div class="flex items-center gap-4 text-sm">
        <button @click="$emit('toggleTheme')" class="opacity-75 hover:opacity-100">
          {{ isDark ? '☀' : '🌙' }}
        </button>
        <template v-if="auth.isLoggedIn()">
          <router-link to="/history">历史</router-link>
          <router-link to="/profile">{{ auth.user?.nickname || '我' }}</router-link>
          <button @click="auth.logout(); $router.push('/')" class="opacity-75">退出</button>
        </template>
        <template v-else>
          <router-link to="/login">登录</router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
defineProps(['isDark'])
defineEmits(['toggleTheme'])
const auth = useAuthStore()
const $router = useRouter()
</script>
