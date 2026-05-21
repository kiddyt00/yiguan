<template>
  <div>
    <div class="page-header">
      <h2>卦象任务管理</h2>
    </div>

    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-bar">
      <el-row :gutter="12">
        <el-col :span="6"><el-input v-model="userIdFilter" placeholder="用户ID" clearable @keyup.enter="load" /></el-col>
        <el-col :span="6"><el-input v-model="searchText" placeholder="搜索问题" clearable @keyup.enter="load" /></el-col>
        <el-col :span="6">
          <el-date-picker v-model="dateRange" type="daterange" range-separator="至"
            start-placeholder="开始" end-placeholder="结束" style="width:100%" @change="load" />
        </el-col>
        <el-col :span="6" class="text-right"><el-button type="primary" @click="load">刷新</el-button></el-col>
      </el-row>
    </el-card>

    <!-- 自定义表格 -->
    <div class="table-wrap">
      <table class="custom-table">
        <thead>
          <tr>
            <th class="col-id">ID</th>
            <th class="col-user">用户</th>
            <th class="col-uid">UID</th>
            <th class="col-q">问题</th>
            <th class="col-gua">本卦</th>
            <th class="col-gua">变卦</th>
            <th class="col-yao">变爻</th>
            <th class="col-time">时间</th>
            <th class="col-action">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="9" class="empty">加载中...</td></tr>
          <tr v-else-if="!items.length"><td colspan="9" class="empty">暂无数据</td></tr>
          <tr v-for="row in items" :key="row.id" @click="showDetail(row)">
            <td class="col-id">{{ row.id }}</td>
            <td class="col-user"><span class="uname">{{ row.nickname || '微信用户' }}</span></td>
            <td class="col-uid">{{ row.user_id }}</td>
            <td class="col-q" :title="row.question">{{ row.question }}</td>
            <td class="col-gua"><span class="tag tag-primary">{{ row.primary_gua }}</span></td>
            <td class="col-gua"><span v-if="row.changing_gua" class="tag tag-change">{{ row.changing_gua }}</span><span v-else class="mute">—</span></td>
            <td class="col-yao">{{ row.yao_positions || '—' }}</td>
            <td class="col-time">{{ formatDate(row.created_at) }}</td>
            <td class="col-action" @click.stop>
              <button class="btn btn-sm" @click="showDetail(row)">详情</button>
              <button class="btn btn-sm btn-danger" @click="remove(row)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="load" background small />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" :title="'卦象详情 #' + (detail?.id || '')" width="720px" top="5vh" destroy-on-close>
      <div v-if="detail" class="detail-wrap">
        <div class="detail-section detail-meta">
          <div class="meta-item"><span class="meta-label">用户</span><span class="meta-value">{{ detail.nickname || '微信用户' }} <span class="mute">#{{ detail.user_id }}</span></span></div>
          <div class="meta-item"><span class="meta-label">时间</span><span class="meta-value">{{ formatDate(detail.created_at) }}</span></div>
        </div>
        <div class="detail-section">
          <div class="section-title">📝 问题</div>
          <div class="question-text">{{ detail.question }}</div>
        </div>
        <div v-if="parsedToss.length" class="detail-section">
          <div class="section-title">🪙 铜钱信息</div>
          <div class="toss-grid">
            <div class="toss-header">
              <span>爻位</span><span>结果</span><span>三钱</span><span>阴阳</span><span>变</span>
            </div>
            <div v-for="t in parsedToss" :key="t.throw" class="toss-row"
              :class="{ changing: t.result === '老阴' || t.result === '老阳' }">
              <span class="fw-600">{{ t.label }}</span>
              <span :class="t.yang ? 'c-gold' : 'c-blue'">{{ t.result }}</span>
              <span class="coins">
                <span v-for="(cv, ci) in t.coin_values" :key="ci" :class="cv === 3 ? 'dot dot-front' : 'dot dot-back'">{{ cv === 3 ? '正' : '反' }}</span>
              </span>
              <span>{{ t.yang ? '⚊ 阳' : '⚋ 阴' }}</span>
              <span>{{ (t.result === '老阴' || t.result === '老阳') ? '● 变' : '—' }}</span>
            </div>
          </div>
        </div>
        <div class="detail-section">
          <div class="section-title">🏷 卦象</div>
          <div class="hex-info">
            <div class="hex-row">
              <span class="badge primary">本卦</span><span class="hex-name">{{ detail.primary_gua }}</span>
              <span v-if="detail.changing_gua" class="hex-arrow">→</span>
              <span v-if="detail.changing_gua"><span class="badge changing">变卦</span><span class="hex-name">{{ detail.changing_gua }}</span></span>
            </div>
            <div v-if="detail.yao_positions" class="mt-1"><span class="mute">变爻：</span>{{ detail.yao_positions }}</div>
            <div v-if="detail.master_yao > 0" class="mt-1"><span class="mute">主变爻：</span><span class="c-red">第 {{ detail.master_yao }} 爻</span></div>
          </div>
        </div>
        <div class="detail-section">
          <div class="section-title">📖 AI 解卦</div>
          <div class="interpretation-wrap"><MarkdownRenderer :content="detail.interpretation" /></div>
        </div>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import MarkdownRenderer from '../components/MarkdownRenderer.vue'

