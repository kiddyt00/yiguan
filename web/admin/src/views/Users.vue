<template>
  <div>
    <!-- 统计概览 -->
    <el-row :gutter="16" class="mb-4">
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总用户</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value" style="color:#67C23A">{{ stats.wechat }}</div><div class="stat-label">微信用户</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value" style="color:#409EFF">{{ stats.phone }}</div><div class="stat-label">手机用户</div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never" class="stat-card"><div class="stat-value" style="color:#E6A23C">{{ stats.active }}</div><div class="stat-label">活跃用户</div></el-card></el-col>
    </el-row>

    <!-- 搜索筛选 -->
    <el-card shadow="never" class="mb-4">
      <el-row :gutter="16">
        <el-col :span="8"><el-input v-model="search" placeholder="搜索昵称或手机号" clearable @input="loadUsers" prefix-icon="Search" /></el-col>
        <el-col :span="4"><el-select v-model="typeFilter" placeholder="用户类型" clearable @change="loadUsers" style="width:100%"><el-option label="全部" value="" /><el-option label="微信" value="wechat" /><el-option label="手机" value="phone" /></el-select></el-col>
        <el-col :span="4"><el-select v-model="statusFilter" placeholder="状态" clearable @change="loadUsers" style="width:100%"><el-option label="全部" value="" /><el-option label="启用" value="active" /><el-option label="禁用" value="disabled" /></el-select></el-col>
        <el-col :span="8" class="text-right"><el-button type="primary" @click="loadUsers">刷新</el-button></el-col>
      </el-row>
    </el-card>

    <!-- 用户列表 -->
    <el-card shadow="never">
      <el-table :data="users" stripe v-loading="loading" empty-text="暂无用户">
        <el-table-column label="用户" min-width="220">
          <template #default="{ row }">
            <div style="display:flex;align-items:center;gap:12px">
              <div class="avatar">{{ row.openid ? '微' : '📱' }}</div>
              <div>
                <div class="user-name">{{ row.openid ? (row.nickname&&row.nickname!=='微信用户'?row.nickname:'微信用户 '+row.id) : (row.nickname||row.phone) }}</div>
                <div class="user-meta">
                  <el-tag :type="row.openid ? 'success' : 'info'" size="small" style="margin-right:4px">{{ row.openid ? '微信' : '手机' }}</el-tag>
                  <span v-if="!row.openid">{{ row.phone }}</span>
                </div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="150"><template #default="{ row }">{{ formatDate(row.created_at) }}</template></el-table-column>
        <el-table-column label="角色" width="70"><template #default="{ row }"><el-tag :type="row.role==='admin'?'danger':'info'" size="small">{{ row.role==='admin'?'管理员':'用户' }}</el-tag></template></el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.is_active?'success':'danger'" size="small" effect="plain">
              <span style="display:inline-block;width:6px;height:6px;border-radius:50%;margin-right:4px;background:{{row.is_active?'#67C23A':'#F56C6C'}}"></span>
              {{ row.is_active ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button size="small" :type="row.is_active?'warning':'success'" plain @click="toggleUser(row)">{{ row.is_active ? '禁用' : '启用' }}</el-button>
            <el-button size="small" type="primary" plain @click="adjustQuota(row)">配额</el-button>
            <el-button size="small" type="info" plain @click="viewHistory(row)">记录</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="loadUsers" /></div>
    </el-card>

    <el-dialog v-model="quotaVisible" title="调整配额" width="360px">
      <div class="mb-2">用户：<strong>{{ currentUser?.nickname||currentUser?.phone }}</strong></div>
      <el-input-number v-model="quotaDelta" :min="-100" :max="100" />
      <div style="font-size:12px;color:#909399;margin-top:8px">正数增加，负数减少</div>
      <template #footer><el-button @click="quotaVisible=false">取消</el-button><el-button type="primary" @click="confirmQuota">确认</el-button></template>
    </el-dialog>

    <el-dialog v-model="historyVisible" :title="'起卦记录'" width="760px">
      <el-table :data="userHistory" size="small" v-loading="historyLoading" empty-text="暂无记录">
        <el-table-column prop="id" label="ID" width="50" />
        <el-table-column prop="question" label="问题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="primary_gua" label="本卦" width="70" />
        <el-table-column prop="changing_gua" label="变卦" width="70" />
        <el-table-column label="时间" width="140"><template #default="{ row }">{{ formatDate(row.created_at) }}</template></el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { adminApi } from '../api'
import { ElMessage } from 'element-plus'

const users=ref([]),total=ref(0),page=ref(1),pageSize=ref(20),loading=ref(false)
const search=ref(''),typeFilter=ref(''),statusFilter=ref(''),currentUser=ref(null)
const quotaVisible=ref(false),quotaDelta=ref(0),historyVisible=ref(false),userHistory=ref([]),historyLoading=ref(false)

const stats=computed(()=>{const u=users.value||[];return{total:total.value,wechat:u.filter(x=>x.openid).length,phone:u.filter(x=>!x.openid).length,active:u.filter(x=>x.is_active).length}})

onMounted(loadUsers)
async function loadUsers(){
  loading.value=true
  try{
    const d=await adminApi.users({limit:pageSize.value,offset:(page.value-1)*pageSize.value})
    let items=d.items||[];total.value=d.total||items.length
    if(typeFilter.value==='wechat')items=items.filter(u=>u.openid)
    if(typeFilter.value==='phone')items=items.filter(u=>!u.openid)
    if(statusFilter.value==='active')items=items.filter(u=>u.is_active)
    if(statusFilter.value==='disabled')items=items.filter(u=>!u.is_active)
    if(search.value){const q=search.value.toLowerCase();items=items.filter(u=>(u.nickname||'').toLowerCase().includes(q)||(u.phone||'').includes(q))}
    users.value=items
  }catch(e){ElMessage.error('加载失败: '+e.message)}finally{loading.value=false}
}
async function toggleUser(u){try{await adminApi.toggleUser(u.id);u.is_active=u.is_active?0:1;ElMessage.success(u.is_active?'已启用':'已禁用')}catch(e){ElMessage.error(e.message)}}
function adjustQuota(u){currentUser.value=u;quotaDelta.value=0;quotaVisible.value=true}
async function confirmQuota(){try{await adminApi.adjustQuota(currentUser.value.id,quotaDelta.value);ElMessage.success('调整成功');quotaVisible.value=false}catch(e){ElMessage.error(e.message)}}
async function viewHistory(u){currentUser.value=u;historyLoading.value=true;historyVisible.value=true;try{const d=await adminApi.userHistory(u.id);userHistory.value=d.items||[]}catch(e){ElMessage.error(e.message)}finally{historyLoading.value=false}}
function formatDate(ts){if(!ts)return '';return new Date(ts).toLocaleString('zh-CN',{month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'})}
</script>

<style scoped>
.stat-card{text-align:center}
.stat-value{font-size:32px;font-weight:700;color:#303133}
.stat-label{font-size:13px;color:#909399;margin-top:4px}
.avatar{width:36px;height:36px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:13px;font-weight:600;background:#ecf5ff;color:#409EFF;flex-shrink:0}
.user-name{font-weight:600;font-size:14px}
.user-meta{font-size:12px;color:#909399;margin-top:2px}
.pagination-wrap{display:flex;justify-content:center;margin-top:16px}
.mb-2{margin-bottom:8px}
.mb-4{margin-bottom:16px}
.text-right{text-align:right}
</style>
