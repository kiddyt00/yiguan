# 易观 (Yi Guan) — 开发进度记录

> 最后更新：2026-07-22
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
| **服务器** | 腾讯云（主机名: VM-0-13-ubuntu） |
| **OS** | Ubuntu 24.04.4 LTS |
| **SSH** | 凭据在 1Password / `.env` |
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
# 此文件包含生产环境密钥，存储在 /root/yiguan/.env，请勿提交到 Git
HTTP_PORT=8080
JWT_SECRET=<your-secret>
LLM_API_KEY=<your-key>
ADMIN_PHONE=<your-phone>
ADMIN_PASSWORD=<your-password>
WX_OPEN_APPID=<your-appid>
WX_OPEN_SECRET=<your-secret>
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

**已清理的旧接口：**
- ~~`GET /api/auth/wechat-qrcode`~~ ✅ 已删除
- ~~`GET /api/auth/wechat-check`~~ ✅ 已删除
- ~~`GET /api/auth/wechat-callback`~~ ✅ 已删除

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

### 2026-07-17 — 支付系统：微信支付修复 + JSAPI 小程序支付 + 支付宝准备

#### 背景
从微信扫码登录上线后，开始推进支付系统。修复了前端支付弹窗二维码渲染问题，新增小程序 JSAPI 支付，并规划支付宝扫码支付接入。

#### 完成事项

| # | 事项 | 文件 | 说明 |
|---|---|---|---|
| 1 | 修复截图离屏渲染 | `ResultView.vue` | 离屏容器渲染解决内容截断和 `overflow:hidden` 问题 |
| 2 | 前端金额显示 ¥0.01 | `Recharge.vue` | 支持测试包 1 分钱金额显示 |
| 3 | 支付二维码渲染修复 | `Recharge.vue` | **Vue `v-if` + `nextTick` 时序坑**（详见下方） |
| 4 | `WX_PAY_APPID` 环境变量 | `docker-compose.yml` | 与 `WX_OPEN_APPID` 解耦，小程序支付独立 appid |
| 5 | `qrcodejs` 本地化 | `web/front/index.html` | 从 CDN → 本地打包，避免国内 CDN 加载失败 |
| 6 | 微信支付响应日志 | `order_handler.go` | 完整记录下单请求/响应日志便于排查 |
| 7 | 新增测试包 (`¥0.01`) | `order_handler.go` | `products` map 加入 `"test"` 商品用于测试 |
| 8 | **JSAPI 小程序支付后端** | `order_handler.go` | 统一下单重构 + 小程序支付接口 |
| 9 | Vue nextTick 陷阱记录 | `AGENTS.md` | 记录典型陷阱供后续 Agent 参考 |

#### 核心坑点：Vue `v-if` + DOM 操作时序

**现象：** 支付弹窗点击后二维码不渲染，控制台无报错。

**根因：**
```
showQR.value = true          ← v-if 触发 DOM 渲染
// Vue 的 DOM 更新是异步的，这里 DOM 还没创建
document.getElementById('pay-qrcode') → null  → 静默失败
```

**修复：**
```js
showQR.value = true
await nextTick()             ← 等 Vue 完成 DOM 更新
document.getElementById('pay-qrcode')  → 找到元素 ✅
```

⚠️ 花了大量时间排查 QR 库/CDN/构建/后端，根因就是少了一个 `nextTick`。

#### JSAPI 小程序支付

| 接口 | 说明 |
|---|---|
| `POST /api/orders/jsapi-create` | 小程序统一下单，返回 JSAPI 调起支付参数（`appId`、`timeStamp`、`nonceStr`、`package`、`paySign`） |

**架构变化：** 原来 `wechatPayNative()` 直接调微信 API，重构为 `unifiedOrder()` 共用方法，参数化 `trade_type=NATIVE|JSAPI` 和 `openid`。

#### 支付宝扫码支付 — 规划中

| 事项 | 状态 |
|---|---|
| 资料准备清单 | ✅ 已整理下发（见下方待办） |
| 后端下单 + 回调 | ⏳ 待资料齐全后开发 |
| 前端支付切换 | ⏳ 待后端完成后对接 |

