<template>
  <header class="shadow-lg bg-transparent">
    <div class="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
      <router-link to="/" class="flex items-center gap-2">
        <img src="../assets/logo-guanjizhai.png" alt="观己斋" class="h-9 w-auto brightness-0 invert" />
      </router-link>
      <div class="flex items-center gap-3 text-sm">
        <button @click="toggleLang" class="text-xs px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 hover:bg-amber-500/30 transition font-medium">
          {{ t('nav.lang.switch') }}
        </button>
        <template v-if="auth.isLoggedIn()">
          <span v-if="quota !== null" class="text-xs px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300">
            {{ quota }}
          </span>
          <router-link to="/history" class="text-gray-300 hover:text-amber-300 transition">{{ t('nav.history') }}</router-link>
          <router-link to="/profile" class="text-gray-300 hover:text-amber-300 transition">{{ auth.user?.nickname || t('nav.profile') }}</router-link>
          <button @click="doLogout" class="text-gray-400 hover:text-amber-300 transition">{{ t('nav.logout') }}</button>
        </template>
        <template v-else>
          <router-link to="/login" class="text-gray-300 hover:text-amber-300 transition">{{ t('nav.login') }}</router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
defineProps(['isDark'])
const { t, locale } = useI18n()
const auth = useAuthStore()
const $router = useRouter()
const quota = ref(null)

function toggleLang() {
  locale.value = locale.value === 'zh' ? 'en' : 'zh'
  localStorage.setItem('lang', locale.value)
}

onMounted(async () => {
  if (auth.isLoggedIn()) {
    try {
      const res = await fetch('/api/user', {
        headers: { Authorization: `Bearer ${auth.token}` }
      })
      if (res.ok) {
        const data = await res.json()
        quota.value = data.remaining_quota
      } else if (res.status === 401) {
        auth.logout()
        $router.push('/login')
      }
    } catch {}
  }
})

function doLogout() {
  auth.logout()
  $router.push('/')
}
</script>
