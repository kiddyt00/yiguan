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
          <h3 class="section-title">🕐 完整开发时间线（共 {{total}} 次提交）</h3>
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
          <div v-for="(g,gi) in todos" :key="gi" class="tg">
            <div class="tp" :style="{color:g.color}">{{ g.label }}<span class="tbadge">{{ g.items.filter(t=>!t.done).length }}</span></div>
            <div v-for="(t,ti) in g.items" :key="ti" class="ti" :class="t.done?'done':''">
              <span>{{ t.done?'✅':'⬜' }}</span>
              <div class="tb"><div class="tt">{{ t.text }}</div><div v-if="t.note" class="tn">{{ t.note }}</div></div>
            </div>
          </div>
          <div class="di">
            <div class="dit">🚀 部署信息</div>
            <div class="dr"><span>首次上线</span><span>2026-06-01</span></div>
            <div class="dr"><span>运行天数</span><span>{{ stats[3].count }}天</span></div>
            <div class="dr"><span>生产域名</span><span>zgjz.insightj.cn</span></div>
            <div class="dr"><span>服务器</span><span>腾讯云 Ubuntu 24.04</span></div>
            <div class="dr"><span>SSL到期</span><span>2026-09-14</span></div>
            <div class="dr"><span>代码提交</span><span>{{ total }}次</span></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref } from 'vue'
