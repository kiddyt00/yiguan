<template>
  <view class="container">
    <view class="card text-center" style="padding-top: 48rpx;">
      <image src="/static/default-avatar.svg" mode="aspectFill"
        style="width:120rpx; height:120rpx; border-radius:60rpx; margin:0 auto;" />
      <view style="font-size: 36rpx; font-weight: 700; margin-top: 16rpx;">{{ profile.nickname || '易友' }}</view>
      <text class="text-muted">{{ profile.phone || '' }}</text>
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

export default {
  data() {
    return { profile: {}, form: { nickname: '', address: '' } }
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
    logout() {
      uni.removeStorageSync('token')
      uni.reLaunch({ url: '/pages/index/index' })
    }
  }
}
</script>
