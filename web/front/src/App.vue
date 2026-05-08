<template>
  <div class="min-h-screen">
    <NavBar :is-dark="isDark" @toggle-theme="toggleTheme" />
    <main class="max-w-2xl mx-auto px-4 py-8">
      <router-view :is-dark="isDark" />
    </main>
    <footer class="text-center text-stone-400 dark:text-slate-500 text-sm py-8">
      易观 · AI 算卦 | 仅供娱乐参考
    </footer>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import NavBar from './components/NavBar.vue'

const isDark = ref(localStorage.getItem('theme') === 'dark')
if (isDark.value) document.documentElement.classList.add('dark')

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
