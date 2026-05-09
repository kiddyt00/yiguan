<template>
  <div class="flex flex-col items-center py-4 mb-6 border-y border-stone-700">
    <div v-for="(line, i) in lines" :key="i" class="flex items-center gap-3 my-1.5">
      <span class="text-xs text-stone-500 w-12 text-right">{{ line.label }}</span>
      <div class="flex gap-1">
        <template v-if="line.yang">
          <!-- 阳爻：一条实线 -->
          <span class="block w-20 h-1 rounded bg-stone-200" :class="{ 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]': line.changing }"></span>
        </template>
        <template v-else>
          <!-- 阴爻：两条短虚线 -->
          <span class="block w-9 h-1 rounded bg-stone-200" :class="{ 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]': line.changing }"></span>
          <span class="block w-9 h-1 rounded bg-stone-200" :class="{ 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]': line.changing }"></span>
        </template>
      </div>
      <span v-if="line.changing" class="text-red-400 text-xs font-bold w-10">○ 变</span>
      <span v-else class="w-10"></span>
    </div>
  </div>
</template>

<script setup>
defineProps(['lines', 'isDark'])
</script>
