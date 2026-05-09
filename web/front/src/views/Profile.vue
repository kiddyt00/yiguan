<template>
  <div class="max-w-md mx-auto">
    <div class="glass-card p-6">
      <h3 class="text-xl font-bold mb-4 text-stone-100">{{ t('profile.title') }}</h3>
      <div class="space-y-3">
        <div><span class="text-sm text-stone-400">{{ t('profile.phone') }}</span><p>{{ user.phone }}</p></div>
        <div>
          <span class="text-sm text-stone-400">{{ t('profile.nickname') }}</span>
          <input v-model="form.nickname" class="w-full border rounded-lg p-2 mt-1 bg-transparent text-stone-100 border-stone-600 focus:border-amber-500 outline-none" />
        </div>
        <div>
          <span class="text-sm text-stone-400">{{ t('profile.address') }}</span>
          <input v-model="form.address" :placeholder="t('profile.address.placeholder')" class="w-full border rounded-lg p-2 mt-1 bg-transparent text-stone-100 border-stone-600 focus:border-amber-500 outline-none" />
        </div>
        <button @click="save" :disabled="saving"
          class="w-full py-3 rounded-lg font-medium transition bg-amber-600 text-white hover:bg-amber-500">
          {{ saving ? t('profile.saving') : t('profile.save') }}
        </button>
        <div v-if="msg" class="text-center text-sm text-green-500">{{ msg }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const auth = useAuthStore()
const user = computed(() => auth.user || {})
const form = ref({ nickname: '', address: '' })
const saving = ref(false)
const msg = ref('')

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
    msg.value = t('profile.saved')
  }
  saving.value = false
}
</script>
