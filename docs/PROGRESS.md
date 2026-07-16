# 易观 (Yi Guan) — 开发进度记录

> 最后更新：2026-07-16
> 用于后续 Agent 快速了解项目状态

---

## 一、项目概览

AI 驱动的周易占筮平台。Go 后端 + Vue 3 SPA 前端，Docker 部署。

| 层 | 技术 |
|---|---|
| 后端 | Go 1.24，标准库 `net/http` |
| 前端 | Vue 3 + Vite + Pinia，Tailwind CSS v4 |
| 管理后台 | Vue 3 + Vite（独立构建） |
| 小程序 | UniApp (Vue 3) + 原生版 (miniapp-native) |
| 数据库 | SQLite (modernc.org/sqlite) |
| AI | 千问 / DeepSeek / 任意 OpenAI 兼容 |
| 部署 | Docker Compose，主机 Nginx 反代 |

---

## 二、生产环境

| 项目 | 值 |
|---|---|
| **服务器** | 腾讯云 `124.223.16.159`（主机名: VM-0-13-ubuntu） |
| **OS** | Ubuntu 24.04.4 LTS |
| **SSH** | `root / Jason1987!@#` |
| **生产域名** | `https://zgjz.insightj.cn` |
| **管理后台** | `https://zgjz.insightj.cn/admin` |
| **旧域名** | `gjz.shadouyou.cloud` → 已不再使用（DNS 仍指向老服务器 49.235.108.61） |
| **SSL 证书** | Let's Encrypt YE2，到期 2026-09-14 |

### 部署架构

```
用户 → https://zgjz.insightj.cn:443
    → 主机 Nginx (SSL卸载 + SSE no-buffer)
        → Docker: yiguan-frontend (Nginx 容器内 80)
            → Docker: yiguan-backend (Go 8080 /health)
```

### 环境变量

通过 `/root/yiguan/.env` 配置：

```bash
HTTP_PORT=8080
JWT_SECRET=yiguan-prod-secret-20260601
LLM_API_KEY=sk-b8bd0f8077764c59b948c503cf1ee5f7
ADMIN_PHONE=13800000000
ADMIN_PASSWORD=admin123
WX_OPEN_APPID=wx4f153ab5be9e4723          # 微信开放平台（网站扫码登录）
WX_OPEN_SECRET=406484c5cd361f7a2c74c8b97d1e3ec2
```

---

## 三、开发时间线

### 2026-07-16 — 微信扫码登录上线

#### 背景
项目自 2026-06-01 部署后处于停滞状态。生产服务器未配置任何微信环境变量，微信扫码登录不可用。

#### 完成事项

| # | 事项 | 文件 | 说明 |
|---|---|---|---|
| 1 | 修复回调地址硬编码 | `auth.go:315` | `49.235.108.61` → `zgjz.insightj.cn` |
| 2 | docker-compose 加 WX_OPEN 变量 | `docker-compose.yml` | 传递 `WX_OPEN_APPID` / `WX_OPEN_SECRET` |
| 3 | 生产配置 WX_OPEN | 服务器 `.env` | 写入 AppID 和 Secret |
| 4 | QR码依赖外部服务 | `Login.vue` | 移除 `api.qrserver.com`，改用本地 `qrcode` 库（后又被官方 SDK 替代） |
| 5 | 修复 DOM 时序 bug | `Login.vue` | `getElementById('qrcode')` 时 DOM 还没渲染 |
| 6 | 修复 `#wechat_redirect` 缺失 | `auth.go:318` | URL fragment 告诉微信弹出授权页而不是显示二维码页 |
| 7 | 修复 `redirect_uri` 未 URL 编码 | `auth.go:316` | 微信 OAuth 要求编码 |
| 8 | **改用 wxLogin.js 官方 SDK** | 多个文件 | 详见下方 |

#### 核心变更：wxLogin.js 官方 SDK

**为什么改：** 微信官方推荐方案，更稳定，支持快速登录。

**架构变化：**

```
旧方案：后端生成 URL → 前端用 qrcode 库画二维码 → 轮询 check 接口
新方案：前端加载 wxLogin.js SDK → 微信官方渲染二维码 → 通过前端回调页拿 code → 后端 API 直接返回 JWT
```

**新增文件：**
- `web/front/src/views/WxCallback.vue` — 微信授权回调页

**新增后端 API：**
- `GET /api/auth/wechat-appid` → 返回 `{"appid": "wx4f153..."}`
- `POST /api/auth/wechat-code` → 接收 `{"code": "xxx"}`，换 openid，返回 JWT

**保留的旧接口（不再被前端调用，待清理）：**
- `GET /api/auth/wechat-qrcode` — 生成扫码 URL
- `GET /api/auth/wechat-check` — 轮询扫码状态
- `GET /api/auth/wechat-callback` — OAuth 回调（返回 HTML）

#### 决策记录

| 决策 | 选择 | 原因 |
|---|---|---|
| 域名 | `zgjz.insightj.cn` | `gjz.shadouyou.cloud` 指向老服务器 |
| 扫码登录方案 | 微信开放平台 qrconnect | 已有审核通过的网站应用 |
| 二维码渲染 | wxLogin.js 官方 SDK | 比自绘二维码更稳定，支持快速登录 |
| 快速登录 | 保留（默认启用） | Windows/Mac 微信客户端可一键确认，无需扫码 |
| 旧 v2 分支 | 已删除 | 落后 main 153 commits，内容已合并到 main |

