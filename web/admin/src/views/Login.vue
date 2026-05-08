<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-slate-800 to-indigo-950">
    <!-- 背景装饰 -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-96 h-96 rounded-full bg-indigo-500/10 blur-3xl"></div>
      <div class="absolute -bottom-40 -left-40 w-96 h-96 rounded-full bg-cyan-500/10 blur-3xl"></div>
    </div>

    <div class="relative z-10 w-full max-w-md px-4">
      <!-- Logo 区 -->
      <div class="text-center mb-8">
        <div class="text-5xl mb-3">☯</div>
        <h1 class="text-2xl font-bold text-white">易观后台管理</h1>
        <p class="text-slate-400 text-sm mt-2">AI 智能解卦平台</p>
      </div>

      <!-- 登录卡片 -->
      <el-card class="!rounded-xl !border-0 !bg-white/95 backdrop-blur shadow-2xl">
        <template #header>
          <div class="text-center text-base font-semibold text-slate-700">管理员登录</div>
        </template>
        <el-form @submit.prevent="handleLogin" class="px-2">
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
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item class="!mb-0">
            <el-button
              type="primary"
              size="large"
              class="w-full !h-11 !text-base !font-medium !rounded-lg"
              :loading="loading"
              @click="handleLogin"
            >
              登 录
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <p class="text-center text-slate-500 text-xs mt-6">仅限管理员访问</p>
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
