<template>
  <div class="max-w-md mx-auto space-y-4">
    <!-- 配额卡片 -->
    <div class="glass-card p-6 text-center">
      <div class="text-sm mb-1" :class="isDark ? 'text-stone-400' : 'text-stone-500'">剩余占卜次数</div>
      <div class="text-5xl font-bold" :class="quota > 0 ? 'text-amber-600' : 'text-red-500'">{{ quota }}</div>
      <div class="text-xs mt-2" :class="isDark ? 'text-stone-500' : 'text-stone-400'">次</div>
      <router-link to="/recharge" class="inline-block mt-3 px-6 py-2 rounded-lg text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500">
        💎 充值
      </router-link>
    </div>

    <!-- 个人信息 -->
    <div class="glass-card p-6">
      <h3 class="text-base font-bold mb-3" :class="isDark ? 'text-stone-200' : 'text-stone-700'">👤 个人信息</h3>
      <div class="space-y-3 text-sm">
        <div class="flex justify-between py-1 border-b" :class="isDark ? 'border-stone-700' : 'border-stone-100'">
          <span :class="isDark ? 'text-stone-400' : 'text-stone-500'">手机号</span>
          <span :class="isDark ? 'text-stone-200' : 'text-stone-800'">{{ user.phone }}</span>
        </div>
        <div>
          <label class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">昵称</label>
          <input v-model="form.nickname" class="w-full border rounded-lg p-2 mt-1 bg-transparent outline-none focus:border-amber-500 text-sm"
            :class="isDark ? 'text-stone-100 border-stone-600' : 'text-stone-800 border-stone-300'" />
        </div>
        <div>
          <label class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">地址</label>
          <input v-model="form.address" placeholder="选填" class="w-full border rounded-lg p-2 mt-1 bg-transparent outline-none focus:border-amber-500 text-sm"
            :class="isDark ? 'text-stone-100 border-stone-600' : 'text-stone-800 border-stone-300'" />
        </div>
        <button @click="save" :disabled="saving"
          class="w-full py-2 rounded-lg text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500 disabled:opacity-50">
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <div v-if="msg" class="text-center text-xs text-green-500">{{ msg }}</div>
      </div>
    </div>

    <!-- 充值记录 -->
    <div class="glass-card p-6">
      <div class="flex items-center justify-between cursor-pointer" @click="showOrders = !showOrders">
        <h3 class="text-base font-bold" :class="isDark ? 'text-stone-200' : 'text-stone-700'">📋 充值记录</h3>
        <span class="text-xs" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ showOrders ? '收起' : '展开' }}</span>
      </div>
      <div v-if="showOrders">
        <div v-if="loadingOrders" class="text-sm text-center py-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">加载中...</div>
        <div v-else-if="orders.length === 0" class="text-sm text-center py-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">暂无记录</div>
        <div v-else class="space-y-1.5 mt-3">
          <div v-for="o in orders" :key="o.id"
            class="flex items-center justify-between py-2 px-3 rounded-lg text-xs"
            :class="isDark ? 'bg-stone-800/40' : 'bg-stone-50'">
            <div>
              <span class="font-medium" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ productName(o.product_id) }}</span>
              <span class="ml-1" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ (o.amount/100).toFixed(0) }}元</span>
              <span class="ml-2" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ formatTime(o.created_at) }}</span>
            </div>
            <span :class="statusClass(o.status)">{{ statusText(o.status) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { apiGetJSON, apiPut } from '../utils/request'

defineProps({ isDark: Boolean })

const auth = useAuthStore()
const user = auth.user || {}
const form = ref({ nickname: '', address: '' })
const saving = ref(false)
const msg = ref('')
const quota = ref(0)
const orders = ref([])
const loadingOrders = ref(false)
const showOrders = ref(false)

function productName(id) { return { trial: '尝鲜包', standard: '标准包', unlimited: '畅享包' }[id] || id }
function statusClass(s) { return s === 'paid' ? 'text-green-500' : s === 'pending' ? 'text-amber-500' : 'text-red-500' }
function statusText(s) { return s === 'paid' ? '✅ 已到账' : s === 'pending' ? '⏳ 待支付' : '❌ 失败' }
function formatTime(ts) { if (!ts) return ''; return new Date(ts).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) }

onMounted(async () => {
  try {
    const data = await apiGetJSON('/api/user')
    quota.value = data.remaining_quota || 0
    form.value.nickname = (data.user || data).nickname || ''
    form.value.address = (data.user || data).address || ''
  } catch (e) {}
  loadingOrders.value = true
  try {
    const ordData = await apiGetJSON('/api/orders')
    orders.value = ordData.items || []
  } catch (e) {}
  loadingOrders.value = false
})

async function save() {
  saving.value = true
  try {
    const json = await apiPut('/api/user', { nickname: form.value.nickname, address: form.value.address }).then(r => r.json())
    auth.setAuth(auth.token, json.user || json)
    msg.value = '保存成功'
  } catch (e) { msg.value = '保存失败' }
  saving.value = false
}
</script>
