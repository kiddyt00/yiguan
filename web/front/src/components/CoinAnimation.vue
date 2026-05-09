<template>
  <div class="text-center py-8">
    <!-- 当前爻位提示 -->
    <p class="text-lg mb-6 text-green-400">
      正在摇卦... 第{{ currentThrow }}爻（{{ yaoName }}）
    </p>

    <!-- 三枚铜钱 -->
    <div class="flex justify-center gap-6 mb-8">
      <div v-for="(coin, i) in displayCoins" :key="i"
        class="w-20 h-20 rounded-full flex items-center justify-center text-2xl font-bold transition-all duration-500 coin-glow"
        :class="{
          'animate-bounce': isAnimating,
          'bg-gradient-to-br from-amber-400 via-amber-500 to-amber-700 text-amber-950': true
        }"
        :style="{ animationDelay: `${i * 0.15}s`, animationDuration: '0.6s' }">
        {{ coin }}
      </div>
    </div>

    <!-- 进度条 1-6 -->
    <div class="flex justify-center gap-3 mb-4">
      <span v-for="n in 6" :key="n"
        class="w-9 h-9 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300"
        :class="n <= currentThrow
          ? 'step-active'
          : 'step-inactive'">
        {{ n }}
      </span>
    </div>

    <!-- 底部提示 -->
    <p class="text-sm text-stone-400">{{ statusText }}</p>

    <!-- 计算结果展示 -->
    <div v-if="showResult" class="mt-4 text-center">
      <p class="text-base font-medium text-amber-300">
        三枚铜钱：{{ coinValues.join('+') }} = {{ sum }} -> {{ resultType }}
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
  return props.coinValues.map(v => v === null ? '?' : (v === 2 ? '正' : '反'))
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
  if (sum.value > 0) return '正在解卦...'
  return '准备起卦...'
})
</script>
