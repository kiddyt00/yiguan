<template>
  <div class="max-w-md mx-auto">
    <div class="glass-card p-6">
      <h3 class="text-xl font-bold mb-5 text-center" :class="isDark ? 'text-stone-100' : 'text-stone-800'">
        {{ showQR ? t('login.qrcode.title') : t('login.title') }}
      </h3>

      <!-- ======== 二维码扫码模式 ======== -->
      <template v-if="showQR">
        <div class="text-center py-4">
          <div v-if="qrStatus === 'loading'" class="text-sm py-8" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
            {{ t('login.qrcode.loading') }}
          </div>
          <div v-else-if="qrStatus === 'pending'" class="space-y-3">
            <div id="qrcode" class="inline-block bg-white p-3 rounded-lg"></div>
            <p class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('login.qrcode.prompt') }}</p>
          </div>
          <div v-else-if="qrStatus === 'ok'" class="py-8">
            <p class="text-green-500 text-lg font-medium">{{ t('login.qrcode.ok') }}</p>
          </div>
          <div v-else-if="qrStatus === 'expired'" class="space-y-3 py-4">
            <p class="text-sm text-red-400">{{ t('login.qrcode.expired') }}</p>
            <button @click="genQRCode" class="text-sm underline" :class="isDark ? 'text-stone-300' : 'text-stone-600'">
              {{ t('login.qrcode.retry') }}
            </button>
          </div>
          <div v-else class="text-sm text-red-400 py-4">{{ qrError }}</div>

          <button @click="closeQR" class="mt-4 text-sm opacity-60 hover:opacity-100 transition"
            :class="isDark ? 'text-stone-400' : 'text-stone-500'">
            ← {{ t('login.qrcode.back') }}
          </button>
        </div>
      </template>

      <!-- ======== 表单模式 ======== -->
      <template v-else>
        <!-- 方法切换 -->
        <div class="flex mb-5 bg-stone-100/10 rounded-lg p-0.5"
          :class="isDark ? 'bg-stone-800/60' : 'bg-stone-100'">
          <button v-for="m in methods" :key="m.key"
            @click="method = m.key; error = ''"
            class="flex-1 py-2 text-sm text-center rounded-md transition font-medium"
            :class="method === m.key
              ? 'bg-amber-600 text-white shadow-sm'
              : (isDark ? 'text-stone-400 hover:text-stone-200' : 'text-stone-500 hover:text-stone-700')">
            {{ t(m.label) }}
          </button>
        </div>

        <!-- 手机号（共用） -->
        <input v-model="phone" :placeholder="t('login.phone.placeholder')" maxlength="11"
          @input="onPhoneInput" @keyup.enter="method === 'sms' && code.length === 6 ? submit() : null"
          class="w-full border rounded-lg p-3 mb-3 bg-transparent outline-none focus:border-amber-500 transition"
          :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />

        <!-- 验证码模式 -->
        <template v-if="method === 'sms'">
          <div class="flex gap-2 mb-4">
            <input v-model="code" :placeholder="t('login.code.placeholder')" maxlength="6"
              @input="onCodeInput" ref="codeInput"
              class="flex-1 border rounded-lg p-3 bg-transparent outline-none focus:border-amber-500 transition"
              :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />
            <button @click="sendSMS" :disabled="smsCountdown > 0 || phone.length !== 11"
              class="px-4 py-3 rounded-lg text-sm font-medium whitespace-nowrap transition disabled:opacity-40"
              :class="isDark ? 'bg-slate-700 text-stone-200 hover:bg-slate-600' : 'bg-stone-200 text-stone-700 hover:bg-stone-300'">
              {{ smsCountdown > 0 ? smsCountdown + 's' : t('login.sms.send') }}
            </button>
          </div>
        </template>

        <!-- 密码模式 -->
        <template v-if="method === 'password'">
          <input v-model="password" type="password" :placeholder="t('login.password.placeholder')"
            @keyup.enter="submit()"
            class="w-full border rounded-lg p-3 mb-3 bg-transparent outline-none focus:border-amber-500 transition"
            :class="isDark ? 'text-stone-100 border-stone-600 placeholder:text-stone-500' : 'text-stone-800 border-stone-300 placeholder:text-stone-400'" />
        </template>

        <!-- CTA 按钮 -->
        <button @click="submit()" :disabled="loading || submitDisabled"
          class="w-full py-3 rounded-lg font-medium transition disabled:opacity-50"
          :class="loading ? 'bg-amber-500 text-white cursor-wait' : 'bg-amber-600 text-white hover:bg-amber-500 active:bg-amber-700'">
          {{ loading ? t('login.submit.loading') : t(method === 'sms' ? 'login.submit.sms' : 'login.submit.password') }}
        </button>

        <!-- 密码模式提示 -->
        <p v-if="method === 'password'" class="text-center text-xs mt-2"
          :class="isDark ? 'text-stone-500' : 'text-stone-400'">
          {{ t('login.password.hint') }}
          <button @click="method = 'sms'" class="underline"
            :class="isDark ? 'text-amber-400' : 'text-amber-600'">{{ t('login.password.switchToSms') }}</button>
        </p>

        <!-- 错误信息 -->
        <div v-if="error" class="text-red-500 text-sm mt-3 text-center">{{ error }}</div>

        <!-- 分割线 + 其他方式 -->
        <div class="flex items-center gap-3 my-4">
          <div class="flex-1 h-px" :class="isDark ? 'bg-stone-700' : 'bg-stone-200'"></div>
          <span class="text-xs" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ t('login.or') }}</span>
          <div class="flex-1 h-px" :class="isDark ? 'bg-stone-700' : 'bg-stone-200'"></div>
        </div>

        <button @click="openQR"
          class="w-full py-2.5 rounded-lg text-sm font-medium border transition"
          :class="isDark
            ? 'border-stone-600 text-stone-300 hover:bg-stone-700/50'
            : 'border-stone-300 text-stone-600 hover:bg-stone-100'">
          💬 {{ t('login.method.wechat') }}
        </button>

        <!-- 新用户提示 -->
        <p class="text-center text-xs mt-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">
          {{ t('login.gift') }}
        </p>
      </template>
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

