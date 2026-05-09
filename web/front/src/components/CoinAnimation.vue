<template>
  <div class="text-center py-8">
    <!-- 当前爻位提示 -->
    <p class="text-lg mb-6 text-green-400">
      {{ t('coin.divining', { n: currentThrow, name: yaoName }) }}
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
        :class="n <= currentThrow ? 'step-active' : 'step-inactive'">
        {{ n }}
      </span>
    </div>

    <!-- 底部提示 -->
    <p class="text-sm text-stone-400">{{ statusText }}</p>

    <!-- 计算结果展示 -->
    <div v-if="showResult" class="mt-4 text-center">
      <p class="text-base font-medium text-amber-300">
        {{ t('coin.back') }}/{{ t('coin.front') }}: {{ coinValues.join('+') }} = {{ sum }} → {{ t(`gua.${resultType}`) }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const props = defineProps({
  currentThrow: { type: Number, default: 1 },
  isAnimating: { type: Boolean, default: true },
  coinValues: { type: Array, default: () => [null, null, null] },
  isDark: { type: Boolean, default: false }
})

const yaoNames = computed(() => [t('yao.1'), t('yao.2'), t('yao.3'), t('yao.4'), t('yao.5'), t('yao.6')])
const yaoName = computed(() => yaoNames.value[props.currentThrow - 1] || '')

const displayCoins = computed(() => {
  return props.coinValues.map(v => v === null ? '?' : (v === 2 ? t('coin.front') : t('coin.back')))
})

const sum = computed(() => props.coinValues.reduce((a, b) => a + (b || 0), 0))

const resultType = computed(() => {
  const s = sum.value
  if (s === 0) return ''
  return s === 6 ? 'old_yin' : s === 7 ? 'young_yang' : s === 8 ? 'young_yin' : 'old_yang'
})

const showResult = computed(() => !props.isAnimating && sum.value > 0)

const statusText = computed(() => {
  if (props.isAnimating) return t('coin.shaking')
  if (sum.value > 0) return t('coin.done', { n: props.currentThrow })
  return t('coin.ready')
})
</script>
