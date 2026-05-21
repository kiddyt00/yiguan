<template>
  <el-container class="min-h-screen">
    <el-aside width="200px" class="bg-[#1a1a2e] text-white flex flex-col">
      <!-- Brand 区 -->
      <div class="sidebar-brand">
        <span class="brand-icon">☯</span>
        <div class="brand-text">易观后台</div>
        <div class="brand-sub">YI GUAN ADMIN</div>
      </div>

      <!-- 导航菜单 -->
      <el-menu
        :default-active="route.path"
        router
        background-color="#1a1a2e"
        text-color="#94a3b8"
        active-text-color="#d4a853"
        class="border-0 flex-1"
      >
        <el-menu-item index="/">
          <el-icon><DataBoard /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/hexagrams">
          <el-icon><List /></el-icon>
          <span>卦象任务</span>
        </el-menu-item>
        <el-menu-item index="/models">
          <el-icon><Cpu /></el-icon>
          <span>模型管理</span>
        </el-menu-item>
        <el-menu-item index="/ads">
          <el-icon><Notification /></el-icon>
          <span>广告管理</span>
        </el-menu-item>
      </el-menu>

      <!-- 底部版本 -->
      <div class="text-center py-3 text-xs" style="color: rgba(212, 168, 83, 0.25); letter-spacing: 1px;">
        v2.1
      </div>
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header class="admin-header">
        <div class="page-title">{{ pageTitle }}</div>
        <el-dropdown @command="handleCommand">
          <span class="flex items-center gap-1.5 cursor-pointer text-sm font-medium text-stone-600 hover:text-amber-600 transition-colors">
            <span class="w-7 h-7 rounded-full bg-amber-100 text-amber-700 flex items-center justify-center text-xs font-bold">
              {{ (user?.nickname || '管理').charAt(0) }}
            </span>
            {{ user?.nickname || '管理员' }}
            <el-icon class="text-xs"><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>

      <!-- 主内容 -->
      <el-main class="admin-main">
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

const titles = {
  '/': '仪表盘',
  '/users': '用户管理',
  '/hexagrams': '卦象任务',
  '/models': '模型管理',
  '/ads': '广告管理',
}
const pageTitle = computed(() => titles[route.path] || '')

function handleCommand(cmd) {
  if (cmd === 'logout') {
    logout()
    router.push('/login')
  }
}
</script>
