<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-xl font-bold">🤖 模型管理</h2>
      <button @click="showForm = !showForm"
        class="px-4 py-2 rounded-lg text-sm font-medium transition"
        :class="showForm ? 'bg-slate-600 text-white' : 'bg-cyan-600 text-white hover:bg-cyan-500'">
        {{ showForm ? '取消' : '+ 添加模型' }}
      </button>
    </div>

    <!-- 添加/编辑表单 -->
    <div v-if="showForm" class="bg-slate-700 rounded-lg p-4 mb-6">
      <h3 class="font-medium mb-3">{{ editing ? '编辑模型' : '添加模型' }}</h3>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-xs opacity-60 mb-1">显示名称</label>
          <input v-model="form.name" class="w-full bg-slate-800 rounded px-3 py-1.5 text-sm border border-slate-600 focus:border-cyan-500 outline-none" placeholder="千问">
        </div>
        <div>
          <label class="block text-xs opacity-60 mb-1">Provider Key</label>
          <input v-model="form.provider" class="w-full bg-slate-800 rounded px-3 py-1.5 text-sm border border-slate-600 focus:border-cyan-500 outline-none" placeholder="qwen">
        </div>
        <div>
          <label class="block text-xs opacity-60 mb-1">API Key</label>
          <input v-model="form.api_key" type="password" class="w-full bg-slate-800 rounded px-3 py-1.5 text-sm border border-slate-600 focus:border-cyan-500 outline-none" placeholder="sk-xxx">
        </div>
        <div>
          <label class="block text-xs opacity-60 mb-1">模型名</label>
          <input v-model="form.model" class="w-full bg-slate-800 rounded px-3 py-1.5 text-sm border border-slate-600 focus:border-cyan-500 outline-none" placeholder="qwen-plus">
        </div>
        <div class="col-span-2">
          <label class="block text-xs opacity-60 mb-1">Endpoint</label>
          <input v-model="form.endpoint" class="w-full bg-slate-800 rounded px-3 py-1.5 text-sm border border-slate-600 focus:border-cyan-500 outline-none" placeholder="https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions">
        </div>
      </div>
      <div class="mt-4 flex gap-2">
        <button @click="saveProvider"
          class="px-4 py-1.5 rounded text-sm font-medium bg-cyan-600 hover:bg-cyan-500 transition"
          :disabled="saving">
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <button v-if="editing" @click="resetForm"
          class="px-4 py-1.5 rounded text-sm bg-slate-600 hover:bg-slate-500 transition">
          取消编辑
        </button>
      </div>
    </div>

    <!-- 列表 -->
    <div v-if="loading" class="text-center opacity-50 py-12">加载中...</div>
    <div v-else-if="providers.length === 0" class="text-center opacity-50 py-12">
      暂无配置的模型，请先添加
    </div>
    <div v-for="p in providers" :key="p.id"
      class="bg-slate-700/80 rounded-lg p-4 mb-3 flex items-center justify-between">
      <div>
        <div class="flex items-center gap-2">
          <span class="font-medium">{{ p.name }}</span>
          <span v-if="p.is_default" class="text-xs bg-cyan-600 text-white px-1.5 py-0.5 rounded">默认</span>
          <span class="text-xs opacity-40">({{ p.provider }})</span>
        </div>
        <div class="text-xs opacity-50 mt-1">{{ p.model }} — {{ p.endpoint }}</div>
      </div>
      <div class="flex gap-2">
        <button v-if="!p.is_default" @click="setDefault(p.id)"
          class="text-xs px-2 py-1 rounded bg-slate-600 hover:bg-cyan-600 transition">
          设为默认
        </button>
        <button @click="editProvider(p)"
          class="text-xs px-2 py-1 rounded bg-slate-600 hover:bg-slate-500 transition">
          编辑
        </button>
        <button @click="deleteProvider(p.id)"
          class="text-xs px-2 py-1 rounded bg-red-900/50 hover:bg-red-700 transition">
          删除
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const providers = ref([])
const loading = ref(true)
const saving = ref(false)
const showForm = ref(false)
const editing = ref(null)

const form = ref({
  name: '',
  provider: '',
  api_key: '',
  endpoint: '',
  model: '',
})

function resetForm() {
  form.value = { name: '', provider: '', api_key: '', endpoint: '', model: '' }
  editing.value = null
  showForm.value = false
}

function editProvider(p) {
  form.value = {
    name: p.name,
    provider: p.provider,
    api_key: p.api_key,
    endpoint: p.endpoint,
    model: p.model,
  }
  editing.value = p.id
  showForm.value = true
}

async function fetchProviders() {
  loading.value = true
  const res = await fetch('/api/admin/llm', {
    headers: { Authorization: `Bearer ${token()}` },
  })
  const data = await res.json()
  if (res.ok) providers.value = data.items || []
  loading.value = false
}

async function saveProvider() {
  if (!form.value.provider || !form.value.endpoint || !form.value.model) {
    alert('Provider Key、Endpoint 和模型名为必填')
    return
  }
  saving.value = true
  const method = editing.value ? 'PUT' : 'POST'
  const url = editing.value ? `/api/admin/llm/${editing.value}` : '/api/admin/llm'
  const res = await fetch(url, {
    method,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token()}`,
    },
    body: JSON.stringify(form.value),
  })
  if (res.ok) {
    resetForm()
    await fetchProviders()
  } else {
    const err = await res.json()
    alert(err.error || '操作失败')
  }
  saving.value = false
}

async function deleteProvider(id) {
  if (!confirm('确定删除？')) return
  const res = await fetch(`/api/admin/llm/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token()}` },
  })
  if (res.ok) await fetchProviders()
  else alert('删除失败')
}

async function setDefault(id) {
  const res = await fetch(`/api/admin/llm/${id}/default`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token()}` },
  })
  if (res.ok) await fetchProviders()
}

function token() {
  return localStorage.getItem('token') || ''
}

onMounted(fetchProviders)
</script>
