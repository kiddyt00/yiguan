const api = require('../../utils/api.js')
const API = 'https://zgjz.insightj.cn/api'

const products = [
  { id: 'single',    name: '单次测算', quota: 1, amount: 990, price: '9.90', icon: '🔮', duration: 0 },
  { id: 'monthly',   name: '月卡',     quota: -1, amount: 2990, price: '29.90', icon: '📅', duration: 30 },
  { id: 'quarterly', name: '季卡',     quota: -1, amount: 4990, price: '49.90', icon: '🌿', duration: 90 },
  { id: 'yearly',    name: '年卡',     quota: -1, amount: 9900, price: '99.00', icon: '👑', duration: 365 },
]

Page({
  data: { profile: {}, form: { nickname: '', address: '' }, binding: false, bindError: '', bindSuccess: false,
    products, selected: '', payLoading: false, membership: {}, inviteCode: '', inviteProgress: {} },
  onShow() { this.loadProfile() },
  onNick(e) { this.setData({ 'form.nickname': e.detail.value }) },
  onAddr(e) { this.setData({ 'form.address': e.detail.value }) },
  // 头像选择（微信头像昵称填写能力：open-type=chooseAvatar）
  onChooseAvatar(e) {
    const tempPath = e.detail.avatarUrl
    if (!tempPath) return
    wx.uploadFile({
      url: API + '/upload/avatar',
      filePath: tempPath,
      name: 'avatar',
      header: { 'Authorization': 'Bearer ' + wx.getStorageSync('token') },
      success: (res) => {
        if (res.statusCode === 200) {
          const data = JSON.parse(res.data)
          this.setData({ 'profile.avatar': data.avatar })
          wx.showToast({ title: '头像已更新' })
        } else {
          wx.showToast({ title: '头像上传失败', icon: 'none' })
        }
      },
      fail: () => wx.showToast({ title: '上传失败', icon: 'none' })
    })
  },
  // 设置分享内容
  onShareAppMessage() {
    const code = this.data.inviteCode
    return {
      title: '📤 来观己斋算一卦吧，测运势、问前程',
      path: '/pages/index/index?invite=' + (code || ''),
      imageUrl: '/images/logo_icon.svg'
    }
  },
  loadProfile() {
    const token = wx.getStorageSync('token')
    const authHeader = { 'Authorization': 'Bearer ' + token }
    Promise.all([
      api.profile(),
      new Promise(resolve => {
        wx.request({ url: API + '/user/membership', header: authHeader, success: r => resolve(r.data || {}), fail: () => resolve({}) })
      }),
      new Promise(resolve => {
        wx.request({ url: API + '/invite/code', header: authHeader, success: r => resolve(r.data || {}), fail: () => resolve({}) })
      }),
      new Promise(resolve => {
        wx.request({ url: API + '/invite/progress', header: authHeader, success: r => resolve(r.data || {}), fail: () => resolve({}) })
      }),
    ]).then(([d, m, c, p]) => {
      this.setData({ profile: d.user || d, 'form.nickname': d.nickname || '', 'form.address': d.address || '', membership: m, inviteCode: c.invite_code || '', inviteProgress: p || {} })
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
    this.setData({ payLoading: true })
    // 审核要求：虚拟商品支付必须接入「小程序虚拟支付」（全终端），统一走 wx.requestVirtualPayment
    this.doVirtualPay()
  },
  // 米大师虚拟支付（iOS）
  doVirtualPay() {
    // 先 wx.login 拿临时 code，后端用它换 session_key 生成用户态签名
    wx.login({
      success: (loginRes) => {
        if (!loginRes.code) {
          wx.showToast({ title: '登录失败，请重试', icon: 'none' })
          this.setData({ payLoading: false })
          return
        }
        this.virtualPayRequest(loginRes.code)
      },
      fail: () => {
        wx.showToast({ title: '登录失败，请重试', icon: 'none' })
        this.setData({ payLoading: false })
      }
    })
  },
  virtualPayRequest(loginCode) {
    api.virtualCreateOrder(this.data.selected, loginCode).then(data => {
      const v = data.virtual
      if (!wx.requestVirtualPayment) {
        wx.showToast({ title: '当前微信版本过低，请升级微信后重试', icon: 'none' })
        this.setData({ payLoading: false })
        return
      }
      wx.requestVirtualPayment({
        signData: v.signData,
        paySig: v.paySig,
        signature: v.signature,
        mode: v.mode,
        success: () => {
          wx.showToast({ title: '支付成功' })
          this.setData({ selected: '' })
          this.loadProfile()
        },
        fail: (err) => {
          if ((err.errMsg || '').indexOf('cancel') === -1) {
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
  // JSAPI 支付（Android / 其他平台）
  doJsapiPay() {
    // 后端用 wx.login code 换当前用户 openid（避免历史逗号分隔脏数据）
    wx.login({
      success: (loginRes) => {
        if (!loginRes.code) {
          wx.showToast({ title: '登录失败，请重试', icon: 'none' })
          this.setData({ payLoading: false })
          return
        }
        this.jsapiPayRequest(loginRes.code)
      },
      fail: () => {
        wx.showToast({ title: '登录失败，请重试', icon: 'none' })
        this.setData({ payLoading: false })
      }
    })
  },
  jsapiPayRequest(loginCode) {
    api.jsapiCreateOrder(this.data.selected, loginCode).then(data => {
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
  shareInvite() {
    // 触发右上角原生分享菜单
    wx.showShareMenu({
      withShareTicket: true,
      menus: ['shareAppMessage', 'shareTimeline']
    })
  },
  goOrders() { wx.navigateTo({ url: "/pages/orders/orders" }) },
  logout() {
    wx.showModal({ title: '确认退出', content: '确定要退出登录吗？', success: (r) => { if (r.confirm) { wx.removeStorageSync('token'); wx.reLaunch({ url: '/pages/index/index' })} } })
  }
})
