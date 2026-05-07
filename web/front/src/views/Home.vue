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
        <span class="text-sm opacity-50">剩余次数：{{ quota }}</span>
        <button @click="divine" :disabled="loading || quota <= 0"
          class="px-8 py-3 rounded-lg font-medium transition disabled:opacity-40"
          :class="isDark ? 'bg-cyan-600 text-white hover:bg-cyan-500' : 'bg-red-800 text-amber-100 hover:bg-red-700'">
          <span v-if="loading">⏳ 起卦中...</span>
          <span v-else-if="quota <= 0">次数不足</span>
          <span v-else>☯ 开始提问</span>
        </button>
      </div>

      <!-- 付费预留 -->
      <div v-if="auth.isLoggedIn() && quota <= 0" class="mt-4 text-center">
        <button disabled class="px-6 py-2 rounded-lg text-sm opacity-30 cursor-not-allowed"
          :class="isDark ? 'bg-cyan-600' : 'bg-red-800'">
          购买次数 (即将上线)
        </button>
      </div>
    </div>

    <!-- 结果 -->
    <Result v-if="result" :data="result" :is-dark="isDark" class="mt-8" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import Result from './Result.vue'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const question = ref('')
const loading = ref(false)
const result = ref(null)
const quota = ref(auth.isLoggedIn() ? 3 : 0)
const isDark = computed(() => document.documentElement.classList.contains('dark'))

async function divine() {
  if (!question.value || !auth.token) return
  loading.value = true
  try {
    const res = await fetch('/api/divine', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${auth.token}` },
      body: JSON.stringify({ question: question.value }),
    })
    const data = await res.json()
    if (res.ok) {
      result.value = data
      quota.value = data.remaining_quota
    } else if (res.status === 402) {
      quota.value = 0
      alert('次数已用完，请获取更多次数')
    } else if (res.status === 401) {
      auth.logout()
      router.push('/login')
    }
  } catch (e) {
    alert('请求失败: ' + e.message)
  } finally {
    loading.value = false
  }
}
</script>
