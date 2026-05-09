<template>
  <div class="min-h-screen flex flex-col">
    <NavBar :is-dark="true" />
    <main class="flex-1 max-w-2xl mx-auto w-full px-4 py-8">
      <router-view :is-dark="true" />
    </main>
    <footer class="text-center text-stone-500 text-xs py-6 px-4">
      <p>{{ t('disclaimer.1') }}</p>
      <p class="mt-1">{{ t('disclaimer.2') }}</p>
      <p class="mt-1">{{ t('disclaimer.3') }}</p>
    </footer>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import NavBar from './components/NavBar.vue'
</script>
