<template>
  <div class="max-w-2xl mx-auto">
    <h3 class="text-xl font-bold mb-2 text-center" :class="isDark ? 'text-stone-100' : 'text-stone-800'">
      💎 充值
    </h3>

    <!-- 会员状态 -->
    <div v-if="membership.is_active" class="text-center mb-4 px-4 py-2 rounded-lg bg-amber-50 border border-amber-200">
      <span class="text-amber-700 font-medium">👑 会员有效 · {{ membership.days_left }}天后到期</span>
    </div>

    <p class="text-sm text-center mb-6" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
      <template v-if="membership.is_active">
        会员期内测算不限次数
      </template>
      <template v-else>
        当前剩余：<span class="font-bold text-amber-600">{{ quota }}</span> 次
      </template>
    </p>

    <!-- 套餐选择 -->
    <div class="grid grid-cols-2 gap-3 mb-6">
      <div v-for="p in products" :key="p.id"
        class="border-2 rounded-xl p-4 text-center cursor-pointer transition-all duration-200 relative"
        :class="selected === p.id
          ? 'border-amber-500 bg-amber-50 shadow-md scale-105'
          : (isDark ? 'border-stone-700 hover:border-amber-600' : 'border-stone-200 hover:border-amber-400')"
        @click="select(p.id)">
        <div v-if="p.badge" class="absolute -top-2 -right-2 bg-red-500 text-white text-xs px-2 py-0.5 rounded-full font-bold">{{ p.badge }}</div>
        <div class="text-2xl mb-1">{{ p.icon }}</div>
        <div class="font-bold" :class="isDark ? 'text-stone-100' : 'text-stone-800'">{{ p.name }}</div>
        <div class="text-2xl font-bold text-amber-600 my-1">{{ price(p.amount) }}<span class="text-sm font-normal">元</span></div>
        <div v-if="p.duration > 0" class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
          {{ p.duration }}天不限次
        </div>
        <div v-else class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
          {{ p.quota }} 次
        </div>
      </div>
    </div>

    <!-- 支付方式 -->
    <div class="mb-6">
      <div class="flex gap-3">
        <button @click="payMethod = 'wechat'"
          class="flex-1 py-3 rounded-xl font-medium border-2 transition"
          :class="payMethod === 'wechat'
            ? 'border-green-500 bg-green-50 text-green-700'
            : (isDark ? 'border-stone-600 text-stone-400' : 'border-stone-200 text-stone-500')">
          💚 微信支付
        </button>
        <button @click="payMethod = 'alipay'"
          class="flex-1 py-3 rounded-xl font-medium border-2 transition"
          :class="payMethod === 'alipay'
            ? 'border-blue-500 bg-blue-50 text-blue-700'
            : (isDark ? 'border-stone-600 text-stone-400' : 'border-stone-200 text-stone-500')">
          💙 支付宝
        </button>
      </div>
    </div>

    <!-- 支付按钮 -->
    <button @click="pay"
      :disabled="!selected || loading"
      class="w-full py-3 rounded-xl font-bold text-lg transition disabled:opacity-40"
      :class="loading
        ? 'bg-amber-500 text-white cursor-wait'
        : 'bg-amber-600 text-white hover:bg-amber-500 active:bg-amber-700'">
      {{ loading ? '处理中...' : (payMethod === 'wechat' ? '微信支付 ' : '支付宝 ') + (selectedProduct ? '¥' + price(selectedProduct.amount) : '') }}
    </button>

    <!-- 微信二维码弹窗 -->
    <div v-if="showQR" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeQR">
      <div class="bg-white rounded-2xl p-8 text-center max-w-sm mx-4 shadow-2xl">
        <div class="text-lg font-bold text-stone-800 mb-1">微信支付</div>
        <div class="text-sm text-stone-500 mb-4">{{ selectedProduct?.name }} · ¥{{ selectedProduct ? price(selectedProduct.amount) : '' }}</div>
        <div id="pay-qrcode" class="inline-block bg-white p-3 rounded-xl border"></div>
        <p class="text-sm text-stone-500 mt-4">请使用微信扫描二维码支付</p>
        <p class="text-xs text-stone-400 mt-1">支付后自动到账，请勿关闭页面</p>
        <div v-if="qrStatus === 'paid'" class="mt-4 text-green-600 font-bold text-lg">✅ 支付成功！</div>
        <button @click="closeQR" class="mt-4 text-sm text-amber-600 underline">关闭</button>
      </div>
    </div>

    <p class="text-xs text-center mt-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">
      支付即代表同意《服务协议》，充值次数永久有效
    </p>

    <div class="flex justify-center gap-6 mt-4 text-xs">
      <router-link to="/profile" class="underline" :class="isDark ? 'text-stone-400 hover:text-stone-300' : 'text-stone-500 hover:text-stone-600'">
        ← 个人信息
      </router-link>
      <router-link to="/profile#invite" class="underline" :class="isDark ? 'text-amber-400 hover:text-amber-300' : 'text-amber-600 hover:text-amber-500'">
        📤 邀请好友赚次数
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useAuthStore } from '../stores/auth'
import { apiGetJSON, apiPostJSON } from '../utils/request'

