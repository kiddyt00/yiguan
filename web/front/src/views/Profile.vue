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
        <button @click="save" :disabled="saving"
          class="w-full py-2.5 rounded-xl text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500 disabled:opacity-50">
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <p v-if="msg" class="text-center text-xs text-green-500">{{ msg }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { apiGetJSON, apiPutJSON } from '../utils/request'

defineProps({ isDark: Boolean })
const auth = useAuthStore()
const user = auth.user || {}
const form = ref({ nickname: '' })
const saving = ref(false)
const msg = ref('')
const quota = ref(0)

onMounted(async () => {
  try {
    const data = await apiGetJSON('/api/user')
    quota.value = data.remaining_quota || 0
    form.value.nickname = (data.user || data).nickname || ''
  } catch (e) {}
})

async function save() {
  saving.value = true
  try {
    const json = await apiPutJSON('/api/user', { nickname: form.value.nickname })
    auth.setAuth(auth.token, json.user || json)
    msg.value = '保存成功'
    setTimeout(() => msg.value = '', 2000)
  } catch (e) { msg.value = '保存失败' }
  saving.value = false
}
</script>
