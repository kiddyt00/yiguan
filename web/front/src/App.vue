<template>
  <div class="min-h-screen" :class="{ dark: isDark }">
    <NavBar :is-dark="isDark" @toggle-theme="isDark = !isDark" />
    <main class="max-w-2xl mx-auto px-4 py-8">
      <router-view />
    </main>
    <footer class="text-center text-stone-400 text-sm py-8">
      易观 · AI 算卦 | 仅供娱乐参考
    </footer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import NavBar from './components/NavBar.vue'
const isDark = ref(localStorage.getItem('theme') === 'dark')
</script>
