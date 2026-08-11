# Yi Guan (易观) — Agent Guide

AI-driven I Ching divination platform. Go backend + Vue 3 SPA, Docker deployed.

## Build / Test / Lint

```bash
# Backend
make dev-backend        # go run ./cmd/server/
make test               # go test ./... -count=1 -v
make test-short         # go test ./... -count=1 -short

# Frontend (in web/front/)
npm run dev             # Vite dev server
npm run build           # Production build

# Admin panel (in web/admin/)
npm run dev
npm run build

# Docker
docker compose up -d --build   # Build + start all containers
docker compose logs -f         # Follow logs

# Prod deploy（⚠️ 升级 = 滚动升级，见下方 Deployment Convention）
make deploy-remote
# 即: git push origin main && ssh root@124.223.16.159 'cd /root/yiguan && ./deploy/rolling-update.sh'
```

## Deployment Convention（⚠️ 升级 = 滚动升级）

**任何时候说"升级 / 部署 / 发布"，统一执行滚动升级，禁止再跑 `docker compose up -d --build`（同时重建所有容器 → 全站 502）。**

- 升级入口：`make deploy-remote`，或 `ssh root@124.223.16.159 'cd /root/yiguan && ./deploy/rolling-update.sh'`
- 流程（deploy/rolling-update.sh）：`git pull --ff-only` → `docker compose build`（不影响运行）→ 逐个替换 `backend-a` → `backend-b` → `frontend-a` → `frontend-b`（`--no-deps --wait` 等 healthy）→ 循环验证 200
- 拓扑：宿主 nginx `frontend_pool`(:8081/:8082) → `frontend-a`/`frontend-b` 容器 nginx `backend_pool` → `backend-a`/`backend-b` → SQLite WAL 共享卷 `/data`（双实例并发写，WAL 已启用）
- 生产宿主 nginx 配置对应 `deploy/host-nginx.conf`（复制到 `/etc/nginx/conf.d/zgjz.insightj.cn.conf` + `nginx -t` + reload）
- 首次从单实例迁移到双实例有一键窗口，之后滚动全程无 502
- 架构图：`docs/archify/rolling-topology.html`、`docs/archify/rolling-flow.html`

### 磁盘清理策略

- **每次构建后温和清理**（已内置 `deploy/rolling-update.sh`）：`docker builder prune -f` + `docker image prune -f`，只清未使用旧层，保留近期缓存（不影响构建速度）
- **每 30 天深度清理**（生产 cron，每月 1 号 3:00 执行 `deploy/docker-cleanup.sh`）：`docker system prune -af` + `journalctl --vacuum-size=200M`，日志在 `/var/log/docker-cleanup.log`
- 手动检查：`ssh root@124.223.16.159 'docker system df; df -h /'`

### 数据库备份

- **每日自动备份**（生产 cron，每天 2:30 执行 `deploy/db-backup.sh`）：宿主机 sqlite3 在线 `.backup` 一致性快照 → `/root/yiguan/data/backup/yiguan-YYYYMMDD_HHMMSS.db` + avatars 打包，保留最近 14 份，完整性 `PRAGMA integrity_check` 校验，日志 `/var/log/db-backup.log`
- 手动备份：`ssh root@124.223.16.159 '/root/yiguan/deploy/db-backup.sh'`
- 备份与 db 同机不同目录（防目录级误删）；建议定期下载一份到本地/异地（防整机故障）

## Design Convention（⚠️ 所有页面设计必须去 AI 味 / anti-slop）

**设计任何页面/组件（含错误页、维护页、落地页、UI 改版）时，必须去除 AI 味，规则如下：**

- **文案直白诚实，禁止俏皮话与强行文化梗**：错误页就写"页面不存在 / 服务暂时不可用"，不要写"此卦象并不存在 / 系统正在渡劫"这类强行主题化文案；`em` 斜体强调字不用于标题
- **克制装饰**：不堆装饰性图形/动画（旋转太极 SVG、呼吸灯、粒子背景等），品牌元素一点即可（角落落款）
- **反万能模板**：不用"居中大标题 + 副标题 + 按钮 + 底部小字"的 AI 模板结构；布局要有真实用途（左对齐、大字错误码、清晰操作路径），参考真实产品错误页（GitHub / Stripe 风格）
- **功能优先**：错误码、说明、操作（返回/重试）信息明确；系统字体栈（不滥用衬线大字报）；保持项目墨色/暖底 token（`#faf8f4` 底 / `#26221c` 墨 / `#8a8173` 弱化）
- 自检：交付前用 hallmark 的 slop-test 自查，任何一条命中即重做

