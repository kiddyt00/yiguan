<template>
  <div class="max-w-lg mx-auto space-y-6">
    <!-- 个人信息 -->
    <div class="glass-card p-6">
      <h3 class="text-xl font-bold mb-4" :class="isDark ? 'text-stone-100' : 'text-stone-800'">{{ t('profile.title') }}</h3>
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">配额</span>
          <span class="font-bold text-lg" :class="quota > 0 ? 'text-amber-600' : 'text-red-500'">{{ quota }} 次</span>
        </div>
        <div><span class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('profile.phone') }}</span><p>{{ user.phone }}</p></div>
        <div>
          <span class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('profile.nickname') }}</span>
          <input v-model="form.nickname" class="w-full border rounded-lg p-2 mt-1 bg-transparent outline-none focus:border-amber-500"
            :class="isDark ? 'text-stone-100 border-stone-600' : 'text-stone-800 border-stone-300'" />
        </div>
        <div>
          <span class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('profile.address') }}</span>
          <input v-model="form.address" :placeholder="t('profile.address.placeholder')" class="w-full border rounded-lg p-2 mt-1 bg-transparent outline-none focus:border-amber-500"
            :class="isDark ? 'text-stone-100 border-stone-600' : 'text-stone-800 border-stone-300'" />
        </div>
        <button @click="save" :disabled="saving"
          class="w-full py-3 rounded-lg font-medium transition bg-amber-600 text-white hover:bg-amber-500">
          {{ saving ? t('profile.saving') : t('profile.save') }}
        </button>
        <router-link to="/recharge" class="block w-full py-2.5 rounded-lg font-medium text-center text-sm transition border"
          :class="isDark ? 'border-stone-600 text-amber-300 hover:bg-stone-800' : 'border-stone-300 text-amber-600 hover:bg-amber-50'">
          💎 去充值
        </router-link>
        <div v-if="msg" class="text-center text-sm text-green-500">{{ msg }}</div>
      </div>
    </div>

    <!-- 充值记录 -->
    <div class="glass-card p-6">
      <h4 class="font-bold mb-3" :class="isDark ? 'text-stone-200' : 'text-stone-700'">📋 充值记录</h4>
      <div v-if="loadingOrders" class="text-sm text-center py-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">加载中...</div>
      <div v-else-if="orders.length === 0" class="text-sm text-center py-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">暂无充值记录</div>
      <div v-else class="space-y-2 max-h-60 overflow-y-auto">
        <div v-for="o in orders" :key="o.id"
          class="flex items-center justify-between p-3 rounded-lg text-sm"
          :class="isDark ? 'bg-stone-800/40' : 'bg-stone-50'">
          <div>
            <span class="font-medium" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ productName(o.product_id) }}</span>
            <span class="mx-1" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ (o.amount / 100).toFixed(0) }}元</span>
            <span class="text-xs" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ formatTime(o.created_at) }}</span>
          </div>
          <span :class="statusClass(o.status)">{{ statusText(o.status) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useI18n } from 'vue-i18n'
import { apiGetJSON, apiPut } from '../utils/request'

defineProps({ isDark: Boolean })

const { t } = useI18n()
const auth = useAuthStore()
const user = computed(() => auth.user || {})
const form = ref({ nickname: '', address: '' })
const saving = ref(false)
const msg = ref('')
const quota = ref(0)
const orders = ref([])
const loadingOrders = ref(false)

function productName(id) {
  return { trial: '尝鲜包', standard: '标准包', unlimited: '畅享包' }[id] || id
}
function statusClass(s) {
  return s === 'paid' ? 'text-green-500' : s === 'pending' ? 'text-amber-500' : 'text-red-500'
}
function statusText(s) {
  return s === 'paid' ? '✅ 已到账' : s === 'pending' ? '⏳ 待支付' : '❌ 失败'
}

onMounted(async () => {
  try {
    const userData = await apiGetJSON('/api/user')
    quota.value = userData.remaining_quota || 0
    form.value.nickname = (userData.user || userData).nickname || ''
    form.value.address = (userData.user || userData).address || ''
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
  const json = await apiPut('/api/user', { nickname: form.value.nickname, address: form.value.address }).then(r => r.json())
  auth.setAuth(auth.token, json.user || json)
  msg.value = t('profile.saved')
  saving.value = false
}
function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>
