<template>
  <div>
    <div class="page-header"><h2>卦象任务管理</h2></div>
    <el-card shadow="never" class="filter-bar">
      <el-row :gutter="8">
        <el-col :span="4"><el-input v-model="kw" placeholder="问题/卦名/用户" clearable @keydown="onEnter" /></el-col>
        <el-col :span="3"><el-input v-model="uid" placeholder="用户ID" clearable @keydown="onEnter" /></el-col>
        <el-col :span="4"><el-input v-model="nickname" placeholder="用户昵称" clearable @keydown="onEnter" /></el-col>
        <el-col :span="4"><el-input v-model="guaName" placeholder="本卦/变卦名" clearable @keydown="onEnter" /></el-col>
        <el-col :span="5"><el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width:100%" @change="load" /></el-col>
        <el-col :span="4" class="text-right"><el-button type="primary" @click="load">搜索</el-button></el-col>
      </el-row>
    </el-card>
    <el-table :data="items" stripe v-loading="loading" empty-text="暂无数据" size="small" style="width:100%" @row-click="showDetail">
      <el-table-column label="ID" prop="id" width="55" align="center" />
      <el-table-column label="用户" min-width="100">
        <template #default="{row}"><span style="font-weight:600">{{ row.nickname||'微信用户' }}</span></template>
      </el-table-column>
      <el-table-column label="UID" prop="user_id" width="55" align="center" />
      <el-table-column label="问题" min-width="150" show-overflow-tooltip>
        <template #default="{row}">{{ row.question }}</template>
      </el-table-column>
      <el-table-column label="本卦" width="70" align="center">
        <template #default="{row}"><el-tag size="small" effect="plain" style="background:#fdf8f0;color:#b8860b;border-color:#d4a853">{{ row.primary_gua }}</el-tag></template>
      </el-table-column>
      <el-table-column label="变卦" width="70" align="center">
        <template #default="{row}"><el-tag v-if="row.changing_gua" size="small" effect="plain" style="background:#fef0f0;color:#c62828;border-color:#fca5a5">{{ row.changing_gua }}</el-tag><span v-else style="color:#999">—</span></template>
      </el-table-column>
      <el-table-column label="变爻" width="140" show-overflow-tooltip>
        <template #default="{row}"><span style="color:#8a6020">{{ row.yao_positions||'—' }}</span></template>
      </el-table-column>
      <el-table-column label="时间" width="110">
        <template #default="{row}"><span style="color:#8a7e72;font-size:12px">{{ formatDate(row.created_at) }}</span></template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{row}">
          <el-button size="small" @click.stop="showDetail(row)" style="padding:4px 10px;font-size:12px">详情</el-button>
          <el-button size="small" type="danger" plain @click.stop="remove(row)" style="padding:4px 10px;font-size:12px">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pg"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" background size="small" /></div>
    <el-dialog v-model="detailVisible" :title="'卦象详情 #' + (detail?.id || '')" width="720px" top="5vh" destroy-on-close>
      <div v-if="detail" class="dw">
        <div class="ds dm"><div class="mi"><span class="ml">用户</span><span class="mv">{{ detail.nickname || '微信用户' }} <span class="g">#{{ detail.user_id }}</span></span></div><div class="mi"><span class="ml">时间</span><span class="mv">{{ formatDate(detail.created_at) }}</span></div></div>
        <div class="ds"><div class="st">📝 问题</div><div class="qt">{{ detail.question }}</div></div>
        <div v-if="pt.length" class="ds"><div class="st">🪙 铜钱信息</div>
          <div class="tg"><div class="th"><span>爻位</span><span>结果</span><span>三钱</span><span>阴阳</span><span>变</span></div>
            <div v-for="t in pt" :key="t.throw" class="tr" :class="{ch:t.result==='老阴'||t.result==='老阳'}">
              <span class="fw6">{{ t.label }}</span><span :class="t.yang?'cg2':'cb'">{{ t.result }}</span>
              <span class="cs"><span v-for="(cv,ci) in t.coin_values" :key="ci" :class="cv===3?'d df':'d db'">{{ cv===3?'正':'反' }}</span></span>
              <span>{{ t.yang?'⚊ 阳':'⚋ 阴' }}</span><span>{{ (t.result==='老阴'||t.result==='老阳')?'● 变':'—' }}</span>
            </div>
          </div>
        </div>
        <div class="ds"><div class="st">🏷 卦象</div>
          <div class="hi"><div class="hr"><span class="b b1">本卦</span><span class="hn">{{ detail.primary_gua }}</span><span v-if="detail.changing_gua" class="ha">→</span><span v-if="detail.changing_gua"><span class="b b2">变卦</span><span class="hn">{{ detail.changing_gua }}</span></span></div>
          <div v-if="detail.yao_positions" class="mt1"><span class="g">变爻：</span>{{ detail.yao_positions }}</div><div v-if="detail.master_yao>0" class="mt1"><span class="g">主变爻：</span><span class="cr">第{{ detail.master_yao }}爻</span></div></div>
        </div>
        <div class="ds"><div class="st">📖 AI 解卦</div><div class="iw"><MarkdownRenderer :content="detail.interpretation" /></div></div>
      </div>
      <template #footer><el-button @click="detailVisible=false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import MarkdownRenderer from '../components/MarkdownRenderer.vue'
