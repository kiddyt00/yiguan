<template>
  <el-container class="min-h-screen">
    <el-aside width="200px" class="bg-slate-800 text-white">
      <div class="text-xl font-bold p-4 text-center">☯ 易观后台</div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#1e293b"
        text-color="#cbd5e1"
        active-text-color="#38bdf8"
        class="border-0"
      >
        <el-menu-item index="/"><el-icon><DataBoard /></el-icon>仪表盘</el-menu-item>
        <el-menu-item index="/users"><el-icon><User /></el-icon>用户管理</el-menu-item>
        <el-menu-item index="/hexagrams"><el-icon><List /></el-icon>卦象任务</el-menu-item>
        <el-menu-item index="/models"><el-icon><Cpu /></el-icon>模型管理</el-menu-item>
        <el-menu-item index="/ads"><el-icon><Notification /></el-icon>广告管理</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="bg-white shadow-sm flex items-center justify-between px-4">
        <span class="text-lg font-medium">{{ pageTitle }}</span>
        <el-dropdown @command="handleCommand">
          <span class="cursor-pointer">{{ user?.nickname || '管理员' }} <el-icon><ArrowDown /></el-icon></span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="bg-gray-50 p-6">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { user } from '../stores/auth'
import { logout } from '../stores/auth'

const route = useRoute()
const router = useRouter()

const titles = { '/': '仪表盘', '/users': '用户管理', '/hexagrams': '卦象任务', '/models': '模型管理', '/ads': '广告管理' }
const pageTitle = computed(() => titles[route.path] || '')

function handleCommand(cmd) {
  if (cmd === 'logout') {
    logout()
    router.push('/login')
  }
}
</script>
