# 易观 v2.0 全栈增强设计

> 日期: 2026-05-08
> 状态: Approved
> 涉及: 数据库 / 后端 API / 前台前端 / 后台前端

---

## 1. 数据库变更

### 1.1 users 表新增字段

```sql
ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user';
ALTER TABLE users ADD COLUMN is_active INTEGER DEFAULT 1;
```

- `role`: `'user'` / `'admin'`
- `is_active`: 0=禁用（无法登录和起卦），1=启用

### 1.2 新增 llm_models 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| name | TEXT NOT NULL | 模型名，如 qwen-plus |
| provider | TEXT NOT NULL | 提供商标识（qwen/deepseek） |
| endpoint | TEXT NOT NULL | API 地址 |
| api_key | TEXT NOT NULL | API key 明文存储 |
| is_default | INTEGER DEFAULT 0 | 是否默认模型 |
| is_enabled | INTEGER DEFAULT 1 | 是否启用 |
| sort_order | INTEGER DEFAULT 0 | 排序 |
| created_at | DATETIME DEFAULT CURRENT_TIMESTAMP | 创建时间 |

### 1.3 新增 ads 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| name | TEXT NOT NULL | 广告名称 |
| description | TEXT DEFAULT '' | 描述 |
| ad_type | TEXT DEFAULT 'iframe' | iframe / redirect |
| content_url | TEXT NOT NULL | 广告展示 URL |
| watch_duration | INTEGER NOT NULL DEFAULT 30 | 需停留秒数 |
| reward_quota | INTEGER NOT NULL DEFAULT 1 | 奖励起卦次数 |
| is_enabled | INTEGER DEFAULT 1 | 开关 |
| sort_order | INTEGER DEFAULT 0 | 排序 |
| created_at | DATETIME DEFAULT CURRENT_TIMESTAMP | 创建时间 |

### 1.4 新增 ad_records 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| user_id | INTEGER NOT NULL | FK → users(id) |
| ad_id | INTEGER NOT NULL | FK → ads(id) |
| watch_duration | INTEGER DEFAULT 0 | 实际观看秒数 |
| status | TEXT DEFAULT 'watching' | watching / completed / abandoned |
| rewarded | INTEGER DEFAULT 0 | 是否已发放奖励 |
| created_at | DATETIME DEFAULT CURRENT_TIMESTAMP | 观看时间 |

---

## 2. 后端架构

### 2.1 文件结构

```
cmd/server/main.go          入口，增加管理员初始化 + SSE 路由
internal/
├── handler/
│   ├── admin.go            增强：Dashboard + Users CRUD + toggle/quota/history
│   ├── auth.go             增强：JWT payload 加入 role
│   ├── divine.go           现有，不变
│   ├── divine_stream.go    新增：SSE 流式起卦
│   ├── model_handler.go    新增：模型 CRUD
│   ├── ad_handler.go       新增：广告 CRUD + 用户端广告接口
│   ├── hexagram_handler.go 新增：起卦记录管理（分页/详情/删除）
│   └── helpers.go          现有，不变
├── llm/
│   ├── client.go           现有 SSE 支持
│   └── router.go           新增：模型热切换路由
├── store/
│   ├── store.go            接口扩展（新方法）
│   └── sqlite/
│       ├── sqlite.go       迁移脚本增加新表/字段
│       ├── users.go        增加 role/is_active 字段查询
│       ├── models.go       新增：llm_models CRUD
│       ├── ads.go          新增：ads/ad_records CRUD
│       └── hexagrams.go    新增：管理员查询全量记录
```

### 2.2 管理员初始化

启动时读取环境变量或 `config.yaml`：

```yaml
admin:
  phone: "13800000000"
  password: "admin123"
```

启动逻辑：
1. 查找该手机号用户，存在则更新 role 为 `admin`
2. 不存在则创建，role 为 `admin`
3. 未配置则跳过（需手动改数据库）

### 2.3 认证中间件增强