const items = ref([]), total = ref(0), page = ref(1), pageSize = ref(20), loading = ref(false)
const userIdFilter = ref(''), searchText = ref(''), dateRange = ref(null)
const detailVisible = ref(false), detail = ref(null)
const parsedToss = computed(() => { if (!detail.value?.toss_data) return []; try { return JSON.parse(detail.value.toss_data) } catch { return [] } })

onMounted(() => load())
async function load() {
  loading.value = true
  try {
    const p = { limit: pageSize.value, offset: (page.value - 1) * pageSize.value }
    if (userIdFilter.value) p.userId = userIdFilter.value
    const d = await adminApi.hexagrams(p)
    items.value = d.items || []; total.value = d.total || 0
  } catch (e) { ElMessage.error('加载失败: ' + e.message) } finally { loading.value = false }
}
function showDetail(r) { detail.value = r; detailVisible.value = true }
async function remove(r) { try { await ElMessageBox.confirm('确定删除？', '确认'); await adminApi.deleteHexagram(r.id); ElMessage.success('已删除'); load() } catch (e) { if (e !== 'cancel') ElMessage.error(e.message) } }
function formatDate(ts) { if (!ts) return ''; return new Date(ts).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) }
</script>

<style scoped>
.filter-bar { margin-bottom: 12px; }
.text-right { text-align: right; }

/* ===== 自定义表格 ===== */
.table-wrap {
  background: #fff;
  border: 1px solid #f0ebe0;
  border-radius: 10px;
  overflow: hidden;
}
.custom-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  table-layout: fixed;
}
.custom-table thead {
  background: #faf8f5;
}
.custom-table th {
  padding: 10px 8px;
  font-weight: 600;
  color: #78716c;
  text-align: left;
  border-bottom: 1px solid #f0ebe0;
  white-space: nowrap;
  font-size: 12px;
}
.custom-table td {
  padding: 10px 8px;
  border-bottom: 1px solid #f5f2ed;
  color: #292524;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.custom-table tbody tr {
  cursor: pointer;
  transition: background 0.12s;
}
.custom-table tbody tr:hover {
  background: #fefaf0;
}
.custom-table tbody tr:last-child td {
  border-bottom: none;
}
.custom-table .empty {
  text-align: center;
  padding: 40px 0;
  color: #909399;
}

