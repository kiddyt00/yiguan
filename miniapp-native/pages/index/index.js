const api = require('../../utils/api.js')
const API = 'https://gjz.shadouyou.cloud/api'

Page({
  data: { question:'', loading:false, showMaster:false, quota:-1, hasToken:false, nickname:'',
    // 登录表单
    loginTabs:['短信登录','密码登录'],activeTab:'短信登录',
    phone:'', code:'', wxLoading:false, smsLoading:false, sendCountdown:0,
    pwdPhone:'', password:'', regNickname:'', pwdLoading:false, pwdError:'', isRegister:false
  },
  get sendDisabled(){return this.data.sendCountdown>0||this.data.phone.length!==11},
  get sendText(){return this.data.sendCountdown>0?this.data.sendCountdown+'s':'获取验证码'},

  onShow(){
    const token=wx.getStorageSync('token')
    this.setData({hasToken:!!token})
    if(token)this.loadQuota()
  },
  onShareAppMessage(){return{title:'观己斋 - 观易知变，见心明境',path:'/pages/index/index'}},
  onInput(e){this.setData({question:e.detail.value})},

  loadQuota(){
    api.profile().then(d=>this.setData({quota:d.remaining_quota??-1,nickname:d.nickname||''})).catch(()=>{})
  },
  startDivine(){
    if(!this.data.question||!this.data.hasToken)return
    if(this.data.quota<=0){wx.showToast({title:'次数已用完',icon:'none'});return}
    this.setData({loading:true})
    getApp().globalData.question=this.data.question
    wx.navigateTo({url:'/pages/result/result',complete:()=>this.setData({loading:false})})
  },
  toggleMaster(){this.setData({showMaster:!this.data.showMaster})},

  // 登录
  switchTab(e){this.setData({activeTab:e.currentTarget.dataset.tab})},
  onPhone(e){this.setData({phone:e.detail.value})},
  onCode(e){this.setData({code:e.detail.value})},
  onPwdPhone(e){this.setData({pwdPhone:e.detail.value})},
  onPassword(e){this.setData({password:e.detail.value})},
  onNick(e){this.setData({nickname:e.detail.value})},

  wechatLogin(){
    this.setData({wxLoading:true})
    wx.login({success:(lr)=>{
      wx.request({url:API+'/auth/wechat-login',method:'POST',data:{code:lr.code},header:{'Content-Type':'application/json'},
        success:r=>{if(r.statusCode===200&&r.data.token){wx.setStorageSync('token',r.data.token);wx.showToast({title:'登录成功'});setTimeout(()=>this.onShow(),800)}else{wx.showToast({title:r.data?.error||'登录失败',icon:'none'})}},
        fail:()=>wx.showToast({title:'网络错误',icon:'none'}),complete:()=>this.setData({wxLoading:false})})},
      fail:()=>{wx.showToast({title:'微信登录失败',icon:'none'});this.setData({wxLoading:false})}
    })
  },

  sendSMS(){
    if(this.data.phone.length!==11)return
    wx.request({url:API+'/auth/sms-send',method:'POST',data:{phone:this.data.phone},header:{'Content-Type':'application/json'},
      success:r=>{if(r.statusCode===200){wx.showToast({title:'验证码已发送'});this.setData({sendCountdown:60});const t=setInterval(()=>{if(this.data.sendCountdown<=0){clearInterval(t);return}this.setData({sendCountdown:this.data.sendCountdown-1})},1000)}else{wx.showToast({title:r.data?.error||'发送失败',icon:'none'})}},
      fail:()=>wx.showToast({title:'网络错误',icon:'none'})})
  },

  smsLogin(){
    if(!this.data.phone||!this.data.code){wx.showToast({title:'请输入手机号和验证码',icon:'none'});return}
    this.setData({smsLoading:true})
    wx.request({url:API+'/auth/sms-login',method:'POST',data:{phone:this.data.phone,code:this.data.code},header:{'Content-Type':'application/json'},
      success:r=>{if(r.statusCode===200&&r.data.token){wx.setStorageSync('token',r.data.token);wx.showToast({title:'登录成功'});setTimeout(()=>this.onShow(),800)}else{wx.showToast({title:r.data?.error||'登录失败',icon:'none'})}},
      fail:()=>wx.showToast({title:'网络错误',icon:'none'}),complete:()=>this.setData({smsLoading:false})})
  },

  toggleMode(){this.setData({isRegister:!this.data.isRegister,pwdError:''})},

  pwdSubmit(){
    const pwdPhone=this.data.pwdPhone,password=this.data.password,isRegister=this.data.isRegister
    if(pwdPhone.length!==11){this.setData({pwdError:'请输入正确的手机号'});return}
    if(password.length<6){this.setData({pwdError:'密码至少6位'});return}
    this.setData({pwdLoading:true,pwdError:''})
    const p=isRegister?api.register(pwdPhone,password):api.login(pwdPhone,password)
    p.then(data=>{wx.setStorageSync('token',data.token);wx.showToast({title:isRegister?'注册成功':'登录成功'});setTimeout(()=>this.onShow(),800)})
    .catch(e=>this.setData({pwdError:e.message}))
    .finally(()=>this.setData({pwdLoading:false}))
  },

  goHistory(){wx.navigateTo({url:'/pages/history/history'})},
  goUser(){wx.navigateTo({url:'/pages/user/user'})},
  goAds(){wx.navigateTo({url:'/pages/ads/ads'})},
  goAbout(){wx.navigateTo({url:'/pages/about/about'})},
})