```go
// middleware/auth.go
func AuthRequired(secret string) func(http.Handler) http.Handler {
    // 现有逻辑不变
}

func AdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w, r) {
        // 从 JWT context 中取出 role
        if role != "admin" {
            writeJSON(w, 403, {"error": "需要管理员权限"})
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### 2.4 模型热切换

```go
// internal/llm/router.go
type ModelRouter struct {
    mu     sync.RWMutex
    store  store.Store
    current *llm.Client  // 当前默认模型客户端
}

func (r *ModelRouter) Get() *llm.Client {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.current
}

func (r *ModelRouter) Reload() error {
    // 从数据库读取 is_default=true 的模型
    // 构建新的 llm.Client
    // 原子替换 r.current
}
```

Admin 修改模型配置后调用 `r.Reload()`，无需重启服务。

### 2.5 SSE 流式协议

**端点**: `POST /api/divine/stream`

**请求**: `{"question": "..."}`

**响应**: `text/event-stream`

| 事件名 | 数据结构 | 说明 |
|--------|---------|------|
| `phase` | `{"phase":"coins","data":{"throw":1,"result":"yin"}}` | 铜钱抛掷，共 6 次 |
| `phase` | `{"phase":"hexagram","data":{"primary_gua":"乾","changing_gua":"姤","yao_positions":"011111"}}` | 卦象计算完成 |
| `ai` | `{"chunk":"根据卦象显示，"}` | AI 解卦文本块，逐 chunk 推送 |
| `done` | `{"id":42,"interpretation":"完整文本","created_at":"..."}` | 全部完成 |
| `error` | `{"error":"描述"}` | 任意阶段出错 |

**流程**:
```
客户端 POST → 200 + text/event-stream
  → coins × 6（铜钱动画）
  → hexagram（显示卦象）
  → ai × N（AI 解卦打字机效果）
  → save to history DB
  → done（完整结果）
  → 连接关闭
```

### 2.6 Admin API 路由

```
POST   /api/admin/users/:id/toggle           启用/禁用
PUT    /api/admin/users/:id                  编辑
POST   /api/admin/users/:id/quota            调整配额 {"delta": 5}
GET    /api/admin/users/:id/history          用户起卦记录

GET    /api/admin/hexagrams                  全量起卦记录（分页+筛选）
GET    /api/admin/hexagrams/:id              单条详情
DELETE /api/admin/hexagrams/:id              删除

GET    /api/admin/models                     模型列表
POST   /api/admin/models                     新增模型
PUT    /api/admin/models/:id                 编辑模型
DELETE /api/admin/models/:id                 删除
POST   /api/admin/models/:id/set-default     设为默认
POST   /api/admin/models/:id/toggle          启用/禁用

GET    /api/admin/ads                        广告列表
POST   /api/admin/ads                        新增广告
PUT    /api/admin/ads/:id                    编辑广告
DELETE /api/admin/ads/:id                    删除广告
POST   /api/admin/ads/:id/toggle             启用/禁用
GET    /api/admin/ads/stats                  播放统计

