import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'

// 三层缓存：L1 Pinia 内存 → L2 localStorage → L3 后端 DB/LLM
const cache = reactive(new Map()) // Map<historyId, Map<lang, { text, ts }>>

function localKey(historyId, lang) {
  return `trans_${historyId}_${lang}`
}

function setCacheL1(historyId, lang, text) {
  if (!cache.has(historyId)) cache.set(historyId, new Map())
  cache.get(historyId).set(lang, { text, ts: Date.now() })
  // L2: localStorage
  try {
    localStorage.setItem(localKey(historyId, lang), text)
  } catch { /* quota exceeded, ignore */ }
}

function getCacheL1(historyId, lang) {
  const m = cache.get(historyId)
  if (m?.has(lang)) return m.get(lang).text
  return null
}

function getCacheL2(historyId, lang) {
  try {
    const v = localStorage.getItem(localKey(historyId, lang))
    if (v) {
      setCacheL1(historyId, lang, v) // promote to L1
      return v
    }
  } catch { /* ignore */ }
  return null
}

export function useTranslation() {
  const { locale } = useI18n()
  const auth = useAuthStore()

  /**
   * 检查历史记录是否需要翻译
   * @param {Object} history - 历史记录对象，需含 lang 字段
   */
  function needsTranslation(history) {
    return history?.lang && locale.value !== history.lang
  }

  /**
   * 获取翻译（按 L1→L2→GET 逐层尝试，都不命中返回 null）
   * @param {number} historyId
   * @param {string} targetLang - 'zh' | 'en'
   */
  async function getTranslation(historyId, targetLang) {
    // L1
    const l1 = getCacheL1(historyId, targetLang)
    if (l1) return l1

    // L2
    const l2 = getCacheL2(historyId, targetLang)
    if (l2) return l2

    // L3: GET 试探后端缓存
    try {
      const res = await fetch(
        `/api/history/${historyId}/translate?target=${targetLang}`,
        { headers: { Authorization: `Bearer ${auth.token}` } }
      )
      if (res.ok) {
        const data = await res.json()
        setCacheL1(historyId, targetLang, data.content)
        return data.content
      }
    } catch { /* network error, fall through */ }
    return null
  }

  /**
   * 生成翻译（POST 调 LLM，幂等——已存在则直接返回）
   */
  async function generateTranslation(historyId, targetLang) {
    try {
      const res = await fetch(
        `/api/history/${historyId}/translate?target=${targetLang}`,
        {
          method: 'POST',
          headers: { Authorization: `Bearer ${auth.token}` }
        }
      )
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.error || '翻译失败')
      }
      const data = await res.json()
      setCacheL1(historyId, targetLang, data.content)
      return data.content
    } catch (e) {
      throw e
    }
  }

  /**
   * 获取目标语言（locale 的反向）
   */
  function targetLang() {
    return locale.value === 'zh' ? 'en' : 'zh'
  }

  return { needsTranslation, getTranslation, generateTranslation, targetLang }
}
