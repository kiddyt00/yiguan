<template>
  <div class="max-w-2xl mx-auto">
    <h3 class="text-lg font-bold mb-4" :class="isDark ? 'text-stone-100' : 'text-stone-800'">📄 充值记录</h3>

    <div v-if="orders.length === 0 && !loading" class="text-center text-sm py-8" :class="isDark ? 'text-stone-400' : 'text-stone-400'">
      暂无充值记录
    </div>

    <div v-for="o in orders" :key="o.id" class="rounded-xl p-4 mb-3"
      :class="isDark ? 'bg-stone-800/60 border border-stone-700' : 'bg-white border border-stone-100 shadow-sm'">
      <div class="flex items-center justify-between">
        <div>
          <div class="font-semibold text-sm" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ productName(o.product_id) }}</div>
          <div class="text-xs mt-0.5" :class="isDark ? 'text-stone-400' : 'text-stone-400'">{{ fmtTime(o.created_at) }}</div>
        </div>
        <div class="text-right">
          <div class="font-bold text-amber-600">¥{{ (o.amount / 100).toFixed(2) }}</div>
          <div class="text-xs mt-0.5">
            <span :class="statusClass(o.status)">{{ statusLabel(o.status) }}</span>
          </div>
        </div>
      </div>

      <div v-if="o.status === 'paid'" class="mt-3 pt-3 border-t" :class="isDark ? 'border-stone-700' : 'border-stone-100'">
        <button v-if="canRefund(o)" @click="requestRefund(o)" :disabled="refundingId === o.id"
          class="text-xs px-3 py-1.5 rounded-lg border transition"
          :class="isDark ? 'border-stone-600 text-stone-300 hover:bg-stone-700' : 'border-stone-200 text-stone-500 hover:bg-stone-50'">
          {{ refundingId === o.id ? '处理中...' : '申请退款' }}
        </button>
        <span v-else class="text-xs text-stone-400">已过退款期</span>
      </div>
    </div>

    <div class="text-center mt-6">
      <router-link to="/profile" class="text-xs underline" :class="isDark ? 'text-stone-400' : 'text-stone-500'">
        ← 返回个人中心
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiGetJSON, apiPostJSON } from '../utils/request'

defineProps({ isDark: Boolean })

const orders = ref([])
const loading = ref(false)
const refundingId = ref(0)

const productMap = { single: '单次测算', monthly: '月卡', quarterly: '季卡', yearly: '年卡' }
function productName(id) { return productMap[id] || id || '-' }
function fmtTime(t) { return t ? t.replace('T', ' ').slice(0, 16) : '' }
function statusClass(s) {
  if (s === 'paid') return 'text-green-600'
  if (s === 'pending') return 'text-amber-600'
  if (s === 'refunded') return 'text-stone-400'
  return 'text-stone-400'
}
function statusLabel(s) {
  if (s === 'paid') return '已支付'
  if (s === 'pending') return '待支付'
  if (s === 'refunded') return '已退款'
  return s || ''
}

function canRefund(order) {
  if (order.status !== 'paid' || !order.paid_at) return false
  const paidAt = new Date(order.paid_at)
  const now = new Date()
  const hoursDiff = (now - paidAt) / (1000 * 60 * 60)
  return hoursDiff <= 24
}

async function loadOrders() {
  loading.value = true
  try {
    const data = await apiGetJSON('/api/orders')
    orders.value = data.items || []
  } catch (e) { console.error(e) }
  loading.value = false
}

async function requestRefund(order) {
  if (!confirm(`确定要退款吗？\n${productName(order.product_id)} · ¥${(order.amount/100).toFixed(2)}`)) return
  refundingId.value = order.id
  try {
    const data = await apiPostJSON(`/api/orders/${order.id}/refund`, {})
    if (data.refund) {
      alert('✅ 退款申请已提交，等待处理')
      loadOrders()
    }
  } catch (e) {
    alert(e.message || '退款失败')
  }
  refundingId.value = 0
}

onMounted(loadOrders)
</script>