GET    /api/admin/dashboard                  增强版仪表盘
```

### 2.7 用户端新增 API

```
GET    /api/ads/active                       获取启用的广告列表
POST   /api/ads/:id/watch                    开始观看
POST   /api/ads/:id/complete                 完成观看 {"duration": 30}
POST   /api/divine/stream                    流式起卦（SSE）
```

### 2.8 错误处理

| 场景 | HTTP 状态 | 行为 |
|------|-----------|------|
| 非管理员访问后台 | 403 | 返回权限错误 |
| 用户被禁用 | 403 | 登录时拒绝 |
| 无起卦配额 | 402 | 返回配额不足，引导看广告 |
| 广告时长不足 | 400 | 不发放配额 |
| LLM API 超时 | SSE error | 推送错误事件，history 标记 interpretation="" |
| 删除最后一个启用模型 | 400 | 拒绝，至少保留一个 |

---

## 3. 前端设计

### 3.1 后台管理面板（admin）

**UI 框架**: Element Plus

**布局**:

```
┌──────────────────────────────────────────┐
│ Logo    ☯ 易观后台管理          [用户]   │
├────────┬─────────────────────────────────┤
│ 📊 仪表盘 │                             │
│ 👥 用户管理 │    内容区域                │
│ 🔮 卦象任务 │    (列表/表单/详情)        │
│ 🤖 模型管理 │                            │
│ 📢 广告管理 │                            │
└────────┴─────────────────────────────────┘
```

**文件结构**:
```
web/admin/src/
├── layout/
│   └── AdminLayout.vue       侧边栏 + 顶栏 + 内容区
├── views/
│   ├── Login.vue             后台登录页
│   ├── Dashboard.vue         数据卡片 + 统计
│   ├── Users.vue             用户列表 + 编辑弹窗 + 配额操作
│   ├── Hexagrams.vue         起卦记录列表 + 详情
│   ├── Models.vue            模型 CRUD
│   └── Ads.vue               广告 CRUD + 播放统计
├── stores/
│   └── admin.js              状态管理
├── router/
│   └── index.js              路由守卫：role=admin
└── api/
    └── index.js              API 请求封装
```

**路由守卫**:
```js
router.beforeEach((to, from, next) => {
  const user = authStore.currentUser;
  if (to.meta.requiresAdmin && user?.role !== 'admin') {
    next('/login');
  } else {
    next();
  }
});
```

### 3.2 前台用户端（front）

**UI 框架**: 保持现有 Tailwind CSS

**新增/修改页面**:

1. **流式起卦** — 修改 `views/Result.vue`
   - 使用 `EventSource` / `fetch` + `ReadableStream` 接收 SSE
   - 铜钱动画阶段：接收 `phase/coins` 逐爻播放
   - 卦象展示：接收 `phase/hexagram` 渲染卦象图
   - AI 解卦：接收 `ai` 事件，逐字打字机效果
   - 完成：接收 `done`，保存记录

2. **广告页面** — 新增 `views/AdCenter.vue`
   - 展示启用广告卡片列表
   - 点击弹出模态框，iframe 加载广告
   - 倒计时组件（"还需观看 N 秒"）
   - 倒计时结束后按钮变为"领取次数"

3. **个人中心** — 修改 `views/Profile.vue`
   - 显示当前可用起卦次数
   - "看广告领次数"入口跳转 AdCenter

---

## 4. Nginx 配置

确保 SSE 不被 nginx 缓冲：

```nginx
location /api/divine/stream {
    proxy_pass http://backend;
    proxy_buffering off;
    proxy_cache off;
    proxy_read_timeout 120s;
    # SSE headers
    add_header X-Accel-Buffering no;
}
```

---

## 5. 部署

管理员账号通过 `.env` 或 `config.yaml` 配置：

```yaml
admin:
  phone: "13800000000"
  password: "admin123"
```

或环境变量：
```env
ADMIN_PHONE=13800000000
ADMIN_PASSWORD=admin123
```

---

## 6. 验收标准

1. [x] 管理员初始化：启动后 admin 用户存在且 role='admin'
2. [x] 非管理员访问 `/api/admin/*` 返回 403
3. [x] `/api/admin/dashboard` 返回完整统计数据
4. [x] 用户管理：列表/编辑/禁用/配额调整/查看记录
5. [x] 卦象管理：列表/详情/删除
6. [x] 模型管理：CRUD + 设为默认 + 热切换
7. [x] 广告管理：CRUD + 启用/禁用 + 播放统计
8. [x] 用户端：`/api/ads/active` + watch + complete 完整流程
9. [x] SSE 流式起卦：coins → hexagram → ai → done 全链路
10. [x] 前台"看广告领次数"功能可正常使用 (API 层面验证通过)
11. [x] SSE 连接不被 nginx 缓冲 (事件流实时推送)
12. [x] 后台登录页拦截非管理员 (401 无token / 403 非admin)