// 方法定义
const methods = [
  { key: 'sms', label: 'login.method.sms' },
  { key: 'password', label: 'login.method.password' },
]
const method = ref('sms')

// 表单字段
const phone = ref('')
const code = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const smsCountdown = ref(0)

// QR 状态
const showQR = ref(false)
const qrStatus = ref('')
const qrError = ref('')
const qrTicket = ref('')
let qrTimer = null
const codeInput = ref(null)

// 提交禁用判断
const submitDisabled = computed(() => {
  if (phone.value.length !== 11) return true
  if (method.value === 'sms') return code.value.length !== 6
  return password.value.length < 6
})

// 手机号输入：只保留数字
function onPhoneInput(e) {
  phone.value = phone.value.replace(/\D/g, '')
}

// 验证码输入：达到6位自动提交
function onCodeInput(e) {
  code.value = code.value.replace(/\D/g, '').slice(0, 6)
  if (code.value.length === 6 && phone.value.length === 11) {
    submit()
  }
}

// 发送短信验证码
async function sendSMS() {
  error.value = ''
  try {
    const res = await fetch('/api/auth/sms-send', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone: phone.value }),
    })
    if (res.ok) {
      smsCountdown.value = 60
      const timer = setInterval(() => {
        smsCountdown.value--
        if (smsCountdown.value <= 0) clearInterval(timer)
      }, 1000)
    } else {
      const data = await res.json()
      error.value = data.error || t('login.network.error')
    }
  } catch (e) {
    error.value = t('login.network.error')
  }
}

// 提交登录
async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (method.value === 'sms') {
      const res = await fetch('/api/auth/sms-login', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: phone.value, code: code.value }),
      })
      const data = await res.json()
      if (res.ok) {
        auth.setAuth(data.token, data.user)
        router.push('/')
      } else {
        error.value = data.error || t('login.network.error')
      }
    } else {
      const res = await fetch('/api/auth/login', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: phone.value, password: password.value }),
      })
      const data = await res.json()
      if (res.ok) {
        auth.setAuth(data.token, data.user)
        router.push('/')
      } else {
        if (res.status === 401) {
          error.value = t('login.error.unauthorized')
        } else {
          error.value = data.error || t('login.network.error')
        }
      }
    }
  } catch (e) {
    error.value = t('login.network.error')
  } finally {
    loading.value = false
  }
}

// ======== 微信扫码 ========

function openQR() {
  showQR.value = true
  error.value = ''
  genQRCode()
}

function closeQR() {
  showQR.value = false
  if (qrTimer) { clearInterval(qrTimer); qrTimer = null }
  qrStatus.value = ''
}

watch(showQR, (val) => {
  if (!val && qrTimer) {
    clearInterval(qrTimer)
    qrTimer = null
  }
})

onUnmounted(() => {
  if (qrTimer) clearInterval(qrTimer)
})

async function genQRCode() {
  qrStatus.value = 'loading'
  qrError.value = ''
  try {
    const res = await fetch('/api/auth/wechat-qrcode')
    const data = await res.json()
    if (!res.ok) {
      qrStatus.value = 'error'
      qrError.value = data.error || t('login.network.error')
      return
    }
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
  } catch (e) {
    qrStatus.value = 'error'
    qrError.value = t('login.network.error')
  }
}

function startPoll() {
  if (qrTimer) clearInterval(qrTimer)
  qrTimer = setInterval(async () => {
    try {
      const res = await fetch('/api/auth/wechat-check?ticket=' + qrTicket.value)
      const data = await res.json()
      if (data.status === 'ok') {
        clearInterval(qrTimer)
        qrTimer = null
        qrStatus.value = 'ok'
        // 扫码成功后需要获取用户信息
        const userRes = await fetch('/api/user', {
          headers: { Authorization: 'Bearer ' + data.token }
        })
        if (userRes.ok) {
          const userData = await userRes.json()
          auth.setAuth(data.token, userData.user || {})
        } else {
          auth.setAuth(data.token, {})
        }
        setTimeout(() => router.push('/'), 1000)
      } else if (data.status === 'expired') {
        clearInterval(qrTimer)
        qrTimer = null
        qrStatus.value = 'expired'
      }
    } catch (e) {
      // ignore poll errors
    }
  }, 2000)
}

function drawQRCode(canvas, url) {
  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.onload = () => {
    const ctx = canvas.getContext('2d')
    ctx.drawImage(img, 0, 0, 200, 200)
  }
  img.src = 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=' + encodeURIComponent(url)
}
</script>
