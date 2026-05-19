const api = require('../../utils/api.js')
const marked = require('../../utils/marked.js')
Page({
  data: { items: [], offset: 0, hasMore: true, loading: false, expandedId: null },
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
  renderMD(text) { return marked.parse(text || '') }
})
