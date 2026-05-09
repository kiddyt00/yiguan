<template>
  <view class="container" style="padding-top: 120rpx;">
    <view class="text-center mb-3">
      <text style="font-size: 80rpx;">☯</text>
      <view style="font-size: 36rpx; font-weight: 700; margin-top: 16rpx;">{{ isRegister ? '注册' : '登录' }}</view>
      <text class="text-muted" style="display: block; margin-top: 8rpx;">新用户注册即赠 3 次免费起卦</text>
    </view>

    <view class="card">
      <input v-model="phone" type="number" placeholder="手机号" maxlength="11"
        style="border-bottom:1rpx solid #E0D6C8; padding:20rpx 0; font-size:30rpx; margin-bottom:24rpx;" />
      <input v-model="password" type="password" placeholder="密码（至少6位）"
        style="border-bottom:1rpx solid #E0D6C8; padding:20rpx 0; font-size:30rpx; margin-bottom:32rpx;" />
      <button class="btn-primary" :loading="loading" @tap="submit" style="width:100%;">
        {{ isRegister ? '注 册' : '登 录' }}
      </button>
    </view>

    <view class="text-center mt-3">
      <text class="text-muted" @tap="isRegister = !isRegister">
        {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
      </text>
    </view>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'

export default {
  data() {
    return { phone: '', password: '', loading: false, isRegister: false }
  },
  methods: {
    async submit() {
      if (!this.phone || !this.password) {
        uni.showToast({ title: '请填写完整', icon: 'none' }); return
      }
      if (this.password.length < 6) {
        uni.showToast({ title: '密码至少6位', icon: 'none' }); return
      }
      this.loading = true
      try {
        const fn = this.isRegister ? api.register : api.login
        const res = await fn(this.phone, this.password)
        uni.setStorageSync('token', res.token)
        uni.showToast({ title: this.isRegister ? '注册成功' : '登录成功' })
        setTimeout(() => uni.reLaunch({ url: '/pages/index/index' }), 800)
      } catch (e) {
        uni.showToast({ title: e.message, icon: 'none' })
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
