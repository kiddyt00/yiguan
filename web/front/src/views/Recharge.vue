<template>
  <div class="max-w-2xl mx-auto">
    <h3 class="text-xl font-bold mb-2 text-center" :class="isDark ? 'text-stone-100' : 'text-stone-800'">
      💎 充值占卜次数
    </h3>
    <p class="text-sm text-center mb-6" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
      当前剩余：<span class="font-bold text-amber-600">{{ quota }}</span> 次
    </p>

    <!-- 套餐选择 -->
    <div class="grid grid-cols-3 gap-3 mb-6">
      <div v-for="p in products" :key="p.id"
        class="border-2 rounded-xl p-4 text-center cursor-pointer transition-all duration-200"
        :class="selected === p.id
          ? 'border-amber-500 bg-amber-50 shadow-md scale-105'
          : (isDark ? 'border-stone-700 hover:border-amber-600' : 'border-stone-200 hover:border-amber-400')"
        @click="select(p.id)">
        <div class="text-2xl mb-1">{{ p.icon }}</div>
        <div class="font-bold text-lg" :class="isDark ? 'text-stone-100' : 'text-stone-800'">{{ p.name }}</div>
        <div class="text-2xl font-bold text-amber-600 my-1">{{ price(p.amount) }}<span class="text-sm font-normal">元</span></div>
        <div class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ p.quota }} 次</div>
        <div class="text-xs mt-1" :class="isDark ? 'text-stone-500' : 'text-stone-400'">≈ {{ (p.amount / p.quota).toFixed(1) }}分/次</div>
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
        <button disabled
          class="flex-1 py-3 rounded-xl font-medium border-2 opacity-40 cursor-not-allowed"
          :class="isDark ? 'border-stone-600 text-stone-500' : 'border-stone-200 text-stone-400'">
          💙 支付宝（即将上线）
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
      {{ loading ? '处理中...' : '微信支付 ' + (selectedProduct ? '¥' + price(selectedProduct.amount) : '') }}
    </button>

    <!-- 二维码弹窗 -->
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

    <div class="text-center mt-4">
      <router-link to="/profile" class="text-xs underline" :class="isDark ? 'text-stone-400 hover:text-stone-300' : 'text-stone-500 hover:text-stone-600'">
        ← 查看充值记录和个人信息
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { apiGetJSON, apiPostJSON } from '../utils/request'

const props = defineProps({ isDark: Boolean })
const auth = useAuthStore()

function price(amount) { return (amount / 100).toFixed(amount < 100 ? 2 : 0) }

const products = [
  { id: 'test', name: '测试包', quota: 1, amount: 1, icon: '🧪' },
  { id: 'trial', name: '尝鲜包', quota: 10, amount: 500, icon: '🔮' },
  { id: 'standard', name: '标准包', quota: 50, amount: 2000, icon: '🌟' },
  { id: 'unlimited', name: '畅享包', quota: 200, amount: 6000, icon: '👑' },
]

const selected = ref('')
const payMethod = ref('wechat')
const loading = ref(false)
const showQR = ref(false)
const qrStatus = ref('')
const orderId = ref(0)
const quota = ref(0)
const selectedProduct = computed(() => products.find(p => p.id === selected.value))

function select(id) { selected.value = id }

onMounted(async () => {
  try {
    const userData = await apiGetJSON('/api/user')
    quota.value = userData.remaining_quota || 0
  } catch (e) {}
})

async function pay() {
  if (!selected.value) return
  loading.value = true
  try {
    const data = await apiPostJSON('/api/orders/create', { product_id: selected.value })
    if (data.code_url) {
      orderId.value = data.id
      showQR.value = true
      qrStatus.value = 'pending'
      await drawQR(data.code_url)
      startPoll()
    } else {
      alert('创建订单失败')
    }
  } catch (e) {
    alert('网络错误: ' + e.message)
  } finally {
    loading.value = false
  }
}

async function drawQR(url) {
  const container = document.getElementById('pay-qrcode')
  if (!container) return
  container.innerHTML = ''
  // 使用 qrcode 生成（已在前端项目中）
  try {
    const QRCode = (await import('qrcode')).default
    const canvas = document.createElement('canvas')
    await QRCode.toCanvas(canvas, url, { width: 200, margin: 2 })
    container.appendChild(canvas)
  } catch {
    // fallback: 使用图片
    const img = document.createElement('img')
    img.src = 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=' + encodeURIComponent(url)
    img.width = 200
    img.height = 200
    container.appendChild(img)
  }
}

function startPoll() {
  const timer = setInterval(async () => {
    try {
      const data = await apiGetJSON('/api/orders/' + orderId.value)
      if (data.status === 'paid') {
        clearInterval(timer)
        qrStatus.value = 'paid'
        // 刷新quota和订单记录
        const userData = await apiGetJSON('/api/user')
        quota.value = userData.remaining_quota || 0

      }
    } catch (e) {}
  }, 2000)
  // 10分钟后停止轮询
  setTimeout(() => { clearInterval(timer); if (qrStatus.value === 'pending') qrStatus.value = 'expired' }, 600000)
}

function closeQR() {
  showQR.value = false
  qrStatus.value = ''
}

</script>
