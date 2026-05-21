<template>
  <div>
    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <div v-for="card in cards" :key="card.label"
        class="stat-card"
        :style="{ '--card-accent': card.color }"
        @click="goTo(card.route)">
        <div class="stat-top">
          <span class="stat-icon">{{ card.icon }}</span>
          <span class="stat-arrow">→</span>
        </div>
        <div class="stat-label">{{ card.label }}</div>
        <div class="stat-value">{{ card.value }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const stats = ref({})

const cards = computed(() => [
  { label: '注册用户', value: stats.value.total_users ?? '-', icon: '👥', color: '#d4a853', route: '/users' },
  { label: '活跃用户', value: stats.value.active_users ?? '-', icon: '🔥', color: '#4ade80', route: '/users' },
  { label: '今日起卦', value: stats.value.today_divines ?? '-', icon: '🔮', color: '#60a5fa', route: '/hexagrams' },
  { label: '总起卦数', value: stats.value.total_divines ?? '-', icon: '📊', color: '#a78bfa', route: '/hexagrams' },
  { label: '今日广告', value: stats.value.ad_watches_today ?? '-', icon: '📺', color: '#f59e0b', route: '/ads' },
  { label: '总广告播放', value: stats.value.total_ads_watched ?? '-', icon: '📈', color: '#f472b6', route: '/ads' },
])

onMounted(async () => {
  try {
    stats.value = await adminApi.dashboard()
  } catch (e) {
    ElMessage.error('加载仪表盘失败: ' + e.message)
  }
})

function goTo(route) {
  router.push(route)
}
</script>
