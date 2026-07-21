<template>
  <div>
    <div class="page-header"><h2>📋 开发进度</h2></div>

    <!-- 概览统计（可点击） -->
    <div class="stats-grid">
      <div class="stat-card clickable" v-for="s in stats" :key="s.label" @click="scrollTo(s.scrollTo)" :style="'--hl:'+s.color">
        <div class="stat-value" :style="{ color: s.color }">{{ s.count }}</div>
        <div class="stat-label">{{ s.label }}</div>
        <div class="stat-hint">点击查看 →</div>
      </div>
    </div>

    <!-- 双栏布局 -->
    <div class="two-col">
      <!-- 左栏：时间线（鱼骨式） -->
      <div class="col-tl">
        <div ref="tlSection" class="tl-section">
          <h3 class="section-title">🕐 开发时间线</h3>
          <div class="tl-spine">
            <div class="tl-inner">
              <div v-for="(item, i) in timeline" :key="i" class="tl-node">
                <div class="tl-connector">
                  <div class="tl-dot" :class="item.dotClass"></div>
                </div>
                <div class="tl-card" :class="item.cardClass">
                  <div class="tl-header">
                    <span class="tl-date">{{ item.date }}</span>
                    <span class="tl-tag" :class="item.tagClass">{{ item.tag }}</span>
                    <span class="tl-count">{{ item.doneCount }}/{{ item.tasks.length }} 项</span>
                  </div>
                  <div class="tl-title">{{ item.title }}</div>
                  <div class="tl-progress"><div class="tl-bar" :style="'width:'+item.pct+'%'"></div></div>
                  <ul class="tl-tasks">
                    <li v-for="(t, ti) in item.tasks" :key="ti" class="tl-task">
                      <span class="tl-check">{{ t.done ? '✅' : '⬜' }}</span>
                      <span :class="t.done ? '' : 'tl-muted'">{{ t.text }}</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右栏：待办事项 -->
      <div class="col-todo">
        <div ref="todoSection" class="todo-wrap">
          <h3 class="todo-heading">📌 待办事项</h3>
          <div v-for="(g, gi) in todos" :key="gi" class="todo-group" :ref="'todo-'+gi">
            <div class="todo-priority" :style="{ color: g.color }">
              {{ g.label }}
              <span class="todo-badge">{{ g.items.filter(t => !t.done).length }}</span>
            </div>
            <div v-for="(t, ti) in g.items" :key="ti" class="todo-item" :class="t.done ? 'done' : ''">
              <span class="todo-check">{{ t.done ? '✅' : '⬜' }}</span>
              <div class="todo-body">
                <div class="todo-text">{{ t.text }}</div>
                <div v-if="t.note" class="todo-note">{{ t.note }}</div>
              </div>
            </div>
          </div>
          <!-- 部署信息 -->
          <div class="deploy-info">
            <div class="deploy-title">🚀 部署信息</div>
            <div class="deploy-row"><span>首次上线</span><span>2026-06-01</span></div>
            <div class="deploy-row"><span>运行天数</span><span>{{ stats[3].count }} 天</span></div>
            <div class="deploy-row"><span>生产域名</span><span>zgjz.insightj.cn</span></div>
            <div class="deploy-row"><span>服务器</span><span>腾讯云 Ubuntu 24.04</span></div>
            <div class="deploy-row"><span>SSL 到期</span><span>2026-09-14</span></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const startDate = new Date('2026-06-01')
const today = new Date()
const daysRunning = Math.floor((today - startDate) / (1000 * 60 * 60 * 24))

