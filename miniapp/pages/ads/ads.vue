<template>
  <view class="container">
    <view class="card">
      <text style="font-size: 30rpx; font-weight: 600;">📢 观看广告获取起卦次数</text>
      <text class="text-muted" style="display: block; margin-top: 8rpx;">每个账号每日最多 3 次</text>
      <view v-if="quota >= 0" style="margin-top: 12rpx; font-size: 26rpx; color: #5D4037;">
        当前剩余：<text style="color:#8B4513; font-weight:600;">{{ quota }}</text> 次
      </view>
    </view>

    <view v-for="ad in ads" :key="ad.id" class="card">
      <view style="font-size: 30rpx; font-weight: 600;">{{ ad.name }}</view>
      <text class="text-muted" style="display: block; margin-top: 4rpx;">{{ ad.description }}</text>
      <view style="display: flex; justify-content: space-between; align-items: center; margin-top: 16rpx;">
        <text style="font-size: 26rpx;">观看 {{ ad.watch_duration }}s 奖励 {{ ad.reward_quota }} 次</text>
        <button class="btn-primary" style="padding:12rpx 32rpx; font-size:26rpx;" @tap="watchAd(ad)">
          观看
        </button>
      </view>
      <!-- 广告内容模拟 -->
      <view v-if="watchingId === ad.id" style="margin-top: 16rpx; padding: 16rpx; background: #F0EBE0; border-radius: 8rpx; text-align: center;">
        <text>广告播放中... {{ countdown }}s</text>
      </view>
    </view>

    <view v-if="ads.length === 0" class="text-center text-muted" style="padding: 120rpx 0;">暂无可用广告</view>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'

export default {
  data() {
    return { ads: [], watchingId: null, countdown: 0, timer: null, quota: -1 }
  },
  onShow() { this.loadAds(); this.loadQuota() },
  onUnload() { if (this.timer) clearInterval(this.timer) },
  methods: {
    async loadQuota() {
      try {
        const data = await api.profile()
        this.quota = data.remaining_quota ?? -1
      } catch {}
    },
    async loadAds() {
      try {
        const data = await api.activeAds()
        this.ads = data.items || data || []
      } catch (e) { uni.showToast({ title: e.message, icon: 'none' }) }
    },
    async watchAd(ad) {
      try {
        await api.watchAd(ad.id)
        this.watchingId = ad.id
        this.countdown = ad.watch_duration
        this.timer = setInterval(async () => {
          this.countdown--
          if (this.countdown <= 0) {
            clearInterval(this.timer)
            try {
              const data = await api.completeAd(ad.id, ad.watch_duration)
              uni.showToast({ title: `获得 ${ad.reward_quota} 次起卦！` })
              this.quota = data.remaining_quota ?? (this.quota + (data.rewarded || 1))
            } catch (e) { uni.showToast({ title: e.message, icon: 'none' }) }
            this.watchingId = null
          }
        }, 1000)
      } catch (e) { uni.showToast({ title: e.message, icon: 'none' }) }
    }
  }
}
</script>
