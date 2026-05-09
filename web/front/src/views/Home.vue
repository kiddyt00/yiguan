<template>
  <div>
    <div class="text-center mb-8">
      <h2 class="text-4xl font-bold mb-3" :class="isDark ? 'text-white' : 'text-stone-800'">{{ t('home.title') }}</h2>
      <p class="text-base" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('home.subtitle') }}</p>
    </div>

    <div class="glass-card p-6">
      <label class="block font-medium mb-2" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ t('home.question.label') }}</label>
      <textarea v-model="question" rows="3" required
        :placeholder="t('home.question.placeholder')"
        class="w-full border rounded-lg p-3 resize-none bg-transparent outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500/30 transition"
        :class="isDark ? 'text-stone-100 placeholder-stone-500 border-stone-600' : 'text-stone-800 placeholder-stone-400 border-stone-300'"></textarea>

      <div v-if="!auth.isLoggedIn()" class="mt-6 text-center">
        <p class="text-sm mb-3" :class="isDark ? 'text-stone-400' : 'text-stone-500'">{{ t('home.login.prompt') }}</p>
        <router-link to="/login"
          class="inline-block px-8 py-3 rounded-lg font-medium transition bg-amber-600 text-white hover:bg-amber-500">
          {{ t('home.login.btn') }}
        </router-link>
      </div>

      <div v-else class="mt-6 text-center">
        <button @click="startDivination" :disabled="!question.trim() || loading"
          class="px-10 py-3 rounded-lg font-medium text-lg transition disabled:opacity-40 bg-amber-600 text-white hover:bg-amber-500 shadow-lg shadow-amber-600/30">
          {{ loading ? t('home.divine.loading') : t('home.divine.btn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

defineProps({ isDark: Boolean })

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const question = ref('')
const loading = ref(false)

function startDivination() {
  if (!question.value.trim() || !auth.token) return
  loading.value = true
  router.push({ path: '/stream', state: { question: question.value.trim() } })
}
</script>
