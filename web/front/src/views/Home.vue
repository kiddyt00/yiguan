<template>
  <div>
    <div class="text-center mb-8">
      <h2 class="text-3xl font-bold mb-2">心有疑虑，问卦于天</h2>
      <p class="opacity-60">默想你的问题，诚心求问，AI 为你解卦</p>
    </div>

    <div class="bg-stone-300/30 rounded-lg h-16 mb-8 flex items-center justify-center text-sm opacity-40">
      广告位 (Banner)
    </div>

    <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
      <label class="block font-medium mb-2">请输入你想问的问题：</label>
      <textarea v-model="question" rows="3" required
        placeholder="例如：我该不该换工作？这段感情能长久吗？..."
        class="w-full border rounded-lg p-3 resize-none bg-transparent"
        :class="isDark ? 'border-slate-600' : 'border-stone-300'"></textarea>

      <!-- 未登录提示 -->
      <div v-if="!auth.isLoggedIn()" class="mt-6 text-center">
        <p class="text-sm opacity-50 mb-3">登录后即可免费使用（新用户赠送3次）</p>
        <router-link to="/login"
          class="inline-block px-8 py-3 rounded-lg font-medium transition"
          :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
          登录 / 注册
        </router-link>
      </div>

      <div v-else class="mt-4">
        <div class="flex gap-3 items-center justify-between flex-wrap">
          <router-link to="/ads" class="text-sm underline" :class="isDark ? 'text-cyan-400' : 'text-amber-600'">📢 看广告领次数</router-link>
          <div class="flex gap-2">
            <button @click="divine(false)" :disabled="!question || loading"
              class="px-6 py-3 rounded-lg font-medium transition disabled:opacity-40"
              :class="isDark ? 'bg-slate-600 text-white hover:bg-slate-500' : 'bg-stone-600 text-white hover:bg-stone-500'">
              {{ loading ? '起卦中...' : '常规算卦' }}
            </button>
            <button @click="divine(true)" :disabled="!question || loading"
              class="px-6 py-3 rounded-lg font-medium transition disabled:opacity-40"
              :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
              {{ loading ? '起卦中...' : '☯ AI 流式算卦' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const question = ref('')
const loading = ref(false)
const isDark = computed(() => document.documentElement.classList.contains('dark'))

const API = import.meta.env.PROD ? '' : ''
function divine(stream) {
  if (!question.value || !auth.token) return
  loading.value = true
  if (stream) {
    router.push({ path: '/stream', state: { question: question.value } })
  } else {
    router.push({ path: '/result', state: { question: question.value } })
  }
}
</script>