**需准备的资料：** AppID、应用私钥、支付宝公钥（企业商户，应用名：真观己斋）

#### 环境变量新增

| 变量 | 用途 |
|---|---|
| `WX_PAY_MCHID` | 微信支付商户号 |
| `WX_PAY_API_KEY` | 微信支付 API 密钥 |
| `WX_PAY_APPID` | 微信支付 AppID（小程序支付用，与 `WX_OPEN_APPID` 可不同） |
| `WX_PAY_NOTIFY_URL` | 微信支付回调地址（默认 `https://zgjz.insightj.cn/api/orders/notify`） |

---

### 2026-07-20 — 小程序 API 地址更新 + 旧接口清理

#### 完成事项

| # | 事项 | 文件 | 说明 |
|---|---|---|---|
| 1 | 小程序 API 地址更新 | `miniapp/` 和 `miniapp-native/` 共 8 个文件 11 处 | `gjz.shadouyou.cloud` → `zgjz.insightj.cn` |
| 2 | 清理旧微信扫码接口 | 路由 + handler 代码 | `wechat-qrcode`、`wechat-check`、`wechat-callback` 已删除 |
| 3 | 卦象管理页点击弹窗修复 | `Hexagrams.vue` | 函数名 + ref 变量名缩短导致的不匹配问题已修复 |
| 4 | 生产部署 | 服务器部署 | 已 SSH 登录生产部署新版 |
| 5 | 【新增】小程序支付（JSAPI） | `miniapp-native/` + `miniapp/` | 用户页面添加充值功能，调起 `wx.requestPayment()` |

---

### 2026-07-21 — 小程序全流程打通 + 生产配置修复

#### 完成事项

| # | 事项 | 说明 |
|---|---|---|
| 1 | 安装 `miniprogram-ci` 上传工具 | 已添加到 Makefile：`make upload-miniapp` |
| 2 | 上传密钥配置 | 已放入 `miniapp-native/` |
| 3 | 上传体验版 v1.0.0 | ✅ 成功上传 |
| 4 | 服务器域名 + IP 白名单配置 | `zgjz.insightj.cn` + `221.12.22.43` |
| 5 | 修复生产环境缺失 `WX_APPID`/`WX_SECRET` | docker-compose 缺少 env 传递，已修复 |
| 6 | 修复 JSAPI 下单 XML 缺少 `openid` | request_struct 漏了 OpenID 字段，微信返回签名错误 |
| 7 | **小程序支付全流程测试通过** | ✅ 微信登录→选套餐→调起支付→支付成功 |
| 8 | `favicon.ico` 404 修复 | Nginx 301 重定向到 SVG |

---

### 2026-07-22 — 支付宝接入准备 + 进度页面完善

#### 完成事项

| # | 事项 | 说明 |
|---|---|---|
| 1 | 支付宝 RSA2 密钥对生成 | 应用私钥 + 应用公钥已保存到项目目录（已 gitignore） |
| 2 | 管理后台进度页面完善 | 完整生命周期 280 次提交 → 10 个阶段鱼骨图 |
| 3 | 进度页统计卡片可点击跳转 | 点击后平滑滚动到时间线/待办/部署区域 |
| 4 | 进度页双栏布局 + 左右交替鱼骨 | 左侧时间线 + 右侧待办事项 |
| 5 | SECRETS.md 更新 | 补充微信小程序 + 支付宝密钥信息 |
| 6 | `.gitignore` 更新 | 添加 `*.pem` 保护支付宝密钥文件 |
| 7 | 微信账号合并（用户6→8） | 历史记录/quota迁移，openid切换到小程序端，禁用重复账号 |
| 8 | **UnionID去重机制** | 数据库加`unionid`字段，`jscode2session`获取unionid，先查unionid再查openid |
| 9 | 微信头像显示修复 | `wx.getUserInfo`无avatar时不覆盖数据库，`fillUserAvatar`补全admin列表 |
| 10 | 上传小程序v1.0.2→v1.0.3 | 修复头像清空bug+用户管理优化 |
| 11 | Web端导航栏显示微信头像 | 替换首字母圆点为头像图片 |
| 12 | **账号合并最终修复** | 清理用户6/7/8/9，保留ops(ID10)，openid支持逗号分隔多值LIKE查询 |
| 13 | `wechatCode` + `wechatOpenToken` unionid支持 | 网页扫码登录也按unionid去重 |

