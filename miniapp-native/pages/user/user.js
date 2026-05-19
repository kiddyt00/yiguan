const api = require('../../utils/api.js')
Page({
  data: { profile: {}, form: { nickname: '', address: '' } },
  onShow() { this.loadProfile() },
  onNick(e) { this.setData({ 'form.nickname': e.detail.value }) },
  onAddr(e) { this.setData({ 'form.address': e.detail.value }) },
  loadProfile() {
    api.profile().then(d => {
      this.setData({ profile: d, 'form.nickname': d.nickname || '', 'form.address': d.address || '' })
    }).catch(e => wx.showToast({ title: e.message, icon: 'none' }))
  },
  save() {
    api.updateProfile(this.data.form).then(() => {
      wx.showToast({ title: '保存成功' }); this.loadProfile()
    }).catch(e => wx.showToast({ title: e.message, icon: 'none' }))
  },
  logout() {
    wx.removeStorageSync('token'); wx.reLaunch({ url: '/pages/index/index' })
  }
})
