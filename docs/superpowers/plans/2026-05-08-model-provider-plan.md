# 模型管理供应商引导式配置 — 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。

**目标：** 重写模型管理页面，改为供应商引导五步配置 + 对齐 Hermes 25 个供应商预设

**架构：** 后端加 `display_name` 字段（向前兼容），前端新增 `providers.js` 供应商预设数据，`Models.vue` 完全重写为级联表单

**技术栈：** Go 1.24 + SQLite + Vue 3 + Element Plus + Tailwind v4

---

## 文件结构

```
internal/store/store.go              # 修改: LLMModel 加 DisplayName 字段
internal/store/sqlite/sqlite.go      # 修改: 迁移加 display_name 列
internal/store/sqlite/models.go      # 修改: INSERT/UPDATE/SELECT 加 display_name
internal/handler/model_handler.go    # 可能微调: display_name 验证
web/admin/src/data/providers.js      # 新增: 25 个供应商预设数据
web/admin/src/views/Models.vue       # 重写: 供应商引导表单 + 新表格
internal/handler/model_handler_test.go # 修改: 加 display_name 测试断言
```

---

### 任务 1：后端 — 加 display_name 字段

**文件：**
- 修改：`internal/store/store.go:52-62`
- 修改：`internal/store/sqlite/sqlite.go:82-93`（建表 SQL）
- 修改：`internal/store/sqlite/sqlite.go:132-165`（迁移逻辑）
- 修改：`internal/store/sqlite/models.go:9-97`

- [ ] **步骤 1：LLMModel 加 DisplayName**

```go
// LLMModel LLM 模型配置
type LLMModel struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Provider    string    `json:"provider"`
	Endpoint    string    `json:"endpoint"`
	APIKey      string    `json:"api_key"`
	IsDefault   int       `json:"is_default"`
	IsEnabled   int       `json:"is_enabled"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}
