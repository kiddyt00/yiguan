# 模型管理 v2.2 — 供应商引导式配置

> 设计日期：2026-05-08

## 目标

将模型管理从四个散装文本框改为供应商引导的五步配置流程，对齐 Hermes Agent 的供应商列表。

## 供应商预设（25 个）

| key | 显示名 | 默认 Base URL | 预设模型 |
|-----|--------|---------------|---------|
| `openrouter` | OpenRouter | `https://openrouter.ai/api/v1` | openai/gpt-4o, anthropic/claude-sonnet-4 |
| `vercel` | Vercel AI Gateway | `https://api.vercel.ai/v1` | gpt-4o, claude-sonnet-4 |
| `anthropic` | Anthropic | `https://api.anthropic.com/v1` | claude-sonnet-4-20250514, claude-haiku-3-5 |
| `mimo` | Xiaomi MiMo | `https://api.mimo.xiaomi.com/v1` | MiMo-V2.5-Pro, MiMo-V2.5-Omni |
| `nvidia` | NVIDIA NIM | `https://integrate.api.nvidia.com/v1` | nemotron-4, llama-3.1-nemotron |
| `huggingface` | Hugging Face | `https://api-inference.huggingface.co/v1` | meta-llama/Llama-3.1-70B, mistralai/Mixtral-8x22B |
| `google` | Google AI Studio | `https://generativelanguage.googleapis.com/v1beta/openai` | gemini-2.5-flash, gemini-2.5-pro |
| `deepseek` | DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat, deepseek-reasoner |
| `xai` | xAI | `https://api.x.ai/v1` | grok-3, grok-3-mini |
| `zhipu` | 智谱 AI | `https://open.bigmodel.cn/api/paas/v4` | glm-4-plus, glm-4-flash |
| `kimi` | Kimi / Moonshot | `https://api.moonshot.cn/v1` | moonshot-v1-8k, moonshot-v1-32k |
| `stepfun` | StepFun | `https://api.stepfun.com/v1` | step-2-16k, step-1-flash |
| `minimax` | MiniMax (国际) | `https://api.minimax.chat/v1` | abab7-chat, abab6.5s-chat |
| `minimax-cn` | MiniMax (国内) | `https://api.minimaxi.com/v1` | abab7-chat, abab6.5s-chat |
| `dashscope` | 阿里云 DashScope | `https://dashscope.aliyuncs.com/compatible-mode/v1` | qwen-plus, qwen-max, qwen-turbo |
| `ollama-cloud` | Ollama Cloud | `https://api.ollama.com/v1` | llama3.2, mistral, qwen2.5 |
| `arcee` | Arcee AI | `https://api.arcee.ai/v1` | trinity, trinity-mini |
| `gmi` | GMI Cloud | `https://api.gmicloud.com/v1` | (手动输入) |
| `kilo` | Kilo Code | `https://api.kilocode.ai/v1` | (手动输入) |
| `opencode-zen` | OpenCode Zen | `https://api.opencode.ai/zen/v1` | (手动输入) |
| `opencode-go` | OpenCode Go | `https://api.opencode.ai/go/v1` | (手动输入) |
| `bedrock` | AWS Bedrock | (手动输入) | claude-sonnet-4, nova-pro |
| `azure` | Azure Foundry | (手动输入) | (手动输入) |
| `qwen-code` | Qwen Coding Plan | `https://coding.dashscope.aliyuncs.com/v1` | qwen3.6-plus |
| `custom` | 自定义 | (手动输入) | (手动输入) |

## 表单流程

```
① 供应商 → 下拉选择，带搜索
② Base URL → 选供应商自动填，可手动修改
③ API Key → 密码输入框
④ 模型选择 → 供应商预设列表 + "✎ 自定义输入..."选项
⑤ 显示名称 → 如"DeepSeek 主力号"
⑥ 启用开关
```

## 数据库变更

- `llm_models` 新增 `display_name TEXT DEFAULT ''`
- 迁移：`UPDATE llm_models SET display_name = provider || ' ' || name WHERE display_name = '' OR display_name IS NULL`

## 表格列调整

```
ID │ 显示名称 │ 供应商 │ 模型 ID │ Base URL │ 状态 │ 默认 │ 操作
```

## 涉及文件

| 层 | 文件 | 变更 |
|----|------|------|
| 数据库 | `internal/store/sqlite/sqlite.go` | 加 display_name 列迁移 |
| 模型 | `internal/store/store.go` | LLMModel 加 DisplayName 字段 |
| 模型 CRUD | `internal/store/sqlite/models.go` | INSERT/UPDATE/SELECT 加 display_name |
| 前端 | `web/admin/src/views/Models.vue` | 完全重写表单和表格 |
| 前端 API | `web/admin/src/api/index.js` | 无需改动（字段自动序列化） |
| 后端 handler | `internal/handler/model_handler.go` | 无需改动（JSON 自动绑定新字段） |
