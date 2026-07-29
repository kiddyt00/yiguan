<template>
  <div class="max-w-sm mx-auto space-y-5">
    <!-- 配额 -->
    <div class="text-center py-8 px-6 rounded-2xl" :class="isDark ? 'bg-stone-800/60' : 'bg-gradient-to-br from-amber-50 to-white border border-amber-100'">
      <div class="text-xs font-medium tracking-widest uppercase" :class="isDark ? 'text-stone-400' : 'text-stone-400'">剩余次数</div>
      <div class="mt-2 text-6xl font-extrabold" :class="quota > 0 ? 'text-amber-600' : 'text-red-400'">{{ quota }}</div>
      <div class="mt-1 text-xs" :class="isDark ? 'text-stone-500' : 'text-stone-400'">次</div>
      <router-link to="/recharge" class="inline-flex items-center gap-1.5 mt-4 px-6 py-2.5 rounded-xl text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500 shadow-lg shadow-amber-600/20">
        💎 充值
      </router-link>
    </div>

    <!-- 个人信息 -->
    <div class="rounded-2xl p-6" :class="isDark ? 'bg-stone-800/60' : 'bg-white border border-stone-100'">
      <h3 class="text-sm font-bold mb-4" :class="isDark ? 'text-stone-200' : 'text-stone-700'">个人信息</h3>
      <div class="space-y-3.5 text-sm">
        <div class="flex items-center justify-between py-2 px-3 rounded-lg" :class="isDark ? 'bg-stone-700/40' : 'bg-stone-50'">
          <span class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">手机号</span>
          <span class="font-medium" :class="isDark ? 'text-stone-200' : 'text-stone-800'">{{ user.phone }}</span>
        </div>
        <div>
          <label class="text-xs" :class="isDark ? 'text-stone-400' : 'text-stone-500'">昵称</label>
          <input v-model="form.nickname"
            class="w-full mt-1 px-3 py-2 rounded-lg border bg-transparent outline-none transition text-sm focus:border-amber-500 focus:ring-1 focus:ring-amber-500/20"
            :class="isDark ? 'border-stone-600 text-stone-200' : 'border-stone-200 text-stone-800'" />
        </div>
        <button @click="save" :disabled="saving"
          class="w-full py-2.5 rounded-xl text-sm font-medium transition bg-amber-600 text-white hover:bg-amber-500 disabled:opacity-50">
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <p v-if="msg" class="text-center text-xs text-green-500">{{ msg }}</p>
      </div>
    </div>

    <!-- 邀请好友 -->
    <div class="rounded-2xl p-6" :class="isDark ? 'bg-stone-800/60' : 'bg-white border border-stone-100'">
      <h3 class="text-sm font-bold mb-1" :class="isDark ? 'text-stone-200' : 'text-stone-700'">📤 邀请好友</h3>
      <p class="text-xs mb-3" :class="isDark ? 'text-stone-400' : 'text-stone-400'">每邀请3位好友注册，其中1人完成测算，得1次免费测算</p>

      <div v-if="inviteCode" class="space-y-3">
        <div class="flex items-center gap-2 py-2 px-3 rounded-lg text-sm font-mono" :class="isDark ? 'bg-stone-700/40 text-stone-300' : 'bg-stone-50 text-stone-700'">
          <span class="flex-1 truncate">{{ inviteCode }}</span>
          <button @click="copyCode" class="text-xs px-2 py-1 rounded bg-amber-600 text-white hover:bg-amber-500">复制</button>
        </div>

        <div class="flex gap-2 text-xs">
          <div class="flex-1 text-center py-2 rounded-lg" :class="isDark ? 'bg-stone-700/40' : 'bg-stone-50'">
            <div class="font-bold text-lg" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ progress.registered_count }}</div>
            <div class="text-stone-400">已注册</div>
          </div>
          <div class="flex-1 text-center py-2 rounded-lg" :class="isDark ? 'bg-stone-700/40' : 'bg-stone-50'">
            <div class="font-bold text-lg" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ progress.divined_count }}</div>
            <div class="text-stone-400">已测算</div>
          </div>
          <div class="flex-1 text-center py-2 rounded-lg" :class="isDark ? 'bg-stone-700/40' : 'bg-stone-50'">
            <div class="font-bold text-lg" :class="isDark ? 'text-stone-200' : 'text-stone-700'">{{ progress.reward_round }}</div>
            <div class="text-stone-400">已奖励</div>
          </div>
        </div>

        <div v-if="progress.pending_reward" class="text-center">
          <button @click="claimReward" :disabled="claiming"
            class="px-4 py-2 rounded-xl text-sm font-medium bg-green-600 text-white hover:bg-green-500 disabled:opacity-50">
            🎁 领取奖励
          </button>
        </div>
        <p v-else class="text-center text-xs text-stone-400">
          还需 {{ Math.max(0, 3 - progress.registered_count) }} 人注册，{{ Math.max(0, 1 - progress.divined_count) }} 人测算
        </p>
      </div>

      <div v-else class="text-center text-sm text-stone-400 py-4">加载中...</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { apiGetJSON, apiPutJSON, apiPostJSON } from '../utils/request'

defineProps({ isDark: Boolean })
const auth = useAuthStore()
const user = auth.user || {}
const form = ref({ nickname: '' })
const saving = ref(false)
const msg = ref('')
const quota = ref(0)
const inviteCode = ref('')
const progress = ref({ registered_count: 0, divined_count: 0, reward_round: 0, pending_reward: false })
const claiming = ref(false)

onMounted(async () => {
  try {
    const data = await apiGetJSON('/api/user')
    quota.value = data.remaining_quota || 0
    form.value.nickname = (data.user || data).nickname || ''
  } catch (e) {}
  loadInvite()
})

async function loadInvite() {
  try {
    const [codeData, progressData] = await Promise.all([
      apiGetJSON('/api/invite/code').catch(() => ({ invite_code: '' })),
      apiGetJSON('/api/invite/progress').catch(() => ({})),
    ])
    inviteCode.value = codeData.invite_code || ''
    progress.value = { registered_count: 0, divined_count: 0, reward_round: 0, pending_reward: false, ...progressData }
  } catch (e) {}
}

function copyCode() {
  const shareUrl = `https://zgjz.insightj.cn?invite=${inviteCode.value}`
  // 优先用 Web Share API（手机浏览器弹出原生分享）
  if (navigator.share) {
    navigator.share({ title: '观己斋·周易占筮', text: '来算一卦吧，测运势、问前程', url: shareUrl })
    return
  }
  // 回退到复制链接
  navigator.clipboard.writeText(shareUrl).then(() => {
    alert('邀请链接已复制，分享给好友即可')
  }).catch(() => {
    alert('邀请码: ' + inviteCode.value)
  })
}

async function claimReward() {
  claiming.value = true
  try {
    const data = await apiPostJSON('/api/invite/claim', {})
    if (data.rewarded) {
      alert('🎉 获得1次免费测算！')
      loadInvite()
      // 刷新quota
      const userData = await apiGetJSON('/api/user')
      quota.value = userData.remaining_quota || 0
    } else {
      alert('暂未达到奖励条件')
    }
  } catch (e) {
    alert('领取失败')
  }
  claiming.value = false
}

async function save() {
  saving.value = true
  try {
    const json = await apiPutJSON('/api/user', { nickname: form.value.nickname })
    auth.setAuth(auth.token, json.user || json)
    msg.value = '保存成功'
    setTimeout(() => msg.value = '', 2000)
  } catch (e) { msg.value = '保存失败' }
  saving.value = false
}
</script>
