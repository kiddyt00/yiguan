# 易观 · Yi Guan

> 观乎天文，以察时变；观乎人文，以化成天下。——《周易·贲卦》

易观是一个 AI 驱动的周易占筮平台，以技术之力梳理古典智慧，提供铜钱起卦、大模型解卦、历史记录等能力。支持多渠道登录、多模型热切换、流式响应，并内置广告变现体系。

## 技术栈

| 层 | 技术 | 说明 |
|---|---|---|
| 后端 | Go 1.24 | 标准库 `net/http`，无第三方框架 |
| 数据库 | SQLite (modernc.org/sqlite) | 纯 Go 实现，零 CGO 依赖 |
| 前端 | Vue 3 + Vite + Pinia | SPA，Tailwind CSS v4 中国风主题 |
| 管理后台 | Vue 3 + Vite | 独立构建，管理员专用 |
| 小程序 | UniApp (Vue 3) | 微信小程序端 |
| AI | 多模型兼容 | 千问 / DeepSeek / 任意 OpenAI 兼容接口 |
| 部署 | Docker Compose | 前后端分离容器，Nginx 反代 |

## 特性

- **铜钱起卦** — 模拟传统三钱六爻起卦法，55 法定主变爻，输出本卦、变卦及变爻位置
- **AI 智能解卦** — 流式 SSE 响应，逐字输出解卦内容，支持多模型热切换和容错链
- **用户体系** — 手机号 / 微信小程序 / 微信扫码三种登录方式，JWT 鉴权
- **Quota 计费** — 注册赠送 3 次免费算卦，后续支持付费购买、转发裂变、广告解锁
- **模型管理** — 后台可视化配置大模型，25 家内置供应商，一键测试连接，动态切换默认模型
- **广告系统** — 首页 + 结果页广告位，观看广告解锁 quota（每日限 3 次），后台管理广告上下架与统计
- **管理后台** — 用户管理、卦象记录、模型配置、广告运营，一站式 Dashboard
- **翻译** — AI 解读中英文互译，三层缓存（内存 → localStorage → 后端 DB）
- **优雅关闭** — 信号监听，HTTP Server 超时控制，健康检查端点

## 快速开始

### 环境要求

- Go >= 1.24
- Node.js >= 18
- Docker & Docker Compose（生产部署）

### 本地开发

```bash
# 克隆项目
git clone git@github.com:kiddyt00/yiguan.git
cd yiguan

# 启动后端（默认端口 8080）
make dev-backend

# 启动前端（另一个终端）
make dev-frontend

# 启动管理后台（可选）
make dev-admin
```

访问 http://localhost:8080 使用后端 API，前端开发服务器见 Vite 输出。

### 运行测试

```bash
make test          # 全部测试
make test-short    # 跳过慢速测试
```

### Docker 部署

```bash
# 配置环境变量（复制 .env.example 或直接 export）
export ADMIN_PHONE=13800000000
export JWT_SECRET=your-secret
export LLM_API_KEY=sk-xxx
export WX_APPID=wx9e87b7216be83619
export WX_SECRET=your-wechat-secret

# 一键构建并启动
make deploy

# 查看日志
make docker-logs

# 停止
make docker-down
```

服务启动后访问 `http://localhost:80`（通过 `HTTP_PORT` 环境变量修改）。

## 项目结构

```
yiguan/
├── cmd/server/main.go          # 程序入口，路由注册，中间件
├── internal/
│   ├── engine/                 # 算卦引擎（铜钱起卦、卦象数据）
│   ├── handler/                # HTTP 处理器（鉴权、算卦、用户、管理、广告、模型、翻译）
│   ├── llm/                    # LLM 客户端、路由器、容错链、流式输出
│   ├── middleware/             # JWT 鉴权、管理员鉴权
│   └── store/
│       └── sqlite/             # SQLite 数据访问层
├── web/
│   ├── front/                  # 用户端 Vue 3 SPA
│   └── admin/                  # 管理后台 Vue 3 SPA
├── miniapp/                    # UniApp 微信小程序
│   ├── pages/                  # 8 个页面（首页/起卦/登录/个人/历史/广告/关于/密码登录）
│   ├── utils/                  # API 封装、配置、Markdown 渲染
│   └── static/                 # 静态资源
├── test/                       # E2E 测试
├── deploy/                     # 部署配置（nginx、二进制）
├── config.yaml                 # 本地开发配置
├── docker-compose.yml          # Docker 编排
├── Dockerfile.backend          # 后端镜像
├── Dockerfile.frontend         # 前端 Nginx 镜像
└── Makefile                    # 构建、测试、部署命令
```

## API 概览

