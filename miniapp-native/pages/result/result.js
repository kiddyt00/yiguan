const marked = require('../../utils/marked.js')
const API = 'https://gjz.shadouyou.cloud/api'

function coinsFromLine(v) {
  const map = { 6: [2,2,2], 7: [2,2,3], 8: [2,3,3], 9: [3,3,3] }
  return map[v] || [0,0,0]
}
function lineType(v) {
  return { 6:'老阴',7:'少阳',8:'少阴',9:'老阳' }[v] || ''
}

Page({
  data: {
    phase: 'coins', coinLabel: '', hexagram: { primary: '', changing: '' },
    aiText: '', error: '', statusMsg: '', showMaster: false,
    dots: '', dotsTimer: null, sseBuffer: '',
    statusMap: { coins:'起卦中...', hexagram:'卦象已现', ai:'AI 解读中...', done:'解读完成', error:'出错了' }
  },
  get showHexagram() { return ['hexagram','ai','done'].includes(this.data.phase) },
  get showAI() { return ['ai','done'].includes(this.data.phase) },
  get aiHtml() {
    let html = marked.parse(this.data.aiText || '')
    if (this.data.phase === 'ai') html += '<span style="animation:blink 1s infinite">▊</span>'
    return html
  },
  onLoad() { this.startDots(); this.startStream() },
  onUnload() { if (this.data.dotsTimer) clearInterval(this.data.dotsTimer) },
  startDots() {
    const dots = ['','.','..','...']
    let i = 0
    const timer = setInterval(() => { this.setData({ dots: dots[i%4] }); i++ }, 500)
    this.data.dotsTimer = timer
  },
  startStream() {
    const question = getApp().globalData?.question || ''
    if (!question) { this.setData({ error:'问题为空', phase:'error' }); return }
    const token = wx.getStorageSync('token')
    const requestTask = wx.request({
      url: API + '/divine/stream',
      method: 'POST',
      header: { 'Content-Type':'application/json', 'Authorization':'Bearer '+token },
      data: { question },
      enableChunked: true,
      responseType: 'text',
      success: () => {},
      fail: (err) => this.setData({ error:'网络连接失败: '+(err.errMsg||''), phase:'error' })
    })
    this.data.sseBuffer = ''
    requestTask.onChunkReceived((res) => {
      let chunk = typeof res.data === 'string' ? res.data : ''
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
        if (d.phase === 'coins') this.setData({ phase:'coins', coinLabel: d.data.label+' — '+d.data.result })
        else if (d.phase === 'hexagram') this.setData({ phase:'hexagram', hexagram:{ primary:d.data.primary_gua, changing:d.data.changing_gua } })
      } else if (event === 'ai') {
        if (this.data.phase === 'hexagram') this.setData({ phase:'ai' })
        this.setData({ aiText: this.data.aiText + (d.chunk || '') })
      } else if (event === 'status') { this.setData({ statusMsg: d.msg || '' }) }
      else if (event === 'done') { this.setData({ phase:'done' }) }
      else if (event === 'error') { this.setData({ error:d.error, phase:'done' }) }
    } catch(e) {}
  },
  toggleMaster() { this.setData({ showMaster: !this.data.showMaster }) },
  goHome() { wx.reLaunch({ url: '/pages/index/index' }) },
  goHistory() { wx.navigateTo({ url: '/pages/history/history' }) },
})