const d0=new Date('2026-06-01'),days=Math.floor((new Date()-d0)/864e5)
const phases=[
  {period:'2026-05-07',tag:'MVP',tagClass:'tv1',dotClass:'dv1',title:'MVP v1.0 — 从零搭建',subtitle:'24次 · Go+HTMX',commits:24,pct:100,tasks:[
    {done:true,text:'项目骨架：go.mod/main.go/config'},{done:true,text:'周易引擎：64卦+铜钱起爻+变卦'},{done:true,text:'千问API：prompt+解卦'},{done:true,text:'HTTP处理器+模板：首页/算卦/结果'},{done:true,text:'CSS动画+E2E测试5/5'},
  ]},
  {period:'2026-05-07~08',tag:'重构',tagClass:'tv2',dotClass:'dv2',title:'v2.0前后端分离',subtitle:'29次 · Go+Vue3 SPA+Admin',commits:29,pct:100,tasks:[
    {done:true,text:'SQLite存储层：用户/Quota/历史/管理'},{done:true,text:'JWT鉴权+注册/登录API'},{done:true,text:'Vue3前台：5页面+双主题'},{done:true,text:'Vue3管理后台+Admin API'},{done:true,text:'多模型LLM（千问/DeepSeek）'},{done:true,text:'Docker Compose+Nginx'},
  ]},
  {period:'2026-05-09',tag:'增强',tagClass:'tv2',dotClass:'dv2',title:'v2.0功能增强',subtitle:'42次 · 模型/广告/i18n',commits:42,pct:100,tasks:[
    {done:true,text:'LLM重试+超时配置'},{done:true,text:'优雅关闭+健康检查+日志'},{done:true,text:'模型管理后台（25家供应商）'},{done:true,text:'广告系统+管理'},{done:true,text:'i18n中英双语+翻译API'},{done:true,text:'限流+Docker优化'},
  ]},
  {period:'2026-05-10',tag:'打磨',tagClass:'tv2',dotClass:'dv2',title:'体验打磨&Bug修复',subtitle:'12次 · 测试/样式',commits:12,pct:100,tasks:[
    {done:true,text:'测试覆盖（auth/handler/engine）'},{done:true,text:'暗黑模式+样式打磨'},{done:true,text:'登录页重做+密码找回'},
  ]},
  {period:'2026-05-12',tag:'完善',tagClass:'tv2',dotClass:'dv2',title:'功能完善',subtitle:'38次 · 历史/分享/广告/短信',commits:38,pct:100,tasks:[
    {done:true,text:'历史记录搜索/分页/翻译'},{done:true,text:'截图生成+分享'},{done:true,text:'广告观看+每日上限'},{done:true,text:'短信验证码（开发模式）'},{done:true,text:'管理后台卦象/用户/广告/模型'},{done:true,text:'数据分析页面'},
  ]},
  {period:'2026-05-19~21',tag:'小程序',tagClass:'tv2',dotClass:'dv2',title:'微信小程序开发',subtitle:'46次 · UniApp+原生双版',commits:46,pct:100,tasks:[
    {done:true,text:'UniApp版（Vue3编译）'},{done:true,text:'原生miniapp-native'},{done:true,text:'微信小程序登录'},{done:true,text:'首页/结果/历史/个人中心'},{done:true,text:'小程序广告'},
  ]},
  {period:'2026-06-01',tag:'上线',tagClass:'tlaunch',dotClass:'dlaunch',title:'v2.1生产部署🚀',subtitle:'5次 · 腾讯云Docker',commits:5,pct:100,tasks:[
    {done:true,text:'腾讯云+域名zgjz.insightj.cn'},{done:true,text:'Docker生产编排+Nginx'},{done:true,text:'7用户·138条占卜'},
  ]},
  {period:'2026-07-16',tag:'登录',tagClass:'tdone',dotClass:'ddone',title:'微信扫码登录+安全加固',subtitle:'46次 · wxLogin.js',commits:46,pct:100,tasks:[
    {done:true,text:'wxLogin.js SDK替代自绘二维码'},{done:true,text:'WeChat OAuth配置'},{done:true,text:'CORS白名单+限流60/min'},{done:true,text:'Docker日志+前端请求封装'},
  ]},
  {period:'2026-07-17',tag:'支付',tagClass:'tdone',dotClass:'ddone',title:'支付系统',subtitle:'23次 · 微信支付+JSAPI',commits:23,pct:100,tasks:[
    {done:true,text:'Web扫码支付(Native)上线'},{done:true,text:'二维码渲染修复(Vue nextTick)'},{done:true,text:'JSAPI小程序支付后端'},{done:true,text:'截图离屏渲染修复'},{done:true,text:'支付宝接入规划'},
  ]},
  {period:'2026-07-20~21',tag:'完成',tagClass:'tdone',dotClass:'ddone',title:'小程序全流程打通',subtitle:'14次 · 支付+管理后台',commits:14,pct:100,tasks:[
    {done:true,text:'小程序API地址更新11处'},{done:true,text:'miniprogram-ci上传工具'},{done:true,text:'miniapp-native+miniapp支付'},{done:true,text:'JSAPI缺少openid修复'},{done:true,text:'WX_APPID/WX_SECRET修复'},{done:true,text:'管理后台进度页'},
  ]},
  {period:'2026-07-22',tag:'完善',tagClass:'tdone',dotClass:'ddone',title:'账号体系+头像+用户管理大修',subtitle:'多轮修复 · 去重/合并/头像',commits:20,pct:100,tasks:[
    {done:true,text:'支付宝RSA2密钥对生成'},{done:true,text:'进度页完整生命周期+鱼骨图'},{done:true,text:'用户管理新设计（紧凑列+弹性表）'},{done:true,text:'全局表头不换行CSS'},{done:true,text:'卦象表格改用el-table'},{done:true,text:'微信头像保存+展示（admin+小程序+Web导航栏）'},{done:true,text:'UnionID去重机制+账号合并'},{done:true,text:'清除重复微信用户+逗号分隔openid'},
  ]},
  {period:'2026-07-29',tag:'二期',tagClass:'tdone',dotClass:'ddone',title:'v2.2 二期定价 + 会员系统 + 支付宝',subtitle:'新定价上线',commits:12,pct:100,tasks:[
    {done:true,text:'memberships表+会员模型/SQLite实现'},{done:true,text:'商品重定义：single/monthly/quarterly/yearly'},{done:true,text:'支付回调适配会员卡创建（顺延叠加）'},{done:true,text:'测算权限改为会员优先再判quota'},{done:true,text:'GET /api/user/membership 会员状态API'},{done:true,text:'前端Recharge.vue新定价UI+grid重构'},{done:true,text:'支付宝扫码支付接入（alipay.trade.page.pay）'},{done:true,text:'miniapp/miniapp-native商品同步更新'},
  ]},
]
const total=phases.reduce((s,p)=>s+p.commits,0)
const todos=[
  {label:'🔴 急迫',color:'#c62828',items:[{done:false,text:'小程序提交审核+正式发布',note:'体验版已测试通过'}]},
  {label:'🟡 待开发',color:'#d4a853',items:[{done:true,text:'支付宝扫码支付接入',note:'✅ 已上线'},{done:false,text:'短信服务接入',note:'当前仅打印日志'}]},
  {label:'🟢 优化项',color:'#667eea',items:[{done:true,text:'订单管理页面'},{done:true,text:'会员管理页面（管理后台）'},{done:false,text:'数据库自动备份'},{done:false,text:'前端SSE流自动重连'}]},
]
const done=phases.flatMap(p=>p.tasks).filter(t=>t.done).length
const todo=todos[0].items.filter(t=>!t.done).length+todos[1].items.filter(t=>!t.done).length
const opt=todos[2].items.filter(t=>!t.done).length
const stats=[
  {label:'已完成',count:done,color:'#d4a853',scrollTo:'tl'},
  {label:'待办',count:todo,color:'#c62828',scrollTo:'todo'},
  {label:'优化',count:opt,color:'#667eea',scrollTo:'opt'},
  {label:'运行天数',count:days,color:'#2e7d32',scrollTo:'deploy'},
]
const tlSec=ref(null),tdSec=ref(null)
function scrollTo(t){const e=t==='tl'?tlSec.value:tdSec.value;if(!e)return;e.scrollIntoView({behavior:'smooth'});e.classList.add('hl');setTimeout(()=>e.classList.remove('hl'),1200)}
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
.tl-section,.todo-wrap{transition:box-shadow .3s;border-radius:10px}
.tl-section.hl,.todo-wrap.hl{box-shadow:0 0 0 3px #d4a85340}

.tl{position:relative}
.tl-spine{position:absolute;left:50%;top:0;bottom:0;width:2px;background:linear-gradient(to bottom,#d4a853,#e5ddd0 80%);transform:translateX(-50%)}
.tl-row{display:flex;margin-bottom:20px}
.tl-row:last-child{margin-bottom:0}
.tl-left{justify-content:flex-start}
.tl-right{justify-content:flex-end}

.tl-card{display:flex;align-items:flex-start;gap:8px;width:48%}
.tl-left .tl-card{flex-direction:row-reverse;text-align:left}
.tl-left .tl-card .tl-dot{flex-shrink:0;margin-top:6px}
.tl-right .tl-card{flex-direction:row;text-align:left}
.tl-right .tl-card .tl-dot{flex-shrink:0;margin-top:6px}

.tl-dot{width:12px;height:12px;border-radius:50%;border:2px solid;flex-shrink:0;margin-top:6px}
.ddone{background:#d4a853;border-color:#d4a853}
.dlaunch{background:#2e7d32;border-color:#2e7d32}
.dv1{background:#3b82f6;border-color:#3b82f6}
.dv2{background:#667eea;border-color:#667eea}

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
