<template>
  <div class="min-h-screen">
    <NavBar :is-dark="isDark" @toggle-theme="toggleTheme" />
    <main class="max-w-2xl mx-auto px-4 py-8">
      <router-view :is-dark="isDark" />
    </main>
    <footer class="text-center text-stone-400 dark:text-slate-500 text-xs py-6 px-4">
      本小程序旨在推广中国传统文化，所有内容基于《周易》公开文献整理，无预测功能，结果仅供学习交流。
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
