<template>
  <div>
    <el-card shadow="never" body-style="padding:0">
      <el-table :data="memberships" stripe v-loading="loading" empty-text="暂无会员记录" size="small" style="width:100%">
        <el-table-column label="ID" width="50" prop="id" />
        <el-table-column label="用户" min-width="100">
          <template #default="{row}"><span :title="'ID:'+row.user_id">{{ row.user_name || '用户'+row.user_id }}</span></template>
        </el-table-column>
        <el-table-column label="类型" width="70">
          <template #default="{row}">{{ productName(row.product_id) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{row}">
            <el-tag :type="row.status==='active'?'success':'info'" size="small">
              {{ row.status==='active'?'有效':(row.status==='terminated'?'已终止':'已退款') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="生效" width="160">
          <template #default="{row}"><span class="t-time">{{ row.start_time.replace('T',' ').slice(0,16) }}</span></template>
        </el-table-column>
        <el-table-column label="到期" width="160">
          <template #default="{row}"><span class="t-time">{{ row.end_time.replace('T',' ').slice(0,16) }}</span></template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{row}"><span class="t-time">{{ row.created_at.replace('T',' ').slice(0,16) }}</span></template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'

const memberships = ref([])
const loading = ref(false)

const productMap = { monthly: '月卡', quarterly: '季卡', yearly: '年卡' }
function productName(id) { return productMap[id] || id || '-' }

async function load() {
  loading.value = true
  try {
    const data = await adminApi.memberships(100, 0)
    memberships.value = data.items || []
  } catch (e) {
    ElMessage.error('加载会员记录失败: ' + (e.message || '未知错误'))
  }
  loading.value = false
}

onMounted(load)
</script>

<style scoped>
.t-time{font-size:12px;color:var(--muted)}
</style>