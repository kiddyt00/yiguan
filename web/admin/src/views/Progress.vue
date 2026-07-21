<template>
  <div>
    <div class="page-header"><h2>📋 开发进度</h2></div>

    <!-- 概览统计 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="s in stats" :key="s.label">
        <div class="stat-value" :style="{ color: s.color }">{{ s.count }}</div>
        <div class="stat-label">{{ s.label }}</div>
      </div>
    </div>

    <!-- 时间线 -->
    <div class="tl-wrap">
      <div v-for="(item, i) in timeline" :key="i" class="tl-item">
        <div class="tl-dot" :class="item.dotClass"></div>
        <div v-if="i < timeline.length - 1" class="tl-line"></div>
        <div class="tl-card">
          <div class="tl-header">
            <span class="tl-date">{{ item.date }}</span>
            <span class="tl-tag" :class="item.tagClass">{{ item.tag }}</span>
          </div>
          <div class="tl-title">{{ item.title }}</div>
          <ul class="tl-tasks">
            <li v-for="(t, ti) in item.tasks" :key="ti" class="tl-task" :class="t.done ? 'done' : 'pending'">
              <span class="tl-status">{{ t.done ? '✅' : '⏳' }}</span>
              <span>{{ t.text }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 待办 -->
    <div class="todo-wrap">
      <h3 class="todo-heading">📌 待办事项</h3>
      <div v-for="(g, gi) in todos" :key="gi" class="todo-group">
        <div class="todo-priority" :style="{ color: g.color }">{{ g.label }}</div>
        <div v-for="(t, ti) in g.items" :key="ti" class="todo-item" :class="t.done ? 'done' : ''">
          <span class="todo-check">{{ t.done ? '✅' : '⬜' }}</span>
          <div class="todo-body">
            <div class="todo-text">{{ t.text }}</div>
            <div v-if="t.note" class="todo-note">{{ t.note }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const stats = [
  { label: '已完成功能', count: 42, color: '#d4a853' },
  { label: '待办事项', count: 8, color: '#c62828' },
  { label: '优化项', count: 7, color: '#667eea' },
  { label: '部署天数', count: 50, color: '#2e7d32' },
]

const timeline = [
  {
    date: '2026-07-21', tag: '完成', tagClass: 'tag-done',
    title: '小程序全流程打通',
    tasks: [
      { done: true, text: '微信小程序登录 + 支付全流程测试通过' },
      { done: true, text: '生产环境 WX_APPID/WX_SECRET 配置修复' },
      { done: true, text: 'JSAPI 下单 XML 缺少 openid 修复' },
      { done: true, text: '体验版 v1.0.0 已上传，待审核发布' },
    ],
  },
  {
    date: '2026-07-20', tag: '完成', tagClass: 'tag-done',
    title: '小程序 API 地址更新 + 多项修复',
    tasks: [
      { done: true, text: '小程序 API 地址 11 处更新（gjz.shadouyou.cloud→zgjz.insightj.cn）' },
      { done: true, text: '安装 miniprogram-ci 命令行上传工具' },
      { done: true, text: '管理后台卦象弹窗修复（函数名+ref 变量名不匹配）' },
      { done: true, text: 'favicon.ico 404 → 301 重定向到 SVG' },
      { done: true, text: '小程序支付前端（miniapp-native + miniapp 双端）' },
    ],
  },
  {
    date: '2026-07-17', tag: '完成', tagClass: 'tag-done',
    title: '支付系统：微信支付修复 + JSAPI + 支付宝规划',
    tasks: [
      { done: true, text: '支付二维码渲染修复（Vue nextTick 时序坑）' },
      { done: true, text: 'JSAPI 小程序支付后端（统一下单重构）' },
      { done: true, text: '截图离屏渲染修复' },
      { done: true, text: '支付宝接入规划（待商户资料）' },
    ],
  },
  {
    date: '2026-07-16', tag: '完成', tagClass: 'tag-done',
    title: '微信扫码登录上线',
    tasks: [
      { done: true, text: 'wxLogin.js 官方 SDK 接入' },
      { done: true, text: '微信开放平台配置' },
      { done: true, text: '回调地址硬编码修复 + URL 编码' },
      { done: true, text: '安全加固：CORS 白名单 / API 限流 / 错误信息清理' },
    ],
  },
  {
    date: '2026-06-01', tag: '上线', tagClass: 'tag-launch',
    title: '易观 v2.1 正式部署',
    tasks: [
      { done: true, text: 'Docker Compose 生产部署' },
      { done: true, text: '腾讯云服务器配置' },
      { done: true, text: '域名 zgjz.insightj.cn + SSL 证书' },
    ],
  },
]

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
</script>

<style scoped>
.page-header { margin-bottom: 20px; }
.page-header h2 { font-size: 20px; font-weight: 700; color: #1c1917; }

/* 统计卡片 */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 24px; }
.stat-card { background: #fff; border: 1px solid #e5ddd0; border-radius: 10px; padding: 18px; text-align: center; }
.stat-value { font-size: 28px; font-weight: 800; line-height: 1; }
.stat-label { font-size: 13px; color: #8a7e72; margin-top: 6px; }

/* 时间线 */
.tl-wrap { position: relative; padding-left: 24px; margin-bottom: 28px; }
.tl-item { position: relative; padding-bottom: 20px; }
.tl-dot { position: absolute; left: -28px; top: 4px; width: 12px; height: 12px; border-radius: 50%; border: 2px solid; z-index: 1; }
.tl-dot.tag-done { background: #d4a853; border-color: #d4a853; }
.tl-dot.tag-launch { background: #2e7d32; border-color: #2e7d32; }
.tl-line { position: absolute; left: -23px; top: 16px; width: 2px; bottom: 0; background: #e5ddd0; }
.tl-card { background: #fff; border: 1px solid #e5ddd0; border-radius: 10px; padding: 16px 20px; }
.tl-header { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.tl-date { font-size: 12px; color: #8a7e72; font-weight: 600; }
.tl-tag { font-size: 11px; padding: 1px 8px; border-radius: 4px; font-weight: 600; }
.tag-done { background: #fdf8f0; color: #b8860b; border: 1px solid #d4a853; }
.tag-launch { background: #e8f5e9; color: #2e7d32; border: 1px solid #a5d6a7; }
.tl-title { font-size: 15px; font-weight: 700; color: #1c1917; margin-bottom: 8px; }
.tl-tasks { list-style: none; padding: 0; margin: 0; }
.tl-task { display: flex; align-items: flex-start; gap: 6px; font-size: 13px; color: #292524; padding: 3px 0; line-height: 1.5; }
.tl-task.pending { color: #8a7e72; }
.tl-status { flex-shrink: 0; font-size: 12px; }

/* 待办 */
.todo-wrap { background: #fff; border: 1px solid #e5ddd0; border-radius: 10px; padding: 20px; }
.todo-heading { font-size: 16px; font-weight: 700; color: #1c1917; margin-bottom: 16px; }
.todo-group { margin-bottom: 16px; }
.todo-group:last-child { margin-bottom: 0; }
.todo-priority { font-size: 14px; font-weight: 700; margin-bottom: 8px; }
.todo-item { display: flex; align-items: flex-start; gap: 8px; padding: 6px 0; border-bottom: 1px solid #f0ebe0; }
.todo-item:last-child { border-bottom: none; }
.todo-item.done { opacity: 0.5; }
.todo-check { flex-shrink: 0; font-size: 13px; }
.todo-body { flex: 1; }
.todo-text { font-size: 13px; color: #292524; }
.todo-item.done .todo-text { text-decoration: line-through; }
.todo-note { font-size: 12px; color: #8a7e72; margin-top: 2px; }
</style>
