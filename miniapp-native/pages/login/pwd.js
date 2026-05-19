const api = require('../../utils/api.js')
Page({
  data: { phone: '', password: '', nickname: '', loading: false, error: '', isRegister: false },
  onPhone(e) { this.setData({ phone: e.detail.value }) },
  onPwd(e) { this.setData({ password: e.detail.value }) },
  onNick(e) { this.setData({ nickname: e.detail.value }) },
  toggleMode() { this.setData({ isRegister: !this.data.isRegister, error: '' }) },
  goBack() { wx.navigateBack() },
  submit() {
    const { phone, password, nickname, isRegister } = this.data
    if (phone.length !== 11) { this.setData({ error: '请输入正确的手机号' }); return }
    if (password.length < 6) { this.setData({ error: '密码至少6位' }); return }
    this.setData({ loading: true, error: '' })
    const p = isRegister
      ? api.register(phone, password)
      : api.login(phone, password)
    p.then(data => {
      wx.setStorageSync('token', data.token)
      wx.showToast({ title: isRegister ? '注册成功' : '登录成功' })
      setTimeout(() => wx.reLaunch({ url: '/pages/index/index' }), 800)
    }).catch(e => this.setData({ error: e.message }))
    .finally(() => this.setData({ loading: false }))
  }
})
