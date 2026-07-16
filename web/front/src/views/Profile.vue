<template>
  <div class="max-w-sm mx-auto space-y-5">
    <!-- 配额 -->
    <div class="text-center py-8 px-6 rounded-2xl" :class="isDark ? 'bg-stone-800/60' : 'bg-gradient-to-br from-amber-50 to-white border border-amber-100'">
      <div class="text-xs font-medium tracking-widest uppercase" :class="isDark ? 'text-stone-400' : 'text-stone-400'">剩余次数</div>
      <div class="mt-2 text-6xl font-extrabold" :class="quota > 0 ? 'text-amber-600' : 'text-red-400'">{{ quota }}</div>
      <div class="mt-1 text-xs" :class="isDark ? 'text-stone-500' : 'text-stone-400'">次</div>
      <router-link to="/recharge" class="inline-flex items-center gap-1.5 mt-4 px-6 py-2.5 rounded-xl text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500 shadow-lg shadow-amber-600/20">
        💎 充值
      </router-link>
    </div>

    <!-- 个人信息 -->
    <div class="rounded-2xl p-6" :class="isDark ? 'bg-stone-800/60' : 'bg-white border border-stone-100'">
      <h3 class="text-sm font-bold mb-4" :class="isDark ? 'text-stone-200' : 'text-stone-700'">个人信息</h3>
      <div class="space-y-3.5 text-sm">
        <div class="flex items-center justify-between py-2 px-3 rounded-lg" :class="isDark ? 'bg-stone-700/40' : 'bg-stone-50'">
          <span class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">手机号</span>
          <span class="font-medium" :class="isDark ? 'text-stone-200' : 'text-stone-800'">{{ user.phone }}</span>
        </div>
        <div>
          <label class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">昵称</label>
          <input v-model="form.nickname"
            class="w-full mt-1 px-3 py-2 rounded-lg border bg-transparent outline-none transition text-sm focus:border-amber-500 focus:ring-1 focus:ring-amber-500/20"
            :class="isDark ? 'border-stone-600 text-stone-200' : 'border-stone-200 text-stone-800'" />
        </div>
        <div>
          <label class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">地址</label>
          <input v-model="form.address" placeholder="选填"
            class="w-full mt-1 px-3 py-2 rounded-lg border bg-transparent outline-none transition text-sm focus:border-amber-500 focus:ring-1 focus:ring-amber-500/20"
            :class="isDark ? 'border-stone-600 text-stone-200' : 'border-stone-200 text-stone-800'" />
        </div>
        <button @click="save" :disabled="saving"
          class="w-full py-2.5 rounded-xl text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500 disabled:opacity-50">
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <p v-if="msg" class="text-center text-xs text-green-500">{{ msg }}</p>
      </div>
    </div>

    <!-- 充值记录（简化） -->
    <div class="rounded-2xl p-6" :class="isDark ? 'bg-stone-800/60' : 'bg-white border border-stone-100'">
      <div class="flex items-center justify-between cursor-pointer" @click="showOrders = !showOrders">
        <h3 class="text-sm font-bold" :class="isDark ? 'text-stone-200' : 'text-stone-700'">充值记录</h3>
        <span class="text-xs" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ showOrders ? '收起' : '展开' }}</span>
      </div>
      <div v-if="showOrders" class="mt-3 space-y-1">
        <div v-if="orders.length === 0" class="text-xs text-center py-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">暂无记录</div>
        <div v-for="o in orders" :key="o.id"
          class="flex items-center justify-between py-2 px-3 rounded-lg text-xs"
          :class="isDark ? 'bg-stone-700/40' : 'bg-stone-50'">
          <div>
            <span class="font-medium" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ {trial:'尝鲜包',standard:'标准包',unlimited:'畅享包'}[o.product_id] || o.product_id }}</span>
            <span class="ml-1.5" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ (o.amount/100).toFixed(0) }}元</span>
            <span class="ml-2 text-2xs opacity-60" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ formatTime(o.created_at) }}</span>
          </div>
          <span :class="o.status==='paid'?'text-green-500':o.status==='pending'?'text-amber-500':'text-red-400'">
            {{ o.status==='paid'?'已到账':o.status==='pending'?'待支付':'失败' }}
          </span>
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
const showOrders = ref(false)

function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  try {
    const data = await apiGetJSON('/api/user')
    quota.value = data.remaining_quota || 0
    form.value.nickname = (data.user || data).nickname || ''
    form.value.address = (data.user || data).address || ''
  } catch (e) {}
  try {
    const d = await apiGetJSON('/api/orders')
    orders.value = d.items || []
  } catch (e) {}
})

async function save() {
  saving.value = true
  try {
    const json = await apiPut('/api/user', { nickname: form.value.nickname, address: form.value.address }).then(r => r.json())
    auth.setAuth(auth.token, json.user || json)
    msg.value = '保存成功'
    setTimeout(() => msg.value = '', 2000)
  } catch (e) { msg.value = '保存失败' }
  saving.value = false
}
</script>
