const api = require('../../utils/api.js')
const API = 'https://gjz.shadouyou.cloud/api'

Page({
  data: {
    loginTabs: ['短信登录','密码登录'],
    activeTab: '短信登录',
    // 短信
    phone:'', code:'', loading:false, wxLoading:false, sendCountdown:0,
    // 密码
    pwdPhone:'', password:'', nickname:'', pwdLoading:false, pwdError:'', isRegister:false
  },
  get sendDisabled(){return this.data.sendCountdown>0||this.data.phone.length!==11},
  get sendText(){return this.data.sendCountdown>0?this.data.sendCountdown+'s':'获取验证码'},

  switchTab(e){this.setData({activeTab:e.currentTarget.dataset.tab})},

  onPhone(e){this.setData({phone:e.detail.value})},
  onCode(e){this.setData({code:e.detail.value})},
  onPwdPhone(e){this.setData({pwdPhone:e.detail.value})},
  onPassword(e){this.setData({password:e.detail.value})},
  onNick(e){this.setData({nickname:e.detail.value})},

  wechatLogin(){
    this.setData({wxLoading:true})
    wx.login({
      success:(lr)=>{
        wx.request({
          url:API+'/auth/wechat-login', method:'POST', data:{code:lr.code},
          header:{'Content-Type':'application/json'},
          success:(res)=>{
            if(res.statusCode===200&&res.data.token){
              wx.setStorageSync('token',res.data.token)
              wx.showToast({title:'登录成功'})
              setTimeout(()=>wx.reLaunch({url:'/pages/index/index'}),800)
            }else{wx.showToast({title:res.data?.error||'登录失败',icon:'none'})}
          },
          fail:()=>wx.showToast({title:'网络错误',icon:'none'}),
          complete:()=>this.setData({wxLoading:false})
        })
      },
      fail:()=>{wx.showToast({title:'微信登录失败',icon:'none'});this.setData({wxLoading:false})}
    })
  },

  sendSMS(){
    if(this.data.phone.length!==11)return
    wx.request({
      url:API+'/auth/sms-send', method:'POST', data:{phone:this.data.phone},
      header:{'Content-Type':'application/json'},
      success:(res)=>{
        if(res.statusCode===200){
          wx.showToast({title:'验证码已发送'})
          this.setData({sendCountdown:60})
          const t=setInterval(()=>{
            if(this.data.sendCountdown<=0){clearInterval(t);return}
            this.setData({sendCountdown:this.data.sendCountdown-1})
          },1000)
        }else{wx.showToast({title:res.data?.error||'发送失败',icon:'none'})}
      },
      fail:()=>wx.showToast({title:'网络错误',icon:'none'})
    })
  },

  smsLogin(){
    if(!this.data.phone||!this.data.code){wx.showToast({title:'请输入手机号和验证码',icon:'none'});return}
    this.setData({loading:true})
    wx.request({
      url:API+'/auth/sms-login', method:'POST', data:{phone:this.data.phone,code:this.data.code},
      header:{'Content-Type':'application/json'},
      success:(res)=>{
        if(res.statusCode===200&&res.data.token){
          wx.setStorageSync('token',res.data.token)
          wx.showToast({title:'登录成功'})
          setTimeout(()=>wx.reLaunch({url:'/pages/index/index'}),800)
        }else{wx.showToast({title:res.data?.error||'登录失败',icon:'none'})}
      },
      fail:()=>wx.showToast({title:'网络错误',icon:'none'}),
      complete:()=>this.setData({loading:false})
    })
  },

  toggleMode(){this.setData({isRegister:!this.data.isRegister,pwdError:''})},

  pwdSubmit(){
    const{pwdPhone,password,nickname,isRegister}=this.data
    if(pwdPhone.length!==11){this.setData({pwdError:'请输入正确的手机号'});return}
    if(password.length<6){this.setData({pwdError:'密码至少6位'});return}
    this.setData({pwdLoading:true,pwdError:''})
    const p=isRegister?api.register(pwdPhone,password):api.login(pwdPhone,password)
    p.then(data=>{
      wx.setStorageSync('token',data.token)
      wx.showToast({title:isRegister?'注册成功':'登录成功'})
      setTimeout(()=>wx.reLaunch({url:'/pages/index/index'}),800)
    }).catch(e=>this.setData({pwdError:e.message}))
    .finally(()=>this.setData({pwdLoading:false}))
  }
})
