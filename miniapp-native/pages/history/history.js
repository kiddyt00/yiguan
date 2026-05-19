const api = require('../../utils/api.js')
const marked = require('../../utils/marked.js')
const API = 'https://gjz.shadouyou.cloud/api'

Page({
  data: {
    items: [], offset: 0, hasMore: true, loading: false, expandedId: null,
    translateCache: {}, translatingMap: {}
  },
  onShow() { this.setData({ items:[], offset:0, hasMore:true }); this.loadMore() },
  loadMore() {
    this.setData({ loading: true })
    api.history(20, this.data.offset).then(d => {
      const items = this.data.items.concat(d.items || [])
      this.setData({ items, offset: this.data.offset + 20, hasMore: d.total > items.length })
    }).catch(e => wx.showToast({ title:e.message, icon:'none' }))
    .finally(() => this.setData({ loading: false }))
  },
  toggleItem(e) {
    const id = e.currentTarget.dataset.id
    this.setData({ expandedId: this.data.expandedId === id ? null : id })
  },
  needsTranslation(h) { return h && h.lang === 'en' },
  getDisplayText(h) {
    const cached = this.data.translateCache[h.id]
    return cached || h.interpretation
  },
  doTranslate(e) {
    const id = e.currentTarget.dataset.id
    const lang = e.currentTarget.dataset.lang
    const target = lang === 'zh' ? 'en' : 'zh'
    if (this.data.translateCache[id]) return
    this.setData({ ['translatingMap.' + id]: true })
    wx.request({
      url: API + '/history/' + id + '/translate?target=' + target,
      method: 'POST',
      header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
      timeout: 60000,
      success: (res) => {
        if (res.statusCode === 200 && res.data.content) {
          this.setData({ ['translateCache.' + id]: res.data.content, ['translatingMap.' + id]: false })
        } else {
          wx.showToast({ title: res.data?.error || '翻译失败', icon: 'none' })
          this.setData({ ['translatingMap.' + id]: false })
        }
      },
      fail: () => { wx.showToast({ title: '网络错误', icon: 'none' }); this.setData({ ['translatingMap.' + id]: false }) }
    })
  },
  renderMD(text) { return marked.parse(text || '') }
})