/* 列宽 */
.col-id { width: 48px; text-align: center; }
.col-user { width: 76px; }
.col-uid { width: 48px; text-align: center; color: #909399; font-size: 12px; }
.col-q { }
.col-gua { width: 62px; text-align: center; }
.col-yao { width: 130px; }
.col-time { width: 110px; color: #78716c; font-size: 12px; }
.col-action { width: 110px; text-align: center; }

/* 用户 */
.uname { font-weight: 600; color: #292524; font-size: 13px; }

/* 卦象标签 */
.tag {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}
.tag-primary { background: #fefaf0; color: #b8860b; border: 1px solid #d4a853; }
.tag-change { background: #fff0f0; color: #dc2626; border: 1px solid #fca5a5; }

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border: 1px solid #d4a853;
  border-radius: 5px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  background: #fff;
  color: #b8860b;
}
.btn:hover { background: #d4a853; color: #fff; }
.btn-danger { border-color: #fca5a5; color: #dc2626; }
.btn-danger:hover { background: #dc2626; color: #fff; border-color: #dc2626; }
.btn + .btn { margin-left: 4px; }

/* 分页 */
.pagination-wrap { display: flex; justify-content: center; margin-top: 14px; }

/* 工具类 */
.mute { color: #909399; font-size: 12px; }
.mt-1 { margin-top: 6px; }
.fw-600 { font-weight: 600; }
.c-gold { color: #d4a853; font-weight: 600; }
.c-blue { color: #667eea; font-weight: 600; }
.c-red { color: #dc2626; font-weight: 600; }

/* ===== 详情弹窗 ===== */
.detail-wrap { max-height: 65vh; overflow-y: auto; padding-right: 4px; }
.detail-section { margin-bottom: 18px; padding-bottom: 18px; border-bottom: 1px solid #f0ebe0; }
.detail-section:last-child { border-bottom: none; margin-bottom: 0; padding-bottom: 0; }
.section-title { font-size: 15px; font-weight: 700; color: #1c1917; margin-bottom: 10px; }
.detail-meta { display: flex; gap: 32px; flex-wrap: wrap; }
.meta-item { display: flex; gap: 8px; align-items: center; }
.meta-label { font-size: 13px; color: #78716c; min-width: 40px; }
.meta-value { font-size: 14px; font-weight: 500; color: #1c1917; }
.question-text { background: #faf8f5; padding: 12px 16px; border-radius: 8px; font-size: 14px; color: #292524; line-height: 1.6; }

/* 铜钱 */
.toss-grid { border: 1px solid #f0ebe0; border-radius: 8px; overflow: hidden; }
.toss-header { display: flex; background: #faf8f5; font-size: 12px; font-weight: 600; color: #78716c; padding: 8px 12px; border-bottom: 1px solid #f0ebe0; }
.toss-header span, .toss-row span { flex: 1; }
.toss-row { display: flex; padding: 7px 12px; font-size: 13px; border-bottom: 1px solid #f5f2ed; align-items: center; }
.toss-row:last-child { border-bottom: none; }
.toss-row.changing { background: #fff8f0; }
.coins { display: flex; gap: 5px; }
.dot { display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; border-radius: 50%; font-size: 11px; font-weight: 600; }
.dot-front { background: #d4a853; color: #fff; }
.dot-back { background: #e8e4d8; color: #78716c; }

/* 卦象 */
.hex-info { padding: 2px 0; }
.hex-row { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; flex-wrap: wrap; }
.hex-arrow { font-size: 18px; color: #b8860b; font-weight: 700; }
.hex-name { font-size: 16px; font-weight: 700; color: #292524; }
.badge { display: inline-block; padding: 2px 7px; border-radius: 4px; font-size: 11px; font-weight: 600; margin-right: 4px; }
.badge.primary { background: #fefaf0; color: #b8860b; border: 1px solid #d4a853; }
.badge.changing { background: #fff0f0; color: #dc2626; border: 1px solid #fca5a5; }

/* Markdown */
.interpretation-wrap { background: #faf8f5; padding: 16px 20px; border-radius: 8px; }

/* 滚动条 */
.detail-wrap::-webkit-scrollbar { width: 4px; }
.detail-wrap::-webkit-scrollbar-thumb { background: #d4a85340; border-radius: 2px; }
</style>