## Architecture

```
User → https://zgjz.insightj.cn (Host Nginx SSL → frontend_pool :8081/:8082)
  → Docker: yiguan-frontend-a / yiguan-frontend-b (Nginx → SPA / API proxy → backend_pool)
  → Docker: yiguan-backend-a / yiguan-backend-b (Go :8080 → SQLite WAL 共享卷 /data)
```

- **Backend**: Go 1.24, stdlib `net/http` (no framework). Routes hardcoded in `cmd/server/main.go`. Go 1.22+ mux with method patterns (`GET /path`, `POST /path`).
- **Frontend**: Vue 3 + Vite + Pinia + Tailwind CSS v4. SPA with `vue-router` (web history mode).
- **Admin panel**: Separate Vue 3 SPA, built independently, served under `/admin`.
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO).
- **AI**: OpenAI-compatible LLM clients (Qwen, DeepSeek, etc.). Model routing with fallback chain.

## Key Files & Directories

| Path | Purpose |
|---|---|
| `cmd/server/main.go` | Entry point, route registration, middleware chain |
| `internal/handler/` | HTTP handlers (auth, divine, history, admin, ads, models, translate) |
| `internal/middleware/` | JWT auth, admin auth, rate limiter |
| `internal/engine/` | I Ching divination logic (coin toss, hexagram generation) |
| `internal/llm/` | LLM client, router, SSE stream helpers |
| `internal/store/` | Store interface + SQLite implementation |
| `internal/qianwen/` | Alibaba DashScope / Qianwen API client |
| `web/front/` | User-facing SPA |
| `web/admin/` | Admin panel SPA |
| `miniapp-native/` | Native WeChat miniprogram (production) |
| `deploy/` | Nginx configs, entrypoint script |
| `docs/PROGRESS.md` | Development timeline, environment info, handoff notes |
| `docs/superpowers/` | Historical design docs and plans |
| `utils/request.js` | Frontend API client (auto-injects Bearer token, handles 401) |

### Config files

- `config.yaml` — Local dev config (port, LLM providers, JWT secret, admin)
- `.env` — Production secrets (JWT_SECRET, LLM_API_KEY, WX_OPEN_*). Live at `/root/yiguan/.env`
- `docker-compose.yml` — Two services: backend + frontend, logging: 10MB x 3 files

## API Overview (User-facing)

All API calls route through frontend Nginx (`/api/` → backend:8080). Auth via `Authorization: Bearer <token>`.

### Public (no auth)
| Method | Path | Purpose |
|---|---|---|
| POST | `/api/auth/register` | Phone + password registration |
| POST | `/api/auth/login` | Phone + password login |
| POST | `/api/auth/sms-send` | Send SMS code (dev: logs only) |
| POST | `/api/auth/sms-login` | SMS code login/register |
| GET | `/api/auth/wechat-appid` | Get WeChat Open Platform AppID (for wxLogin.js) |
| POST | `/api/auth/wechat-code` | Exchange WeChat auth code → JWT |
| GET | `/api/ads/active` | List active ads |

### Authenticated (Bearer token)
| Method | Path | Purpose |
|---|---|---|
| GET/PUT | `/api/user` | Get/update profile |
| POST | `/api/user/bind-wechat` | Bind WeChat to existing account |
| POST | `/api/divine/stream` | SSE stream: toss coins + AI interpretation |
| GET | `/api/history` | List divination history |
| GET/POST | `/api/history/{id}/translate` | Translate AI interpretation |
| POST | `/api/ads/{id}/watch\|complete` | Ad watching flow |

## Middleware Chain

```
Request → RateLimit (60/min/IP) → Logging → Route → corsWrap → authMW/adminMW
```

