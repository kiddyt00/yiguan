<template>
  <view class="container">
    <!-- 状态标题 -->
    <view class="text-center mb-3" style="font-size: 36rpx; font-weight: 700;">
      {{ statusMap[phase] }}
    </view>

    <!-- 铜钱动画 -->
    <view v-if="phase === 'coins'" class="text-center card">
      <text style="font-size: 60rpx;">🪙</text>
      <view style="font-size: 28rpx; margin-top: 16rpx;">{{ coinLabel }}</view>
    </view>

    <!-- 卦象展示 -->
    <view v-if="showHexagram" class="flex gap-2">
      <view class="card text-center" style="flex:1;">
        <text class="text-muted">本卦</text>
        <view style="font-size: 40rpx; font-weight: 700; margin-top: 8rpx;">{{ hexagram.primary }}</view>
      </view>
      <view class="card text-center" style="flex:1;">
        <text class="text-muted">变卦</text>
        <view style="font-size: 40rpx; font-weight: 700; margin-top: 8rpx;">{{ hexagram.changing }}</view>
      </view>
    </view>

    <!-- 加载动画 -->
    <view v-if="phase === 'hexagram'" class="text-center card">
      <text style="font-size: 40rpx;">🤔</text>
      <view style="margin-top: 16rpx; font-size: 28rpx; color: #8B4513;">
        AI 正在思考中{{ dots }}
      </view>
      <text v-if="statusMsg" class="text-muted" style="display: block; margin-top: 8rpx;">{{ statusMsg }}</text>
    </view>

    <!-- AI 解卦流式 -->
    <view v-if="showAI" class="card">
      <view style="font-size: 30rpx; font-weight: 600; margin-bottom: 16rpx;">🤖 AI 解读</view>
      <rich-text :nodes="aiHtml" style="font-size: 28rpx; line-height: 1.8;" />
    </view>

    <!-- 错误 -->
    <view v-if="error" class="card text-center" style="color: #C62828;">{{ error }}</view>

    <!-- 完成 -->
    <view v-if="phase === 'done'" class="text-center mt-3">
      <button class="btn-primary" @tap="goHome">返回首页</button>
      <view class="card mt-3" @tap="showMaster = true">
        <text style="font-size: 28rpx; color: #8B4513;">🎓 周易大师一对一详解</text>
      </view>
    </view>

    <!-- 大师二维码 -->
    <view v-if="showMaster" class="card text-center mt-3">
      <text style="font-size: 30rpx; font-weight: 600;">周易大师 · 一对一深度交流</text>
      <image src="/static/master-qr.png" mode="widthFix" style="width: 400rpx; margin: 24rpx auto;" />
      <text class="text-muted">长按识别二维码添加大师微信</text>
      <button class="btn-secondary mt-3" @tap="showMaster = false">关闭</button>
    </view>
  </view>
</template>

<script>
import { marked } from '../../utils/marked.min.js'

export default {
  data() {
    return {
      phase: 'coins',
      coinLabel: '',
      hexagram: { primary: '', changing: '' },
      aiText: '',
      error: '',
      statusMsg: '',
      showMaster: false,
      dots: '',
      dotsTimer: null,
      statusMap: { coins: '起卦中...', hexagram: '卦象已现', ai: 'AI 解读中...', done: '解读完成', error: '出错了' }
    }
  },
  computed: {
    showHexagram() { return ['hexagram','ai','done'].includes(this.phase) },
    showAI() { return ['ai','done'].includes(this.phase) },
    aiHtml() {
      let html = marked.parse(this.aiText || '')
      if (this.phase === 'ai') html += '<span style="animation: blink 1s infinite">▊</span>'
      return html.replace(/\n/g, '<br/>')
    }
  },
  onLoad() {
    this.startDots()
    this.startStream()
  },
  onUnload() {
    if (this.dotsTimer) clearInterval(this.dotsTimer)
  },
  methods: {
    startDots() {
      const dots = ['', '.', '..', '...']
      let i = 0
      this.dotsTimer = setInterval(() => {
        this.dots = dots[i % dots.length]
        i++
      }, 500)
    },
    async startStream() {
      const question = getApp().globalData?.question || ''
      if (!question) { this.error = '问题为空'; return }

      const token = uni.getStorageSync('token')
      try {
        const task = uni.request({
          url: 'https://49.235.108.61/api/divine/stream',
          method: 'POST',
          header: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          },
          data: { question },
          enableChunked: true,
          responseType: 'text'
        })

        task.onChunkReceived(res => {
          const text = typeof res.data === 'string' ? res.data :
            new TextDecoder().decode(new Uint8Array(res.data))
          this.parseSSE(text)
        })
      } catch (e) {
        this.error = '网络连接失败'
        this.phase = 'error'
      }
    },
    parseSSE(text) {
      const lines = text.split('\n')
      let event = ''
      for (const line of lines) {
        if (line.startsWith('event:')) event = line.slice(6).trim()
        else if (line.startsWith('data:')) {
          try {
            const data = JSON.parse(line.slice(5).trim())
            if (event === 'phase') {
              if (data.phase === 'coins') {
                this.phase = 'coins'
                this.coinLabel = `${data.data.label} — ${data.data.result}`
              } else if (data.phase === 'hexagram') {
                this.phase = 'hexagram'
                this.hexagram = { primary: data.data.primary_gua, changing: data.data.changing_gua }
              }
            } else if (event === 'ai') {
              if (this.phase === 'hexagram') this.phase = 'ai'
              this.aiText += data.chunk
            } else if (event === 'status') {
              this.statusMsg = data.msg || ''
            } else if (event === 'done') {
              this.phase = 'done'
            } else if (event === 'error') {
              this.error = data.error
              this.phase = 'done'
            }
          } catch(e) {}
        }
      }
    },
    goHome() { uni.reLaunch({ url: '/pages/index/index' }) }
  }
}
</script>
