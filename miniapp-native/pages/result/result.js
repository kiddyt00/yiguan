const api = require('../../utils/api.js')
const marked = require('../../utils/marked.js')
const API = 'https://gjz.shadouyou.cloud/api'

Page({
  data: {
    mode: 'new',
    phase: 'coins',
    tossResults: [],
    yaoNames: ['初爻','二爻','三爻','四爻','五爻','上爻'],
    hexagram: { primary: '', changing: '' },
    aiText: '', error: '', statusMsg: '', showMaster: false,
    dots: '', dotsTimer: null, sseBuffer: '',
    loadingFromDB: false,
    remainingQuota: -1,
    statusMap: { coins:'起卦中...', hexagram:'卦象已现', ai:'AI 解读中...', done:'解读完成', error:'出错了' }
  },
  get showHexagram() { return ['hexagram','ai','done'].includes(this.data.phase) },
  get showAI() { return ['ai','done'].includes(this.data.phase) },
  get aiHtml() {
    let html = marked.parse(this.data.aiText || '')
    if (this.data.phase === 'ai') html += '<span style="animation:blink 1s infinite">▊</span>'
    return html
  },
  onLoad() {
    // 检查是否有历史记录ID（从history页跳转回看）
    const histId = this.data.historyId
    if (histId) { this.loadHistory(histId); return }
    this.startDots()
    this.startStream()
  },
  onUnload() { if (this.data.dotsTimer) clearInterval(this.data.dotsTimer) },
  onShareAppMessage() {
    return {
      title: '我在观己斋占了一卦「' + this.data.hexagram.primary + '」→「' + this.data.hexagram.changing + '」',
      path: '/pages/index/index'
    }
  },
  startDots() {
    const dots = ['','.','..','...']; let i = 0
    const timer = setInterval(() => { this.setData({ dots: dots[i%4] }); i++ }, 500)
    this.data.dotsTimer = timer
  },
  // 从历史记录加载已有结果
  loadHistory(id) {
    this.setData({ loadingFromDB: true, mode: 'db' })
    wx.request({
      url: API + '/admin/hexagrams/' + id,
      header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
      success: (res) => {
        if (res.statusCode === 200) {
          const d = res.data
          this.setData({
            phase: 'done', loadingFromDB: false,
            hexagram: { primary: d.primary_gua, changing: d.changing_gua },
            aiText: d.interpretation || ''
          })
          getApp().globalData.question = d.question
        }
      },
      fail: () => this.setData({ error:'加载失败', phase:'error', loadingFromDB: false })
    })
  },
  startStream() {
    const question = getApp().globalData?.question || ''
    if (!question) { this.setData({ error:'问题为空', phase:'error' }); return }
    const token = wx.getStorageSync('token')
    const requestTask = wx.request({
      url: API + '/divine/stream', method: 'POST',
      header: { 'Content-Type':'application/json', 'Authorization':'Bearer '+token },
      data: { question }, enableChunked: true, responseType: 'text',
      success: () => {},
      fail: (err) => this.setData({ error:'网络连接失败: '+(err.errMsg||''), phase:'error' })
    })
    this.data.sseBuffer = ''
    requestTask.onChunkReceived((res) => {
      let chunk = ''
      if (typeof res.data === 'string') {
        chunk = res.data
      } else if (res.data instanceof ArrayBuffer) {
        chunk = new TextDecoder('utf-8').decode(new Uint8Array(res.data))
      } else if (res.data instanceof Uint8Array) {
        chunk = new TextDecoder('utf-8').decode(res.data)
      }
      this.data.sseBuffer += chunk
      this.drainBuffer()
    })
  },
  drainBuffer() {
    let idx
    while ((idx = this.data.sseBuffer.indexOf('\n\n')) !== -1) {
      const block = this.data.sseBuffer.slice(0, idx)
      this.data.sseBuffer = this.data.sseBuffer.slice(idx + 2)
      this.parseBlock(block)
    }
  },
  parseBlock(block) {
    const lines = block.split('\n')
    let event = '', dataStr = ''
    for (const l of lines) {
      if (l.startsWith('event:')) event = l.slice(6).trim()
      else if (l.startsWith('data:')) dataStr = l.slice(5).trim()
    }
    if (!dataStr) return
    try {
      const d = JSON.parse(dataStr)
      if (event === 'phase') {
        if (d.phase === 'coins') {
          const toss = d.data
          const tossResults = [...(this.data.tossResults || [])]
          // coin_values 从后端是 ['反','反','正'] 格式，转为数字表示
          // 后端传来的是 ["反","反","正"]，用前端逻辑还原
          const coinNums = Array.isArray(toss.coin_values)
            ? toss.coin_values.map(c => c === '正' ? 3 : 2)
            : [0,0,0]
          tossResults.push({
            throw: toss.throw, label: toss.label,
            result: toss.result, sum: toss.sum,
            coin_values: toss.coin_values || [],
            is_changing: toss.result === '老阴' || toss.result === '老阳',
            yaoValue: toss.result === '少阳' || toss.result === '老阳'
          })
          this.setData({ phase: 'coins', tossResults })
        } else if (d.phase === 'hexagram') {
          this.setData({ phase: 'hexagram', hexagram: { primary: d.data.primary_gua, changing: d.data.changing_gua } })
        }
      } else if (event === 'ai') {
        if (this.data.phase === 'hexagram') this.setData({ phase: 'ai' })
        this.setData({ aiText: this.data.aiText + (d.chunk || '') })
      } else if (event === 'status') { this.setData({ statusMsg: d.msg || '' }) }
      else if (event === 'done') {
        // 完成后获取剩余配额
        api.profile().then(p => this.setData({ remainingQuota: p.remaining_quota ?? -1 })).catch(() => {})
        this.setData({ phase: 'done' })
      }
      else if (event === 'error') { this.setData({ error: d.error, phase: 'done' }) }
    } catch(e) {}
  },
  toggleMaster() { this.setData({ showMaster: !this.data.showMaster }) },
  shareResult() { this.onShareAppMessage(); /* 微信触发原生分享菜单 */ },
  goHome() { wx.reLaunch({ url: '/pages/index/index' }) },
  goHistory() { wx.navigateTo({ url: '/pages/history/history' }) },
})
