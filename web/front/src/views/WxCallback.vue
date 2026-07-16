<template>
  <div class="min-h-screen flex items-center justify-center" :class="isDark ? 'bg-stone-900' : 'bg-stone-50'">
    <div class="text-center">
      <div v-if="loading" class="space-y-4">
        <div class="animate-spin text-4xl">☯</div>
        <p class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">微信登录中...</p>
      </div>
      <div v-else-if="success" class="space-y-3">
        <p class="text-green-500 text-2xl">✅</p>
        <p class="text-green-600 font-medium">登录成功</p>
        <p class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
          {{ autoClose ? '窗口即将自动关闭...' : '请关闭此窗口' }}
        </p>
      </div>
      <div v-else class="space-y-3">
        <p class="text-red-500 text-2xl">❌</p>
        <p class="text-red-500 text-sm">{{ error }}</p>
        <button @click="tryClose" class="text-xs underline text-amber-600">关闭窗口</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const loading = ref(true)
const success = ref(false)
const error = ref('')
const autoClose = ref(false)
const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches

onMounted(async () => {
  const code = route.query.code
  const state = route.query.state

  if (!code) {
    error.value = '缺少授权参数'
    loading.value = false
    return
  }

  // 验证 state（CSRF 防护）
  const savedState = sessionStorage.getItem('wx_login_state')
  if (savedState && state !== savedState) {
    error.value = '安全验证失败，请重试'
    loading.value = false
    return
  }
  sessionStorage.removeItem('wx_login_state')

  try {
    const res = await fetch('/api/auth/wechat-code', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    })
    const data = await res.json()
    if (!res.ok) {
      error.value = data.error || '登录失败'
      loading.value = false
      return
    }

    // 登录成功
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user || {}))

    // 通知父窗口（iframe 场景）
    if (window.parent !== window) {
      window.parent.postMessage({ type: 'wx-login', token: data.token, user: data.user }, '*')
    }

    success.value = true
    loading.value = false

    // 自动关闭
    autoClose.value = true
    setTimeout(() => tryClose(), 1500)
  } catch (e) {
    error.value = '网络错误'
    loading.value = false
  }
})

function tryClose() {
  if (window.parent !== window) {
    // iframe 内，让父窗口关闭
    window.parent.postMessage({ type: 'wx-login-close' }, '*')
  } else {
    window.close()
  }
}
</script>
