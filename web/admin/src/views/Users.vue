<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">用户管理</h2>
    <el-table :data="users" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="类型" width="70">
        <template #default="{ row }">
          <el-tag :type="row.openid ? 'success' : 'info'" size="small">{{ row.openid ? '微信' : '手机' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column label="微信昵称" width="120">
        <template #default="{ row }">
          <span v-if="row.openid && row.nickname !== '微信用户'">{{ row.nickname }}</span>
          <span v-else class="opacity-40">—</span>
        </template>
      </el-table-column>
      <el-table-column label="显示昵称" width="120">
        <template #default="{ row }">
          {{ row.nickname || '—' }}
        </template>
      </el-table-column>
      <el-table-column prop="role" label="角色" width="70">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">{{ row.role }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="70">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">{{ row.is_active ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button size="small" @click="toggleUser(row)">{{ row.is_active ? '禁用' : '启用' }}</el-button>
          <el-button size="small" type="primary" @click="adjustQuota(row)">配额</el-button>
          <el-button size="small" type="info" @click="viewHistory(row)">记录</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="quotaVisible" title="调整配额" width="400px">
      <el-input-number v-model="quotaDelta" :min="-10000" :max="10000" />
      <template #footer>
        <el-button @click="quotaVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmQuota">确认</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="historyVisible" :title="'用户起卦记录'" width="800px">
      <el-table :data="userHistory" size="small">
        <el-table-column prop="id" label="ID" width="50" />
        <el-table-column prop="question" label="问题" />
        <el-table-column prop="primary_gua" label="本卦" width="70" />
        <el-table-column prop="changing_gua" label="变卦" width="70" />
        <el-table-column prop="created_at" label="时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'
const users = ref([])
const quotaVisible = ref(false)
const quotaDelta = ref(0)
const currentUser = ref(null)
const historyVisible = ref(false)
const userHistory = ref([])
onMounted(loadUsers)
async function loadUsers(){try{const d=await adminApi.users({limit:100});users.value=d.items||[]}catch(e){ElMessage.error('加载失败: '+e.message)}}
async function toggleUser(u){try{await adminApi.toggleUser(u.id);u.is_active=u.is_active?0:1;ElMessage.success(u.is_active?'已启用':'已禁用')}catch(e){ElMessage.error(e.message)}}
function adjustQuota(u){currentUser.value=u;quotaDelta.value=0;quotaVisible.value=true}
async function confirmQuota(){try{await adminApi.adjustQuota(currentUser.value.id,quotaDelta.value);ElMessage.success('已调整');quotaVisible.value=false}catch(e){ElMessage.error(e.message)}}
async function viewHistory(u){try{const d=await adminApi.userHistory(u.id);userHistory.value=d.items||[];historyVisible.value=true}catch(e){ElMessage.error(e.message)}}
function formatDate(ts){if(!ts)return'';return new Date(ts).toLocaleString()}
</script>
