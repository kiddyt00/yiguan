import { createI18n } from 'vue-i18n'
import zh from './locales/zh.json'
import en from './locales/en.json'

const saved = localStorage.getItem('lang') || 'zh'

const i18n = createI18n({
  legacy: false,
  locale: saved,
  fallbackLocale: 'zh',
  messages: { zh, en },
})

export default i18n
