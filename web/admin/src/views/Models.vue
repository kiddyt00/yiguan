<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-2xl font-bold">模型管理</h2>
      <el-button type="primary" @click="openCreate">新增模型</el-button>
    </div>

    <!-- 模型列表 -->
    <el-table :data="models" stripe>
      <el-table-column prop="id" label="ID" width="50" />
      <el-table-column prop="display_name" label="显示名称" min-width="140" />
      <el-table-column prop="provider_label" label="供应商" width="130" />
      <el-table-column prop="name" label="模型 ID" width="160" />
      <el-table-column prop="endpoint" label="Base URL" min-width="220" show-overflow-tooltip />
      <el-table-column prop="is_enabled" label="启用" width="65" align="center">
        <template #default="{ row }">
          <el-switch :model-value="row.is_enabled === 1"
            @change="toggleModel(row)" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="is_default" label="默认" width="65" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" type="warning" size="small">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="190">
        <template #default="{ row }">
          <el-button size="small" v-if="!row.is_default" @click="setDefault(row)">设为默认</el-button>
          <el-button size="small" type="primary" @click="editModel(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑模型' : '新增模型'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="100px" class="pr-4">
        <el-form-item label="① 供应商" required>
          <el-select v-model="form.provider" class="w-full" filterable placeholder="选择供应商"
            @change="onProviderChange">
            <el-option v-for="p in PROVIDERS" :key="p.key" :label="p.label" :value="p.key" />
          </el-select>
        </el-form-item>

        <el-form-item label="② Base URL" required>
          <el-input v-model="form.endpoint" placeholder="API 端点地址，选供应商自动填入" />
        </el-form-item>

        <el-form-item label="③ API Key" required>
          <div class="flex gap-2">
            <el-input v-model="form.api_key" class="flex-1" type="password" show-password placeholder="sk-..." />
            <el-button :loading="testingConn" @click="testConnection" :disabled="!form.endpoint || !form.api_key">
              测试连接
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="④ 模型选择" required>
          <div class="flex gap-2">
            <el-select v-model="form.modelSelect" class="flex-1" filterable allow-create
              placeholder="选择或输入模型名" @change="onModelSelect">
              <el-option v-for="m in currentModels" :key="m" :label="m" :value="m" />
              <el-option v-for="m in fetchedModels" :key="'f-'+m" :label="m" :value="m" />
              <el-option label="✎ 自定义输入..." value="__custom__" />
            </el-select>
            <el-button :loading="fetchingModels" @click="fetchModelList" :disabled="!form.endpoint || !form.api_key">
              刷新
            </el-button>
          </div>
          <el-input v-if="showCustomModel" v-model="form.customModel" class="mt-2"
            placeholder="输入模型 ID" />
        </el-form-item>

        <el-form-item label="⑤ 显示名称" required>
          <el-input v-model="form.display_name" placeholder="如 DeepSeek 主力号、千问极速版" />
          <div class="text-xs text-gray-400 mt-1">留空则自动生成「供应商 模型ID」</div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { PROVIDERS, findProvider } from '../data/providers'

const models = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const saving = ref(false)
const showCustomModel = ref(false)
const fetchedModels = ref([])
const fetchingModels = ref(false)
const testingConn = ref(false)

const form = ref({
  provider: '', endpoint: '', api_key: '', name: '',
  modelSelect: '', customModel: '', display_name: '',
})

const currentModels = computed(() => {
  const p = findProvider(form.value.provider)
  return p ? p.models : []
})

function providerLabel(key) {
  const p = findProvider(key)
  return p ? p.label : key
}

function onProviderChange(key) {
  const p = findProvider(key)
  if (p && p.baseURL) form.value.endpoint = p.baseURL
  form.value.modelSelect = ''
  form.value.name = ''
  form.value.customModel = ''
  showCustomModel.value = false
  fetchedModels.value = []
}

