const api = require('../../utils/api.js')
const API = 'https://zgjz.insightj.cn/api'

const products = [
  { id: 'test', name: '测试包', quota: 1, amount: 1, price: '0.01', icon: '🧪' },
  { id: 'trial', name: '尝鲜包', quota: 10, amount: 500, price: '5.00', icon: '🔮' },
  { id: 'standard', name: '标准包', quota: 50, amount: 2000, price: '20.00', icon: '🌟' },
  { id: 'unlimited', name: '畅享包', quota: 200, amount: 6000, price: '60.00', icon: '👑' },
]

Page({
  data: { profile: {}, form: { nickname: '', address: '' }, binding: false, bindError: '', bindSuccess: false,
    products, selected: '', payLoading: false },
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
  // 支付
  selectProduct(e) {
    this.setData({ selected: e.currentTarget.dataset.id })
  },
  doPay() {
    if (!this.data.selected) return
    const openid = this.data.profile.openid
    if (!openid) { wx.showToast({ title: '请先绑定微信', icon: 'none' }); return }
    this.setData({ payLoading: true })
    api.jsapiCreateOrder(this.data.selected, openid).then(data => {
      const pay = data.payment
      wx.requestPayment({
        timeStamp: pay.timeStamp,
        nonceStr: pay.nonceStr,
        package: pay.package,
        signType: pay.signType,
        paySign: pay.paySign,
        success: () => {
          wx.showToast({ title: '支付成功' })
          this.setData({ selected: '' })
          this.loadProfile()
        },
        fail: (err) => {
          if (err.errMsg.indexOf('cancel') === -1) {
            wx.showToast({ title: '支付失败: ' + (err.errMsg || '未知错误'), icon: 'none' })
          }
        },
        complete: () => { this.setData({ payLoading: false }) }
      })
    }).catch(e => {
      wx.showToast({ title: e.message || '下单失败', icon: 'none' })
      this.setData({ payLoading: false })
    })
  },
  logout() {
    wx.removeStorageSync('token'); wx.reLaunch({ url: '/pages/index/index' })
  }
})
