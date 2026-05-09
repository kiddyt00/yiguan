<template>
  <div class="flex flex-col items-center py-4 mb-6 border-y border-stone-700">
    <div v-for="(line, i) in lines" :key="i" class="flex items-center gap-3 my-1.5">
      <span class="text-xs text-stone-500 w-14 text-right">{{ line.label }}</span>
      <div class="flex gap-1">
        <template v-if="line.yang">
          <span class="block w-20 h-1 rounded bg-stone-200" :class="{ 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]': line.changing }"></span>
        </template>
        <template v-else>
          <span class="block w-9 h-1 rounded bg-stone-200" :class="{ 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]': line.changing }"></span>
          <span class="block w-9 h-1 rounded bg-stone-200" :class="{ 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]': line.changing }"></span>
        </template>
      </div>
      <span v-if="line.changing" class="text-red-400 text-xs font-bold w-14">{{ t('yao.changing') }}</span>
      <span v-else class="w-14"></span>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
defineProps(['lines', 'isDark'])
const { t } = useI18n()
</script>
