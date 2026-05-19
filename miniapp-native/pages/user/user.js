const api = require('../../utils/api.js')
const API = 'https://gjz.shadouyou.cloud/api'

Page({
  data: { profile: {}, form: { nickname: '', address: '' }, binding: false, bindError: '', bindSuccess: false },
  onShow() { this.loadProfile() },
  onNick(e) { this.setData({ 'form.nickname': e.detail.value }) },
  onAddr(e) { this.setData({ 'form.address': e.detail.value }) },
  loadProfile() {
    api.profile().then(d => {
      this.setData({ profile: d.user || d, 'form.nickname': d.nickname || '', 'form.address': d.address || '' })
    }).catch(e => wx.showToast({ title: e.message, icon: 'none' }))
  },
  save() {
    api.updateProfile(this.data.form).then(() => {
      wx.showToast({ title: '保存成功' }); this.loadProfile()
    }).catch(e => wx.showToast({ title: e.message, icon: 'none' }))
  },
  bindWechat() {
    this.setData({ binding: true, bindError: '', bindSuccess: false })
    wx.login({
      success: (loginRes) => {
        wx.request({
          url: API + '/user/bind-wechat',
          method: 'POST',
          header: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
          data: { code: loginRes.code },
          success: (res) => {
            if (res.statusCode === 200 && res.data.bound) {
              this.setData({ bindSuccess: true, binding: false, 'profile.openid': res.data.openid })
              wx.showToast({ title: '绑定成功' })
            } else {
              this.setData({ bindError: res.data?.error || '绑定失败', binding: false })
            }
          },
          fail: () => this.setData({ bindError: '网络错误', binding: false })
        })
      },
      fail: () => this.setData({ bindError: '微信登录失败', binding: false })
    })
  },
  logout() {
    wx.removeStorageSync('token'); wx.reLaunch({ url: '/pages/index/index' })
  }
})
