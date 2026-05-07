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
      <div v-if="!auth.isLoggedIn()" class="mt-4 text-center text-sm opacity-60">
        请先 <router-link to="/login" class="underline text-red-600">登录</router-link> 后使用（新用户赠送3次免费）
      </div>

      <div v-else class="mt-4 flex items-center justify-between">
        <div class="flex gap-3 items-center">
          <router-link to="/ads" class="text-sm underline" :class="isDark ? 'text-cyan-400' : 'text-amber-600'">📢 看广告领次数</router-link>
        </div>
        <button @click="divine" :disabled="!question"
          class="px-8 py-3 rounded-lg font-medium transition disabled:opacity-40"
          :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
          ☯ 开始提问
        </button>
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
const isDark = computed(() => document.documentElement.classList.contains('dark'))

function divine() {
  if (!question.value || !auth.token) return
  router.push({ path: '/stream', state: { question: question.value } })
}
</script>
