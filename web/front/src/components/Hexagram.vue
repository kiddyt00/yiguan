<template>
  <div class="flex flex-col items-center py-4 mb-6 border-y" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
    <div v-for="(line, i) in lines" :key="i" class="flex items-center gap-3 my-1.5">
      <span class="text-xs opacity-40 w-12 text-right">{{ line.label }}</span>
      <div class="flex gap-1">
        <template v-if="line.yang">
          <span class="block w-20 h-2 rounded" :class="isDark ? 'bg-cyan-500' : 'bg-red-800'"></span>
        </template>
        <template v-else>
          <span class="block w-9 h-2 rounded" :class="isDark ? 'bg-slate-400' : 'bg-stone-700'"></span>
          <span class="block w-9 h-2 rounded" :class="isDark ? 'bg-slate-400' : 'bg-stone-700'"></span>
        </template>
      </div>
      <span v-if="line.changing" class="text-red-500 text-xs font-bold w-10">○ 变</span>
      <span v-else class="w-10"></span>
    </div>
  </div>
</template>

<script setup>
defineProps(['lines', 'isDark'])
</script>
