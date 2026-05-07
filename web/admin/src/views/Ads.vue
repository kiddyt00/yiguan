<template>
  <div>
    <h2 class="text-2xl font-bold mb-4 flex justify-between items-center">
      广告管理
      <el-button type="primary" @click="showCreate = true">新增广告</el-button>
    </h2>
    <el-table :data="ads" stripe>
      <el-table-column prop="id" label="ID" width="50" />
      <el-table-column prop="name" label="名称" width="120" />
      <el-table-column prop="content_url" label="URL" min-width="200" show-overflow-tooltip />
      <el-table-column prop="watch_duration" label="时长(秒)" width="90" />
      <el-table-column prop="reward_quota" label="奖励次数" width="90" />
      <el-table-column prop="is_enabled" label="状态" width="60">
        <template #default="{ row }">
          <el-switch v-model="row.is_enabled" :active-value="1" :inactive-value="0"
            @change="toggleAd(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="editAd(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-divider>广告播放统计</el-divider>
    <el-table :data="stats" stripe>
      <el-table-column prop="ad_name" label="广告" width="120" />
      <el-table-column prop="total" label="总播放" width="80" />
      <el-table-column prop="completed" label="完成" width="80" />
      <el-table-column prop="reward_total" label="发放奖励" width="80" />
    </el-table>

    <el-dialog v-model="showCreate" :title="editTarget ? '编辑广告' : '新增广告'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" /></el-form-item>
        <el-form-item label="广告URL"><el-input v-model="form.content_url" /></el-form-item>
        <el-form-item label="观看时长"><el-input-number v-model="form.watch_duration" :min="5" :max="300" /></el-form-item>
        <el-form-item label="奖励次数"><el-input-number v-model="form.reward_quota" :min="1" :max="10" /></el-form-item>
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

const ads = ref([])
const stats = ref([])
const showCreate = ref(false)
const editTarget = ref(null)
const form = ref({ name: '', description: '', content_url: '', watch_duration: 30, reward_quota: 1 })

onMounted(load)

async function load() {
  try {
    const data = await adminApi.ads()
    ads.value = data.items
  } catch (e) { ElMessage.error('加载失败') }
  try {
    const data = await adminApi.adStats()
    stats.value = data.items
  } catch (e) {}
}

async function toggleAd(row) {
  try {
    await adminApi.toggleAd(row.id, row.is_enabled === 1)
    ElMessage.success('操作成功')
  } catch (e) {
    row.is_enabled = row.is_enabled ? 0 : 1
  }
}

function editAd(row) {
  editTarget.value = row
  form.value = { name: row.name, description: row.description, content_url: row.content_url, watch_duration: row.watch_duration, reward_quota: row.reward_quota }
  showCreate.value = true
}

async function save() {
  try {
    if (editTarget.value) {
      await adminApi.updateAd(editTarget.value.id, form.value)
    } else {
      await adminApi.createAd(form.value)
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
    await adminApi.deleteAd(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message)
  }
}
</script>
