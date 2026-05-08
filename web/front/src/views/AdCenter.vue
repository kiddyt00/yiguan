<template>
  <div class="bg-white/80 backdrop-blur rounded-xl shadow-md p-6" :class="{ '!bg-slate-800/80': isDark }">
    <h3 class="text-xl font-bold mb-4 text-center">📢 看广告领次数</h3>

    <div v-if="loading" class="text-center py-8 text-gray-400">加载中...</div>

    <div v-else-if="ads.length === 0" class="text-center text-gray-400 py-8">暂无可用广告</div>

    <div v-for="ad in ads" :key="ad.id" class="mb-4 p-4 rounded-lg border" :class="isDark ? 'border-slate-600 bg-slate-700' : 'border-stone-200'">
      <div class="flex justify-between items-center">
        <div>
          <h4 class="font-bold">{{ ad.name }}</h4>
          <p class="text-sm opacity-60">{{ ad.description }}</p>
          <p class="text-sm mt-1">观看 <span class="font-bold">{{ ad.watch_duration }}</span> 秒，奖励 <span class="text-amber-600 font-bold">{{ ad.reward_quota }}</span> 次起卦</p>
        </div>
        <button @click="watchAd(ad)"
          class="px-4 py-2 rounded-lg font-medium transition"
          :class="isDark ? 'bg-cyan-600 hover:bg-cyan-500 text-white' : 'bg-amber-600 hover:bg-amber-500 text-white'">
          观看
        </button>
      </div>
    </div>

    <!-- 广告弹窗 -->
    <div v-if="watchingAd" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div class="bg-white rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-auto" :class="isDark ? 'bg-slate-800 text-white' : ''">
        <div class="flex justify-between items-center mb-4">
          <h4 class="font-bold text-lg">{{ watchingAd.name }}</h4>
          <button @click="closeAd" class="text-gray-400 text-2xl hover:text-gray-600">&times;</button>
        </div>
        <iframe :src="watchingAd.content_url" class="w-full h-64 border rounded" frameborder="0" />
        <div class="mt-4 text-center">
          <p class="text-lg mb-3">
            <template v-if="countdown > 0">还需观看 <span class="text-amber-600 font-bold">{{ countdown }}</span> 秒</template>
            <template v-else>✅ 观看完成！</template>
          </p>
          <button v-if="countdown <= 0" @click="claimReward"
            class="px-8 py-3 rounded-lg font-medium bg-green-600 text-white hover:bg-green-500 transition">
            🎁 领取 {{ watchingAd.reward_quota }} 次起卦
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const props = defineProps(['isDark'])
const emit = defineEmits(['rewarded'])

const ads = ref([])
const loading = ref(true)
const watchingAd = ref(null)
const countdown = ref(0)
let timer = null

onMounted(async () => {
  try {
    const res = await fetch('/api/ads/active')
    const data = await res.json()
    ads.value = data.items || []
  } catch (e) {
    console.error('Failed to load ads:', e)
  } finally {
    loading.value = false
  }
})

async function watchAd(ad) {
  try {
    const res = await fetch(`/api/ads/${ad.id}/watch`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${auth.token}` },
    })
    if (res.status === 401) { auth.logout(); router.push('/login'); return }
    watchingAd.value = ad
    countdown.value = ad.watch_duration
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) {
    alert('开始观看失败')
  }
}

async function claimReward() {
  try {
    const res = await fetch(`/api/ads/${watchingAd.value.id}/complete`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ duration: watchingAd.value.watch_duration }),
    })
    const data = await res.json()
    if (data.rewarded) {
      emit('rewarded', data)
    } else if (data.error) {
      alert(data.error)
    }
  } catch (e) {
    alert('领取失败')
  }
  closeAd()
}

function closeAd() {
  watchingAd.value = null
  if (timer) clearInterval(timer)
  countdown.value = 0
}
</script>
