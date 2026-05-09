<template>
  <div class="max-w-md mx-auto">
    <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
      <h3 class="text-xl font-bold mb-4 text-center">{{ tabLabel }}</h3>

      <div class="flex mb-4 border-b" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
        <button @click="tab = 'login'" class="flex-1 py-2 text-center text-sm"
          :class="tab === 'login' ? 'border-b-2 font-medium ' + activeTabClass : 'opacity-50'">密码登录</button>
        <button @click="tab = 'sms'" class="flex-1 py-2 text-center text-sm"
          :class="tab === 'sms' ? 'border-b-2 font-medium ' + activeTabClass : 'opacity-50'">短信登录</button>
        <button @click="tab = 'register'" class="flex-1 py-2 text-center text-sm"
          :class="tab === 'register' ? 'border-b-2 font-medium ' + activeTabClass : 'opacity-50'">注册</button>
      </div>

      <!-- 短信登录 -->
      <template v-if="tab === 'sms'">
        <div class="flex gap-2 mb-4">
          <input v-model="phone" placeholder="手机号"
            class="flex-1 border rounded-lg p-3 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
          <button @click="sendSMS" :disabled="smsCountdown > 0 || phone.length !== 11"
            class="px-4 py-3 rounded-lg text-sm font-medium whitespace-nowrap transition"
            :class="isDark ? 'bg-slate-600 text-white' : 'bg-stone-200 text-stone-700'">
            {{ smsCountdown > 0 ? smsCountdown + 's' : '获取验证码' }}
          </button>
        </div>
        <input v-model="code" placeholder="验证码" maxlength="6"
          class="w-full border rounded-lg p-3 mb-4 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
        <button @click="smsLogin" :disabled="loading"
          class="w-full py-3 rounded-lg font-medium transition"
          :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
          {{ loading ? '处理中...' : '登录 / 注册' }}
        </button>
      </template>

      <!-- 密码登录 / 注册 -->
      <template v-else>
        <input v-model="phone" placeholder="手机号"
          class="w-full border rounded-lg p-3 mb-3 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
        <input v-model="password" type="password" :placeholder="tab === 'register' ? '密码（至少6位）' : '密码'"
          class="w-full border rounded-lg p-3 mb-3 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
        <input v-if="tab === 'register'" v-model="nickname" placeholder="昵称"
          class="w-full border rounded-lg p-3 mb-4 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />

        <button @click="submit" :disabled="loading"
          class="w-full py-3 rounded-lg font-medium transition"
          :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
          {{ loading ? '处理中...' : (tab === 'register' ? '注册' : '登录') }}
        </button>

        <p v-if="tab === 'register'" class="text-center text-xs opacity-50 mt-3">新用户注册即赠 3 次免费起卦</p>
      </template>

      <div v-if="error" class="text-red-500 text-sm mt-3 text-center">{{ error }}</div>
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
const code = ref('')
const tab = ref('login')
const error = ref('')
const loading = ref(false)
const smsCountdown = ref(0)
const isDark = computed(() => document.documentElement.classList.contains('dark'))
const activeTabClass = computed(() => isDark.value ? 'border-cyan-400 text-cyan-400' : 'border-red-800 text-red-800')
const tabLabel = computed(() => ({ login: '密码登录', sms: '短信登录', register: '注册' }[tab.value]))

async function submit() {
  error.value = ''
  loading.value = true
  const url = tab.value === 'register' ? '/api/auth/register' : '/api/auth/login'
  const body = tab.value === 'register'
    ? { phone: phone.value, password: password.value, nickname: nickname.value }
    : { phone: phone.value, password: password.value }
  try {
    const res = await fetch(url, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
    })
    const data = await res.json()
    if (res.ok) { auth.setAuth(data.token, data.user); router.push('/') }
    else { error.value = data.error || '操作失败' }
  } catch (e) { error.value = '网络错误' }
  finally { loading.value = false }
}

async function sendSMS() {
  try {
    const res = await fetch('/api/auth/sms-send', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone: phone.value }),
    })
    if (res.ok) {
      smsCountdown.value = 60
      const timer = setInterval(() => { smsCountdown.value--; if (smsCountdown.value <= 0) clearInterval(timer) }, 1000)
    } else {
      const data = await res.json()
      error.value = data.error || '发送失败'
    }
  } catch (e) { error.value = '网络错误' }
}

async function smsLogin() {
  error.value = ''
  loading.value = true
  try {
    const res = await fetch('/api/auth/sms-login', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone: phone.value, code: code.value }),
    })
    const data = await res.json()
    if (res.ok) { auth.setAuth(data.token, data.user); router.push('/') }
    else { error.value = data.error || '验证失败' }
  } catch (e) { error.value = '网络错误' }
  finally { loading.value = false }
}
</script>
