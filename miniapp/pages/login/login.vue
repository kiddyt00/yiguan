<template>
  <view class="container" style="padding-top: 120rpx;">
    <view class="text-center mb-3">
      <text style="font-size: 80rpx;">☯</text>
      <view style="font-size: 36rpx; font-weight: 700; margin-top: 16rpx;">观己斋</view>
      <text class="text-muted" style="display: block; margin-top: 8rpx;">观易知变，见心明境</text>
    </view>

    <!-- 微信一键登录 -->
    <view class="card text-center">
      <button
        class="btn-primary"
        style="width: 100%; background: #07C160;"
        :loading="wxLoading"
        @tap="wechatLogin"
        open-type="getPhoneNumber"
        @getphonenumber="onGetPhone"
      >
        <text style="margin-right: 8rpx;">📱</text>微信授权登录
      </button>
      <text class="text-muted" style="display: block; margin-top: 8rpx;">一键授权，无需注册</text>
    </view>

    <!-- 分隔线 -->
    <view style="display: flex; align-items: center; margin: 32rpx 0;">
      <view style="flex: 1; height: 1rpx; background: #E0D6C8;"></view>
      <text class="text-muted" style="margin: 0 24rpx;">或</text>
      <view style="flex: 1; height: 1rpx; background: #E0D6C8;"></view>
    </view>

    <!-- 手机号验证码登录 -->
    <view class="card">
      <view style="display: flex; gap: 16rpx; margin-bottom: 24rpx;">
        <input v-model="phone" type="number" placeholder="手机号" maxlength="11"
          style="flex: 1; border-bottom: 1rpx solid #E0D6C8; padding: 16rpx 0; font-size: 30rpx;" />
        <button
          class="btn-secondary"
          style="padding: 12rpx 24rpx; font-size: 24rpx; white-space: nowrap;"
          :disabled="sendDisabled"
          @tap="sendSMS"
        >
          {{ sendText }}
        </button>
      </view>
      <input v-model="code" type="number" placeholder="验证码" maxlength="6"
        style="border-bottom: 1rpx solid #E0D6C8; padding: 16rpx 0; font-size: 30rpx; margin-bottom: 32rpx;" />
      <button class="btn-primary" style="width: 100%;" :loading="loading" @tap="smsLogin">
        登录 / 注册
      </button>
    </view>

    <!-- 密码登录入口 -->
    <view class="text-center mt-3">
      <text class="text-muted" @tap="goPwdLogin">密码登录</text>
    </view>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'

export default {
  data() {
    return {
      phone: '',
      code: '',
      loading: false,
      wxLoading: false,
      sendCountdown: 0,
      timer: null
    }
  },
  computed: {
    sendDisabled() { return this.sendCountdown > 0 || this.phone.length !== 11 },
    sendText() { return this.sendCountdown > 0 ? `${this.sendCountdown}s` : '获取验证码' }
  },
  onUnload() { if (this.timer) clearInterval(this.timer) },
  methods: {
    // 微信登录
    wechatLogin() {
      this.wxLoading = true
      uni.login({
        provider: 'weixin',
        success: async (loginRes) => {
          try {
            const res = await uni.request({
              url: 'https://gjz.shadouyou.cloud/api/auth/wechat-login',
              method: 'POST',
              data: { code: loginRes.code },
              header: { 'Content-Type': 'application/json' }
            })
            if (res.statusCode === 200 && res.data.token) {
              uni.setStorageSync('token', res.data.token)
              uni.showToast({ title: '登录成功' })
              setTimeout(() => uni.reLaunch({ url: '/pages/index/index' }), 800)
            } else {
              uni.showToast({ title: res.data?.error || '登录失败', icon: 'none' })
            }
          } catch (e) {
            uni.showToast({ title: '网络错误', icon: 'none' })
          }
        },
        fail: () => {
          uni.showToast({ title: '微信登录失败', icon: 'none' })
        },
        complete: () => { this.wxLoading = false }
      })
    },

    // 发送短信验证码
    async sendSMS() {
      if (this.phone.length !== 11) return
      try {
        const res = await uni.request({
          url: 'https://gjz.shadouyou.cloud/api/auth/sms-send',
          method: 'POST',
          data: { phone: this.phone },
          header: { 'Content-Type': 'application/json' }
        })
        if (res.statusCode === 200) {
          uni.showToast({ title: '验证码已发送' })
          this.sendCountdown = 60
          this.timer = setInterval(() => {
            this.sendCountdown--
            if (this.sendCountdown <= 0) clearInterval(this.timer)
          }, 1000)
        } else {
          uni.showToast({ title: res.data?.error || '发送失败', icon: 'none' })
        }
      } catch (e) {
        uni.showToast({ title: '网络错误', icon: 'none' })
      }
    },

    // 短信验证码登录
    async smsLogin() {
      if (!this.phone || !this.code) {
        uni.showToast({ title: '请输入手机号和验证码', icon: 'none' }); return
      }
      this.loading = true
      try {
        const res = await uni.request({
          url: 'https://gjz.shadouyou.cloud/api/auth/sms-login',
          method: 'POST',
          data: { phone: this.phone, code: this.code },
          header: { 'Content-Type': 'application/json' }
        })
        if (res.statusCode === 200 && res.data.token) {
          uni.setStorageSync('token', res.data.token)
          uni.showToast({ title: '登录成功' })
          setTimeout(() => uni.reLaunch({ url: '/pages/index/index' }), 800)
        } else {
          uni.showToast({ title: res.data?.error || '登录失败', icon: 'none' })
        }
      } catch (e) {
        uni.showToast({ title: '网络错误', icon: 'none' })
      } finally { this.loading = false }
    },

    goPwdLogin() {
      uni.navigateTo({ url: '/pages/login/pwd' })
    }
  }
}
</script>
