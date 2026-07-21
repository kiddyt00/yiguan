<template>
  <div>
    <div class="page-header"><h2>📋 开发进度</h2></div>
    <div class="stats-grid">
      <div class="stat-card clickable" v-for="s in stats" :key="s.label" @click="scrollTo(s.scrollTo)" :style="'--hl:'+s.color">
        <div class="stat-value" :style="{ color: s.color }">{{ s.count }}</div>
        <div class="stat-label">{{ s.label }}</div>
        <div class="stat-hint">点击查看 →</div>
      </div>
    </div>
    <div class="two-col">
      <div class="col-tl">
        <div ref="tlSection" class="tl-section">
          <h3 class="section-title">🕐 完整开发时间线（共 {{totalCommits}} 次提交）</h3>
          <div class="tl">
            <div class="tl-spine"></div>
            <div v-for="(phase, i) in phases" :key="i" class="tl-row" :class="i%2===0?'tl-left':'tl-right'">
              <div class="tl-card">
                <div class="tl-dot" :class="phase.dotClass"></div>
                <div class="tl-body">
                  <div class="tl-h"><span class="tl-date">{{ phase.period }}</span><span class="tl-tag" :class="phase.tagClass">{{ phase.tag }}</span><span class="tl-cnt">{{ phase.commits }}次</span></div>
                  <div class="tl-title">{{ phase.title }}</div>
                  <div v-if="phase.subtitle" class="tl-sub">{{ phase.subtitle }}</div>
                  <div class="tl-pbar"><div class="tl-pfill" :style="'width:'+phase.pct+'%'"></div></div>
                  <ul class="tl-items">
                    <li v-for="t in phase.tasks" :key="t.text" class="tl-item"><span class="tl-chk">{{ t.done?'✅':'⬜' }}</span>{{ t.text }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="col-todo">
        <div ref="todoSection" class="todo-wrap">
          <h3>📌 待办事项</h3>
          <div v-for="(g,gi) in todos" :key="gi" class="tg"><div class="tp" :style="{color:g.color}">{{ g.label }}<span class="tbadge">{{ g.items.filter(t=>!t.done).length }}</span></div>
            <div v-for="(t,ti) in g.items" :key="ti" class="ti" :class="t.done?'done':''"><span>{{ t.done?'✅':'⬜' }}</span><div class="tb"><div class="tt">{{ t.text }}</div><div v-if="t.note" class="tn">{{ t.note }}</div></div></div>
          </div>
          <div class="di"><div class="dit">🚀 部署信息</div><div class="dr"><span>首次上线</span><span>2026-06-01</span></div><div class="dr"><span>运行天数</span><span>{{ stats[3].count }}天</span></div><div class="dr"><span>生产域名</span><span>zgjz.insightj.cn</span></div><div class="dr"><span>服务器</span><span>腾讯云 Ubuntu 24.04</span></div><div class="dr"><span>SSL到期</span><span>2026-09-14</span></div><div class="dr"><span>代码提交</span><span>{{ totalCommits }}次</span></div></div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref } from 'vue'
