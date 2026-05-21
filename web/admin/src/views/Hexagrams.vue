<template>
  <div>
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>卦象任务管理</h2>
    </div>

    <!-- 搜索筛选 -->
    <el-card shadow="never" class="mb-4">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input v-model="userIdFilter" placeholder="用户ID" clearable @keyup.enter="load" />
        </el-col>
        <el-col :span="6">
          <el-input v-model="searchText" placeholder="搜索问题/卦象" clearable @keyup.enter="load" />
        </el-col>
        <el-col :span="6">
          <el-date-picker v-model="dateRange" type="daterange" range-separator="至"
            start-placeholder="开始日期" end-placeholder="结束日期" style="width:100%"
            @change="load" />
        </el-col>
        <el-col :span="6" class="text-right">
          <el-button type="primary" @click="load">刷新</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 记录列表 -->
    <el-card shadow="never" class="mb-4">
      <el-table :data="items" stripe v-loading="loading" empty-text="暂无数据">
        <el-table-column label="ID" width="70" prop="id" align="center" />
        <el-table-column label="用户" width="140">
          <template #default="{ row }">
            <span class="user-name">{{ row.nickname || '微信用户' }}</span>
            <span class="user-id" style="margin-left:4px">#{{ row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="问题" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.question }}</template>
        </el-table-column>
        <el-table-column label="本卦" width="78" align="center">
          <template #default="{ row }">
            <el-tag effect="plain" class="gua-tag">{{ row.primary_gua }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变卦" width="78" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.changing_gua" effect="plain" type="warning" class="gua-tag">{{ row.changing_gua }}</el-tag>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="变爻" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.yao_positions" class="yao-text nowrap">{{ row.yao_positions }}</span>
            <span v-else class="text-muted">无</span>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="130">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <span class="action-cell">
              <el-button size="small" @click="showDetail(row)">详情</el-button>
              <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
            </span>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total"
          layout="total, prev, pager, next" @current-change="load" />
      </div>
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" :title="'卦象详情 #' + (detail?.id || '')" width="720px" top="5vh"
      destroy-on-close>
      <div v-if="detail" class="detail-wrap">
        <!-- 基本信息 -->
        <div class="detail-section detail-meta">
          <div class="meta-item">
            <span class="meta-label">用户</span>
            <span class="meta-value">{{ detail.nickname || '用户' + detail.user_id }} <span class="text-muted">#{{ detail.user_id }}</span></span>
          </div>
          <div class="meta-item">
            <span class="meta-label">时间</span>
            <span class="meta-value">{{ formatDate(detail.created_at) }}</span>
          </div>
        </div>

        <!-- 问题 -->
        <div class="detail-section">
          <div class="section-title">📝 问题</div>
          <div class="question-text">{{ detail.question }}</div>
        </div>

        <!-- 铜钱信息 -->
        <div v-if="parsedToss.length" class="detail-section">
          <div class="section-title">🪙 铜钱信息</div>
          <div class="toss-grid">
            <div class="toss-header">
              <span class="toss-cell">爻位</span>
              <span class="toss-cell">结果</span>
              <span class="toss-cell">三钱</span>
              <span class="toss-cell">阴阳</span>
              <span class="toss-cell">变</span>
            </div>
            <div v-for="t in parsedToss" :key="t.throw" class="toss-row"
              :class="{ 'is-changing': t.result === '老阴' || t.result === '老阳' }">
              <span class="toss-cell toss-label">{{ t.label }}</span>
              <span class="toss-cell" :class="t.yang ? 'text-yang' : 'text-yin'">{{ t.result }}</span>
              <span class="toss-cell toss-coins">
                <span v-for="(cv, ci) in t.coin_values" :key="ci"
                  class="coin-dot" :class="cv === 3 ? 'coin-front' : 'coin-back'">
                  {{ cv === 3 ? '正' : '反' }}
                </span>
              </span>
              <span class="toss-cell">{{ t.yang ? '⚊ 阳' : '⚋ 阴' }}</span>
              <span class="toss-cell">{{ (t.result === '老阴' || t.result === '老阳') ? '● 变' : '—' }}</span>
            </div>
          </div>
        </div>

        <!-- 卦象 -->
        <div class="detail-section">
          <div class="section-title">🏷 卦象</div>
          <div class="hex-info">
            <div class="hex-row">
              <span class="hex-labels">
                <span class="hex-badge primary">本卦</span>
                <span class="hex-name">{{ detail.primary_gua }}</span>
              </span>
              <span v-if="detail.changing_gua" class="hex-arrow">→</span>
              <span v-if="detail.changing_gua" class="hex-labels">
                <span class="hex-badge changing">变卦</span>
                <span class="hex-name">{{ detail.changing_gua }}</span>
              </span>
            </div>
            <div v-if="detail.yao_positions" class="hex-yao">
              <span class="yao-label">变爻：</span>
              <span class="yao-value">{{ detail.yao_positions }}</span>
            </div>
            <div v-if="detail.master_yao > 0" class="hex-master">
              <span class="yao-label">主变爻：</span>
              <span class="yao-value master">第 {{ detail.master_yao }} 爻（最重要）</span>
            </div>
          </div>
        </div>

        <!-- AI 解卦 -->
        <div class="detail-section">
          <div class="section-title">📖 AI 解卦</div>
          <div class="interpretation-wrap">
            <MarkdownRenderer :content="detail.interpretation" />
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import MarkdownRenderer from '../components/MarkdownRenderer.vue'

const items = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const userIdFilter = ref('')
const searchText = ref('')
const dateRange = ref(null)
const detailVisible = ref(false)
const detail = ref(null)

// 解析 toss_data JSON
const parsedToss = computed(() => {
  if (!detail.value?.toss_data) return []
  try {
    return JSON.parse(detail.value.toss_data)
  } catch {
    return []
  }
})

onMounted(() => load())

async function load() {
  loading.value = true
  try {
    const params = { limit: pageSize.value, offset: (page.value - 1) * pageSize.value }
    if (userIdFilter.value) params.userId = userIdFilter.value
    const data = await adminApi.hexagrams(params)
    items.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载失败: ' + e.message)
  } finally {
    loading.value = false
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

function formatDate(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}
</script>

<style scoped>
.mb-4 { margin-bottom: 16px; }
.text-right { text-align: right; }
.text-muted { color: #909399; font-size: 12px; }

/* 用户 */
.user-name { font-weight: 600; font-size: 14px; color: #1c1917; }
.user-id { font-size: 11px; color: #909399; white-space: nowrap; }

/* 卦象标签 */
.gua-tag { font-family: 'Noto Sans SC', sans-serif; font-weight: 600; }
.yao-text { font-size: 13px; color: #b8860b; }
.nowrap { white-space: nowrap; }

/* 操作按钮行 */
.action-cell { display: inline-flex; gap: 6px; white-space: nowrap; }

/* 分页 */
.pagination-wrap { display: flex; justify-content: center; margin-top: 16px; }

/* ===== 详情弹窗 ===== */
.detail-wrap { max-height: 65vh; overflow-y: auto; padding-right: 4px; }

.detail-section {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0ebe0;
}
.detail-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #1c1917;
  margin-bottom: 10px;
}

/* 基本信息 */
.detail-meta {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}
.meta-item { display: flex; gap: 8px; align-items: center; }
.meta-label { font-size: 13px; color: #78716c; min-width: 40px; }
.meta-value { font-size: 14px; font-weight: 500; color: #1c1917; }

/* 问题 */
.question-text {
  background: #faf8f5;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 14px;
  color: #292524;
  line-height: 1.6;
}

/* 铜钱表格 */
.toss-grid {
  border: 1px solid #f0ebe0;
  border-radius: 8px;
  overflow: hidden;
}
.toss-header {
  display: flex;
  background: #faf8f5;
  font-size: 12px;
  font-weight: 600;
  color: #78716c;
  padding: 8px 12px;
  border-bottom: 1px solid #f0ebe0;
}
.toss-row {
  display: flex;
  padding: 8px 12px;
  font-size: 13px;
  border-bottom: 1px solid #f5f2ed;
  transition: background 0.15s;
}
.toss-row:last-child { border-bottom: none; }
.toss-row.is-changing { background: #fff8f0; }
.toss-cell { flex: 1; display: flex; align-items: center; gap: 4px; }
.toss-label { font-weight: 600; color: #292524; }
.text-yang { color: #d4a853; font-weight: 600; }
.text-yin { color: #667eea; font-weight: 600; }

/* 铜钱点 */
.toss-coins { gap: 6px; }
.coin-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  font-size: 11px;
  font-weight: 600;
}
.coin-front { background: #d4a853; color: #fff; }
.coin-back { background: #e8e4d8; color: #78716c; }

/* 卦象信息 */
.hex-info { padding: 4px 0; }
.hex-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 10px;
}
.hex-labels { display: flex; align-items: center; gap: 8px; }
.hex-arrow { font-size: 20px; color: #b8860b; font-weight: 700; }
.hex-name { font-size: 16px; font-weight: 700; color: #292524; }
.hex-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}
.hex-badge.primary { background: #fefaf0; color: #b8860b; border: 1px solid #d4a853; }
.hex-badge.changing { background: #fff0f0; color: #dc2626; border: 1px solid #fca5a5; }

.hex-yao, .hex-master { margin-top: 6px; font-size: 14px; }
.yao-label { color: #78716c; }
.yao-value { color: #292524; }
.yao-value.master { color: #dc2626; font-weight: 600; }

/* 解卦 */
.interpretation-wrap {
  background: #faf8f5;
  padding: 16px 20px;
  border-radius: 8px;
}

/* 滚动条 */
.detail-wrap::-webkit-scrollbar { width: 4px; }
.detail-wrap::-webkit-scrollbar-thumb { background: #d4a85340; border-radius: 2px; }
</style>
