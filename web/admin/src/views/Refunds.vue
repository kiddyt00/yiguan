<template>
  <div>
    <el-row :gutter="16" class="mb-4">
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value">{{ total }}</div><div class="stat-label">退款申请</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value" style="color:#E6A23C">{{ pendingCount }}</div><div class="stat-label">待处理</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value" style="color:#67C23A">{{ completedCount }}</div><div class="stat-label">已退款</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value" style="color:#F56C6C">{{ refundTotal }}</div><div class="stat-label">退款总额(元)</div></el-card></el-col>
    </el-row>

    <el-card shadow="never" body-style="padding:0">
      <el-table :data="refunds" stripe v-loading="loading" empty-text="暂无退款申请" size="small" style="width:100%">
        <el-table-column label="ID" width="60" prop="id" />
        <el-table-column label="用户ID" width="80" prop="user_id" />
        <el-table-column label="订单ID" width="80" prop="order_id" />
        <el-table-column label="金额" width="90" align="right">
          <template #default="{row}">¥{{ (row.amount/100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="原因" min-width="140" prop="reason" show-overflow-tooltip />
        <el-table-column label="状态" width="90" align="center">
          <template #default="{row}">
            <el-tag :type="refundStatusType(row.status)" size="small">{{ refundStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160">
          <template #default="{row}">{{ row.created_at.replace('T',' ').slice(0,16) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center">
          <template #default="{row}">
            <template v-if="row.status === 'pending'">
              <el-button type="success" size="small" :loading="actingId === row.id" @click="approve(row)">批准退款</el-button>
              <el-button type="danger" size="small" plain :loading="actingId === row.id" @click="showReject(row)">驳回</el-button>
            </template>
            <span v-else style="color:#909399;font-size:12px">已处理</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="rejectVisible" title="驳回退款申请" width="420px">
      <el-input v-model="rejectReason" type="textarea" :rows="3" placeholder="请输入驳回原因（可选）" />
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="actingId !== null" @click="reject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../api'

const refunds = ref([])
const loading = ref(false)
const actingId = ref(null)
const rejectVisible = ref(false)
const rejectReason = ref('')
const rejectTarget = ref(null)

const total = computed(() => refunds.value.length)
const pendingCount = computed(() => refunds.value.filter(r => r.status === 'pending').length)
const completedCount = computed(() => refunds.value.filter(r => r.status === 'completed').length)
const refundTotal = computed(() => refunds.value.filter(r => r.status === 'completed').reduce((s, r) => s + r.amount, 0) / 100)

function refundStatusType(s) {
  if (s === 'completed') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'approved') return 'primary'
  if (s === 'rejected') return 'info'
  return 'danger'
}
function refundStatusLabel(s) {
  if (s === 'pending') return '待处理'
  if (s === 'approved') return '已批准'
  if (s === 'completed') return '已退款'
  if (s === 'rejected') return '已驳回'
  return s || '未知'
}

async function load() {
  loading.value = true
  try {
    const res = await adminApi.refunds()
    refunds.value = res.items || []
  } catch (e) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function approve(row) {
  try {
    await ElMessageBox.confirm(`确认向用户（ID:${row.user_id}）退款 ¥${(row.amount/100).toFixed(2)}？退款成功后对应会员/次数将被回收。`, '批准退款', {
      type: 'warning',
      confirmButtonText: '确认退款',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  actingId.value = row.id
  try {
    await adminApi.approveRefund(row.id)
    ElMessage.success('退款成功')
    await load()
  } catch (e) {
    ElMessage.error(e.message || '退款失败')
  } finally {
    actingId.value = null
  }
}

function showReject(row) {
  rejectTarget.value = row
  rejectReason.value = ''
  rejectVisible.value = true
}

async function reject() {
  if (!rejectTarget.value) return
  actingId.value = rejectTarget.value.id
  try {
    await adminApi.rejectRefund(rejectTarget.value.id, rejectReason.value)
    ElMessage.success('已驳回')
    rejectVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.message || '驳回失败')
  } finally {
    actingId.value = null
  }
}

onMounted(load)
</script>
