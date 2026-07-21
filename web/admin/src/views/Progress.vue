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
            <div v-for="(phase, i) in phases" :key="i" class="tl-row">
              <div class="tl-side">
                <div v-if="i%2===0" class="tl-card">
                  <div class="tl-dot" :class="phase.dotClass"></div>
                  <div class="tl-cbody">
                    <div class="tl-h">
                      <span class="tl-d">{{ phase.period }}</span>
                      <span class="tl-t" :class="phase.tagClass">{{ phase.tag }}</span>
                      <span class="tl-c">{{ phase.commits }}次</span>
                    </div>
                    <div class="tl-ti">{{ phase.title }}</div>
                    <div v-if="phase.subtitle" class="tl-su">{{ phase.subtitle }}</div>
                    <div class="tl-bar-bg"><div class="tl-bar" :style="'width:'+phase.pct+'%'"></div></div>
                    <ul class="tl-l">
                      <li v-for="t in phase.tasks" :key="t.text" class="tl-li"><span class="tl-ck">{{ t.done?'✅':'⬜' }}</span>{{ t.text }}</li>
                    </ul>
                  </div>
                </div>
              </div>
              <div class="tl-gap">
                <div v-if="i%2!==0" class="tl-dot" :class="phase.dotClass"></div>
              </div>
              <div class="tl-side">
                <div v-if="i%2!==0" class="tl-card">
                  <div class="tl-cbody">
                    <div class="tl-h">
                      <span class="tl-d">{{ phase.period }}</span>
                      <span class="tl-t" :class="phase.tagClass">{{ phase.tag }}</span>
                      <span class="tl-c">{{ phase.commits }}次</span>
                    </div>
                    <div class="tl-ti">{{ phase.title }}</div>
                    <div v-if="phase.subtitle" class="tl-su">{{ phase.subtitle }}</div>
                    <div class="tl-bar-bg"><div class="tl-bar" :style="'width:'+phase.pct+'%'"></div></div>
                    <ul class="tl-l">
                      <li v-for="t in phase.tasks" :key="t.text" class="tl-li"><span class="tl-ck">{{ t.done?'✅':'⬜' }}</span>{{ t.text }}</li>
                    </ul>
                  </div>
                  <div class="tl-dot" :class="phase.dotClass"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="col-todo">
        <div ref="todoSection" class="todo-wrap">
          <h3 class="todo-heading">📌 待办事项</h3>
          <div v-for="(g, gi) in todos" :key="gi" class="todo-group">
            <div class="todo-priority" :style="{ color: g.color }">{{ g.label }}<span class="todo-badge">{{ g.items.filter(t=>!t.done).length }}</span></div>
            <div v-for="(t, ti) in g.items" :key="ti" class="todo-item" :class="t.done?'done':''">
              <span class="todo-check">{{ t.done?'✅':'⬜' }}</span>
              <div class="todo-body"><div class="todo-text">{{ t.text }}</div><div v-if="t.note" class="todo-note">{{ t.note }}</div></div>
            </div>
          </div>
          <div class="deploy-info">
            <div class="deploy-title">🚀 部署信息</div>
            <div class="deploy-row"><span>首次上线</span><span>2026-06-01</span></div>
            <div class="deploy-row"><span>运行天数</span><span>{{ stats[3].count }}天</span></div>
            <div class="deploy-row"><span>生产域名</span><span>zgjz.insightj.cn</span></div>
            <div class="deploy-row"><span>服务器</span><span>腾讯云 Ubuntu 24.04</span></div>
            <div class="deploy-row"><span>SSL到期</span><span>2026-09-14</span></div>
            <div class="deploy-row"><span>代码提交</span><span>{{ totalCommits }}次</span></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
