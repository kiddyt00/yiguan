<template>
  <view class="container">
    <!-- 头部 -->
    <view class="text-center mb-3" style="padding-top: 60rpx;">
      <text style="font-size: 80rpx;">☯</text>
      <view style="font-size: 44rpx; font-weight: 700; margin-top: 16rpx;">观己斋</view>
      <text class="text-muted" style="display: block; margin-top: 8rpx;">观易知变，见心明境</text>
    </view>

    <!-- 输入区 -->
    <view class="card">
      <view style="font-size: 30rpx; font-weight: 600; margin-bottom: 16rpx;">请默想你的问题：</view>
      <textarea
        v-model="question"
        placeholder="例如：最近有些迷茫，想听听《周易》的启发..."
        :maxlength="200"
        style="width: 100%; height: 160rpx; font-size: 28rpx; line-height: 1.6;"
      />
      <view class="flex gap-2 mt-3">
        <button class="btn-secondary" style="flex:1;" @tap="goHistory">历史记录</button>
        <button class="btn-primary" style="flex:2;" :disabled="!question" @tap="startDivine">
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
      <image src="/static/master-qr.png" mode="widthFix" style="width: 400rpx; margin: 24rpx auto;" />
      <text class="text-muted">长按识别二维码添加大师微信</text>
      <button class="btn-secondary mt-3" @tap="showMaster = false">关闭</button>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return { question: '', loading: false, showMaster: false }
  },
  methods: {
    startDivine() {
      if (!this.question) return
      const token = uni.getStorageSync('token')
      if (!token) {
        uni.navigateTo({ url: '/pages/login/login' })
        return
      }
      this.loading = true
      uni.navigateTo({
        url: '/pages/result/result',
        success: () => { this.loading = false },
        fail: () => { this.loading = false }
      })
      // 通过全局变量传递问题
      getApp().globalData = { question: this.question }
    },
    goHistory() { uni.navigateTo({ url: '/pages/history/history' }) },
    goUser() { uni.navigateTo({ url: '/pages/user/user' }) },
    goAds() { uni.navigateTo({ url: '/pages/ads/ads' }) },
    goAbout() { uni.navigateTo({ url: '/pages/about/about' }) }
  }
}
</script>
