<template>
  <div class="max-w-md mx-auto">
    <div class="glass-card p-6">
      <h3 class="text-xl font-bold mb-4 text-center" :class="isDark ? 'text-stone-100' : 'text-stone-800'">{{ tabLabel }}</h3>

      <div class="flex mb-4 border-b" :class="isDark ? 'border-stone-700' : 'border-stone-300'">
        <button v-for="tabKey in ['login','sms','qrcode','register']" :key="tabKey"
          @click="tab = tabKey" class="flex-1 py-2 text-center text-sm"
          :class="tab === tabKey ? 'border-b-2 font-medium border-amber-500 text-amber-400' : (isDark ? 'text-stone-400' : 'text-stone-500')">
          {{ t(`login.tab.${tabKey}`) }}
        </button>
      </div>

      <!-- 微信扫码登录 -->
      <template v-if="tab === 'qrcode'">
        <div class="text-center py-4">
          <div v-if="qrStatus === 'loading'" class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('login.qrcode.loading') }}</div>
          <div v-else-if="qrStatus === 'pending'" class="space-y-3">
            <div id="qrcode" class="inline-block bg-white p-3 rounded-lg"></div>
            <p class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('login.qrcode.prompt') }}</p>
          </div>
          <div v-else-if="qrStatus === 'ok'" class="text-green-500 font-medium">{{ t('login.qrcode.ok') }}</div>
          <div v-else-if="qrStatus === 'expired'" class="space-y-3">
            <p class="text-sm text-red-400">{{ t('login.qrcode.expired') }}</p>
            <button @click="genQRCode" class="text-sm underline" :class="isDark ? 'text-stone-300' : 'text-stone-600'">{{ t('login.qrcode.retry') }}</button>
          </div>
          <div v-else class="text-sm text-red-400">{{ qrError }}</div>
        </div>
      </template>

      <!-- 短信登录 -->
      <template v-if="tab === 'sms'">
        <div class="flex gap-2 mb-4">
          <input v-model="phone" :placeholder="t('login.phone.placeholder')"
            class="flex-1 border rounded-lg p-3 bg-transparent outline-none focus:border-amber-500"
            :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />
          <button @click="sendSMS" :disabled="smsCountdown > 0 || phone.length !== 11"
            class="px-4 py-3 rounded-lg text-sm font-medium whitespace-nowrap transition disabled:opacity-40"
            :class="isDark ? 'bg-slate-700 text-stone-200 hover:bg-slate-600' : 'bg-stone-200 text-stone-700 hover:bg-stone-300'">
            {{ smsCountdown > 0 ? smsCountdown + 's' : t('login.sms.send') }}
          </button>
        </div>
        <input v-model="code" :placeholder="t('login.code.placeholder')" maxlength="6"
          class="w-full border rounded-lg p-3 mb-4 bg-transparent outline-none focus:border-amber-500"
          :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />
        <button @click="smsLogin" :disabled="loading"
          class="w-full py-3 rounded-lg font-medium transition bg-amber-600 text-white hover:bg-amber-500">
          {{ loading ? t('login.submit.loading') : t('login.sms.login') }}
        </button>
      </template>

      <!-- 密码登录 / 注册 -->
      <template v-else-if="tab === 'login' || tab === 'register'">
        <input v-model="phone" :placeholder="t('login.phone.placeholder')"
          class="w-full border rounded-lg p-3 mb-3 bg-transparent outline-none focus:border-amber-500"
          :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />
        <input v-model="password" type="password"
          :placeholder="tab === 'register' ? t('login.password.register') : t('login.password.placeholder')"
          class="w-full border rounded-lg p-3 mb-3 bg-transparent outline-none focus:border-amber-500"
          :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />
        <input v-if="tab === 'register'" v-model="nickname" :placeholder="t('login.nickname.placeholder')"
          class="w-full border rounded-lg p-3 mb-4 bg-transparent outline-none focus:border-amber-500"
          :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />

        <button @click="submit" :disabled="loading"
          class="w-full py-3 rounded-lg font-medium transition bg-amber-600 text-white hover:bg-amber-500">
          {{ loading ? t('login.submit.loading') : t(`login.submit.${tab}`) }}
        </button>

        <p v-if="tab === 'register'" class="text-center text-xs mt-3" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('login.gift') }}</p>
      </template>

      <div v-if="error && tab !== 'qrcode'" class="text-red-500 text-sm mt-3 text-center">{{ error }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

defineProps({ isDark: Boolean })

const { t } = useI18n()
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
const qrStatus = ref('')
const qrError = ref('')
const qrTicket = ref('')
let qrTimer = null
const tabLabel = computed(() => t(`login.tab.${tab.value}`))

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
    else { error.value = data.error || t('login.network.error') }
  } catch (e) { error.value = t('login.network.error') }
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
      error.value = data.error || t('login.network.error')
    }
  } catch (e) { error.value = t('login.network.error') }
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
    else { error.value = data.error || t('login.network.error') }
  } catch (e) { error.value = t('login.network.error') }
  finally { loading.value = false }
}

watch(tab, (val) => {
  error.value = ''
  if (val === 'qrcode') genQRCode()
  else { if (qrTimer) { clearInterval(qrTimer); qrTimer = null } }
})

onUnmounted(() => { if (qrTimer) clearInterval(qrTimer) })

async function genQRCode() {
  qrStatus.value = 'loading'
  qrError.value = ''
  try {
    const res = await fetch('/api/auth/wechat-qrcode')
    const data = await res.json()
    if (!res.ok) { qrStatus.value = 'error'; qrError.value = data.error; return }
    qrTicket.value = data.ticket
    const container = document.getElementById('qrcode')
    if (container) {
      container.innerHTML = ''
      const canvas = document.createElement('canvas')
      canvas.width = canvas.height = 200
      container.appendChild(canvas)
      drawQRCode(canvas, data.qrcode_url)
    }
    qrStatus.value = 'pending'
    startPoll()
  } catch (e) { qrStatus.value = 'error'; qrError.value = t('login.network.error') }
}

function startPoll() {
  if (qrTimer) clearInterval(qrTimer)
  qrTimer = setInterval(async () => {
    try {
      const res = await fetch('/api/auth/wechat-check?ticket=' + qrTicket.value)
      const data = await res.json()
      if (data.status === 'ok') { clearInterval(qrTimer); qrStatus.value = 'ok'; auth.setAuth(data.token, {}); setTimeout(() => router.push('/'), 800) }
      else if (data.status === 'expired') { clearInterval(qrTimer); qrStatus.value = 'expired' }
    } catch (e) { /* ignore poll errors */ }
  }, 2000)
}

function drawQRCode(canvas, url) {
  const img = new Image()
  img.onload = () => { const ctx = canvas.getContext('2d'); ctx.drawImage(img, 0, 0, 200, 200) }
  img.src = 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=' + encodeURIComponent(url)
}
</script>