### 用户端

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| POST | `/api/auth/register` | 手机号注册 | 否 |
| POST | `/api/auth/login` | 手机号登录 | 否 |
| POST | `/api/auth/wechat-login` | 微信小程序登录 | 否 |
| GET | `/api/auth/wechat-qrcode` | 微信扫码登录二维码 | 否 |
| GET | `/api/auth/wechat-check` | 扫码状态轮询 | 否 |
| POST | `/api/auth/sms-send` | 发送短信验证码 | 否 |
| POST | `/api/auth/sms-login` | 短信验证码登录 | 否 |
| GET | `/api/user` | 获取用户信息 | Bearer Token |
| PUT | `/api/user` | 更新用户信息 | Bearer Token |
| POST | `/api/divine` | 起卦并解卦 | Bearer Token |
| POST | `/api/divine/stream` | 流式起卦解卦 (SSE) | Bearer Token |
| GET | `/api/history` | 历史算卦记录 | Bearer Token |
| GET/POST | `/api/history/{id}/translate` | 翻译 AI 解读 | Bearer Token |
| GET | `/api/ads/active` | 获取启用广告 | 否 |
| POST | `/api/ads/{id}/watch` | 开始观看广告 | Bearer Token |
| POST | `/api/ads/{id}/complete` | 广告观看完成 | Bearer Token |

### 管理后台

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/dashboard` | 仪表盘数据 |
| GET | `/api/admin/users` | 用户列表 |
| POST | `/api/admin/users/{id}/toggle` | 启用/禁用用户 |
| POST | `/api/admin/users/{id}/quota` | 调整用户配额 |
| GET | `/api/admin/users/{id}/history` | 用户算卦记录 |
| GET | `/api/admin/hexagrams` | 卦象记录列表 |
| DELETE | `/api/admin/hexagrams/{id}` | 删除卦象记录 |
| GET | `/api/admin/models` | 模型列表 |
| POST | `/api/admin/models` | 添加模型 |
| PUT | `/api/admin/models/{id}` | 更新模型 |
| DELETE | `/api/admin/models/{id}` | 删除模型 |
| POST | `/api/admin/models/{id}/set-default` | 设为默认模型 |
| POST | `/api/admin/models/{id}/toggle` | 启用/禁用模型 |
| POST | `/api/admin/models/fetch` | 从供应商拉取模型列表 |
| POST | `/api/admin/models/test` | 测试模型连接 |
| GET | `/api/admin/ads` | 广告列表 |
| POST | `/api/admin/ads` | 创建广告 |
| PUT | `/api/admin/ads/{id}` | 更新广告 |
| DELETE | `/api/admin/ads/{id}` | 删除广告 |
| POST | `/api/admin/ads/{id}/toggle` | 启用/禁用广告 |
| GET | `/api/admin/ads/stats` | 广告统计数据 |

## 微信小程序

小程序端使用 UniApp 开发，支持以下功能：

- **微信一键登录** — 调用 wx.login + 后端 code2session
- **短信验证码登录** — 手机号+验证码
- **密码登录/注册** — 传统手机号+密码
- **流式起卦** — SSE 推演动画，3 阶段展示（铜钱 → 卦象 → AI 解读）
- **历史记录** — 分页列表 + Markdown 渲染解卦内容
- **广告解锁** — 看 30s 广告获取 1 次起卦（每日限 3 次）
- **分享** — 支持转发好友 / 朋友圈

使用 HBuilder X 打开 `miniapp/` 目录，运行到微信开发者工具即可调试。

**小程序 AppID:** `wx9e87b7216be83619`

## 配置

`config.yaml`（本地开发）：

```yaml
server:
  port: 8080
jwt_secret: "change-me-in-production"
db_path: "yiguan.db"

admin:
  phone: "13800000000"
  password: "admin123"

llm:
  default: "qwen"
  providers:
    qwen:
      api_key: "sk-xxx"
      endpoint: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
      model: "qwen-plus"
    deepseek:
      api_key: ""
      endpoint: "https://api.deepseek.com/v1/chat/completions"
      model: "deepseek-chat"
```

生产环境通过环境变量覆盖敏感配置：

| 变量 | 说明 |
|---|---|
| `SERVER_PORT` | 服务端口 |
| `JWT_SECRET` | JWT 签名密钥 |
| `LLM_API_KEY` | 大模型 API Key |
| `DB_PATH` | 数据库文件路径 |
| `ADMIN_PHONE` | 管理员手机号 |
| `ADMIN_PASSWORD` | 管理员密码 |
| `WX_APPID` / `WX_SECRET` | 微信小程序凭证 |
| `WX_OPEN_APPID` / `WX_OPEN_SECRET` | 微信开放平台凭证（扫码登录） |
| `HTTP_PORT` | 前端 Nginx 对外端口（默认 80） |

## 算卦逻辑

1. **起爻** — 模拟 3 枚铜钱抛掷 6 次，得 6 个爻值（6/7/8/9）
2. **变卦** — 6（老阴）和 9（老阳）为变爻，自动生成变卦
3. **55 法定主变爻** — 6 次抛掷总值减 55 取模，指向最关键的变爻
4. **AI 解读** — 将本卦、变卦、变爻信息构造成 prompt，调用大模型生成结构化解卦
5. **容错链** — 默认模型失败时自动切换到下一个已启用的模型

## 生产部署

当前生产环境地址：**https://gjz.shadouyou.cloud**

部署流程：
1. 配置环境变量（`.env` 或 export）
2. `make deploy` — Docker Compose 一键构建启动
3. Nginx 反向代理 80/443 → 前端容器（8080）
4. SSL 证书通过 acme.sh 自动管理
