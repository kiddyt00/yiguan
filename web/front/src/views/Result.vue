<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
    <h3 class="text-xl font-bold mb-4 text-center">📜 卦象结果</h3>

    <div class="grid grid-cols-2 gap-6 mb-6">
      <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-amber-50'">
        <span class="text-sm opacity-60">本卦</span>
        <div class="text-2xl font-bold mt-1" :class="isDark ? 'text-cyan-400' : 'text-red-900'">{{ data.primary.name }}</div>
        <div class="text-3xl">{{ data.primary.symbol }}</div>
        <p class="text-xs opacity-40 mt-1">{{ data.primary.gua_ci }}</p>
      </div>
      <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700' : 'bg-stone-50'">
        <span class="text-sm opacity-60">变卦</span>
        <div class="text-2xl font-bold mt-1">{{ data.changing.name }}</div>
        <div class="text-3xl">{{ data.changing.symbol }}</div>
        <p class="text-xs opacity-40 mt-1">{{ data.changing.gua_ci }}</p>
      </div>
    </div>

    <!-- 变爻 -->
    <div class="text-center mb-6">
      <span class="text-sm opacity-60">变爻：</span>
      <span v-for="y in data.yao_positions" :key="y.position"
        class="inline-block px-2 py-0.5 rounded text-sm mx-0.5"
        :class="y.is_master ? 'bg-red-700 text-amber-100' : 'bg-red-100 text-red-700'">
        {{ y.label }}{{ y.is_master ? ' ★主' : '' }}
      </span>
    </div>

    <!-- 卦象图 -->
    <Hexagram :lines="hexagramLines" :is-dark="isDark" />

    <!-- AI 解卦 -->
    <div class="border-t pt-6 mt-6" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <h4 class="text-lg font-medium mb-3">🤖 AI 解卦</h4>
      <div class="leading-relaxed whitespace-pre-wrap opacity-80">{{ data.interpretation }}</div>
    </div>

    <!-- 广告 -->
    <div class="bg-stone-300/30 rounded-lg h-16 my-6 flex items-center justify-center text-sm opacity-40">广告位 (结果页)</div>

    <!-- 大师 -->
    <div class="border-t pt-6 text-center" :class="isDark ? 'border-slate-600' : 'border-stone-200'">
      <p class="opacity-60 mb-3">想获得更深入的解读吗？</p>
      <button @click="showQR = !showQR"
        class="px-8 py-3 rounded-lg font-medium transition"
        :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-amber-600 text-white hover:bg-amber-500'">
        🔮 周易大师一对一详解
      </button>
      <div v-if="showQR" class="mt-4">
        <div class="mx-auto w-48 h-48 rounded-lg flex items-center justify-center text-sm opacity-40" :class="isDark ? 'bg-slate-600' : 'bg-stone-300'">
          大师微信二维码
        </div>
        <p class="text-sm opacity-40 mt-2">扫码添加周易大师微信</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import Hexagram from '../components/Hexagram.vue'
const props = defineProps(['data', 'isDark'])
const showQR = ref(false)

const hexagramLines = computed(() => {
  const names = ['上爻', '五爻', '四爻', '三爻', '二爻', '初爻']
  // 根据 primary YaoDesc 构建
  const desc = props.data.primary.yao_desc || ''
  const changing = props.data.yao_positions.map(y => y.position)
  return names.map((name, i) => ({
    label: name,
    yang: desc[5 - i] === '1',
    changing: changing.includes(5 - i),
  }))
})
</script>
