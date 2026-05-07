<template>
  <div>
    <h2 class="text-2xl font-bold mb-4 flex justify-between items-center">
      模型管理
      <el-button type="primary" @click="showCreate = true">新增模型</el-button>
    </h2>
    <el-table :data="models" stripe>
      <el-table-column prop="id" label="ID" width="50" />
      <el-table-column prop="name" label="模型名" width="120" />
      <el-table-column prop="provider" label="提供商" width="100" />
      <el-table-column prop="endpoint" label="Endpoint" min-width="200" show-overflow-tooltip />
      <el-table-column prop="is_default" label="默认" width="60">
        <template #default="{ row }"><el-tag v-if="row.is_default" type="warning">默认</el-tag></template>
      </el-table-column>
      <el-table-column prop="is_enabled" label="状态" width="60">
        <template #default="{ row }">
          <el-switch v-model="row.is_enabled" :active-value="1" :inactive-value="0"
            @change="toggleModel(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" v-if="!row.is_default" @click="setDefault(row)">设为默认</el-button>
          <el-button size="small" type="primary" @click="editModel(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" :title="editTarget ? '编辑模型' : '新增模型'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="模型名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="提供商"><el-input v-model="form.provider" /></el-form-item>
        <el-form-item label="Endpoint"><el-input v-model="form.endpoint" /></el-form-item>
        <el-form-item label="API Key"><el-input v-model="form.api_key" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const models = ref([])
const showCreate = ref(false)
const editTarget = ref(null)
const form = ref({ name: '', provider: '', endpoint: '', api_key: '' })

onMounted(load)

async function load() {
  try {
    const data = await adminApi.models()
    models.value = data.items
  } catch (e) {
    ElMessage.error('加载失败')
  }
}

async function toggleModel(row) {
  try {
    await adminApi.toggleModel(row.id, row.is_enabled === 1)
    ElMessage.success('操作成功')
  } catch (e) {
    ElMessage.error(e.message)
    row.is_enabled = row.is_enabled ? 0 : 1
  }
}

async function setDefault(row) {
  try {
    await adminApi.setDefaultModel(row.id)
    ElMessage.success('已设为默认')
    load()
  } catch (e) {
    ElMessage.error(e.message)
  }
}

function editModel(row) {
  editTarget.value = row
  form.value = { name: row.name, provider: row.provider, endpoint: row.endpoint, api_key: row.api_key }
  showCreate.value = true
}

async function save() {
  try {
    if (editTarget.value) {
      await adminApi.updateModel(editTarget.value.id, form.value)
    } else {
      await adminApi.createModel(form.value)
    }
    ElMessage.success('保存成功')
    showCreate.value = false
    editTarget.value = null
    load()
  } catch (e) {
    ElMessage.error(e.message)
  }
}

async function remove(row) {
  try {
    await ElMessageBox.confirm('确定删除？', '确认')
    await adminApi.deleteModel(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message)
  }
}
</script>
