<template>
  <div class="max-w-md mx-auto">
    <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
      <h3 class="text-xl font-bold mb-4 text-center">{{ isRegister ? '注册' : '登录' }}</h3>

      <div class="flex mb-4 border-b" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
        <button @click="isRegister = false" class="flex-1 py-2 text-center"
          :class="!isRegister ? 'border-b-2 font-medium ' + (isDark ? 'border-cyan-400 text-cyan-400' : 'border-red-800 text-red-800') : 'opacity-50'">登录</button>
        <button @click="isRegister = true" class="flex-1 py-2 text-center"
          :class="isRegister ? 'border-b-2 font-medium ' + (isDark ? 'border-cyan-400 text-cyan-400' : 'border-red-800 text-red-800') : 'opacity-50'">注册</button>
      </div>

      <input v-model="phone" placeholder="手机号"
        class="w-full border rounded-lg p-3 mb-3 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
      <input v-model="password" type="password" placeholder="密码（至少6位）"
        class="w-full border rounded-lg p-3 mb-3 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
      <input v-if="isRegister" v-model="nickname" placeholder="昵称（选填）"
        class="w-full border rounded-lg p-3 mb-4 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />

      <div v-if="error" class="text-red-500 text-sm mb-3 text-center">{{ error }}</div>

      <button @click="submit" :disabled="loading"
        class="w-full py-3 rounded-lg font-medium transition"
        :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
        {{ loading ? '处理中...' : (isRegister ? '注册' : '登录') }}
      </button>

      <p v-if="isRegister" class="text-center text-xs opacity-50 mt-3">新用户注册即赠 3 次免费算卦</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const phone = ref('')
const password = ref('')
const nickname = ref('')
const isRegister = ref(false)
const error = ref('')
const loading = ref(false)
const isDark = computed(() => document.documentElement.classList.contains('dark'))

async function submit() {
  error.value = ''
  loading.value = true
  const url = isRegister.value ? '/api/auth/register' : '/api/auth/login'
  const body = isRegister.value
    ? { phone: phone.value, password: password.value, nickname: nickname.value }
    : { phone: phone.value, password: password.value }
  try {
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const data = await res.json()
    if (res.ok) {
      auth.setAuth(data.token, data.user)
      router.push('/')
    } else {
      error.value = data.error || '操作失败'
    }
  } catch (e) {
    error.value = '网络错误'
  } finally {
    loading.value = false
  }
}
</script>
