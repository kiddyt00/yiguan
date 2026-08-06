<template>
  <div class="text-center py-8">
    <!-- 当前爻位提示 -->
    <p class="text-lg mb-6" :class="isDark ? 'text-green-400' : 'text-green-600'">
      {{ t('coin.divining', { n: currentThrow, name: yaoName }) }}
    </p>

    <!-- 三枚古铜钱(双面 3D,正面刻"易" 背面刻"观") -->
    <div class="flex justify-center gap-6 mb-8">
      <div v-for="(v, i) in coinValues" :key="i" class="coin3d">
        <div ref="coinEls" class="coin-inner" :aria-label="coinLabels[i]">
          <!-- 正面: 易 -->
          <div class="coin-face">
            <svg viewBox="0 0 100 100" width="80" height="80">
              <circle cx="50" cy="50" r="47" fill="#d4a853" stroke="#8a6320" stroke-width="3"/>
              <circle cx="50" cy="50" r="40" fill="none" stroke="#b08a3e" stroke-width="2" stroke-dasharray="4 4"/>
              <rect x="39" y="39" width="22" height="22" rx="4" fill="#4a3410"/>
              <text x="50" y="58" font-size="20" text-anchor="middle" fill="#f5d98c" font-weight="bold" font-family="serif">易</text>
            </svg>
          </div>
          <!-- 背面: 观 -->
          <div class="coin-face coin-back">
            <svg viewBox="0 0 100 100" width="80" height="80">
              <circle cx="50" cy="50" r="47" fill="#c89b45" stroke="#7a5a1c" stroke-width="3"/>
              <circle cx="50" cy="50" r="40" fill="none" stroke="#a8843a" stroke-width="2" stroke-dasharray="4 4"/>
              <rect x="39" y="39" width="22" height="22" rx="4" fill="#3a2808"/>
              <text x="50" y="58" font-size="20" text-anchor="middle" fill="#f5d98c" font-weight="bold" font-family="serif">观</text>
            </svg>
          </div>
        </div>
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
const currentAngles = [0, 0, 0] // 每枚铜钱当前朝向(0=正面 / 180=背面)

const yaoNames = computed(() => [t('yao.1'), t('yao.2'), t('yao.3'), t('yao.4'), t('yao.5'), t('yao.6')])
const yaoName = computed(() => yaoNames.value[props.currentThrow - 1] || '')

// 无障碍标签: 每枚铜钱当前朝向
const coinLabels = computed(() => props.coinValues.map(v =>
  v === null ? t('coin.shaking') : (v === 2 ? t('coin.front') : t('coin.back'))
))

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

// 推演中:铜钱轻柔摆动(小角度往返,像待定中的翻动)
function wobble() {
  const els = coinEls.value
  if (!els || !els.length || prefersReducedMotion) return
  gsap.killTweensOf(els)
  gsap.to(els, {
    rotateY: '+=30',
    duration: 0.6, ease: 'sine.inOut',
    yoyo: true, repeat: -1,
    stagger: { each: 0.2, from: 'start' }
  })
}

// 揭晓:铜钱翻到对应面(正→正面0°, 反→背面180°)+ 轻微放大闪亮
function settleCoins() {
  const els = coinEls.value
  if (!els || !els.length) return
  if (prefersReducedMotion) {
    gsap.set(els, { rotateY: (i) => props.coinValues[i] === 3 ? 180 : 0 })
    return
  }
  gsap.killTweensOf(els)
  els.forEach((el, i) => {
    const target = props.coinValues[i] === 3 ? 180 : 0
    const from = currentAngles[i]
    let to = target
    if (to === from) to = from + 360 // 同面也要翻一整圈,保证每爻都有翻转动作
    gsap.fromTo(el, { rotateY: from }, {
      rotateY: to, duration: 0.6, ease: 'power2.inOut', delay: i * 0.08,
      onComplete: () => { currentAngles[i] = to % 360 }
    })
  })
  // 点睛:轻微放大 + 亮度
  gsap.fromTo(els, { scale: 1, filter: 'brightness(1)' }, {
    scale: 1.08, filter: 'brightness(1.35)',
    duration: 0.18, ease: 'power2.out', stagger: 0.08,
    yoyo: true, repeat: 1, delay: 0.6, overwrite: true
  })
}

// coinValues 变化: 全 null → 摆动待机;有值 → 翻面揭晓
watch(() => props.coinValues, (v) => {
  nextTick(() => {
    if (v.every(x => x === null)) wobble()
    else settleCoins()
  })
}, { deep: true })

onMounted(() => { nextTick(() => { if (props.isAnimating) wobble() }) })
onUnmounted(() => { if (coinEls.value.length) gsap.killTweensOf(coinEls.value) })
</script>

<style scoped>
.coin3d {
  width: 80px;
  height: 80px;
  perspective: 500px;
}
.coin-inner {
  width: 100%;
  height: 100%;
  position: relative;
  transform-style: preserve-3d;
}
.coin-face {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  border-radius: 50%;
  filter: drop-shadow(0 0 10px rgba(212, 168, 83, 0.35));
}
.coin-back {
  transform: rotateY(180deg);
}
</style>
