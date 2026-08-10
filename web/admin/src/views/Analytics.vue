<template>
  <div>
    <div class="page-header"><h2>数据分析</h2></div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="mb-4">
      <el-col :span="12">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon-circle" style="background:oklch(94% 0.03 80);color:var(--accent-deep)"><el-icon :size="22"><Key /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ analytics.total_logins ?? '-' }}</div>
            <div class="stat-label">总登录次数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon-circle" style="background:oklch(94% 0.03 80);color:var(--accent-deep)"><el-icon :size="22"><Calendar /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ analytics.today_logins ?? '-' }}</div>
            <div class="stat-label">今日登录</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 登录趋势 -->
    <el-card shadow="never" class="mb-4 chart-card">
      <div class="chart-title">登录趋势（近7天）</div>
      <v-chart :option="loginTrendOption" autoresize style="height:260px" />
    </el-card>

    <!-- 时段分布 -->
    <el-card shadow="never" class="mb-4 chart-card">
      <div class="chart-title">时段分布</div>
      <v-chart :option="hourOption" autoresize style="height:240px" />
    </el-card>

    <!-- 设备 & 操作系统 -->
    <el-row :gutter="16" class="mb-4">
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <div class="chart-title">设备分布</div>
          <v-chart :option="deviceOption" autoresize style="height:240px" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <div class="chart-title">操作系统</div>
          <v-chart :option="osOption" autoresize style="height:240px" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 城市分布 -->
    <el-card shadow="never" class="chart-card">
      <div class="chart-title">城市分布</div>
      <div v-if="cityList.length" class="city-grid">
        <div v-for="c in cityList" :key="c.city" class="city-item">
          <span class="city-name">{{ c.city }}</span>
          <span class="city-count">{{ c.count }}</span>
        </div>
      </div>
      <div v-else class="empty-text">暂无数据</div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
use([BarChart, LineChart, PieChart, GridComponent, TooltipComponent, CanvasRenderer])

const analytics = ref({})

const loginTrendOption = computed(() => {
  const data = analytics.value.daily_trend || {}
  return {
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category', data: Object.keys(data) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{ type: 'line', data: Object.values(data), smooth: true, lineStyle: { color: 'var(--accent)', width: 2 }, itemStyle: { color: 'var(--accent)' }, areaStyle: { color: 'oklch(70% 0.12 75 / 0.12)' } }],
    tooltip: { trigger: 'axis' },
  }
})

const hourOption = computed(() => {
  const data = analytics.value.login_by_hour || {}
  return {
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category', data: Object.keys(data) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{ type: 'bar', data: Object.values(data), itemStyle: { color: 'oklch(60% 0.08 170)', borderRadius: [4, 4, 0, 0] }, barWidth: 20 }],
    tooltip: { trigger: 'axis' },
  }
})

const deviceOption = computed(() => {
  const data = analytics.value.login_by_device || {}
  const entries = Object.entries(data)
  if (!entries.length) return {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [{
      type: 'pie', radius: ['36%', '60%'], center: ['50%', '50%'],
      data: entries.map(([k, v]) => ({ name: k, value: v })),
      itemStyle: { borderRadius: 4 },
      label: { show: true, formatter: '{b}\n{d}%', fontSize: 11 },
    }],
  }
})

const osOption = computed(() => {
  const data = analytics.value.login_by_os || {}
  const entries = Object.entries(data)
  if (!entries.length) return {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [{
      type: 'pie', radius: ['36%', '60%'], center: ['50%', '50%'],
      data: entries.map(([k, v]) => ({ name: k, value: v })),
      itemStyle: { borderRadius: 4 },
      label: { show: true, formatter: '{b}\n{d}%', fontSize: 11 },
    }],
  }
})

const cityList = computed(() => {
  const data = analytics.value.login_by_city || {}
  return Object.entries(data)
    .map(([city, count]) => ({ city, count }))
    .sort((a, b) => b.count - a.count)
})

onMounted(async () => {
  try {
    analytics.value = await adminApi.analytics()
  } catch (e) {
    ElMessage.error('加载分析数据失败: ' + e.message)
  }
})
</script>

<style scoped>
.page-header h2 { font-size: 22px; font-weight: 700; color: var(--ink); margin-bottom: 16px; }
.mb-4 { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; gap: 16px; padding: 8px 0; }
.stat-icon-circle { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 22px; flex-shrink: 0; }
.stat-info { flex: 1; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--ink); line-height: 1.2; }
.stat-label { font-size: 13px; color: var(--muted); margin-top: 2px; }
.chart-card { border-radius: 10px; }
.chart-title { font-size: 15px; font-weight: 700; color: var(--ink); margin-bottom: 12px; }
.city-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.city-item { display: flex; align-items: center; justify-content: space-between; background: var(--paper-3); padding: 8px 14px; border-radius: 8px; min-width: 140px; }
.city-name { font-size: 14px; font-weight: 500; color: var(--ink-2); }
.city-count { font-size: 16px; font-weight: 700; color: var(--accent-deep); }
.empty-text { text-align: center; padding: 32px 0; color: var(--faint); font-size: 14px; }
</style>
