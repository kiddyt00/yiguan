<template>
  <div class="min-h-screen flex bg-[var(--paper-2)]">
    <!-- ===== 左栏 · 品牌区（墨色） ===== -->
    <div class="hidden md:flex md:w-[42%] lg:w-[46%] bg-[var(--sidebar)] text-white flex-col justify-between p-12 relative overflow-hidden">
      <!-- 氛围光（固定径向，atmospheric 允许） -->
      <div class="absolute -top-32 -right-32 w-80 h-80 rounded-full bg-[oklch(70%_0.10_75)]/10 blur-3xl pointer-events-none"></div>
      <div class="absolute -bottom-40 -left-24 w-96 h-96 rounded-full bg-[oklch(70%_0.10_75)]/8 blur-3xl pointer-events-none"></div>

      <div class="relative">
        <!-- 乾卦印章 -->
        <svg viewBox="0 0 48 36" class="w-14 h-11 text-[var(--accent)] mb-8" fill="currentColor" aria-hidden="true">
          <rect x="2" y="2" width="44" height="7" rx="3.5" />
          <rect x="2" y="14.5" width="44" height="7" rx="3.5" />
          <rect x="2" y="27" width="44" height="7" rx="3.5" />
        </svg>

        <h1 class="text-3xl font-bold tracking-[0.2em] text-[var(--accent-2)]">真观己斋</h1>
        <p class="mt-3 text-sm tracking-[0.35em] text-[var(--sidebar-text)]">周易占筮 · 管理平台</p>
      </div>

      <div class="relative">
        <p class="text-xs leading-6 tracking-wider text-[var(--sidebar-text)]/80">
          观己观心，知时知命。<br />
          后台仅供授权管理员访问。
        </p>
        <p class="mt-6 text-[11px] tracking-widest" style="color: oklch(45% 0.02 75);">ZHEN GUAN JI ZHAI · v2.5</p>
      </div>
    </div>

    <!-- ===== 右栏 · 表单区（宣纸色） ===== -->
    <div class="flex-1 flex items-center justify-center px-6 py-16">
      <div class="w-full max-w-sm">
        <!-- 移动端品牌（左栏隐藏时） -->
        <div class="md:hidden text-center mb-10">
          <svg viewBox="0 0 48 36" class="w-12 h-9 mx-auto text-[var(--accent)] mb-4" fill="currentColor" aria-hidden="true">
            <rect x="2" y="2" width="44" height="7" rx="3.5" />
            <rect x="2" y="14.5" width="44" height="7" rx="3.5" />
            <rect x="2" y="27" width="44" height="7" rx="3.5" />
          </svg>
          <h1 class="text-2xl font-bold tracking-[0.18em] text-[var(--ink)]">真观己斋</h1>
          <p class="mt-2 text-xs tracking-[0.3em] text-[var(--muted)]">周易占筮 · 管理平台</p>
        </div>

        <div class="mb-8">
          <h2 class="text-xl font-bold text-[var(--ink)] tracking-wider">管理员登录</h2>
          <p class="mt-2 text-sm text-[var(--muted)]">请输入账号密码进入后台</p>
        </div>

        <el-form @submit.prevent="handleLogin" class="px-0">
          <el-form-item>
            <el-input
              v-model="phone"
              placeholder="手机号"
              size="large"
              :prefix-icon="Phone"
            />
          </el-form-item>
          <el-form-item>
            <el-input
              v-model="password"
              type="password"
              placeholder="密码"
              size="large"
              :prefix-icon="Lock"
              show-password
              @keydown="e=>e.key==='Enter'&&handleLogin()"
            />
          </el-form-item>
          <el-form-item class="!mb-0 mt-8">
            <el-button
              type="primary"
              size="large"
              class="w-full !h-11 !text-base !font-medium !rounded-lg tracking-widest"
              :loading="loading"
              @click="handleLogin"
            >
              登 录
            </el-button>
          </el-form-item>
        </el-form>

        <p class="text-center text-xs mt-8 text-[var(--faint)]">仅限管理员访问 · 登录即记录审计日志</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Phone, Lock } from '@element-plus/icons-vue'
import { login } from '../stores/auth'
import { ElMessage } from 'element-plus'

const phone = ref('')
const password = ref('')
const loading = ref(false)
const router = useRouter()

async function handleLogin() {
  if (!phone.value || !password.value) {
    ElMessage.warning('请输入手机号和密码')
    return
  }
  loading.value = true
  try {
    await login(phone.value, password.value)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>
