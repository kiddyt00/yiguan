<template>
  <header class="shadow-lg bg-transparent">
    <div class="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
      <router-link to="/" class="flex items-center gap-2">
        <img src="/logo-icon.svg" alt="观己斋" class="h-8 w-auto" />
      </router-link>
      <div class="flex items-center gap-3 text-sm">
        <!-- 主题切换 -->
        <button @click="$emit('toggle-theme')" class="text-lg leading-none opacity-70 hover:opacity-100 transition"
          :title="isDark ? '切换浅色' : '切换深色'">
          {{ isDark ? '☀️' : '🌙' }}
        </button>
        <!-- 语言切换 -->
        <div class="relative">
          <button @click="langOpen = !langOpen" class="text-xs px-2 py-0.5 rounded-full font-medium flex items-center gap-1 transition"
            :class="isDark ? 'bg-amber-500/20 text-amber-300 hover:bg-amber-500/30' : 'bg-amber-100 text-amber-700 hover:bg-amber-200'">
            {{ locale === 'zh' ? '中文' : 'EN' }}
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
          </button>
          <div v-if="langOpen" class="absolute top-8 right-0 border rounded-lg shadow-xl z-50 overflow-hidden"
            :class="isDark ? 'bg-slate-800 border-stone-700' : 'bg-white border-stone-200'">
            <button v-for="opt in [{v:'zh',l:'中文'},{v:'en',l:'English'}]" :key="opt.v"
              @click="setLocale(opt.v)"
              class="block w-full text-left px-4 py-2 text-sm hover:bg-amber-500/20 transition"
              :class="locale === opt.v ? 'text-amber-600 bg-amber-50 font-medium' : (isDark ? 'text-stone-200' : 'text-stone-700')">
              {{ opt.l }}
            </button>
          </div>
        </div>
        <template v-if="auth.isLoggedIn()">
          <span v-if="quota !== null" class="text-xs px-2 py-0.5 rounded-full"
            :class="isDark ? 'bg-amber-500/20 text-amber-300' : 'bg-amber-100 text-amber-700'">
            {{ quota > 0 ? t('quota.remaining', { n: quota }) : t('quota.depleted') }}
          </span>
          <router-link to="/history" class="transition"
            :class="isDark ? 'text-gray-300 hover:text-amber-300' : 'text-stone-600 hover:text-amber-600'">{{ t('nav.history') }}</router-link>
          <router-link to="/profile" class="transition"
            :class="isDark ? 'text-gray-300 hover:text-amber-300' : 'text-stone-600 hover:text-amber-600'">{{ auth.user?.nickname || t('nav.profile') }}</router-link>
          <button @click="doLogout" class="transition"
            :class="isDark ? 'text-gray-400 hover:text-amber-300' : 'text-stone-400 hover:text-amber-600'">{{ t('nav.logout') }}</button>
        </template>
        <template v-else>
          <router-link to="/login" class="transition"
            :class="isDark ? 'text-gray-300 hover:text-amber-300' : 'text-stone-600 hover:text-amber-600'">{{ t('nav.login') }}</router-link>
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
defineEmits(['toggle-theme'])
const { t, locale } = useI18n()
const auth = useAuthStore()
const $router = useRouter()
const quota = ref(null)
const langOpen = ref(false)

function setLocale(v) {
  locale.value = v
  localStorage.setItem('lang', v)
  langOpen.value = false
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
