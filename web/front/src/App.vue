<template>
  <div class="min-h-screen flex flex-col">
    <NavBar :is-dark="isDark" @toggle-theme="isDark = !isDark" />
    <main class="flex-1 max-w-2xl mx-auto w-full px-4 py-8">
      <router-view :is-dark="isDark" />
    </main>
    <footer class="text-center text-xs py-6 px-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'">
      <p>{{ t('disclaimer.1') }}</p>
      <p class="mt-1">{{ t('disclaimer.2') }}</p>
      <p class="mt-1">{{ t('disclaimer.3') }}</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import NavBar from './components/NavBar.vue'

const isDark = ref(localStorage.getItem('theme') !== 'light')

function applyTheme(dark) {
  if (dark) {
    document.documentElement.classList.remove('light')
  } else {
    document.documentElement.classList.add('light')
  }
}

onMounted(() => applyTheme(isDark.value))

watch(isDark, (val) => {
  localStorage.setItem('theme', val ? 'dark' : 'light')
  applyTheme(val)
})
</script>
