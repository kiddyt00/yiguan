<template>
  <el-container class="min-h-screen">
    <el-aside width="200px" class="bg-[var(--sidebar)] text-white flex flex-col">
      <!-- Brand 区 -->
      <div class="sidebar-brand">
        <span class="brand-icon">
          <svg viewBox="0 0 48 36" class="w-11 h-8" fill="currentColor" aria-hidden="true">
            <rect x="2" y="2" width="44" height="7" rx="3.5" />
            <rect x="2" y="14.5" width="44" height="7" rx="3.5" />
            <rect x="2" y="27" width="44" height="7" rx="3.5" />
          </svg>
        </span>
        <div class="brand-text">真观己斋后台</div>
        <div class="brand-sub">ZHEN GUAN JI ZHAI</div>
      </div>

      <!-- 导航菜单（配色由 --el-menu-* 变量控制） -->
      <el-menu
        :default-active="route.path"
        router
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
        <el-menu-item index="/analytics">
          <el-icon><DataAnalysis /></el-icon>
          <span>数据分析</span>
        </el-menu-item>
        <el-menu-item index="/models">
          <el-icon><Cpu /></el-icon>
          <span>模型管理</span>
        </el-menu-item>
        <el-menu-item index="/ads">
          <el-icon><Notification /></el-icon>
          <span>广告管理</span>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><Document /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
        <el-menu-item index="/refunds">
          <el-icon><RefreshLeft /></el-icon>
          <span>退款管理</span>
        </el-menu-item>
        <el-menu-item index="/memberships">
          <el-icon><Crown /></el-icon>
          <span>会员管理</span>
        </el-menu-item>
        <el-menu-item index="/progress">
          <el-icon><List /></el-icon>
          <span>开发进度</span>
        </el-menu-item>
      </el-menu>

      <!-- 底部版本 -->
      <div class="text-center py-3 text-xs" style="color: oklch(45% 0.02 75); letter-spacing: 1px;">
        v2.3
      </div>
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header class="admin-header">
        <div class="page-title">{{ pageTitle }}</div>
        <el-dropdown @command="handleCommand">
          <span class="flex items-center gap-1.5 cursor-pointer text-sm font-medium text-[var(--ink-2)] hover:text-[var(--accent-deep)] transition-colors">
            <span class="w-7 h-7 rounded-full bg-[oklch(93%_0.03_80)] text-[var(--accent-deep)] flex items-center justify-center text-xs font-bold">
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
  '/analytics': '数据分析',
  '/models': '模型管理',
  '/ads': '广告管理',
  '/progress': '开发进度',
  '/orders': '订单管理',
  '/refunds': '退款管理',
  '/memberships': '会员管理',
}
const pageTitle = computed(() => titles[route.path] || '')

function handleCommand(cmd) {
  if (cmd === 'logout') {
    logout()
    router.push('/login')
  }
}
</script>