- **RateLimit** — `internal/middleware/ratelimit.go`. Sliding window per IP, auto-cleanup every 5min.
- **corsWrap** — `main.go`. Whitelist: `zgjz.insightj.cn`, `localhost:5173`, `localhost:8080`.
- **authMW** — `internal/middleware/auth.go`. Extracts JWT `user_id` into request context.
- **adminMW** — `internal/middleware/admin.go`. Verifies `role == "admin"`.

## WeChat Login Flow

Uses wxLogin.js official SDK (NOT the deprecated polling approach):

1. Frontend loads `https://res.wx.qq.com/connect/zh_CN/htmledition/js/wxLogin.js`
2. Creates `new WxLogin({appid, redirect_uri, scope:'snsapi_login', self_redirect:true})`
3. User scans QR code → WeChat auth dialog on phone → redirects to `/wx-callback?code=xxx`
4. `WxCallback.vue` calls `POST /api/auth/wechat-code` → backend exchanges code for openid → returns JWT
5. Callback page `postMessage`s token to parent window → auto-login

## Production Environment

| Item | Value |
|---|---|
| Server | `***REMOVED***` (Tencent Cloud, Ubuntu 24.04) |
| Domain | `https://zgjz.insightj.cn` |
| Admin | `https://zgjz.insightj.cn/admin` |
| SSH | `root@<server-ip>` (credentials in 1Password / `.env`) |
| Deploy dir | `/root/yiguan` |
| DB | `/root/yiguan/data/yiguan.db` (SQLite, ~864KB) |
| WeChat Open | AppID in `.env` |
| SSL | Let's Encrypt via acme.sh, auto-renewed |

## Coding Conventions (observed)

- **Go**: stdlib `net/http` only. Handlers return JSON via `writeJSON()` helper. Errors logged server-side, generic messages to client.
- **Frontend**: Vue 3 Composition API (`<script setup>`). Pinia stores. Vue Router lazy-loaded routes.
- **API calls**: Use `utils/request.js` helpers (`apiGet`, `apiPost`, `apiGetJSON`, `apiPostJSON`). Never raw `fetch()`.
- **i18n**: `vue-i18n` with translation keys (e.g., `t('login.qrcode.title')`).
- **Tests**: `*_test.go` adjacent to source. Engine tests are thorough; handler tests exist for auth/model/ad but not for divine/SSE/translate.

## Gotchas for AI Agents

- **`.env.example` is outdated** — missing WX_OPEN_* vars and SMS vars. Check docs/PROGRESS.md for current env list.
- **Miniapp API URLs are hardcoded** — both `miniapp/` and `miniapp-native/` point to old domain `gjz.shadouyou.cloud` which no longer serves the app. Needs updating to `zgjz.insightj.cn`.
- **QR code `#wechat_redirect`** — if this hash fragment is missing from the qrconnect URL, WeChat won't show the auth dialog. The current wxLogin.js SDK handles this internally.
- **Vue `v-if` + DOM 操作时序陷阱** — 设置 `ref` 触发 `v-if` 渲染后，Vue 的 DOM 更新是异步的。如果紧接着 `document.getElementById()` 会拿到 `null` 而静默失败。典型：`Recharge.vue` 中 `showQR.value = true` 后立即 `drawQR()` 调用 `document.getElementById('pay-qrcode')` 返回 `null`，`if (!el) return` 跳过，二维码永不渲染。修复：`import { nextTick } from 'vue'`，`await nextTick()` 后再操作 DOM。模式：`showQR.value = true; await nextTick(); document.getElementById(...)`。⚠️ 2026-07-17 实际案例：支付弹窗二维码不出就是此问题，花了大量时间排查 QR 库/CDN/构建/后端，根因就是少了一个 `nextTick`。
- **SMS is fake** — `sms-send` only logs codes. No real SMS provider configured.
- **Docker build uses Chinese mirrors** — `deploy.sh` has `ensure_image()` that pulls from `docker.1ms.run/library` as proxy. If this goes down, builds fail.
- **Rate limiter is global (60/min/IP)** — no per-route granularity. Could be tightened for SMS/register endpoints.
- **Git working tree has `tmp/` files** — `tmp/*.png` are tracked by git and show as deleted in status. They're artifacts, ignore them.
