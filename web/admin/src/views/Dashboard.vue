<template>
  <div>
    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <div v-for="card in cards" :key="card.label"
        class="stat-card"
        :style="{ '--card-accent': card.color }"
        @click="goTo(card.route)">
        <div class="stat-top">
          <span class="stat-icon">
            <el-icon :size="22"><component :is="card.icon" /></el-icon>
          </span>
          <span class="stat-arrow">→</span>
        </div>
        <div class="stat-label">{{ card.label }}</div>
        <div class="stat-value">{{ card.value }}</div>
      </div>
    </div>

    <!-- 起卦趋势图表 -->
    <div class="chart-card">
      <div class="chart-title">起卦趋势（近7天）</div>
      <v-chart :option="chartOption" autoresize style="height:280px" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import { User, UserFilled, Timer, Histogram, VideoPlay, TrendCharts } from '@element-plus/icons-vue'
use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

const router = useRouter()
const stats = ref({})
const trend = ref({})

const cards = computed(() => [
  { label: '注册用户', value: stats.value.total_users ?? '-', icon: User, color: 'var(--accent)', route: '/users' },
  { label: '活跃用户', value: stats.value.active_users ?? '-', icon: UserFilled, color: 'var(--accent)', route: '/users' },
  { label: '今日起卦', value: stats.value.today_divines ?? '-', icon: Timer, color: 'var(--accent)', route: '/hexagrams' },
  { label: '总起卦数', value: stats.value.total_divines ?? '-', icon: Histogram, color: 'var(--accent)', route: '/hexagrams' },
  { label: '今日广告', value: stats.value.ad_watches_today ?? '-', icon: VideoPlay, color: 'var(--accent)', route: '/ads' },
  { label: '总广告播放', value: stats.value.total_ads_watched ?? '-', icon: TrendCharts, color: 'var(--accent)', route: '/ads' },
])

const chartOption = computed(() => ({
  grid: { left: 40, right: 20, top: 20, bottom: 30 },
  xAxis: { type: 'category', data: Object.keys(trend.value), axisLine: { lineStyle: { color: 'var(--rule)' } }, axisLabel: { color: 'var(--muted)' } },
  yAxis: { type: 'value', minInterval: 1, splitLine: { lineStyle: { color: 'var(--rule)' } }, axisLabel: { color: 'var(--muted)' } },
  series: [{ type: 'bar', data: Object.values(trend.value), itemStyle: { color: 'var(--accent)', borderRadius: [4, 4, 0, 0] }, barWidth: 24 }],
  tooltip: { trigger: 'axis' },
}))

onMounted(async () => {
  try {
    const data = await adminApi.dashboard()
    stats.value = data
    if (data.daily_divine_trend) {
      trend.value = data.daily_divine_trend
    }
  } catch (e) {
    ElMessage.error('加载仪表盘失败: ' + e.message)
  }
})

function goTo(route) {
  router.push(route)
}
</script>

<style scoped>
.chart-card {
  background: var(--paper);
  border: 1px solid var(--rule);
  border-radius: 10px;
  padding: 16px;
  margin-top: 16px;
}
.chart-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--ink);
  margin-bottom: 12px;
  letter-spacing: 0.5px;
}
</style>