const startDate = new Date('2026-06-01')
const daysRunning = Math.floor((new Date() - startDate) / 864e5)
const phases = [
  {period:'2026-05-07',tag:'MVP',tagClass:'tag-v1',dotClass:'dot-v1',title:'MVP v1.0 — 从零搭建',subtitle:'24次提交 · Go + HTMX 快速原型',commits:24,pct:100,tasks:[
    {done:true,text:'项目骨架：go.mod / main.go / config.yaml'},{done:true,text:'周易引擎：64卦数据 + 铜钱起爻 + 变卦构建'},{done:true,text:'千问 API 客户端：prompt 构建 + 解卦调用'},{done:true,text:'HTTP 处理器 + HTML 模板：首页/算卦/结果'},{done:true,text:'CSS 动画 + E2E 集成测试 5/5 通过'},
  ]},
  {period:'2026-05-07~08',tag:'重构',tagClass:'tag-v2',dotClass:'dot-v2',title:'v2.0 前后端分离重构',subtitle:'29次提交 · Go 后端 + Vue 3 SPA',commits:29,pct:100,tasks:[
    {done:true,text:'SQLite 存储层：用户/Quota/历史/管理表'},{done:true,text:'JWT 鉴权 + 注册/登录 API'},{done:true,text:'Vue 3 前台 SPA：5页面 + 双主题'},{done:true,text:'Vue 3 管理后台 + Admin API'},{done:true,text:'多模型 LLM 客户端（千问/DeepSeek切换）'},{done:true,text:'Docker Compose 部署 + Nginx 反代'},
  ]},
  {period:'2026-05-09',tag:'增强',tagClass:'tag-v2',dotClass:'dot-v2',title:'v2.0 功能增强',subtitle:'42次提交 · 模型/广告/i18n/部署',commits:42,pct:100,tasks:[
    {done:true,text:'LLM 自动重试（3次指数退避）+ 超时配置'},{done:true,text:'优雅关闭 + 健康检查 + 请求日志'},{done:true,text:'模型管理后台（25家供应商）'},{done:true,text:'广告系统 + 管理后台'},{done:true,text:'前端 i18n 中英双语 + 翻译 API'},{done:true,text:'限流 + Docker 优化'},
  ]},
  {period:'2026-05-10',tag:'打磨',tagClass:'tag-v2',dotClass:'dot-v2',title:'体验打磨 & Bug 修复',subtitle:'12次提交 · 测试/样式/体验',commits:12,pct:100,tasks:[
    {done:true,text:'测试覆盖（auth/handler/engine 单元测试）'},{done:true,text:'暗黑模式适配 + 样式打磨'},{done:true,text:'登录页重做 + 注册/密码找回'},
  ]},
  {period:'2026-05-12',tag:'完善',tagClass:'tag-v2',dotClass:'dot-v2',title:'功能完善 — 历史/分享/广告/短信',subtitle:'38次提交 · 核心功能补齐',commits:38,pct:100,tasks:[
    {done:true,text:'历史记录搜索/分页/翻译'},{done:true,text:'截图生成 + 分享功能'},{done:true,text:'广告观看 + 每日上限'},{done:true,text:'短信验证码（开发模式）'},{done:true,text:'管理后台卦象/用户/广告/模型管理'},{done:true,text:'数据分析页面'},
  ]},
  {period:'2026-05-19~21',tag:'小程序',tagClass:'tag-v2',dotClass:'dot-v2',title:'微信小程序开发',subtitle:'46次提交 · UniApp + 原生双版本',commits:46,pct:100,tasks:[
    {done:true,text:'UniApp 版小程序（Vue 3 编译）'},{done:true,text:'原生微信小程序（miniapp-native）'},{done:true,text:'微信小程序登录'},{done:true,text:'小程序首页/结果/历史/个人中心'},{done:true,text:'小程序广告系统'},
  ]},
  {period:'2026-06-01',tag:'上线',tagClass:'tag-launch',dotClass:'dot-launch',title:'易观 v2.1 生产部署 🚀',subtitle:'5次提交 · 腾讯云 Docker 上线',commits:5,pct:100,tasks:[
    {done:true,text:'腾讯云服务器 + 域名 zgjz.insightj.cn'},{done:true,text:'Docker Compose 生产编排 + Nginx 反代'},{done:true,text:'7注册用户 · 138条占卜记录'},
  ]},
  {period:'2026-07-16',tag:'登录',tagClass:'tag-done',dotClass:'dot-done',title:'微信扫码登录 + 安全加固',subtitle:'46次提交 · wxLogin.js + 安全',commits:46,pct:100,tasks:[
    {done:true,text:'wxLogin.js 官方 SDK 替代自绘二维码'},{done:true,text:'WeChat OAuth 配置'},{done:true,text:'安全加固：CORS白名单 / 限流60/min'},{done:true,text:'Docker 日志轮转 + 前端统一请求封装'},
  ]},
  {period:'2026-07-17',tag:'支付',tagClass:'tag-done',dotClass:'dot-done',title:'支付系统：微信支付 + 小程序 JSAPI',subtitle:'23次提交 · 二维码/nextTick/截图',commits:23,pct:100,tasks:[
    {done:true,text:'Web 扫码支付（Native）上线'},{done:true,text:'支付二维码渲染修复（Vue nextTick）'},{done:true,text:'JSAPI 小程序支付后端（统一下单）'},{done:true,text:'截图离屏渲染修复'},{done:true,text:'支付宝接入规划'},
  ]},
  {period:'2026-07-20~21',tag:'完成',tagClass:'tag-done',dotClass:'dot-done',title:'小程序全流程打通',subtitle:'14次提交 · 支付上线 + 管理后台',commits:14,pct:100,tasks:[
    {done:true,text:'小程序 API 地址更新 11 处'},{done:true,text:'miniprogram-ci 上传工具'},{done:true,text:'miniapp-native + miniapp 双端支付'},{done:true,text:'JSAPI 缺少 openid 修复'},{done:true,text:'环境变量 WX_APPID/WX_SECRET 修复'},{done:true,text:'管理后台进度页面'},
  ]},
]
const totalCommits = phases.reduce((s,p)=>s+p.commits,0)
const todos = [
  {label:'🔴 急迫',color:'#c62828',items:[{done:false,text:'小程序提交审核 + 正式发布',note:'体验版已测试通过'}]},
  {label:'🟡 待开发',color:'#d4a853',items:[{done:false,text:'支付宝扫码支付接入',note:'等商户资料'},{done:false,text:'短信服务接入',note:'当前仅打印日志'}]},
  {label:'🟢 优化项',color:'#667eea',items:[{done:false,text:'前端 SSE 流自动重连'},{done:false,text:'零停机部署方案'},{done:false,text:'订单后台管理页面'},{done:false,text:'数据库自动备份'}]},
]
const doneCount = phases.flatMap(p=>p.tasks).filter(t=>t.done).length
const todoCount = todos[0].items.filter(t=>!t.done).length + todos[1].items.filter(t=>!t.done).length
const optCount = todos[2].items.filter(t=>!t.done).length
const stats = [
  {label:'已完成功能',count:doneCount,color:'#d4a853',scrollTo:'tl'},
  {label:'待办事项',count:todoCount,color:'#c62828',scrollTo:'todo'},
  {label:'优化项',count:optCount,color:'#667eea',scrollTo:'opt'},
  {label:'运行天数',count:daysRunning,color:'#2e7d32',scrollTo:'deploy'},
]
const tlSection = ref(null), todoSection = ref(null)
function scrollTo(t){const el=t==='tl'?tlSection.value:todoSection.value;if(!el)return;el.scrollIntoView({behavior:'smooth',block:'start'});el.classList.add('highlight');setTimeout(()=>el.classList.remove('highlight'),1200)}
</script>

