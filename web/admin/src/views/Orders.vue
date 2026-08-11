<template>
  <div>
    <el-row :gutter="16" class="mb-4">
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value">{{ total }}</div><div class="stat-label">总订单</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value">{{ paid }}</div><div class="stat-label">已支付</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value">{{ pending }}</div><div class="stat-label">待支付</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value">{{ revenue }}</div><div class="stat-label">收入(元)</div></el-card></el-col>
    </el-row>

    <el-card shadow="never" body-style="padding:0">
      <el-table :data="orders" stripe v-loading="loading" empty-text="暂无订单" size="small" style="width:100%">
        <el-table-column label="ID" width="50" prop="id" />
        <el-table-column label="用户" min-width="100">
          <template #default="{row}"><span :title="'ID:'+row.user_id">{{ row.user_name || '用户'+row.user_id }}</span></template>
        </el-table-column>
        <el-table-column label="商品" width="100">
          <template #default="{row}">{{ productName(row.product_id) }}</template>
        </el-table-column>
        <el-table-column label="金额" width="90" align="right">
          <template #default="{row}"><span class="amount">¥{{ (row.amount/100).toFixed(2) }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{row}">
            <el-tag :type="orderStatusType(row.status)" size="small">{{ orderStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="支付时间" width="160">
          <template #default="{row}"><span class="t-time">{{ row.paid_at ? row.paid_at.replace('T',' ').slice(0,16) : '-' }}</span></template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{row}"><span class="t-time">{{ row.created_at.replace('T',' ').slice(0,16) }}</span></template>
        </el-table-column>
      </el-table>
      <div class="pg" v-if="total > limit">
        <el-pagination background layout="prev,pager,next" :total="total" :page-size="limit" :current-page="page" @current-change="onPageChange" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'

const orders = ref([])
const total = ref(0)
const loading = ref(false)
const limit = 50
const page = ref(1)

const paid = computed(() => orders.value.filter(o => o.status === 'paid').length)
const pending = computed(() => orders.value.filter(o => o.status === 'pending').length)
const revenue = computed(() => orders.value.filter(o => o.status === 'paid').reduce((s, o) => s + o.amount, 0) / 100)

const productMap = { single: '单次', monthly: '月卡', quarterly: '季卡', yearly: '年卡' }
function productName(id) { return productMap[id] || id || '-' }
function orderStatusType(s) {
  if (s === 'paid') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'refunded') return 'info'
  return 'danger'
}
function orderStatusLabel(s) {
  if (s === 'paid') return '已支付'
  if (s === 'pending') return '待支付'
  if (s === 'refunded') return '已退款'
  return s || '未知'
}

async function loadOrders() {
  loading.value = true
  try {
    const data = await adminApi.orders(limit, (page.value - 1) * limit)
    orders.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载订单失败: ' + (e.message || '未知错误'))
  }
  loading.value = false
}

function onPageChange(p) { page.value = p; loadOrders() }

onMounted(loadOrders)
</script>

<style scoped>
.stat-card{text-align:center}
.stat-value{font-size:32px;font-weight:700;color:var(--ink);font-variant-numeric:tabular-nums}
.stat-label{font-size:13px;color:var(--muted);margin-top:4px}
.amount{font-weight:600;color:var(--accent-deep);font-variant-numeric:tabular-nums}
.t-time{font-size:12px;color:var(--muted)}
.pg{display:flex;justify-content:center;padding:12px 0}
</style>
