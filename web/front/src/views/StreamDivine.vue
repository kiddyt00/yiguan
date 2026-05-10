<template>
  <div class="glass-card p-6">

    <!-- 阶段 1: 摇卦中 -->
    <div v-if="phase === 'coins'">
      <h3 class="text-xl font-bold mb-4 text-center" :class="isDark ? 'text-white' : 'text-stone-800'">{{ t('stream.divining') }}</h3>
      <CoinAnimation
        :current-throw="currentThrow"
        :is-animating="isAnimating"
        :coin-values="currentCoins"
        :is-dark="isDark"
      />
    </div>

    <!-- 阶段 2: 结果展示 + 阶段 3: 解读 + 阶段 4: 感谢页 -->
    <div v-if="phase === 'result' || phase === 'interpretation' || phase === 'done'">
      <h3 class="text-sm font-medium text-center mb-1" :class="isDark ? 'text-green-400' : 'text-green-600'">{{ t('stream.result.label') }}</h3>
      <h2 class="text-xl font-bold mb-6 text-center" :class="isDark ? 'text-stone-100' : 'text-stone-800'">{{ t('stream.result.title') }}</h2>

      <!-- 问题 -->
      <div class="mb-6">
        <p class="text-sm mb-1" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('stream.question') }}</p>
        <div class="rounded-lg p-3 text-sm" :class="isDark ? 'bg-slate-800/60 text-stone-200' : 'bg-stone-100 text-stone-700'">
          {{ question }}
        </div>
      </div>

      <!-- 6 次摇卦结果列表 -->
      <div class="mb-6">
        <h4 class="text-sm mb-3" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('stream.toss.history') }}</h4>
        <div class="space-y-2">
          <div v-for="toss in tossResults" :key="toss.throw"
            class="flex items-center gap-3 rounded-lg p-3"
            :class="isDark ? 'bg-slate-800/40' : 'bg-stone-100'">
            <!-- 爻位 -->
            <span class="text-sm font-medium w-14" :class="isDark ? 'text-stone-300' : 'text-stone-600'">{{ t(`yao.${toss.throw}`) }}</span>
            <!-- 3 枚铜钱 -->
            <div class="flex gap-1.5">
              <span v-for="(c, i) in toss.coins" :key="i"
                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold"
                :class="c === 'front'
                  ? 'bg-amber-500/80 text-amber-950'
                  : (isDark ? 'bg-slate-600/60 text-stone-300' : 'bg-stone-200 text-stone-600')">
                {{ t(c === 'front' ? 'coin.front' : 'coin.back') }}
              </span>
            </div>
            <!-- 总和与类型 -->
            <span class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">= {{ toss.sum }}</span>
            <span class="text-xs font-medium px-2 py-0.5 rounded-full"
              :class="toss.result === 'old_yang' || toss.result === 'old_yin'
                ? 'bg-red-500/20 text-red-300'
                : (isDark ? 'bg-amber-500/10 text-amber-300' : 'bg-amber-100 text-amber-700')">
              {{ t(`gua.${toss.result}`) }}
            </span>
            <!-- 爻线 -->
            <span class="ml-auto">
              <span v-if="toss.yang" class="block w-12 h-1 rounded"
                :class="toss.result === 'old_yang'
                  ? 'bg-red-500 shadow-[0_0_6px_rgba(239,68,68,0.5)]'
                  : (isDark ? 'bg-stone-200' : 'bg-stone-500')"></span>
              <span v-else class="flex gap-1">
                <span class="block w-5 h-1 rounded"
                  :class="toss.result === 'old_yin'
                    ? 'bg-red-500 shadow-[0_0_6px_rgba(239,68,68,0.5)]'
                    : (isDark ? 'bg-stone-200' : 'bg-stone-500')"></span>
                <span class="block w-5 h-1 rounded"
                  :class="toss.result === 'old_yin'
                    ? 'bg-red-500 shadow-[0_0_6px_rgba(239,68,68,0.5)]'
                    : (isDark ? 'bg-stone-200' : 'bg-stone-500')"></span>
              </span>
            </span>
          </div>
        </div>
      </div>

      <!-- 本卦 + 变卦 -->
      <div class="grid grid-cols-2 gap-4 mb-6">
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-800/50' : 'bg-stone-100'">
          <span class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('stream.hex.primary') }}</span>
          <div class="text-2xl font-bold mt-1 text-amber-400">{{ t('gua.' + guaResult.primary?.name) }}</div>
          <div class="text-2xl my-1">{{ guaResult.primary?.symbol }}</div>
          <p v-if="guaResult.primary?.gua_ci" class="text-xs mt-1" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ guaResult.primary.gua_ci }}</p>
        </div>
        <div class="rounded-lg p-4 text-center" :class="isDark ? 'bg-slate-700/40' : 'bg-stone-200/60'">
          <span class="text-sm" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('stream.hex.changing') }}</span>
          <div class="text-2xl font-bold mt-1 text-amber-400">{{ t('gua.' + guaResult.changing?.name) }}</div>
          <div class="text-2xl my-1">{{ guaResult.changing?.symbol }}</div>
          <p v-if="guaResult.changing?.gua_ci" class="text-xs mt-1" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ guaResult.changing.gua_ci }}</p>
        </div>
      </div>

      <!-- 变爻提示 -->
      <div v-if="guaResult.yaoPositions?.length" class="rounded-lg p-3 mb-6 text-center text-sm"
        :class="isDark ? 'bg-red-900/40 text-amber-100' : 'bg-red-50 text-red-800'">
        {{ t('stream.yao.changing.label') }}{{ yaoLabels }}
        <template v-if="guaResult.masterYao !== null"> | {{ t('stream.yao.master', { n: guaResult.masterYao + 1 }) }}</template>
      </div>

      <!-- Hexagram 爻线展示 -->
      <hexagram :lines="hexLines" :is-dark="isDark" />
    </div>

    <!-- 加载中 (等待 AI) -->
    <div v-if="showLoading" class="text-center py-6">
      <div class="inline-flex items-center gap-3 px-5 py-3 rounded-lg" :class="isDark ? 'bg-amber-500/10 text-amber-400' : 'bg-amber-100 text-amber-700'">
        <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span class="text-sm font-medium">{{ t('stream.ai.thinking') }}{{ loadingDots }}</span>
      </div>
    </div>

    <!-- 阶段 3: 解读 -->
    <div v-if="phase === 'interpretation' || phase === 'done'" class="border-t pt-6 mt-6" :class="isDark ? 'border-stone-700' : 'border-stone-200'">
      <h3 class="text-sm font-medium text-center mb-1" :class="isDark ? 'text-green-400' : 'text-green-600'">{{ t('stream.interpret.label') }}</h3>
      <h2 class="text-xl font-bold mb-4 text-center" :class="isDark ? 'text-stone-100' : 'text-stone-800'">{{ t('stream.interpret.title') }}</h2>

      <!-- 翻译提示条 -->
      <div v-if="showTranslateBanner" class="mb-4 px-4 py-2.5 rounded-lg flex items-center justify-between gap-3"
        :class="isDark ? 'bg-amber-500/10 border border-amber-500/30' : 'bg-amber-50 border border-amber-200'">
        <span class="text-xs" :class="isDark ? 'text-amber-300' : 'text-amber-700'">
          {{ historyLang === 'zh' ? '⚠ 此解读以中文生成' : '⚠ This reading was generated in English' }}
        </span>
        <button @click="doTranslate" :disabled="isTranslating"
          class="text-xs px-3 py-1 rounded-full font-medium transition border disabled:opacity-50"
          :class="isDark ? 'border-amber-500/50 text-amber-400 hover:bg-amber-500/20' : 'border-amber-400 text-amber-600 hover:bg-amber-50'">
          {{ isTranslating ? '...' : (historyLang === 'zh' ? '翻译为 English →' : '翻译为 中文 →') }}
        </button>
      </div>
      <div v-if="translateError" class="mb-4 text-xs text-red-400 text-center">{{ translateError }}</div>

      <div class="markdown-body leading-relaxed" v-html="renderedTranslation"></div>
    </div>

    <!-- 阶段 4: 感谢页 -->
    <div v-if="phase === 'done'" class="border-t pt-8 mt-8 text-center" :class="isDark ? 'border-stone-700' : 'border-stone-200'">
      <img src="../assets/laozi.jpeg" alt="Laozi" class="w-64 h-auto mx-auto mb-4 rounded-lg" />
      <p class="text-lg font-medium mb-2" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ t('stream.thanks.title') }}</p>
      <p class="text-sm" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ t('stream.thanks.blessing') }}</p>

      <button @click="goHome" class="mt-6 px-8 py-3 rounded-lg font-medium border-2 transition"
        :class="isDark ? 'border-amber-600 text-amber-400 hover:bg-amber-600 hover:text-white' : 'border-amber-500 text-amber-600 hover:bg-amber-500 hover:text-white'">
        {{ t('stream.retry') }}
      </button>
    </div>

    <!-- 错误 -->
    <div v-if="error" class="text-center text-red-400 py-4">{{ error }}</div>

    <!-- 大师入口 -->
    <div class="mt-4 pt-4 border-t" :class="isDark ? 'border-stone-700' : 'border-stone-200'">
      <button @click="showMaster = !showMaster"
        class="w-full py-3 rounded-lg text-center font-medium transition flex items-center justify-center gap-2"
        :class="isDark ? 'bg-amber-500/10 text-amber-400 hover:bg-amber-500/20' : 'bg-amber-50 text-amber-600 hover:bg-amber-100'">
        <span class="text-xl">🎓</span>
        {{ t('stream.master.title') }}
      </button>
      <div v-if="showMaster" class="mt-3 text-center">
        <img src="/qr-master.png" alt="QR" class="w-48 h-auto mx-auto rounded-lg border" :class="isDark ? 'border-stone-600' : 'border-stone-200'" />
        <p class="text-xs mt-2" :class="isDark ? 'text-stone-500' : 'text-stone-400'">{{ t('stream.master.qr') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { marked } from 'marked'
import { useI18n } from 'vue-i18n'
import { useTranslation } from '../composables/useTranslation'
import CoinAnimation from '../components/CoinAnimation.vue'
import Hexagram from '../components/Hexagram.vue'

defineProps({ isDark: Boolean })

const router = useRouter()
const auth = useAuthStore()
const { t, locale } = useI18n()
const { needsTranslation, getTranslation, generateTranslation, targetLang } = useTranslation()
const question = history.state?.question || ''

if (!question) { router.push('/') }

// 状态
const phase = ref('coins')
const currentThrow = ref(1)
const isAnimating = ref(true)
const currentCoins = ref([null, null, null])
const tossResults = ref([])
const guaResult = ref({ primary: null, changing: null, yaoPositions: [], masterYao: null })
const aiText = ref('')
const error = ref('')
const showMaster = ref(false)
const loadingDots = ref('...')
const historyId = ref(null)
const historyLang = ref('zh')
const translationText = ref('')
const isTranslating = ref(false)
const translateError = ref('')

const showLoading = computed(() => phase.value === 'result' && aiText.value === '')

const yaoLabels = computed(() => {
  const joiner = locale.value === 'zh' ? '、' : ', '
  return (guaResult.value.yaoPositions || [])
    .map(y => t(`yao.${y.position + 1}`))
    .join(joiner)
})

const hexLines = computed(() => {
  const desc = guaResult.value.primary?.yao_desc || ''
  const changing = (guaResult.value.yaoPositions || []).map(y => y.position)
  const names = [t('yao.6'), t('yao.5'), t('yao.4'), t('yao.3'), t('yao.2'), t('yao.1')]
  return names.map((name, i) => ({
    label: name,
    yang: desc[5 - i] === '1',
    changing: changing.includes(5 - i),
  }))
})

const renderedAI = computed(() => {
  if (!aiText.value) return ''
  let html = marked.parse(aiText.value)
  if (phase.value === 'interpretation') html += '<span class="animate-pulse">▊</span>'
  return html
})

const showTranslateBanner = computed(() => {
  return (phase.value === 'interpretation' || phase.value === 'done') &&
    historyLang.value && locale.value !== historyLang.value &&
    !translationText.value
})

const displayedInterpretation = computed(() => {
  if (translationText.value) return translationText.value
  return aiText.value
})

const renderedTranslation = computed(() => {
  if (!displayedInterpretation.value) return ''
  let html = marked.parse(displayedInterpretation.value)
  if (phase.value === 'interpretation' && !translationText.value) html += '<span class="animate-pulse">▊</span>'
  return html
})

async function doTranslate() {
  if (!historyId.value || isTranslating.value) return
  isTranslating.value = true
  translateError.value = ''
  try {
    const text = await generateTranslation(historyId.value, targetLang())
    translationText.value = text
  } catch (e) {
    translateError.value = e.message
  } finally {
    isTranslating.value = false
  }
}

let dotsTimer = null
function startDots() {
  dotsTimer = setInterval(() => {
    const dots = ['.', '..', '...', '....']
    const i = dots.indexOf(loadingDots.value)
    loadingDots.value = dots[(i + 1) % dots.length]
  }, 500)
}
function stopDots() {
  if (dotsTimer) { clearInterval(dotsTimer); dotsTimer = null }
}

function goHome() {
  router.push('/')
}

async function startStream() {
  try {
    const resp = await fetch('/api/divine/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token}`,
        'Accept-Language': locale.value,
      },
      body: JSON.stringify({ question }),
    })

    if (!resp.ok) {
      if (resp.status === 401) { auth.logout(); router.push('/login'); return }
      const err = await resp.json()
      error.value = err.error || t('stream.error.divination')
      phase.value = 'done'
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let currentEvent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event:')) {
          currentEvent = line.replace('event:', '').trim()
        } else if (line.startsWith('data:')) {
          const dataStr = line.replace('data:', '').trim()
          if (!dataStr) continue

          try {
            const data = JSON.parse(dataStr)

            if (currentEvent === 'phase') {
              if (data.phase === 'coins') {
                currentThrow.value = data.data.throw
                const val = data.data.yang ? 3 : 2
                currentCoins.value = [null, null, null]
                isAnimating.value = true
                await new Promise(r => setTimeout(r, 800))
                currentCoins.value = [val, val, val]
                isAnimating.value = false
                const coins = (data.data.coin_values || ['?', '?', '?']).map(v => v === '正' ? 'front' : 'back')
                tossResults.value.push({
                  throw: data.data.throw,
                  label: data.data.label,
                  coins,
                  sum: data.data.sum,
                  result: data.data.result,
                  yang: data.data.yang,
                })
                await new Promise(r => setTimeout(r, 400))
              } else if (data.phase === 'hexagram') {
                isAnimating.value = false
                phase.value = 'result'
                guaResult.value = {
                  primary: {
                    name: data.data.primary_gua,
                    gua_ci: data.data.primary_gua_ci || '',
                    symbol: data.data.primary_symbol || '',
                    yao_desc: data.data.primary_yao_desc || ''
                  },
                  changing: {
                    name: data.data.changing_gua,
                    gua_ci: data.data.changing_gua_ci || '',
                    symbol: data.data.changing_symbol || ''
                  },
                  yaoPositions: data.data.yao_positions || [],
                  masterYao: (() => {
                    const mp = (data.data.yao_positions || []).find(y => y.is_master)
                    return mp ? mp.position : null
                  })()
                }
                startDots()
              }
            } else if (currentEvent === 'ai') {
              stopDots()
              if (phase.value === 'result') phase.value = 'interpretation'
              aiText.value += data.chunk
            } else if (currentEvent === 'done') {
              stopDots()
              phase.value = 'done'
              if (data.id) historyId.value = data.id
              if (data.lang) historyLang.value = data.lang
            } else if (currentEvent === 'error') {
              stopDots()
              error.value = data.error
              phase.value = 'done'
            }
          } catch (e) {
            console.error('Parse SSE error:', e)
          }
        }
      }
    }
  } catch (e) {
    stopDots()
    error.value = t('stream.error.network.prefix') + e.message
    phase.value = 'done'
  }
}

onMounted(startStream)
onUnmounted(stopDots)
</script>

<style scoped>
.markdown-body :deep(h2) { font-size: 1.1rem; font-weight: 600; margin-top: 1.25rem; margin-bottom: 0.5rem; padding-bottom: 0.25rem; border-bottom: 1px solid rgba(255,255,255,0.1); }
.markdown-body :deep(h3) { font-size: 1rem; font-weight: 600; margin-top: 1rem; margin-bottom: 0.4rem; }
.markdown-body :deep(p) { margin-bottom: 0.6rem; }
.markdown-body :deep(strong) { font-weight: 700; }
.markdown-body :deep(blockquote) { border-left: 3px solid rgba(212,168,83,0.3); padding-left: 0.75rem; margin: 0.5rem 0; }
</style>
