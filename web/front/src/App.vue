<template>
  <div class="min-h-screen flex flex-col">
    <NavBar :is-dark="true" />
    <main class="flex-1 max-w-2xl mx-auto w-full px-4 py-8">
      <router-view :is-dark="true" />
    </main>
    <footer class="text-center text-stone-500 text-xs py-6 px-4">
      <p>本占卜仅供娱乐参考，不构成任何专业建议</p>
      <p class="mt-1">请理性看待结果，人生抉择请结合自身实际情况</p>
      <p class="mt-1">如遇重大问题，请咨询相关专业人士</p>
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