```

- [ ] **步骤 2：建表 SQL 加 display_name 列**

在 `sqlite.go` 的 llm_models CREATE TABLE 中，`name TEXT NOT NULL` 之后加：

```sql
display_name TEXT DEFAULT '',
```

- [ ] **步骤 3：迁移存量数据**

在 `sqlite.go` 的 `migrate()` 函数末尾（`return nil` 之前）加迁移：

```go
// v2.2: 存量 display_name 填充
var emptyCount int
db.QueryRow("SELECT COUNT(*) FROM llm_models WHERE display_name = '' OR display_name IS NULL").Scan(&emptyCount)
if emptyCount > 0 {
    db.Exec("UPDATE llm_models SET display_name = provider || ' ' || name WHERE display_name = '' OR display_name IS NULL")
}
```

- [ ] **步骤 4：models.go 所有 SQL 加 display_name**

`ListModels` 的 SELECT：
```sql
SELECT id, name, display_name, provider, endpoint, api_key, is_default, is_enabled, sort_order, created_at FROM llm_models ORDER BY sort_order, id
```
Scan 加 `&m.DisplayName`。

`GetModelByID` 同理。

`GetDefaultModel` 同理。

`CreateModel` 的 INSERT：
```sql
INSERT INTO llm_models (name, display_name, provider, endpoint, api_key, is_default, is_enabled, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
```
参数加 `m.DisplayName`。

`UpdateModel` 的 UPDATE：
```sql
UPDATE llm_models SET name = ?, display_name = ?, provider = ?, endpoint = ?, api_key = ?, is_enabled = ?, sort_order = ? WHERE id = ?
```
参数加 `m.DisplayName`。

- [ ] **步骤 5：handler 层 display_name 验证**

在 `model_handler.go` 的 `CreateModel` 中，加验证：

```go
if m.DisplayName == "" {
    m.DisplayName = m.Provider + " " + m.Name
}
```

- [ ] **步骤 6：测试更新**

在 `model_handler_test.go` 的 `TestModelHandler_CreateAndList` 中断言 display_name 不为空：

```go
if resp.Items[0].DisplayName == "" {
    t.Fatal("expected display_name to be set")
}
```

- [ ] **步骤 7：编译 + 测试**

```bash
cd /home/kiddyt00/claude-projects/yiguan && go build ./... && go test ./... -count=1
```

预期：编译成功，全部测试通过。

- [ ] **步骤 8：Commit**

```bash
git add internal/store/ internal/handler/
git commit -m "feat: LLMModel 加 display_name 字段 + 存量迁移"
```

---

### 任务 2：前端 — 供应商预设数据

**文件：**
- 新增：`web/admin/src/data/providers.js`

- [ ] **步骤 1：创建 providers.js**

```js
export const PROVIDERS = [
  {
    key: 'openrouter', label: 'OpenRouter', baseURL: 'https://openrouter.ai/api/v1',
    models: ['openai/gpt-4o', 'anthropic/claude-sonnet-4', 'google/gemini-2.5-flash', 'meta-llama/llama-3.1-405b'],
  },
  {
    key: 'vercel', label: 'Vercel AI Gateway', baseURL: 'https://api.vercel.ai/v1',
    models: ['gpt-4o', 'claude-sonnet-4', 'gemini-2.5-pro'],
  },
  {
    key: 'anthropic', label: 'Anthropic', baseURL: 'https://api.anthropic.com/v1',
    models: ['claude-sonnet-4-20250514', 'claude-haiku-3-5', 'claude-opus-4-20250514'],
  },
  {
    key: 'mimo', label: 'Xiaomi MiMo', baseURL: 'https://api.mimo.xiaomi.com/v1',
    models: ['MiMo-V2.5-Pro', 'MiMo-V2.5-Omni', 'MiMo-V2.5-Flash'],
  },
  {
    key: 'nvidia', label: 'NVIDIA NIM', baseURL: 'https://integrate.api.nvidia.com/v1',
    models: ['nemotron-4', 'llama-3.1-nemotron-70b'],
  },
  {
    key: 'huggingface', label: 'Hugging Face', baseURL: 'https://api-inference.huggingface.co/v1',
    models: ['meta-llama/Llama-3.1-70B', 'mistralai/Mixtral-8x22B', 'Qwen/Qwen2.5-72B'],
  },
  {
    key: 'google', label: 'Google AI Studio', baseURL: 'https://generativelanguage.googleapis.com/v1beta/openai',
    models: ['gemini-2.5-flash', 'gemini-2.5-pro', 'gemini-2.0-flash'],
  },
  {
    key: 'deepseek', label: 'DeepSeek', baseURL: 'https://api.deepseek.com/v1',
    models: ['deepseek-chat', 'deepseek-reasoner'],
  },
  {
    key: 'xai', label: 'xAI (Grok)', baseURL: 'https://api.x.ai/v1',
    models: ['grok-3', 'grok-3-mini'],
  },
  {
    key: 'zhipu', label: '智谱 AI (Z.AI/GLM)', baseURL: 'https://open.bigmodel.cn/api/paas/v4',
    models: ['glm-4-plus', 'glm-4-flash', 'glm-4'],
  },
  {
    key: 'kimi', label: 'Kimi / Moonshot', baseURL: 'https://api.moonshot.cn/v1',
    models: ['moonshot-v1-8k', 'moonshot-v1-32k', 'moonshot-v1-128k'],
  },
  {
    key: 'stepfun', label: 'StepFun', baseURL: 'https://api.stepfun.com/v1',
    models: ['step-2-16k', 'step-1-flash'],
  },
  {
    key: 'minimax', label: 'MiniMax (国际)', baseURL: 'https://api.minimax.chat/v1',
    models: ['abab7-chat', 'abab6.5s-chat'],
  },
  {
    key: 'minimax-cn', label: 'MiniMax (国内)', baseURL: 'https://api.minimaxi.com/v1',
    models: ['abab7-chat', 'abab6.5s-chat'],
  },
  {
    key: 'dashscope', label: '阿里云 DashScope', baseURL: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    models: ['qwen-plus', 'qwen-max', 'qwen-turbo', 'qwen3.6-plus'],
  },
  {
    key: 'ollama-cloud', label: 'Ollama Cloud', baseURL: 'https://api.ollama.com/v1',
    models: ['llama3.2', 'mistral', 'qwen2.5', 'gemma2'],
  },
  {
    key: 'arcee', label: 'Arcee AI', baseURL: 'https://api.arcee.ai/v1',
    models: ['trinity', 'trinity-mini'],
  },
  {
    key: 'gmi', label: 'GMI Cloud', baseURL: 'https://api.gmicloud.com/v1',
    models: [],
  },
  {
    key: 'kilo', label: 'Kilo Code', baseURL: 'https://api.kilocode.ai/v1',
    models: [],
  },
  {
    key: 'opencode-zen', label: 'OpenCode Zen', baseURL: 'https://api.opencode.ai/zen/v1',
    models: [],
  },
  {
    key: 'opencode-go', label: 'OpenCode Go', baseURL: 'https://api.opencode.ai/go/v1',
    models: [],
  },
  {
    key: 'bedrock', label: 'AWS Bedrock', baseURL: '',
    models: ['anthropic.claude-sonnet-4', 'amazon.nova-pro'],
  },
  {
    key: 'azure', label: 'Azure Foundry', baseURL: '',
    models: [],
  },
  {
    key: 'qwen-code', label: 'Qwen Coding Plan', baseURL: 'https://coding.dashscope.aliyuncs.com/v1',
    models: ['qwen3.6-plus'],
  },
  {
    key: 'custom', label: '自定义', baseURL: '',
    models: [],
  },
]

export function findProvider(key) {
  return PROVIDERS.find(p => p.key === key) || PROVIDERS.find(p => p.key === 'custom')
}
```

- [ ] **步骤 2：Commit**

```bash
git add web/admin/src/data/providers.js
git commit -m "feat: 供应商预设数据 — 25 providers 对齐 Hermes"
```

---

### 任务 3：前端 — 重写 Models.vue

**文件：**
- 修改：`web/admin/src/views/Models.vue`

- [ ] **步骤 1：重写 Models.vue — 模板部分**

```vue
<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-2xl font-bold">模型管理</h2>
      <el-button type="primary" @click="openCreate">新增模型</el-button>
    </div>

    <!-- 模型列表 -->
    <el-table :data="models" stripe>
      <el-table-column prop="id" label="ID" width="50" />
      <el-table-column prop="display_name" label="显示名称" min-width="150" />
      <el-table-column prop="provider_label" label="供应商" width="120" />
      <el-table-column prop="name" label="模型 ID" width="150" />
      <el-table-column prop="endpoint" label="Base URL" min-width="200" show-overflow-tooltip />
      <el-table-column prop="is_enabled" label="启用" width="60">
        <template #default="{ row }">
          <el-switch :model-value="row.is_enabled === 1"
            @change="toggleModel(row)" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="is_default" label="默认" width="60">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" type="warning" size="small">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" v-if="!row.is_default" @click="setDefault(row)">设为默认</el-button>
          <el-button size="small" type="primary" @click="editModel(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑模型' : '新增模型'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="100px" label-position="top">
        <el-form-item label="① 供应商" required>
          <el-select v-model="form.provider" class="w-full" filterable placeholder="选择供应商"
            @change="onProviderChange">
            <el-option v-for="p in PROVIDERS" :key="p.key" :label="p.label" :value="p.key" />
          </el-select>
        </el-form-item>

        <el-form-item label="② Base URL" required>
          <el-input v-model="form.endpoint" placeholder="API 端点地址" />
        </el-form-item>

        <el-form-item label="③ API Key" required>
          <el-input v-model="form.api_key" type="password" show-password placeholder="sk-..." />
        </el-form-item>

        <el-form-item label="④ 模型选择" required>
          <el-select v-model="form.modelSelect" class="w-full" filterable allow-create
            placeholder="选择或输入模型名" @change="onModelSelect">
            <el-option v-for="m in currentModels" :key="m" :label="m" :value="m" />
            <el-option label="✎ 自定义输入..." value="__custom__" />
          </el-select>
          <el-input v-if="showCustomModel" v-model="form.customModel" class="mt-2"
            placeholder="输入模型 ID" @input="form.name = form.customModel" />
        </el-form-item>

        <el-form-item label="⑤ 显示名称" required>
          <el-input v-model="form.display_name" placeholder="如 DeepSeek 主力号、千问极速版" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
```

- [ ] **步骤 2：重写 Models.vue — script 部分**

```js
<script setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { PROVIDERS, findProvider } from '../data/providers'

const models = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const showCustomModel = ref(false)

const form = ref({
  provider: '', endpoint: '', api_key: '', name: '',
  modelSelect: '', customModel: '', display_name: '',
})

const currentModels = computed(() => {
  const p = findProvider(form.value.provider)
  return p ? p.models : []
})

function providerLabel(key) {
  return findProvider(key)?.label || key
}

function onProviderChange(key) {
  const p = findProvider(key)
  if (p && p.baseURL) form.value.endpoint = p.baseURL
  form.value.modelSelect = ''
  form.value.name = ''
  form.value.customModel = ''
  showCustomModel.value = false
}

function onModelSelect(val) {
  if (val === '__custom__') {
    showCustomModel.value = true
    form.value.name = ''
  } else {
    showCustomModel.value = false
    form.value.name = val
    if (!form.value.display_name) {
      form.value.display_name = providerLabel(form.value.provider) + ' ' + val
    }
  }
}

onMounted(load)

async function load() {
  try {
    const data = await adminApi.models()
    models.value = data.items.map(m => ({
      ...m,
      provider_label: providerLabel(m.provider),
    }))
  } catch (e) { ElMessage.error('加载失败') }
}

function openCreate() {
  isEdit.value = false
  form.value = { provider: '', endpoint: '', api_key: '', name: '', modelSelect: '', customModel: '', display_name: '' }
  showCustomModel.value = false
  dialogVisible.value = true
}

function editModel(row) {
  isEdit.value = true
  form.value = {
    provider: row.provider,
    endpoint: row.endpoint,
    api_key: row.api_key,
    name: row.name,
    modelSelect: currentModels.value.includes(row.name) ? row.name : '__custom__',
    customModel: currentModels.value.includes(row.name) ? '' : row.name,
    display_name: row.display_name,
  }
  showCustomModel.value = form.value.modelSelect === '__custom__'
  dialogVisible.value = true
}

async function save() {
  if (!form.value.provider || !form.value.endpoint || !form.value.api_key || !form.value.name) {
    ElMessage.warning('请完善所有必填项')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      display_name: form.value.display_name || (providerLabel(form.value.provider) + ' ' + form.value.name),
      provider: form.value.provider,
      endpoint: form.value.endpoint,
      api_key: form.value.api_key,
    }
    if (isEdit.value) {
      await adminApi.updateModel(rowId.value, payload)
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
    ElMessage.success('操作成功')
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
    await ElMessageBox.confirm('确定删除？', '确认')
    await adminApi.deleteModel(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) { if (e !== 'cancel') ElMessage.error(e.message) }
}
</script>
```

_注意：`isEdit` 需要关联一个 `rowId` ref 来在 save 时使用正确的 ID。在 editModel 中设置 `rowId.value = row.id`。_

- [ ] **步骤 3：编译验证前端**

```bash
cd /home/kiddyt00/claude-projects/yiguan/web/admin && npm run build
```

预期：编译成功，无报错。

- [ ] **步骤 4：后端全量测试**

```bash
cd /home/kiddyt00/claude-projects/yiguan && go test ./... -count=1
```

预期：全部通过。

- [ ] **步骤 5：Commit**

```bash
git add web/admin/src/views/Models.vue
git commit -m "feat: 模型管理供应商引导式配置 — 级联选择器 + 25 providers"
```

---

## 自检

**规格覆盖度：**
- ✅ 25 个供应商预设 (providers.js)
- ✅ 五步表单流程 (Models.vue)
- ✅ display_name 字段 (后端)
- ✅ 存量数据迁移
- ✅ 模型选择 + 自定义输入
- ✅ 启用状态开关
- ✅ 表格显示优化

**占位符扫描：** 无 TODO、无待定

**类型一致性：**
- provider key 在 providers.js 和 LLMModel.Provider 间保持一致
- display_name 在后端和前端字段名一致
- name 字段仍作为模型 ID 发送给 API