const timeline = [
  {
    date: '2026-07-21', tag: '完成', tagClass: 'tag-done', dotClass: 'dot-done',
    title: '小程序全流程打通',
    tasks: [
      { done: true, text: '微信小程序登录 + 支付全流程测试通过' },
      { done: true, text: '生产环境 WX_APPID/WX_SECRET 配置修复' },
      { done: true, text: 'JSAPI 下单 XML 缺少 openid 修复' },
      { done: true, text: '体验版 v1.0.0 已上传，待审核发布' },
    ],
  },
  {
    date: '2026-07-20', tag: '完成', tagClass: 'tag-done', dotClass: 'dot-done',
    title: '小程序 API 地址更新 + 多项修复',
    tasks: [
      { done: true, text: '小程序 API 地址 11 处更新（gjz.shadouyou.cloud→zgjz.insightj.cn）' },
      { done: true, text: '安装 miniprogram-ci 命令行上传工具' },
      { done: true, text: '管理后台卦象弹窗修复' },
      { done: true, text: 'favicon.ico 404 → 301 重定向到 SVG' },
      { done: true, text: '小程序支付前端（miniapp-native + miniapp 双端）' },
    ],
  },
  {
    date: '2026-07-17', tag: '完成', tagClass: 'tag-done', dotClass: 'dot-done',
    title: '支付系统：微信支付修复 + JSAPI + 支付宝规划',
    tasks: [
      { done: true, text: '支付二维码渲染修复（Vue nextTick 时序坑）' },
      { done: true, text: 'JSAPI 小程序支付后端（统一下单重构）' },
      { done: true, text: '截图离屏渲染修复' },
      { done: false, text: '支付宝接入规划（待商户资料）' },
    ],
  },
  {
    date: '2026-07-16', tag: '完成', tagClass: 'tag-done', dotClass: 'dot-done',
    title: '微信扫码登录上线',
    tasks: [
      { done: true, text: 'wxLogin.js 官方 SDK 接入' },
      { done: true, text: '微信开放平台配置' },
      { done: true, text: '回调地址硬编码修复 + URL 编码' },
      { done: true, text: '安全加固：CORS 白名单 / API 限流 / 错误信息清理' },
    ],
  },
  {
    date: '2026-06-01', tag: '上线', tagClass: 'tag-launch', dotClass: 'dot-launch',
    title: '易观 v2.1 正式部署',
    tasks: [
      { done: true, text: 'Docker Compose 生产部署' },
      { done: true, text: '腾讯云服务器配置' },
      { done: true, text: '域名 zgjz.insightj.cn + SSL 证书' },
    ],
  },
].map(item => ({
  ...item,
  doneCount: item.tasks.filter(t => t.done).length,
  pct: Math.round(item.tasks.filter(t => t.done).length / item.tasks.length * 100),
}))

const todos = [
  {
    label: '🔴 急迫', color: '#c62828',
    items: [
      { done: false, text: '小程序提交审核 + 正式发布', note: '体验版已测试通过，提交审核即可发布' },
    ],
  },
  {
    label: '🟡 待开发', color: '#d4a853',
    items: [
      { done: false, text: '支付宝扫码支付接入', note: '等商户资料：AppID、应用私钥、支付宝公钥' },
      { done: false, text: '短信服务接入（阿里云/腾讯云）', note: '当前验证码仅打印日志' },
    ],
  },
  {
    label: '🟢 优化项', color: '#667eea',
    items: [
      { done: false, text: '前端 SSE 流自动重连', note: '部署时正在解卦的流不会中断' },
      { done: false, text: '零停机部署方案', note: 'Nginx 健康检查 + 滚动更新' },
      { done: false, text: '订单后台管理页面', note: '目前订单只有 API 接口' },
      { done: false, text: '数据库自动备份', note: '定时备份到对象存储' },
    ],
  },
]

const allTasks = timeline.flatMap(i => i.tasks)
const doneCount = allTasks.filter(t => t.done).length
const todoCount = todos[0].items.filter(t => !t.done).length + todos[1].items.filter(t => !t.done).length
const optCount = todos[2].items.filter(t => !t.done).length

const stats = [
  { label: '已完成功能', count: doneCount, color: '#d4a853', scrollTo: 'tl' },
  { label: '待办事项', count: todoCount, color: '#c62828', scrollTo: 'todo' },
  { label: '优化项', count: optCount, color: '#667eea', scrollTo: 'opt' },
  { label: '部署天数', count: daysRunning, color: '#2e7d32', scrollTo: 'deploy' },
]

// 滚动到目标区域
const tlSection = ref(null)
const todoSection = ref(null)

function scrollTo(target) {
  let el = null
  if (target === 'tl') el = tlSection.value
  else if (target === 'todo' || target === 'opt' || target === 'deploy') el = todoSection.value
  if (!el) return
  el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  // 高亮闪烁效果
  el.classList.add('highlight')
  setTimeout(() => el.classList.remove('highlight'), 1200)
}
</script>

