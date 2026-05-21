<template>
  <div>
    <div class="page-header"><h2>卦象任务管理</h2></div>
    <el-card shadow="never" class="filter-bar">
      <el-row :gutter="12">
        <el-col :span="6"><el-input v-model="userIdFilter" placeholder="用户ID" clearable @keyup.enter="load" /></el-col>
        <el-col :span="6"><el-input v-model="searchText" placeholder="搜索问题" clearable @keyup.enter="load" /></el-col>
        <el-col :span="6"><el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width:100%" @change="load" /></el-col>
        <el-col :span="6" class="text-right"><el-button type="primary" @click="load">刷新</el-button></el-col>
      </el-row>
    </el-card>
    <div class="table-wrap">
      <table class="custom-table">
        <thead>
          <tr>
            <th class="cid">ID</th>
            <th class="cu">用户</th>
            <th class="cuid">UID</th>
            <th class="cq">问题</th>
            <th class="cg">本卦</th>
            <th class="cg">变卦</th>
            <th class="cy">变爻</th>
            <th class="ct">时间</th>
            <th class="ca">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="9" class="empty">加载中...</td></tr>
          <tr v-else-if="!items.length"><td colspan="9" class="empty">暂无数据</td></tr>
          <tr v-for="row in items" :key="row.id" @click="showDetail(row)">
            <td class="cid">{{ row.id }}</td>
            <td class="cu"><span class="un">{{ row.nickname || '微信用户' }}</span></td>
            <td class="cuid">{{ row.user_id }}</td>
            <td class="cq" :title="row.question">{{ row.question }}</td>
            <td class="cg"><span class="t t1">{{ row.primary_gua }}</span></td>
            <td class="cg"><span v-if="row.changing_gua" class="t t2">{{ row.changing_gua }}</span><span v-else class="g">—</span></td>
            <td class="cy"><span class="yao-text">{{ row.yao_positions || '—' }}</span></td>
            <td class="ct">{{ formatDate(row.created_at) }}</td>
            <td class="ca" @click.stop><button class="b" @click="showDetail(row)">详情</button><button class="b bd" @click="remove(row)">删除</button></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="pg"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" background small /></div>
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
const items=ref([]),total=ref(0),page=ref(1),ps=ref(20),loading=ref(false)
const uf=ref(''),st=ref(''),dr=ref(null),dv=ref(false),dl=ref(null)
const pt=computed(()=>{if(!dl.value?.toss_data)return[];try{return JSON.parse(dl.value.toss_data)}catch{return[]}})
onMounted(()=>ld())
async function ld(){loading.value=true;try{const p={limit:ps.value,offset:(page.value-1)*ps.value};if(uf.value)p.userId=uf.value;const d=await adminApi.hexagrams(p);items.value=d.items||[];total.value=d.total||0}catch(e){ElMessage.error('加载失败: '+e.message)}finally{loading.value=false}}
function sd(r){dl.value=r;dv.value=true}
async function rm(r){try{await ElMessageBox.confirm('确定删除？','确认');await adminApi.deleteHexagram(r.id);ElMessage.success('已删除');ld()}catch(e){if(e!=='cancel')ElMessage.error(e.message)}}
function fd(ts){if(!ts)return'';return new Date(ts).toLocaleString('zh-CN',{month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'})}
</script>
<style scoped>
.filter-bar{margin-bottom:12px}.text-right{text-align:right}
.table-wrap{background:#fff;border:1px solid #e5ddd0;border-radius:10px;overflow:hidden}
.custom-table{width:100%;border-collapse:collapse;font-size:13px;table-layout:fixed}
.custom-table thead{background:#f8f5f0}
.custom-table th{padding:9px 6px;font-weight:600;color:#8a7e72;text-align:left;border-bottom:1px solid #e5ddd0;white-space:nowrap;font-size:12px}
.custom-table td{padding:9px 6px;border-bottom:1px solid #eee8e0;color:#292524;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.custom-table tbody tr{cursor:pointer;transition:background .1s}
.custom-table tbody tr:hover{background:#fdf8f0}
.custom-table tbody tr:last-child td{border-bottom:none}
.empty{text-align:center;padding:36px 0;color:#999}
.cid{width:44px;text-align:center}
.cu{width:76px}.cuid{width:44px;text-align:center;color:#999;font-size:12px}
.cq{width:170px}.cg{width:64px;text-align:center}
.cy{width:200px;overflow:visible;text-overflow:clip}
.ct{width:105px;color:#8a7e72;font-size:12px}
.ca{width:120px;text-align:center;white-space:nowrap}
.un{font-weight:600;color:#292524;font-size:13px}
.t{display:inline-block;padding:1px 7px;border-radius:4px;font-size:12px;font-weight:600}
.t1{background:#fdf8f0;color:#b8860b;border:1px solid #d4a853}
.t2{background:#fef0f0;color:#c62828;border:1px solid #fca5a5}
.g{color:#999;font-size:12px}
.yao-text{font-size:13px;color:#8a6020}
.b{display:inline-flex;align-items:center;padding:4px 10px;border:1px solid #d4a853;border-radius:5px;font-size:12px;font-weight:500;cursor:pointer;transition:all .12s;background:#fff;color:#b8860b}
.b:hover{background:#d4a853;color:#fff}
.bd{border-color:#fca5a5;color:#c62828}
.bd:hover{background:#c62828;color:#fff;border-color:#c62828}
.b+.b{margin-left:5px}
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
