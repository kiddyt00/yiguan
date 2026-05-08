<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">卦象任务管理</h2>
    <div class="mb-4 flex gap-2">
      <el-input v-model="userIdFilter" placeholder="用户ID筛选" style="width:150px" @keyup.enter="load" />
      <el-button @click="load">筛选</el-button>
    </div>
    <el-table :data="items" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="nickname" label="用户" width="100" />
      <el-table-column prop="user_id" label="用户ID" width="80" />
      <el-table-column prop="question" label="问题" min-width="200" show-overflow-tooltip />
      <el-table-column prop="primary_gua" label="本卦" width="80" />
      <el-table-column prop="changing_gua" label="变卦" width="80" />
      <el-table-column prop="created_at" label="时间" width="170">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" @click="showDetail(row)">详情</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="detailVisible" title="卦象详情" width="700px">
      <div v-if="detail" class="space-y-3">
        <p><strong>问题：</strong>{{ detail.question }}</p>
        <p><strong>本卦：</strong>{{ detail.primary_gua }} <strong>变卦：</strong>{{ detail.changing_gua }}</p>
        <p><strong>变爻：</strong>{{ detail.yao_positions }}</p>
        <p><strong>AI解卦：</strong></p>
        <div class="bg-gray-50 p-3 rounded whitespace-pre-wrap">{{ detail.interpretation }}</div>
        <p class="text-sm text-gray-400">{{ formatDate(detail.created_at) }}</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const items = ref([])
const detail = ref(null)
const detailVisible = ref(false)
const userIdFilter = ref('')

onMounted(() => load())

async function load() {
  try {
    const params = { limit: 50 }
    if (userIdFilter.value) params.userId = userIdFilter.value
    const data = await adminApi.hexagrams(params)
    items.value = data.items
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  }
}

function showDetail(row) {
  detail.value = row
  detailVisible.value = true
}

async function remove(row) {
  try {
    await ElMessageBox.confirm('确定删除此记录？', '确认')
    await adminApi.deleteHexagram(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message)
  }
}

function formatDate(d) { return new Date(d).toLocaleString('zh-CN') }
</script>
