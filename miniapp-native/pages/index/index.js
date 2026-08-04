const api = require('../../utils/api.js')

Page({
  data: { question:'', loading:false, showMaster:false, quota:-1, isMember:false, hasToken:false, nickname:'' },

  onLoad(options){
    if (options && options.invite) wx.setStorageSync('invite', options.invite)
    if (options && options.historyId) wx.redirectTo({ url: '/pages/result/result?historyId=' + options.historyId })
  },
  onShow(){
    const token = wx.getStorageSync('token')
    this.setData({hasToken: !!token})
    if(token) this.loadQuota()
  },
  onShareAppMessage(){return{title:'观己斋 - 观易知变，见心明境',path:'/pages/index/index'}},
  onShareTimeline(){return{title:'观己斋 - 观易知变，见心明境',query:''}},
  onInput(e){this.setData({question:e.detail.value})},

  loadQuota(){
    api.profile().then(d=>this.setData({quota:d.remaining_quota??-1,nickname:d.nickname||'',isMember:!!(d.membership&&d.membership.is_active)})).catch(()=>{})
  },
  startDivine(){
    if(this.data.loading)return
    if(!this.data.question||!this.data.hasToken)return
    if(!this.data.isMember&&this.data.quota<=0){wx.showToast({title:'次数已用完',icon:'none'});return}
    this.setData({loading:true})
    getApp().globalData.question=this.data.question
    wx.navigateTo({url:'/pages/result/result',complete:()=>this.setData({loading:false})})
  },
  toggleMaster(){this.setData({showMaster:!this.data.showMaster})},

  goLogin(){wx.navigateTo({url:'/pages/login/login'})},
  goHistory(){wx.navigateTo({url:'/pages/history/history'})},
  goUser(){wx.navigateTo({url:'/pages/user/user'})},
  goAds(){wx.navigateTo({url:'/pages/ads/ads'})},
  goAbout(){wx.navigateTo({url:'/pages/about/about'})},
})
