<template>
  <div class="max-w-md mx-auto">
    <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
      <h3 class="text-xl font-bold mb-4">个人信息</h3>
      <div class="space-y-3">
        <div><span class="text-sm opacity-50">手机号</span><p>{{ user.phone }}</p></div>
        <div>
          <span class="text-sm opacity-50">昵称</span>
          <input v-model="form.nickname" class="w-full border rounded-lg p-2 mt-1 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
        </div>
        <div>
          <span class="text-sm opacity-50">地址</span>
          <input v-model="form.address" placeholder="选填" class="w-full border rounded-lg p-2 mt-1 bg-transparent" :class="isDark ? 'border-slate-600' : 'border-stone-300'" />
        </div>
        <button @click="save" :disabled="saving"
          class="w-full py-3 rounded-lg font-medium transition"
          :class="isDark ? 'bg-cyan-600 text-white' : 'bg-red-800 text-amber-100'">
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <div v-if="msg" class="text-center text-sm text-green-500">{{ msg }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const user = computed(() => auth.user || {})
const form = ref({ nickname: '', address: '' })
const saving = ref(false)
const msg = ref('')
const isDark = computed(() => document.documentElement.classList.contains('dark'))

onMounted(async () => {
  const res = await fetch('/api/user', { headers: { Authorization: `Bearer ${auth.token}` } })
  const json = await res.json()
  if (res.ok) {
    const data = json.user || json
    form.value.nickname = data.nickname
    form.value.address = data.address || ''
  }
})

async function save() {
  saving.value = true
  const res = await fetch('/api/user', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${auth.token}` },
    body: JSON.stringify(form.value),
  })
  if (res.ok) {
    const json = await res.json()
    auth.setAuth(auth.token, json.user || json)
    msg.value = '保存成功'
  }
  saving.value = false
}
</script>
