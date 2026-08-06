<template>
  <div class="text-center py-8">
    <!-- 当前爻位提示 -->
    <p class="text-lg mb-6" :class="isDark ? 'text-green-400' : 'text-green-600'">
      {{ t('coin.divining', { n: currentThrow, name: yaoName }) }}
    </p>

    <!-- 三枚铜钱(GSAP 抛掷) -->
    <div class="flex justify-center gap-6 mb-8" style="perspective: 800px;">
      <div v-for="(coin, i) in displayCoins" :key="i" ref="coinEls"
        class="w-20 h-20 rounded-full flex items-center justify-center text-2xl font-bold coin-glow"
        :class="isDark
          ? 'bg-gradient-to-br from-amber-400 via-amber-500 to-amber-700 text-amber-950'
          : 'bg-gradient-to-br from-amber-300 via-amber-400 to-amber-600 text-amber-900'"
        style="will-change: transform;">
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
    <p class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ statusText }}</p>

    <!-- 计算结果展示 -->
    <div v-if="showResult" class="mt-4 text-center">
      <p class="text-base font-medium" :class="isDark ? 'text-amber-300' : 'text-amber-600'">
        {{ t('coin.back') }}/{{ t('coin.front') }}: {{ coinValues.join('+') }} = {{ sum }} → {{ t(`gua.${resultType}`) }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import gsap from 'gsap'

const { t } = useI18n()

const props = defineProps({
  currentThrow: { type: Number, default: 1 },
  isAnimating: { type: Boolean, default: true },
  coinValues: { type: Array, default: () => [null, null, null] },
  isDark: { type: Boolean, default: false }
})

const coinEls = ref([])

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

// 尊重用户"减少动效"偏好
const prefersReducedMotion = typeof window !== 'undefined' &&
  !!window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches

// 待机:铜钱柔和悬浮(缓慢上下浮动 + 轻微旋转),模拟"静待揭晓"
function playIdle() {
  const els = coinEls.value
  if (!els || !els.length || prefersReducedMotion) return
  gsap.killTweensOf(els)
  gsap.fromTo(els, { y: 0, rotation: 0 }, {
    y: '+=13', rotation: 5,
    duration: 1.3, ease: 'sine.inOut',
    yoyo: true, repeat: -1,
    stagger: { each: 0.35, from: 'start' }
  })
}

// 揭晓:落地瞬间轻微放大 + 光晕闪亮(克制点睛)
function playReveal() {
  const els = coinEls.value
  if (!els || !els.length || prefersReducedMotion) return
  gsap.killTweensOf(els)
  gsap.fromTo(els, { scale: 1, filter: 'brightness(1)' }, {
    scale: 1.1, filter: 'brightness(1.4)',
    duration: 0.16, ease: 'power2.out', stagger: 0.06,
    yoyo: true, repeat: 1, overwrite: true
  })
}

// SSE 时序:isAnimating=true → 悬浮待机;false 且已有结果 → 揭晓点睛
watch(() => props.isAnimating, (v) => {
  nextTick(() => {
    if (v) playIdle()
    else if (props.coinValues.some(x => x)) playReveal()
  })
})

onMounted(() => { nextTick(() => { if (props.isAnimating) playIdle() }) })
onUnmounted(() => { if (coinEls.value.length) gsap.killTweensOf(coinEls.value) })
</script>