const props = defineProps({ isDark: Boolean })
const auth = useAuthStore()

function price(amount) { return (amount / 100).toString() }

const products = [
  { id: 'single',    name: '单次测算',   quota: 1,   amount: 990,   duration: 0,   icon: '🔮' },
  { id: 'monthly',   name: '月卡',       quota: -1,  amount: 2990,  duration: 30,  icon: '📅' },
  { id: 'quarterly', name: '季卡',       quota: -1,  amount: 4990,  duration: 90,  icon: '🌿' },
  { id: 'yearly',    name: '年卡',       quota: -1,  amount: 9900,  duration: 365, icon: '👑', badge: '最超值' },
]

const selected = ref('')
const payMethod = ref('wechat')
const loading = ref(false)
const showQR = ref(false)
const qrStatus = ref('')
const orderId = ref(0)
const quota = ref(0)
const membership = ref({ is_active: false, days_left: 0 })
const selectedProduct = computed(() => products.find(p => p.id === selected.value))

function select(id) { selected.value = id }

onMounted(async () => {
  try {
    const [userData, membershipData] = await Promise.all([
      apiGetJSON('/api/user'),
      apiGetJSON('/api/user/membership').catch(() => null),
    ])
    quota.value = userData.remaining_quota || 0
    if (membershipData) membership.value = membershipData
  } catch (e) {}
})

async function pay() {
  if (!selected.value) return
  loading.value = true
  try {
    if (payMethod.value === 'wechat') {
      await payWechat()
    } else {
      await payAlipay()
    }
  } catch (e) {
    alert('网络错误: ' + e.message)
  } finally {
    loading.value = false
  }
}

async function payWechat() {
  const data = await apiPostJSON('/api/orders/create', { product_id: selected.value })
  const codeUrl = data.code_url || (data.order && data.order.code_url)
  if (codeUrl) {
    orderId.value = data.id || (data.order && data.order.id)
    showQR.value = true
    qrStatus.value = 'pending'
    await nextTick()
    drawQR(codeUrl)
    startPoll()
  } else {
    alert('创建订单失败')
  }
}

async function payAlipay() {
  const data = await apiPostJSON('/api/orders/alipay-create', { product_id: selected.value })
  const payURL = data.pay_url || (data.order && data.order.pay_url)
  if (payURL) {
    // 打开支付宝收银台页面
    window.open(payURL, '_blank')
    // 保存订单号用于轮询
    orderId.value = data.id || (data.order && data.order.id)
    // 开始轮询订单状态
    startAlipayPoll()
  } else if (payURL === '' || payURL === undefined) {
    alert('支付宝暂未配置')
  } else {
    alert('创建订单失败')
  }
}

function drawQR(codeUrl) {
  const el = document.getElementById('pay-qrcode')
  if (!el) return
  try {
    new window.QRCode(el, { text: codeUrl, width: 200, height: 200 })
  } catch(e) {
    el.textContent = '二维码加载失败'
  }
}

function startPoll() {
  const timer = setInterval(async () => {
    try {
      const data = await apiGetJSON('/api/orders/' + orderId.value)
      if (data.order) {
        if (data.order.status === 'paid') {
          clearInterval(timer)
          qrStatus.value = 'paid'
          const userData = await apiGetJSON('/api/user')
          quota.value = userData.remaining_quota || 0
        }
      } else if (data.status === 'paid') {
        clearInterval(timer)
        qrStatus.value = 'paid'
        // 刷新quota和会员状态
        const [ud, md] = await Promise.all([
          apiGetJSON('/api/user'),
          apiGetJSON('/api/user/membership').catch(() => null),
        ])
        quota.value = ud.remaining_quota || 0
        if (md) membership.value = md
      }
    } catch (e) {}
  }, 2000)
  setTimeout(() => { clearInterval(timer); if (qrStatus.value === 'pending') qrStatus.value = 'expired' }, 600000)
}

function startAlipayPoll() {
  const timer = setInterval(async () => {
    try {
      const data = await apiGetJSON('/api/orders/' + orderId.value)
      let status = data.status
      if (data.order) {
        status = data.order.status
      }
      if (status === 'paid') {
        clearInterval(timer)
        qrStatus.value = 'paid'
        const userData = await apiGetJSON('/api/user')
        quota.value = userData.remaining_quota || 0
        alert('✅ 支付成功！次数已到账')
      }
    } catch (e) {}
  }, 3000)
  // 30分钟后停止
  setTimeout(() => { clearInterval(timer) }, 1800000)
}

function closeQR() {
  showQR.value = false
  qrStatus.value = ''
}

</script>
