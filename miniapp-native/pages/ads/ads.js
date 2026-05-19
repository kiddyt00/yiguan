const api = require('../../utils/api.js')
Page({
  data: { ads: [], watchingId: null, countdown: 0, quota: -1 },
  onShow() { this.loadAds(); this.loadQuota() },
  onUnload() { if (this.data.timer) clearInterval(this.data.timer) },
  loadQuota() {
    api.profile().then(d => this.setData({ quota: d.remaining_quota ?? -1 })).catch(() => {})
  },
  loadAds() {
    api.activeAds().then(d => this.setData({ ads: d.items || d || [] }))
      .catch(e => wx.showToast({ title: e.message, icon: 'none' }))
  },
  watchAd(e) {
    const id = e.currentTarget.dataset.id
    const ad = this.data.ads.find(a => a.id === id)
    if (!ad) return
    api.watchAd(id).then(() => {
      this.setData({ watchingId: id, countdown: ad.watch_duration })
      const timer = setInterval(() => {
        if (this.data.countdown <= 0) {
          clearInterval(timer)
          api.completeAd(id, ad.watch_duration).then(data => {
            wx.showToast({ title: '获得 ' + (data.rewarded || ad.reward_quota) + ' 次起卦！' })
            this.setData({ quota: data.remaining_quota ?? (this.data.quota + (data.rewarded || 1)) })
          }).catch(e => wx.showToast({ title: e.message, icon: 'none' }))
          this.setData({ watchingId: null })
          return
        }
        this.setData({ countdown: this.data.countdown - 1 })
      }, 1000)
      this.data.timer = timer
    }).catch(e => wx.showToast({ title: e.message, icon: 'none' }))
  }
})
