<template>
  <view class="container">
    <!-- 头部 -->
    <view class="text-center mb-3" style="padding-top: 60rpx;">
      <text style="font-size: 80rpx;">☯</text>
      <view style="font-size: 44rpx; font-weight: 700; margin-top: 16rpx;">观己斋</view>
      <text class="text-muted" style="display: block; margin-top: 8rpx;">观易知变，见心明境</text>
    </view>

    <!-- 登录/配额状态 -->
    <view v-if="isLoggedIn" class="text-center mb-3">
      <text v-if="quota >= 0" style="font-size: 26rpx;"
        :class="quota > 0 ? 'text-muted' : ''"
        :style="quota <= 0 ? 'color:#C62828;' : ''">
        {{ quota > 0 ? '剩余 ' + quota + ' 次提问' : '次数已用完' }}
      </text>
      <text v-else class="text-muted">加载中...</text>
    </view>
    <view v-else class="card text-center">
      <text class="text-muted">登录后即可免费使用（新用户赠送 3 次）</text>
      <button class="btn-primary mt-3" style="width:100%;" @tap="goLogin">登录 / 注册</button>
    </view>

    <!-- 输入区 -->
    <view v-if="isLoggedIn" class="card">
      <view style="font-size: 30rpx; font-weight: 600; margin-bottom: 16rpx;">请默想你的问题：</view>
      <textarea
        v-model="question"
        placeholder="例如：最近有些迷茫，想听听《周易》的启发..."
        :maxlength="200"
        style="width: 100%; height: 160rpx; font-size: 28rpx; line-height: 1.6;"
      />
      <view class="flex gap-2 mt-3">
        <button class="btn-secondary" style="flex:1;" @tap="goHistory">历史记录</button>
        <button class="btn-primary" style="flex:2;" :disabled="!question || quota <= 0" @tap="startDivine">
          {{ loading ? '起卦中...' : '开始提问' }}
        </button>
      </view>
    </view>

    <!-- 底部入口 -->
    <view class="flex gap-2" style="justify-content: center;">
      <text class="text-muted" @tap="goUser">个人中心</text>
      <text class="text-muted">|</text>
      <text class="text-muted" @tap="goAds">📢 看广告领次数</text>
      <text class="text-muted">|</text>
      <text class="text-muted" @tap="goAbout">关于我们</text>
    </view>

    <!-- 大师入口 -->
    <view class="card text-center mt-3" @tap="showMaster = true">
      <text style="font-size: 28rpx; color: #8B4513;">🎓 周易大师一对一详解</text>
    </view>

    <!-- 大师二维码弹窗 -->
    <view v-if="showMaster" class="card text-center mt-3">
      <text style="font-size: 30rpx; font-weight: 600;">周易大师 · 一对一深度交流</text>
      <image src="/static/master-qr.svg" mode="widthFix" style="width: 400rpx; margin: 24rpx auto;" />
      <text class="text-muted">长按识别二维码添加大师微信</text>
      <button class="btn-secondary mt-3" @tap="showMaster = false">关闭</button>
    </view>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'
export default {
  data() {
    return { question: '', loading: false, showMaster: false, quota: -1, isLoggedIn: false }
  },
  onShow() {
    this.isLoggedIn = !!uni.getStorageSync('token')
    if (this.isLoggedIn) this.loadQuota()
  },
  onShareAppMessage() {
    return { title: '观己斋 - 观易知变，见心明境', path: '/pages/index/index' }
  },
  onShareTimeline() {
    return { title: '观己斋 - 观易知变，见心明境', query: '' }
  },
  methods: {
    async loadQuota() {
      try {
        const data = await api.profile()
        this.quota = data.remaining_quota ?? -1
      } catch { this.quota = -1 }
    },
    startDivine() {
      if (!this.question) return
      const token = uni.getStorageSync('token')
      if (!token) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      if (this.quota <= 0) {
        uni.showToast({ title: '次数已用完，请先获取次数', icon: 'none' })
        return
      }
      this.loading = true
      getApp().globalData = { question: this.question }
      uni.navigateTo({
        url: '/pages/result/result',
        complete: () => { this.loading = false }
      })
    },
    goLogin() { uni.navigateTo({ url: '/pages/login/login' }) },
    goHistory() { uni.navigateTo({ url: '/pages/history/history' }) },
    goUser() { uni.navigateTo({ url: '/pages/user/user' }) },
    goAds() { uni.navigateTo({ url: '/pages/ads/ads' }) },
    goAbout() { uni.navigateTo({ url: '/pages/about/about' }) }
  }
}
</script>