const d0=new Date('2026-06-01'),days=Math.floor((new Date()-d0)/864e5)
const phases=[
  {period:'2026-05-07',tag:'MVP',tc:'tv1',dc:'dv1',title:'MVP v1.0 — 从零搭建',sub:'24次 · Go+HTMX',commits:24,tasks:[
    {d:1,t:'项目骨架：go.mod/main.go/config'},{d:1,t:'周易引擎：64卦+铜钱起爻+变卦'},{d:1,t:'千问API：prompt+解卦'},{d:1,t:'HTTP处理器+模板：首页/算卦/结果'},{d:1,t:'CSS动画+E2E测试5/5'},
  ]},
  {period:'2026-05-07~08',tag:'重构',tc:'tv2',dc:'dv2',title:'v2.0前后端分离',sub:'29次 · Go+Vue3 SPA+Admin',commits:29,tasks:[
    {d:1,t:'SQLite存储层：用户/Quota/历史/管理'},{d:1,t:'JWT鉴权+注册/登录API'},{d:1,t:'Vue3前台：5页面+双主题'},{d:1,t:'Vue3管理后台+Admin API'},{d:1,t:'多模型LLM（千问/DeepSeek）'},{d:1,t:'Docker Compose+Nginx'},
  ]},
  {period:'2026-05-09',tag:'增强',tc:'tv2',dc:'dv2',title:'v2.0功能增强',sub:'42次 · 模型/广告/i18n',commits:42,tasks:[
    {d:1,t:'LLM重试+超时配置'},{d:1,t:'优雅关闭+健康检查+日志'},{d:1,t:'模型管理后台（25家供应商）'},{d:1,t:'广告系统+管理'},{d:1,t:'i18n中英双语+翻译API'},{d:1,t:'限流+Docker优化'},
  ]},
  {period:'2026-05-10',tag:'打磨',tc:'tv2',dc:'dv2',title:'体验打磨&Bug修复',sub:'12次 · 测试/样式',commits:12,tasks:[
    {d:1,t:'测试覆盖（auth/handler/engine）'},{d:1,t:'暗黑模式+样式打磨'},{d:1,t:'登录页重做+密码找回'},
  ]},
  {period:'2026-05-12',tag:'完善',tc:'tv2',dc:'dv2',title:'功能完善',sub:'38次 · 历史/分享/广告/短信',commits:38,tasks:[
    {d:1,t:'历史记录搜索/分页/翻译'},{d:1,t:'截图生成+分享'},{d:1,t:'广告观看+每日上限'},{d:1,t:'短信验证码（开发模式）'},{d:1,t:'管理后台卦象/用户/广告/模型'},{d:1,t:'数据分析页面'},
  ]},
  {period:'2026-05-19~21',tag:'小程序',tc:'tv2',dc:'dv2',title:'微信小程序开发',sub:'46次 · UniApp+原生双版',commits:46,tasks:[
    {d:1,t:'UniApp版（Vue3编译）'},{d:1,t:'原生miniapp-native'},{d:1,t:'微信小程序登录'},{d:1,t:'首页/结果/历史/个人中心'},{d:1,t:'小程序广告'},
  ]},
  {period:'2026-06-01',tag:'上线',tc:'tlaunch',dc:'dlaunch',title:'v2.1生产部署🚀',sub:'5次 · 腾讯云Docker',commits:5,tasks:[
    {d:1,t:'腾讯云+域名zgjz.insightj.cn'},{d:1,t:'Docker生产编排+Nginx'},{d:1,t:'7用户·138条占卜'},
  ]},
  {period:'2026-07-16',tag:'登录',tc:'tdone',dc:'ddone',title:'微信扫码登录+安全加固',sub:'46次 · wxLogin.js',commits:46,tasks:[
    {d:1,t:'wxLogin.js SDK替代自绘二维码'},{d:1,t:'WeChat OAuth配置'},{d:1,t:'CORS白名单+限流60/min'},{d:1,t:'Docker日志+前端请求封装'},
  ]},
  {period:'2026-07-17',tag:'支付',tc:'tdone',dc:'ddone',title:'支付系统',sub:'23次 · 微信支付+JSAPI',commits:23,tasks:[
    {d:1,t:'Web扫码支付(Native)上线'},{d:1,t:'二维码渲染修复(Vue nextTick)'},{d:1,t:'JSAPI小程序支付后端'},{d:1,t:'截图离屏渲染修复'},{d:1,t:'支付宝接入规划'},
  ]},
  {period:'2026-07-20~21',tag:'完成',tc:'tdone',dc:'ddone',title:'小程序全流程打通',sub:'14次 · 支付+管理后台',commits:14,tasks:[
    {d:1,t:'小程序API地址更新11处'},{d:1,t:'miniprogram-ci上传工具'},{d:1,t:'miniapp-native+miniapp支付'},{d:1,t:'JSAPI缺少openid修复'},{d:1,t:'WX_APPID/WX_SECRET修复'},{d:1,t:'管理后台进度页'},
  ]},
]
const total=phases.reduce((s,p)=>s+p.commits,0)
const todos=[
  {l:'🔴 急迫',c:'#c62828',is:[{d:0,t:'小程序提交审核+正式发布',n:'体验版已测试通过'}]},
  {l:'🟡 待开发',c:'#d4a853',is:[{d:0,t:'支付宝扫码支付接入',n:'等商户资料'},{d:0,t:'短信服务接入',n:'当前仅打印日志'}]},
  {l:'🟢 优化项',c:'#667eea',is:[{d:0,t:'前端SSE流自动重连'},{d:0,t:'零停机部署'},{d:0,t:'订单管理页面'},{d:0,t:'数据库自动备份'}]},
]
const done=phases.flatMap(p=>p.tasks).filter(t=>t.d).length
const todo=todos[0].is.filter(t=>!t.d).length+todos[1].is.filter(t=>!t.d).length
const opt=todos[2].is.filter(t=>!t.d).length
const stats=[{l:'已完成',c:done,cl:'#d4a853',s:'tl'},{l:'待办',c:todo,cl:'#c62828',s:'todo'},{l:'优化',c:opt,cl:'#667eea',s:'opt'},{l:'运行天数',c:days,cl:'#2e7d32',s:'deploy'}]
const tl=ref(null),ts=ref(null)
function scrollTo(t){const e=t==='tl'?tl.value:ts.value;if(!e)return;e.scrollIntoView({behavior:'smooth'});e.classList.add('hl');setTimeout(()=>e.classList.remove('hl'),1200)}
</script>
<style scoped>
.page-header h2{font-size:20px;font-weight:700;color:#1c1917;margin-bottom:20px}
.stats-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:24px}
.stat-card{background:#fff;border:1px solid #e5ddd0;border-radius:10px;padding:18px;text-align:center;cursor:pointer;transition:all .2s;position:relative}
.stat-card:hover{transform:translateY(-2px);box-shadow:0 4px 16px rgba(0,0,0,.08);border-color:var(--hl)}
.stat-value{font-size:28px;font-weight:800;line-height:1}
.stat-label{font-size:13px;color:#8a7e72;margin-top:6px}
.stat-hint{font-size:11px;color:#ccc;margin-top:4px;opacity:0}
.stat-card:hover .stat-hint{opacity:1}
.two-col{display:flex;gap:20px;align-items:flex-start}
.col-tl{flex:5;min-width:0}
.col-todo{flex:3;min-width:260px}
.section-title{font-size:15px;font-weight:700;color:#1c1917;margin-bottom:16px}
.tl-section,.todo-wrap{transition:box-shadow .3s,border-color .3s;border-radius:10px}
.tl-section.hl,.todo-wrap.hl{box-shadow:0 0 0 3px #d4a85340}

/* 时间线 - 脊柱 */
.tl{position:relative}
.tl-spine{position:absolute;left:50%;top:0;bottom:0;width:2px;background:linear-gradient(to bottom,#d4a853,#e5ddd0 80%);transform:translateX(-50%)}

/* 每一行 = flex容器,卡片在左侧或右侧半区 */
.tl-row{display:flex;margin-bottom:20px}
.tl-row:last-child{margin-bottom:0}
.tl-left{justify-content:flex-start}   /* 卡片靠左 */
.tl-right{justify-content:flex-end}    /* 卡片靠右 */

/* 卡片本身：flex容器，包含dot+内容 */
.tl-card{display:flex;align-items:flex-start;gap:8px;width:48%}
/* 左侧卡片：正常排列（dot在右） */
.tl-left .tl-card{flex-direction:row-reverse;text-align:right}
.tl-left .tl-card .tl-dot{flex-shrink:0;margin-top:6px}
.tl-left .tl-card .tl-body{text-align:left}
/* 右侧卡片：正常排列（dot在左） */
.tl-right .tl-card{flex-direction:row;text-align:left}
.tl-right .tl-card .tl-dot{flex-shrink:0;margin-top:6px}
.tl-right .tl-card .tl-body{text-align:left}

/* 圆点 */
.tl-dot{width:12px;height:12px;border-radius:50%;border:2px solid;flex-shrink:0;margin-top:6px}
.ddone{background:#d4a853;border-color:#d4a853}
.dlaunch{background:#2e7d32;border-color:#2e7d32}
.dv1{background:#3b82f6;border-color:#3b82f6}
.dv2{background:#667eea;border-color:#667eea}

/* 卡片内容 */
.tl-body{background:#fff;border:1px solid #e5ddd0;border-radius:10px;padding:12px 16px;flex:1}
.tl-body:hover{border-color:#d4a853;box-shadow:0 2px 8px rgba(212,168,83,.1)}
.tl-h{display:flex;align-items:center;gap:8px;margin-bottom:6px;flex-wrap:wrap}
.tl-date{font-size:12px;color:#8a7e72;font-weight:600}
.tl-tag{font-size:11px;padding:1px 8px;border-radius:4px;font-weight:600}
.tdone{background:#fdf8f0;color:#b8860b;border:1px solid #d4a853}
.tlaunch{background:#e8f5e9;color:#2e7d32;border:1px solid #a5d6a7}
.tv1{background:#dbeafe;color:#2563eb;border:1px solid #93c5fd}
.tv2{background:#e0e7ff;color:#4f46e5;border:1px solid #a5b4fc}
.tl-cnt{margin-left:auto;font-size:11px;color:#8a7e72;white-space:nowrap}
.tl-title{font-size:14px;font-weight:700;color:#1c1917;margin-bottom:2px}
.tl-sub{font-size:12px;color:#8a7e72;margin-bottom:4px}
.tl-pbar{height:3px;background:#f0ebe0;border-radius:2px;margin-bottom:8px;overflow:hidden}
.tl-pfill{height:100%;background:#d4a853;border-radius:2px}
.tl-items{list-style:none;padding:0;margin:0}
.tl-item{display:flex;gap:6px;font-size:13px;color:#292524;padding:2px 0;line-height:1.5}
.tl-chk{flex-shrink:0;font-size:12px}

/* 待办 */
.todo-wrap{background:#fff;border:1px solid #e5ddd0;border-radius:10px;padding:18px;position:sticky;top:12px}
.todo-wrap h3{font-size:16px;font-weight:700;color:#1c1917;margin-bottom:16px;padding-bottom:12px;border-bottom:2px solid #f0ebe0}
.tg{margin-bottom:14px}
.tg:last-child{margin-bottom:0}
.tp{font-size:14px;font-weight:700;margin-bottom:6px;display:flex;align-items:center;gap:8px}
.tbadge{font-size:11px;background:#f0ebe0;color:#8a7e72;padding:0 6px;border-radius:8px}
.ti{display:flex;align-items:flex-start;gap:8px;padding:5px 0;border-bottom:1px solid #f5f0e8}
.ti:last-child{border-bottom:none}
.ti.done{opacity:.5}
.ti.done .tt{text-decoration:line-through}
.tb{flex:1}
.tt{font-size:13px;color:#292524}
.tn{font-size:12px;color:#8a7e72;margin-top:2px;background:#f8f5f0;padding:2px 6px;border-radius:3px;display:inline-block}
.di{margin-top:16px;padding-top:14px;border-top:2px solid #f0ebe0}
.dit{font-size:14px;font-weight:700;color:#1c1917;margin-bottom:8px}
.dr{display:flex;justify-content:space-between;font-size:12px;padding:3px 0;color:#5a4e42}
.dr span:first-child{color:#8a7e72}
</style>
