<template>
  <div>
    <div class="text-center mb-8">
      <h2 class="text-4xl font-bold mb-3 text-white">观己斋</h2>
      <p class="text-stone-400 text-base">我们以三枚铜钱的起卦方式，还原古人"观象玩辞"的从容与觉知</p>
    </div>

    <div class="glass-card p-6">
      <label class="block font-medium mb-2 text-stone-200">请输入你想问的问题：</label>
      <textarea v-model="question" rows="3" required
        placeholder="例如：我该不该换工作？这段感情能长久吗？..."
        class="w-full border rounded-lg p-3 resize-none bg-transparent text-stone-100 placeholder-stone-500 border-stone-600 focus:border-amber-500 focus:ring-1 focus:ring-amber-500/30 outline-none transition"></textarea>

      <!-- 未登录提示 -->
      <div v-if="!auth.isLoggedIn()" class="mt-6 text-center">
        <p class="text-sm text-stone-400 mb-3">登录后即可免费使用（新用户赠送3次）</p>
        <router-link to="/login"
          class="inline-block px-8 py-3 rounded-lg font-medium transition bg-amber-600 text-white hover:bg-amber-500">
          登录 / 注册
        </router-link>
      </div>

      <div v-else class="mt-6 text-center">
        <button @click="startDivination" :disabled="!question.trim() || loading"
          class="px-10 py-3 rounded-lg font-medium text-lg transition disabled:opacity-40 bg-amber-600 text-white hover:bg-amber-500 shadow-lg shadow-amber-600/30">
          {{ loading ? '起卦中...' : '开始提问' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

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