<style scoped>
.page-header { margin-bottom: 20px; }
.page-header h2 { font-size: 20px; font-weight: 700; color: #1c1917; }

/* ——— 统计卡片（可点击） ——— */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 24px; }
.stat-card { background: #fff; border: 1px solid #e5ddd0; border-radius: 10px; padding: 18px; text-align: center; position: relative; }
.stat-card.clickable { cursor: pointer; transition: all .2s; }
.stat-card.clickable:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,.08); border-color: var(--hl); }
.stat-value { font-size: 28px; font-weight: 800; line-height: 1; }
.stat-label { font-size: 13px; color: #8a7e72; margin-top: 6px; }
.stat-hint { font-size: 11px; color: #ccc; margin-top: 4px; opacity: 0; transition: opacity .2s; }
.stat-card:hover .stat-hint { opacity: 1; }

/* ——— 双栏布局 ——— */
.two-col { display: flex; gap: 20px; align-items: flex-start; }
.col-tl { flex: 3; min-width: 0; }
.col-todo { flex: 2; min-width: 280px; }

/* ——— 区域标题 ——— */
.section-title { font-size: 15px; font-weight: 700; color: #1c1917; margin-bottom: 14px; }

/* ——— 高亮闪烁（滚动目标） ——— */
.tl-section, .todo-wrap { transition: box-shadow .3s, border-color .3s; border-radius: 10px; }
.tl-section.highlight { box-shadow: 0 0 0 3px #d4a85340; }
.todo-wrap.highlight { box-shadow: 0 0 0 3px #d4a85340; }

/* ——— 鱼骨时间线 ——— */
.tl-spine { position: relative; padding: 0; }
.tl-inner { position: relative; }
.tl-inner::before {
  content: ''; position: absolute; left: 17px; top: 6px; bottom: 6px;
  width: 2px; background: linear-gradient(to bottom, #d4a853 0%, #e5ddd0 60%, #e5ddd0 100%);
  border-radius: 1px;
}
.tl-node { position: relative; padding-left: 42px; padding-bottom: 18px; }
.tl-node:last-child { padding-bottom: 0; }
.tl-connector { position: absolute; left: 0; top: 2px; display: flex; align-items: center; width: 42px; height: 22px; }
.tl-connector::before {
  content: ''; flex: 1; height: 2px; background: #d4a853; margin-right: -1px;
}
.tl-dot {
  width: 12px; height: 12px; border-radius: 50%; border: 2px solid; flex-shrink: 0;
  position: relative; z-index: 1;
}
.tl-dot.dot-done { background: #d4a853; border-color: #d4a853; }
.tl-dot.dot-launch { background: #2e7d32; border-color: #2e7d32; }
.tl-card { background: #fff; border: 1px solid #e5ddd0; border-radius: 10px; padding: 12px 16px; }
.tl-card:hover { border-color: #d4a853; box-shadow: 0 2px 8px rgba(212,168,83,0.1); }
.tl-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.tl-date { font-size: 12px; color: #8a7e72; font-weight: 600; }
.tl-tag { font-size: 11px; padding: 1px 8px; border-radius: 4px; font-weight: 600; }
.tag-done { background: #fdf8f0; color: #b8860b; border: 1px solid #d4a853; }
.tag-launch { background: #e8f5e9; color: #2e7d32; border: 1px solid #a5d6a7; }
.tl-count { margin-left: auto; font-size: 11px; color: #8a7e72; }
.tl-title { font-size: 14px; font-weight: 700; color: #1c1917; margin-bottom: 4px; }
.tl-progress { height: 3px; background: #f0ebe0; border-radius: 2px; margin-bottom: 8px; overflow: hidden; }
.tl-bar { height: 100%; background: #d4a853; border-radius: 2px; transition: width .6s; }
.tl-tasks { list-style: none; padding: 0; margin: 0; }
.tl-task { display: flex; align-items: flex-start; gap: 6px; font-size: 13px; color: #292524; padding: 2px 0; line-height: 1.5; }
.tl-check { flex-shrink: 0; font-size: 12px; }
.tl-muted { color: #8a7e72; }

/* ——— 待办 ——— */
.todo-wrap { background: #fff; border: 1px solid #e5ddd0; border-radius: 10px; padding: 18px; position: sticky; top: 12px; }
.todo-heading { font-size: 16px; font-weight: 700; color: #1c1917; margin-bottom: 16px; padding-bottom: 12px; border-bottom: 2px solid #f0ebe0; }
.todo-group { margin-bottom: 14px; }
.todo-group:last-child { margin-bottom: 0; }
.todo-priority { font-size: 14px; font-weight: 700; margin-bottom: 6px; display: flex; align-items: center; gap: 8px; }
.todo-badge { font-size: 11px; background: #f0ebe0; color: #8a7e72; padding: 0 6px; border-radius: 8px; }
.todo-item { display: flex; align-items: flex-start; gap: 8px; padding: 5px 0; border-bottom: 1px solid #f5f0e8; }
.todo-item:last-child { border-bottom: none; }
.todo-item.done { opacity: 0.5; }
.todo-check { flex-shrink: 0; font-size: 13px; }
.todo-body { flex: 1; }
.todo-text { font-size: 13px; color: #292524; }
.todo-item.done .todo-text { text-decoration: line-through; }
.todo-note { font-size: 12px; color: #8a7e72; margin-top: 2px; background: #f8f5f0; padding: 2px 6px; border-radius: 3px; display: inline-block; }

/* ——— 部署信息 ——— */
.deploy-info { margin-top: 16px; padding-top: 14px; border-top: 2px solid #f0ebe0; }
.deploy-title { font-size: 14px; font-weight: 700; color: #1c1917; margin-bottom: 8px; }
.deploy-row { display: flex; justify-content: space-between; font-size: 12px; padding: 3px 0; color: #5a4e42; }
.deploy-row span:first-child { color: #8a7e72; }
</style>