#### 支付宝接入状态

| 步骤 | 状态 |
|---|---|
| 商户号 `2088780753525097` | ✅ 已确认 |
| 电脑网站支付签约 | ⏳ 等待用户在商家中心操作 |
| 开放平台创建应用（真观己斋） | ⏳ 等待用户操作 |
| 应用公钥上传 | ⏳ 等待用户操作 |
| 后端代码开发 | ⏳ 等资料齐全后开始 |

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
| POST | `/api/orders/create` | 微信支付 Native 下单（Web 扫码） | Bearer Token |
| POST | `/api/orders/jsapi-create` | 微信支付 JSAPI 下单（小程序） | Bearer Token |
| GET | `/api/orders/{id}` | 查询订单详情 | Bearer Token |
| GET | `/api/orders` | 订单列表 | Bearer Token |
| POST | `/api/orders/notify` | 微信支付回调通知 | 否（微信服务器回调） |

管理后台 API 见后端代码 `cmd/server/main.go`。

---

## 五、待办与注意事项

### ⚠️ 急迫事项
- [ ] **SSL 证书续期** — `zgjz.insightj.cn` 证书 2026-09-14 到期（acme.sh 自动续期，无需手动）
- [ ] **短信服务接入** — 当前验证码仅打印日志，未接入真实短信服务
- [ ] **支付宝扫码支付接入** — 后端下单 + 回调通知 + 前端支付切换
  - 商户号：`2088780753525097` ✅
  - 应用私钥：已生成 ✅
  - 待获取：AppID + 支付宝公钥（用户在开放平台操作中）
- [x] ~~小程序对接微信支付~~ ✅ 已完成（miniapp-native + miniapp 双端）

### ✅ 安全加固
- [x] ~~修复 5 处 err.Error() 泄露~~ ✅ 已统一替换为通用提示
- [x] ~~CORS 全开 `*`~~ ✅ 改为白名单模式
- [x] ~~API 限流~~ ✅ 新增 60次/分钟/IP 滑动窗口限流
- [x] ~~Docker 日志轮转~~ ✅ 每个容器限 10MB x 3 个文件
- [x] ~~前端统一 fetch 封装~~ ✅ utils/request.js + 8 个视图文件改造（401 自动拦截）

### 📝 开发备忘
- 后端无第三方框架，路由在 `cmd/server/main.go` 中硬编码
- 微信配置/支付配置通过环境变量注入，不用 `config.yaml`
- 支付环境变量：`WX_PAY_MCHID`、`WX_PAY_API_KEY`、`WX_PAY_APPID`、`WX_PAY_NOTIFY_URL`
- 小程序端 `miniapp/` 和 `miniapp-native/` 两个版本存在，API 地址已更新为 `zgjz.insightj.cn`
- 本地开发：`make dev-backend` / `make dev-frontend` / `make dev-admin`
- Docker 部署：`docker compose up -d --build`
- 生产部署目录：`/root/yiguan`
- 数据库文件：`/root/yiguan/data/yiguan.db`

### 给下一个 Agent 的交接信息

```
当前最新 commit: 84e1693 (2026-07-20)
当前分支: main
本地工作区: 干净
远端仓库: git@github.com:kiddyt00/yiguan.git
生产服务器: 凭据在 1Password
生产域名: https://zgjz.insightj.cn
微信开放平台已配置: ✅
微信扫码登录: ✅（wxLogin.js 官方 SDK）
微信支付（Web 扫码）: ✅ 已上线
微信支付（小程序 JSAPI）: ✅ 已上线（miniapp-native + miniapp 双端）
支付宝扫码支付: ⏳ 规划中，待资料齐全后开发
微信小程序已配置: ❌
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
│   │   ├── translate_handler.go     # 翻译
│   │   └── order_handler.go         # 订单（微信支付 Native + JSAPI）
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
