<template>
  <div class="min-h-screen flex">
    <aside class="w-48 bg-slate-800 text-white p-4">
      <div class="text-xl font-bold mb-6">☯ 易观后台</div>
      <nav class="space-y-2 text-sm">
        <router-link to="/" class="block py-1 px-2 rounded hover:bg-slate-700">仪表盘</router-link>
        <router-link to="/users" class="block py-1 px-2 rounded hover:bg-slate-700">用户管理</router-link>
        <router-link to="/llm" class="block py-1 px-2 rounded hover:bg-slate-700">模型管理</router-link>
      </nav>
    </aside>
    <main class="flex-1 p-6">
      <router-view />
    </main>
  </div>
</template>
