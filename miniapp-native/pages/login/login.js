Page({
  data: { phone: '', code: '', loading: false, wxLoading: false, sendCountdown: 0 },
  onPhone(e) { this.setData({ phone: e.detail.value }) },
  onCode(e) { this.setData({ code: e.detail.value }) },
  get sendDisabled() { return this.data.sendCountdown > 0 || this.data.phone.length !== 11 },
  get sendText() { return this.data.sendCountdown > 0 ? this.data.sendCountdown + 's' : '获取验证码' },
  wechatLogin() {
    this.setData({ wxLoading: true })
    wx.login({
      success: (loginRes) => {
        wx.request({
          url: 'https://gjz.shadouyou.cloud/api/auth/wechat-login',
          method: 'POST', data: { code: loginRes.code },
          header: { 'Content-Type': 'application/json' },
          success: (res) => {
            if (res.statusCode === 200 && res.data.token) {
              wx.setStorageSync('token', res.data.token)
              wx.showToast({ title: '登录成功' })
              setTimeout(() => wx.reLaunch({ url: '/pages/index/index' }), 800)
            } else { wx.showToast({ title: res.data?.error || '登录失败', icon: 'none' }) }
          },
          fail: () => wx.showToast({ title: '网络错误', icon: 'none' }),
          complete: () => this.setData({ wxLoading: false })
        })
      },
      fail: () => { wx.showToast({ title: '微信登录失败', icon: 'none' }); this.setData({ wxLoading: false }) }
    })
  },
  sendSMS() {
    if (this.data.phone.length !== 11) return
    wx.request({
      url: 'https://gjz.shadouyou.cloud/api/auth/sms-send',
      method: 'POST', data: { phone: this.data.phone },
      header: { 'Content-Type': 'application/json' },
      success: (res) => {
        if (res.statusCode === 200) {
          wx.showToast({ title: '验证码已发送' })
          this.setData({ sendCountdown: 60 })
          const timer = setInterval(() => {
            if (this.data.sendCountdown <= 0) { clearInterval(timer); return }
            this.setData({ sendCountdown: this.data.sendCountdown - 1 })
          }, 1000)
        } else { wx.showToast({ title: res.data?.error || '发送失败', icon: 'none' }) }
      },
      fail: () => wx.showToast({ title: '网络错误', icon: 'none' })
    })
  },
  smsLogin() {
    if (!this.data.phone || !this.data.code) { wx.showToast({ title:'请输入手机号和验证码', icon:'none' }); return }
    this.setData({ loading: true })
    wx.request({
      url: 'https://gjz.shadouyou.cloud/api/auth/sms-login',
      method: 'POST', data: { phone: this.data.phone, code: this.data.code },
      header: { 'Content-Type': 'application/json' },
      success: (res) => {
        if (res.statusCode === 200 && res.data.token) {
          wx.setStorageSync('token', res.data.token)
          wx.showToast({ title: '登录成功' })
          setTimeout(() => wx.reLaunch({ url: '/pages/index/index' }), 800)
        } else { wx.showToast({ title: res.data?.error || '登录失败', icon: 'none' }) }
      },
      fail: () => wx.showToast({ title: '网络错误', icon: 'none' }),
      complete: () => this.setData({ loading: false })
    })
  },
  goPwd() { wx.navigateTo({ url: '/pages/login/pwd' }) },
})
