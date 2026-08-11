<template>
  <div class="min-h-screen flex flex-col">
    <NavBar :is-dark="isDark" @toggle-theme="isDark = !isDark" />
    <main class="flex-1 max-w-2xl mx-auto w-full px-4 py-8">
      <router-view :is-dark="isDark" />
    </main>
    <footer class="text-center py-6 px-4" :class="isDark ? 'text-stone-500' : 'text-stone-400'" style="font-size: 0.7rem;">
      <div class="space-x-3">
        <a href="/about.html" class="hover:text-amber-500 transition">关于我们</a>
        <a href="/contact.html" class="hover:text-amber-500 transition">联系我们</a>
        <a href="/terms.html" class="hover:text-amber-500 transition">服务协议</a>
        <a href="/privacy.html" class="hover:text-amber-500 transition">隐私政策</a>
      </div>
      <p class="mt-2">{{ t('disclaimer.1') }}</p>
      <p class="mt-1">{{ t('disclaimer.2') }}</p>
      <p class="mt-1">{{ t('disclaimer.3') }}</p>
      <div class="mt-3">© 2026 北京丰弥科技有限公司 All Rights Reserved.</div>
      <div class="mt-1"><a href="https://beian.miit.gov.cn/" target="_blank" rel="nofollow" class="hover:text-amber-500 transition">京ICP备2026035156号-1</a></div>
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