const items=ref([]),total=ref(0),page=ref(1),pageSize=ref(20),loading=ref(false)
const kw=ref(''),uid=ref(''),nickname=ref(''),guaName=ref(''),dateRange=ref(null),detailVisible=ref(false),detail=ref(null)
const pt=computed(()=>{if(!detail.value?.toss_data)return[];try{return JSON.parse(detail.value.toss_data)}catch{return[]}})
onMounted(()=>load())
function onEnter(e){if(e.key==='Enter')load()}
async function load(){loading.value=true;try{const p={limit:pageSize.value,offset:(page.value-1)*pageSize.value};if(kw.value||nickname.value||guaName.value){p.keyword=[kw.value,nickname.value,guaName.value].filter(Boolean).join(' ')};if(uid.value)p.user_id=uid.value;if(dateRange.value){p.date_from=formatDateStr(dateRange.value[0]);p.date_to=formatDateStr(dateRange.value[1])};const d=await adminApi.hexagrams(p);items.value=d.items||[];total.value=d.total||0}catch(e){ElMessage.error('加载失败: '+e.message)}finally{loading.value=false}}
function formatDateStr(d){const y=d.getFullYear(),m=String(d.getMonth()+1).padStart(2,'0'),day=String(d.getDate()).padStart(2,'0');return y+'-'+m+'-'+day}
function showDetail(row){detail.value=row;detailVisible.value=true}
async function remove(row){try{await ElMessageBox.confirm('确定删除？','确认');await adminApi.deleteHexagram(row.id);ElMessage.success('已删除');load()}catch(e){if(e!=='cancel')ElMessage.error(e.message)}}
function formatDate(ts){if(!ts)return'';return new Date(ts).toLocaleString('zh-CN',{month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'})}
</script>
<style scoped>
.filter-bar{margin-bottom:12px}.text-right{text-align:right}
.pg{display:flex;justify-content:center;margin-top:12px}
.dw{max-height:65vh;overflow-y:auto;padding-right:4px}
.ds{margin-bottom:16px;padding-bottom:16px;border-bottom:1px solid #eee8e0}
.ds:last-child{border-bottom:none;margin-bottom:0;padding-bottom:0}
.st{font-size:15px;font-weight:700;color:#1c1917;margin-bottom:10px}
.dm{display:flex;gap:28px;flex-wrap:wrap}
.mi{display:flex;gap:8px;align-items:center}
.ml{font-size:13px;color:#8a7e72;min-width:36px}
.mv{font-size:14px;font-weight:500;color:#1c1917}
.qt{background:#f8f5f0;padding:12px 16px;border-radius:8px;font-size:14px;color:#292524;line-height:1.6}
.tg{border:1px solid #e5ddd0;border-radius:8px;overflow:hidden}
.th{display:flex;background:#f8f5f0;font-size:12px;font-weight:600;color:#8a7e72;padding:8px 12px;border-bottom:1px solid #e5ddd0}
.th span,.tr span{flex:1}
.tr{display:flex;padding:7px 12px;font-size:13px;border-bottom:1px solid #eee8e0;align-items:center}
.tr:last-child{border-bottom:none}
.tr.ch{background:#fdf8f0}
.cs{display:flex;gap:5px}
.d{display:inline-flex;align-items:center;justify-content:center;width:23px;height:23px;border-radius:50%;font-size:11px;font-weight:600}
.df{background:#d4a853;color:#fff}
.db{background:#e5ddd0;color:#8a7e72}
.hi{padding:2px 0}
.hr{display:flex;align-items:center;gap:10px;margin-bottom:8px;flex-wrap:wrap}
.ha{font-size:18px;color:#b8860b;font-weight:700}
.hn{font-size:16px;font-weight:700;color:#292524}
.bl{display:inline-block;padding:2px 7px;border-radius:4px;font-size:11px;font-weight:600;margin-right:4px}
.bl.b1{background:#fdf8f0;color:#b8860b;border:1px solid #d4a853}
.bl.b2{background:#fef0f0;color:#c62828;border:1px solid #fca5a5}
.mt1{margin-top:6px}
.cr{color:#c62828;font-weight:600}
.cg2{color:#d4a853;font-weight:600}
.cb{color:#667eea;font-weight:600}
.fw6{font-weight:600}
.iw{background:#f8f5f0;padding:16px 20px;border-radius:8px}
.dw::-webkit-scrollbar{width:4px}
.dw::-webkit-scrollbar-thumb{background:#d4a85340;border-radius:2px}
</style>
