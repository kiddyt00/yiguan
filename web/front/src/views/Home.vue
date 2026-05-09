<template>
  <div>
    <div class="text-center mb-8">
      <h2 class="text-3xl font-bold mb-2">观己斋</h2>
      <p class="opacity-60">我们以三枚铜钱的起卦方式，还原古人"观象玩辞"的从容与觉知</p>
    </div>

    <!-- 广告位暂时隐藏 -->

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

      <div v-else class="mt-6 text-center">
        <button @click="startDivination" :disabled="!question.trim() || loading"
          class="px-10 py-3 rounded-lg font-medium text-lg transition disabled:opacity-40"
          :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
          {{ loading ? '起卦中...' : '开始提问' }}
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
const loading = ref(false)
const isDark = computed(() => document.documentElement.classList.contains('dark'))

const API = import.meta.env.PROD ? '' : ''
function startDivination() {
  if (!question.value.trim() || !auth.token) return
  loading.value = true
  // 统一使用流式起卦流程（带铜钱动画）
  router.push({ path: '/stream', state: { question: question.value.trim() } })
}
</script>
