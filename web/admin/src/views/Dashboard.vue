<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">仪表盘</h2>
    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      <el-card>
        <div class="text-sm text-gray-500">注册用户</div>
        <div class="text-3xl font-bold mt-2">{{ stats.total_users }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">活跃用户</div>
        <div class="text-3xl font-bold mt-2 text-green-600">{{ stats.active_users }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">今日算卦</div>
        <div class="text-3xl font-bold mt-2">{{ stats.today_divines }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">总起卦数</div>
        <div class="text-3xl font-bold mt-2">{{ stats.total_divines }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">今日广告播放</div>
        <div class="text-3xl font-bold mt-2 text-blue-600">{{ stats.ad_watches_today }}</div>
      </el-card>
      <el-card>
        <div class="text-sm text-gray-500">总广告播放</div>
        <div class="text-3xl font-bold mt-2">{{ stats.total_ads_watched }}</div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'

const stats = ref({})
onMounted(async () => {
  try {
    stats.value = await adminApi.dashboard()
  } catch (e) {
    ElMessage.error('加载仪表盘失败: ' + e.message)
  }
})
</script>