<style scoped>
.page-header h2{font-size:20px;font-weight:700;color:#1c1917;margin-bottom:20px}
.stats-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:24px}
.stat-card{background:#fff;border:1px solid #e5ddd0;border-radius:10px;padding:18px;text-align:center;cursor:pointer;transition:all .2s;position:relative}
.stat-card:hover{transform:translateY(-2px);box-shadow:0 4px 16px rgba(0,0,0,.08);border-color:var(--hl)}
.stat-value{font-size:28px;font-weight:800;line-height:1}
.stat-label{font-size:13px;color:#8a7e72;margin-top:6px}
.stat-hint{font-size:11px;color:#ccc;margin-top:4px;opacity:0;transition:opacity .2s}
.stat-card:hover .stat-hint{opacity:1}
.two-col{display:flex;gap:20px;align-items:flex-start}
.col-tl{flex:5;min-width:0}
.col-todo{flex:3;min-width:260px}
.section-title{font-size:15px;font-weight:700;color:#1c1917;margin-bottom:16px}
.tl-section,.todo-wrap{transition:box-shadow .3s,border-color .3s;border-radius:10px}
.tl-section.highlight,.todo-wrap.highlight{box-shadow:0 0 0 3px #d4a85340}

/* ——— 时间线：一行三列布局 ——— */
.tl{position:relative}
.tl-spine{position:absolute;left:50%;top:0;bottom:0;width:2px;background:linear-gradient(to bottom,#d4a853,#e5ddd0 80%);transform:translateX(-50%)}
.tl-row{display:flex;margin-bottom:20px}
.tl-row:last-child{margin-bottom:0}
.tl-side{flex:1;min-width:0;display:flex}
.tl-gap{width:28px;flex-shrink:0;display:flex;justify-content:center;padding-top:10px;position:relative}
/* dot on gap (right cards) */
.tl-gap .tl-dot{position:relative;z-index:2}

/* 左侧卡片 - 贴在右边对齐中心 */
.tl-row:nth-child(odd) .tl-card{margin-right:14px}
.tl-row:nth-child(even) .tl-card{margin-left:14px}

/* 每张卡片布局：dot + 内容 或 内容 + dot */
.tl-card{display:flex;align-items:flex-start;gap:10px;width:100%}
/* 左侧卡片：dot在右边缘 */
.tl-row:nth-child(odd) .tl-card{flex-direction:row;justify-content:flex-end}
.tl-row:nth-child(odd) .tl-card .tl-dot{order:2;flex-shrink:0}
.tl-row:nth-child(odd) .tl-card .tl-cbody{order:1}
/* 右侧卡片：dot在左边缘 */
.tl-row:nth-child(even){flex-direction:row-reverse}
.tl-row:nth-child(even) .tl-card{flex-direction:row}
.tl-row:nth-child(even) .tl-card .tl-dot{order:0;flex-shrink:0}
.tl-row:nth-child(even) .tl-card .tl-cbody{order:1}

/* 圆点 */
.tl-dot{width:12px;height:12px;border-radius:50%;border:2px solid;margin-top:6px}
.dot-done{background:#d4a853;border-color:#d4a853}
.dot-launch{background:#2e7d32;border-color:#2e7d32}
.dot-v1{background:#3b82f6;border-color:#3b82f6}
.dot-v2{background:#667eea;border-color:#667eea}

/* 卡片内容 */
.tl-cbody{background:#fff;border:1px solid #e5ddd0;border-radius:10px;padding:12px 16px;min-width:280px;max-width:420px}
.tl-cbody:hover{border-color:#d4a853;box-shadow:0 2px 8px rgba(212,168,83,.1)}
.tl-h{display:flex;align-items:center;gap:8px;margin-bottom:6px;flex-wrap:wrap}
.tl-d{font-size:12px;color:#8a7e72;font-weight:600}
.tl-t{font-size:11px;padding:1px 8px;border-radius:4px;font-weight:600}
.tag-done{background:#fdf8f0;color:#b8860b;border:1px solid #d4a853}
.tag-launch{background:#e8f5e9;color:#2e7d32;border:1px solid #a5d6a7}
.tag-v1{background:#dbeafe;color:#2563eb;border:1px solid #93c5fd}
.tag-v2{background:#e0e7ff;color:#4f46e5;border:1px solid #a5b4fc}
.tl-c{margin-left:auto;font-size:11px;color:#8a7e72;white-space:nowrap}
.tl-ti{font-size:14px;font-weight:700;color:#1c1917;margin-bottom:2px}
.tl-su{font-size:12px;color:#8a7e72;margin-bottom:4px}
.tl-bar-bg{height:3px;background:#f0ebe0;border-radius:2px;margin-bottom:8px;overflow:hidden}
.tl-bar{height:100%;background:#d4a853;border-radius:2px}
.tl-l{list-style:none;padding:0;margin:0}
.tl-li{display:flex;gap:6px;font-size:13px;color:#292524;padding:2px 0;line-height:1.5}
.tl-ck{flex-shrink:0;font-size:12px}

/* ——— 待办 ——— */
.todo-wrap{background:#fff;border:1px solid #e5ddd0;border-radius:10px;padding:18px;position:sticky;top:12px}
.todo-heading{font-size:16px;font-weight:700;color:#1c1917;margin-bottom:16px;padding-bottom:12px;border-bottom:2px solid #f0ebe0}
.todo-group{margin-bottom:14px}
.todo-group:last-child{margin-bottom:0}
.todo-priority{font-size:14px;font-weight:700;margin-bottom:6px;display:flex;align-items:center;gap:8px}
.todo-badge{font-size:11px;background:#f0ebe0;color:#8a7e72;padding:0 6px;border-radius:8px}
.todo-item{display:flex;align-items:flex-start;gap:8px;padding:5px 0;border-bottom:1px solid #f5f0e8}
.todo-item:last-child{border-bottom:none}
.todo-item.done{opacity:.5}
.todo-check{flex-shrink:0;font-size:13px}
.todo-body{flex:1}
.todo-text{font-size:13px;color:#292524}
.todo-item.done .todo-text{text-decoration:line-through}
.todo-note{font-size:12px;color:#8a7e72;margin-top:2px;background:#f8f5f0;padding:2px 6px;border-radius:3px;display:inline-block}
.deploy-info{margin-top:16px;padding-top:14px;border-top:2px solid #f0ebe0}
.deploy-title{font-size:14px;font-weight:700;color:#1c1917;margin-bottom:8px}
.deploy-row{display:flex;justify-content:space-between;font-size:12px;padding:3px 0;color:#5a4e42}
.deploy-row span:first-child{color:#8a7e72}
</style>
