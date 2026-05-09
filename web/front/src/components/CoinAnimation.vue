<template>
  <div class="text-center py-8">
    <!-- 当前爻位提示 -->
    <p class="text-lg mb-6" :class="isDark ? 'text-cyan-300' : 'text-amber-700'">
      正在摇卦... 第{{ currentThrow }}爻（{{ yaoName }}）
    </p>

    <!-- 三枚铜钱 -->
    <div class="flex justify-center gap-6 mb-8">
      <div v-for="(coin, i) in displayCoins" :key="i"
        class="w-20 h-20 rounded-full flex items-center justify-center text-2xl font-bold shadow-lg transition-all duration-500"
        :class="{
          'animate-bounce': isAnimating,
          'bg-gradient-to-br from-amber-500 to-amber-700 text-amber-100': !isDark,
          'bg-gradient-to-br from-cyan-600 to-cyan-800 text-cyan-100': isDark
        }"
        :style="{ animationDelay: `${i * 0.15}s`, animationDuration: '0.6s' }">
        {{ coin }}
      </div>
    </div>

    <!-- 进度条 1-6 -->
    <div class="flex justify-center gap-3 mb-4">
      <span v-for="n in 6" :key="n"
        class="w-9 h-9 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300"
        :class="n < currentThrow
          ? (isDark ? 'bg-cyan-600 text-white' : 'bg-red-800 text-amber-100')
          : n === currentThrow
            ? (isDark ? 'bg-cyan-500 text-white ring-2 ring-cyan-300' : 'bg-red-700 text-amber-100 ring-2 ring-red-400')
            : (isDark ? 'bg-slate-700 text-slate-500' : 'bg-stone-200 text-stone-400')">
        {{ n }}
      </span>
    </div>

    <!-- 底部提示 -->
    <p class="text-sm opacity-50">{{ statusText }}</p>

    <!-- 计算结果展示 -->
    <div v-if="showResult" class="mt-4 text-center">
      <p class="text-base font-medium" :class="isDark ? 'text-cyan-400' : 'text-red-800'">
        {{ coinValues.join(' + ') }} = {{ sum }} → {{ resultType }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  currentThrow: { type: Number, default: 1 },
  isAnimating: { type: Boolean, default: true },
  coinValues: { type: Array, default: () => [null, null, null] },
  isDark: { type: Boolean, default: false }
})

const yaoNames = ['初爻', '二爻', '三爻', '四爻', '五爻', '上爻']
const yaoName = computed(() => yaoNames[props.currentThrow - 1] || '')

const displayCoins = computed(() => {
  return props.coinValues.map(v => v === null ? '?' : (v === 2 ? '字' : '背'))
})

const sum = computed(() => props.coinValues.reduce((a, b) => a + (b || 0), 0))

const resultType = computed(() => {
  const s = sum.value
  if (s === 0) return ''
  return s === 6 ? '老阴' : s === 7 ? '少阳' : s === 8 ? '少阴' : '老阳'
})

const showResult = computed(() => !props.isAnimating && sum.value > 0)

const statusText = computed(() => {
  if (props.isAnimating) return '正在摇铜钱...'
  if (sum.value > 0) return `第${props.currentThrow}爻已定`
  return '准备起卦...'
})
</script>
