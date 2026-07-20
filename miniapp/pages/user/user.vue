<template>
  <view class="container">
    <view class="card text-center" style="padding-top: 48rpx;">
      <image src="/static/default-avatar.svg" mode="aspectFill"
        style="width:120rpx; height:120rpx; border-radius:60rpx; margin:0 auto;" />
      <view style="font-size: 36rpx; font-weight: 700; margin-top: 16rpx;">{{ profile.nickname || '易友' }}</view>
      <text class="text-muted">{{ profile.phone || '' }}</text>
    </view>

    <!-- 充值 -->
    <view class="card">
      <view style="font-size:28rpx;font-weight:600;margin-bottom:8rpx;">💎 充值</view>
      <text style="font-size:26rpx;color:#9E8C7A;display:block;margin-bottom:16rpx;">当前剩余：{{profile.remaining_quota || 0}} 次</text>
      <view style="display:flex;gap:12rpx;flex-wrap:wrap;margin-bottom:20rpx;">
        <view v-for="p in products" :key="p.id" @tap="selectProduct(p.id)"
          :style="'flex:1;min-width:120rpx;padding:16rpx 8rpx;border-radius:12rpx;text-align:center;border:2rpx solid '+(selected===p.id?'#D4A853':'#E0D6C8')+';background:'+(selected===p.id?'#FFF8F0':'#FFF')">
          <view style="font-size:24rpx;">{{p.icon}}</view>
          <view style="font-size:24rpx;font-weight:600;margin-top:4rpx;">{{p.name}}</view>
          <view style="font-size:28rpx;font-weight:700;color:#D4A853;">¥{{p.price}}</view>
          <view style="font-size:20rpx;color:#9E8C7A;">{{p.quota}} 次</view>
        </view>
      </view>
      <button class="btn-primary" style="width:100%;" :loading="payLoading" :disabled="!selected" @tap="doPay">
        {{payLoading ? '处理中...' : '微信支付'}}
      </button>
    </view>

    <view class="card">
      <view style="font-size: 28rpx; font-weight: 600; margin-bottom: 16rpx;">编辑资料</view>
      <input v-model="form.nickname" placeholder="昵称"
        style="border-bottom:1rpx solid #E0D6C8; padding:16rpx 0; font-size:28rpx; margin-bottom:16rpx;" />
      <input v-model="form.address" placeholder="地址（选填）"
        style="border-bottom:1rpx solid #E0D6C8; padding:16rpx 0; font-size:28rpx; margin-bottom:24rpx;" />
      <button class="btn-primary" style="width:100%;" @tap="save">保存</button>
    </view>

    <button class="btn-secondary mt-3" style="width:100%;" @tap="logout">退出登录</button>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'

const products = [
  { id: 'test', name: '测试包', quota: 1, amount: 1, price: '0.01', icon: '🧪' },
  { id: 'trial', name: '尝鲜包', quota: 10, amount: 500, price: '5.00', icon: '🔮' },
  { id: 'standard', name: '标准包', quota: 50, amount: 2000, price: '20.00', icon: '🌟' },
  { id: 'unlimited', name: '畅享包', quota: 200, amount: 6000, price: '60.00', icon: '👑' },
]

export default {
  data() {
    return { profile: {}, form: { nickname: '', address: '' }, products, selected: '', payLoading: false }
  },
  onShow() { this.loadProfile() },
  methods: {
    async loadProfile() {
      try {
        const data = await api.profile()
        this.profile = data
        this.form = { nickname: data.nickname || '', address: data.address || '' }
      } catch (e) { uni.showToast({ title: e.message, icon: 'none' }) }
    },
    async save() {
      try {
        await api.updateProfile(this.form)
        uni.showToast({ title: '保存成功' })
        this.loadProfile()
      } catch (e) { uni.showToast({ title: e.message, icon: 'none' }) }
    },
    selectProduct(id) { this.selected = id },
    async doPay() {
      if (!this.selected) return
      const openid = this.profile.openid
      if (!openid) { uni.showToast({ title: '请先绑定微信', icon: 'none' }); return }
      this.payLoading = true
      try {
        const data = await api.jsapiCreateOrder(this.selected, openid)
        const pay = data.payment
        uni.requestPayment({
          provider: 'wxpay',
          timeStamp: pay.timeStamp,
          nonceStr: pay.nonceStr,
          package: pay.package,
          signType: pay.signType,
          paySign: pay.paySign,
          success: () => {
            uni.showToast({ title: '支付成功' })
            this.selected = ''
            this.loadProfile()
          },
          fail: (err) => {
            if (err.errMsg && err.errMsg.indexOf('cancel') === -1) {
              uni.showToast({ title: '支付失败', icon: 'none' })
            }
          }
        })
      } catch (e) {
        uni.showToast({ title: e.message || '下单失败', icon: 'none' })
      } finally { this.payLoading = false }
    },
    logout() {
      uni.removeStorageSync('token')
      uni.reLaunch({ url: '/pages/index/index' })
    }
  }
}
</script>
