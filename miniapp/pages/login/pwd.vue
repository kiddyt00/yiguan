<template>
  <view class="container" style="padding-top: 120rpx;">
    <view class="text-center mb-3">
      <text style="font-size: 80rpx;">☯</text>
      <view style="font-size: 36rpx; font-weight: 700; margin-top: 16rpx;">观己斋</view>
    </view>

    <view class="card">
      <input v-model="phone" type="number" placeholder="手机号" maxlength="11"
        style="border-bottom:1rpx solid #E0D6C8; padding:16rpx 0; font-size:30rpx; margin-bottom:24rpx;" />
      <input v-model="password" type="password" placeholder="密码"
        style="border-bottom:1rpx solid #E0D6C8; padding:16rpx 0; font-size:30rpx; margin-bottom:32rpx;" />
      <view v-if="isRegister" class="mb-3">
        <input v-model="nickname" placeholder="昵称（选填）"
          style="border-bottom:1rpx solid #E0D6C8; padding:16rpx 0; font-size:30rpx; margin-bottom:32rpx;" />
      </view>
      <view v-if="error" class="text-muted mb-3" style="color:#C62828; text-align:center;">{{ error }}</view>
      <button class="btn-primary" style="width:100%;" :loading="loading" @tap="submit">
        {{ isRegister ? '注册' : '登录' }}
      </button>
      <view class="text-center mt-3">
        <text class="text-muted" @tap="toggleMode">
          {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
        </text>
      </view>
    </view>

    <view class="text-center mt-3">
      <text class="text-muted" @tap="goBack">返回其他登录方式</text>
    </view>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'

export default {
  data() {
    return { phone: '', password: '', nickname: '', loading: false, error: '', isRegister: false }
  },
  methods: {
    toggleMode() {
      this.isRegister = !this.isRegister
      this.error = ''
    },
    async submit() {
      if (!this.phone || this.phone.length !== 11) {
        this.error = '请输入正确的手机号'
        return
      }
      if (!this.password || this.password.length < 6) {
        this.error = '密码至少6位'
        return
      }
      this.loading = true
      this.error = ''
      try {
        let data
        if (this.isRegister) {
          data = await api.register(this.phone, this.password)
        } else {
          data = await api.login(this.phone, this.password)
        }
        uni.setStorageSync('token', data.token)
        uni.showToast({ title: this.isRegister ? '注册成功' : '登录成功' })
        setTimeout(() => uni.reLaunch({ url: '/pages/index/index' }), 800)
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    goBack() { uni.navigateBack() }
  }
}
</script>