#### 线上数据

| 指标 | 数据 |
|---|---|
| 注册用户 | 7 人 |
| 占卜记录 | 138 条 |
| Quota 记录 | 2338 条 |
| 数据库大小 | ~864KB |
| 容器运行时长 | 约 6 周（自 2026-06-01） |

---

## 四、API 概览（用户端）

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| POST | `/api/auth/register` | 手机号注册 | 否 |
| POST | `/api/auth/login` | 手机号登录 | 否 |
| POST | `/api/auth/sms-send` | 发送短信验证码（开发模式仅打印日志） | 否 |
| POST | `/api/auth/sms-login` | 短信验证码登录 | 否 |
| **GET** | **`/api/auth/wechat-appid`** | **获取微信 AppID（供 wxLogin.js）** | **否** |
| **POST** | **`/api/auth/wechat-code`** | **微信 code 换 JWT** | **否** |
| POST | `/api/auth/wechat-login` | 微信小程序登录 | 否 |
| POST | `/api/user/bind-wechat` | 绑定微信 | Bearer Token |
| GET | `/api/user` | 获取用户信息 | Bearer Token |
| POST | `/api/divine/stream` | 流式起卦解卦 (SSE) | Bearer Token |
| GET | `/api/history` | 历史算卦记录 | Bearer Token |
| GET/POST | `/api/history/{id}/translate` | 翻译 AI 解读 | Bearer Token |

管理后台 API 见后端代码 `cmd/server/main.go`。

---

## 五、待办与注意事项

### ⚠️ 急迫事项
- [ ] **SSL 证书续期** — `zgjz.insightj.cn` 证书 2026-09-14 到期，约 2 个月后过期
- [ ] **短信服务接入** — 当前验证码仅打印日志，未接入真实短信服务

### 🗑️ 可清理项
- [x] ~~清理旧接口 `wechatQRCode`、`wechatCheck`、`wechatCallback`~~ ✅ 已删除
- [x] ~~删除 `qrcode` npm 依赖~~ ✅ 已卸载
- [x] ~~更新 Makefile 过时目标~~ ✅ `deploy-remote` 已指向新域名
- [ ] 生产服务器上的 `gjz.shadouyou.cloud` Nginx 配置可清理

### 📝 开发备忘
- 后端无第三方框架，路由在 `cmd/server/main.go` 中硬编码
- 微信配置通过环境变量注入，不用 `config.yaml`
- 小程序端 `miniapp/` 和 `miniapp-native/` 两个版本存在，API 地址写死为 `gjz.shadouyou.cloud`（需更新）
- 本地开发：`make dev-backend` / `make dev-frontend` / `make dev-admin`
- Docker 部署：`docker compose up -d --build`
- 生产部署目录：`/root/yiguan`
- 数据库文件：`/root/yiguan/data/yiguan.db`

### 给下一个 Agent 的交接信息

```
当前最新 commit: 1a71f60 (2026-07-16)
当前分支: main
本地工作区: 干净
远端仓库: git@github.com:kiddyt00/yiguan.git
生产服务器: root@124.223.16.159 (密码: Jason1987!@#)
生产域名: https://zgjz.insightj.cn
微信开放平台已配置: ✅
微信小程序已配置: ❌ (用户有 appid/secret 但未提供)
```

---

## 六、项目结构（关键路径）

```
yiguan/
├── cmd/server/main.go              # 入口 + 路由
├── internal/
│   ├── handler/
│   │   ├── auth.go                  # 认证（含微信扫码）
│   │   ├── user.go                  # 用户（含绑定微信）
│   │   ├── divine.go                # 起卦
│   │   ├── divine_stream.go         # SSE 流式起卦
│   │   ├── history.go               # 历史记录
│   │   ├── hexagram_handler.go      # 后台卦象管理
│   │   ├── model_handler.go         # 后台模型管理
│   │   ├── ad_handler.go            # 广告
│   │   ├── admin.go                 # 后台用户管理
│   │   ├── analytics.go             # 后台统计数据
│   │   └── translate_handler.go     # 翻译
│   ├── middleware/
│   │   └── auth.go                  # JWT 鉴权
│   ├── llm/
│   │   ├── client.go                # LLM 客户端
│   │   ├── router.go                # 模型路由/容错链
│   │   └── stream.go                # 流式输出
│   └── store/
│       ├── store.go                 # 接口定义
│       └── sqlite/                  # SQLite 实现
├── web/
│   ├── front/                       # 用户端 SPA
│   │   └── src/views/
│   │       ├── Login.vue            # 登录页（含微信扫码）
│   │       └── WxCallback.vue       # 微信授权回调页
│   └── admin/                       # 管理后台 SPA
├── miniapp/                         # UniApp 小程序
├── miniapp-native/                  # 原生小程序
├── docker-compose.yml
├── Dockerfile.backend
├── Dockerfile.frontend
└── deploy/
    ├── nginx.conf                   # 容器内 Nginx 配置
    └── host-nginx.conf              # 主机 Nginx 配置
```