async function fetchModelList() {
  if (!form.value.endpoint || !form.value.api_key) return
  fetchingModels.value = true
  try {
    const data = await adminApi.fetchModels(form.value.endpoint, form.value.api_key)
    fetchedModels.value = data.models || []
    ElMessage.success(`获取到 ${fetchedModels.value.length} 个模型`)
  } catch (e) {
    ElMessage.error(e.message || '获取模型列表失败')
  } finally {
    fetchingModels.value = false
  }
}

async function testConnection() {
  if (!form.value.endpoint || !form.value.api_key) return
  testingConn.value = true
  try {
    const data = await adminApi.testConnection(form.value.endpoint, form.value.api_key)
    if (data.ok) {
      ElMessage.success(`连接成功 (${data.latency_ms}ms)`)
      // 自动拉取模型列表
      await fetchModelList()
    } else {
      ElMessage.warning(`连接失败: ${data.error || '未知错误'} (${data.latency_ms}ms)`)
    }
  } catch (e) {
    ElMessage.error(e.message || '测试连接失败')
  } finally {
    testingConn.value = false
  }
}

function onModelSelect(val) {
  if (val === '__custom__') {
    showCustomModel.value = true
    form.value.name = form.value.customModel || ''
  } else {
    showCustomModel.value = false
    form.value.name = val
    form.value.customModel = ''
    if (!form.value.display_name) {
      form.value.display_name = providerLabel(form.value.provider) + ' ' + val
    }
  }
}

onMounted(load)

async function load() {
  try {
    const data = await adminApi.models()
    models.value = (data.items || []).map(m => ({
      ...m,
      provider_label: providerLabel(m.provider),
    }))
  } catch (e) { ElMessage.error('加载失败') }
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  form.value = { provider: '', endpoint: '', api_key: '', name: '', modelSelect: '', customModel: '', display_name: '' }
  showCustomModel.value = false
  dialogVisible.value = true
}

function editModel(row) {
  isEdit.value = true
  editId.value = row.id
  const inList = currentModels.value.includes(row.name) || !currentModels.value.length
  form.value = {
    provider: row.provider,
    endpoint: row.endpoint,
    api_key: row.api_key,
    name: row.name,
    modelSelect: inList ? row.name : '__custom__',
    customModel: inList ? '' : row.name,
    display_name: row.display_name || '',
  }
  showCustomModel.value = form.value.modelSelect === '__custom__'
  dialogVisible.value = true
}

async function save() {
  if (!form.value.provider || !form.value.endpoint || !form.value.api_key) {
    ElMessage.warning('请完善供应商、Base URL 和 API Key')
    return
  }
  // Resolve model name
  const modelName = showCustomModel.value ? form.value.customModel : form.value.name
  if (!modelName) {
    ElMessage.warning('请选择或输入模型')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: modelName,
      display_name: form.value.display_name || (providerLabel(form.value.provider) + ' ' + modelName),
      provider: form.value.provider,
      endpoint: form.value.endpoint,
      api_key: form.value.api_key,
    }
    if (isEdit.value) {
      await adminApi.updateModel(editId.value, payload)
    } else {
      await adminApi.createModel(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } catch (e) { ElMessage.error(e.message) }
  finally { saving.value = false }
}

async function toggleModel(row) {
  try {
    await adminApi.toggleModel(row.id, row.is_enabled !== 1)
    ElMessage.success(row.is_enabled ? '已禁用' : '已启用')
    load()
  } catch (e) { ElMessage.error(e.message) }
}

async function setDefault(row) {
  try {
    await adminApi.setDefaultModel(row.id)
    ElMessage.success('已设为默认')
    load()
  } catch (e) { ElMessage.error(e.message) }
}

async function remove(row) {
  try {
    await ElMessageBox.confirm('确定删除该模型？', '确认')
    await adminApi.deleteModel(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) { if (e !== 'cancel') ElMessage.error(e.message) }
}
</script>
