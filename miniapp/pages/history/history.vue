<template>
  <view class="container">
    <view v-if="items.length === 0" class="text-center text-muted" style="padding: 120rpx 0;">暂无记录</view>
    <view v-for="h in items" :key="h.id" class="card" @tap="toggleItem(h)">
      <view style="display:flex; justify-content:space-between; align-items:flex-start;">
        <view>
          <text style="font-weight:600;">{{ h.primary_gua }}</text>
          <text class="text-muted"> → {{ h.changing_gua }}</text>
        </view>
        <text class="text-muted" style="font-size:22rpx;">{{ h.created_at?.slice(0,10) }}</text>
      </view>
      <text class="text-muted" style="display:block; margin-top:8rpx;">问：{{ h.question }}</text>
      <view v-if="expandedId === h.id" style="margin-top:16rpx; padding-top:16rpx; border-top:1rpx solid #E0D6C8;">
        <rich-text :nodes="renderMD(h.interpretation)" style="font-size:26rpx; line-height:1.8; color:#5D4037;" />
      </view>
    </view>
    <view v-if="hasMore" class="text-center mt-3">
      <button class="btn-secondary" @tap="loadMore" :loading="loading">加载更多</button>
    </view>
  </view>
</template>

<script>
import { api } from '../../utils/api.js'
import marked from '../../utils/marked.js'

export default {
  data() {
    return { items: [], offset: 0, hasMore: true, loading: false, expandedId: null }
  },
  onShow() {
    this.items = []; this.offset = 0; this.hasMore = true; this.loadMore()
  },
  methods: {
    async loadMore() {
      this.loading = true
      try {
        const data = await api.history(20, this.offset)
        this.items.push(...(data.items || []))
        this.offset += 20
        this.hasMore = data.total > this.items.length
      } catch (e) { uni.showToast({ title: e.message, icon: 'none' }) }
      finally { this.loading = false }
    },
    toggleItem(h) {
      this.expandedId = this.expandedId === h.id ? null : h.id
    },
    renderMD(text) {
      return marked.parse(text || '').replace(/\n/g, '<br/>')
    }
  }
}
</script>
