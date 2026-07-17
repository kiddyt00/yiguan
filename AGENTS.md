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

# Prod deploy
git push origin main
ssh root@<server-ip>
cd /root/yiguan && git pull && docker compose up -d --build
```

## Architecture

```
User → https://zgjz.insightj.cn (Host Nginx SSL)
  → Docker: yiguan-frontend (Nginx → SPA / API proxy)
  → Docker: yiguan-backend (Go :8080 → SQLite)
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
| `miniapp/` | UniApp WeChat miniprogram (Vue 3) |
| `miniapp-native/` | Native WeChat miniprogram |
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
- **Vue `v-if` + DOM 操作时序陷阱** — 设置 `ref` 触发 `v-if` 渲染后，Vue 的 DOM 更新是异步的。如果紧接着 `document.getElementById()` 会拿到 `null` 而静默失败。典型：`Recharge.vue` 中 `showQR.value = true` 后立即 `drawQR()` 调用 `document.getElementById('pay-qrcode')` 返回 `null`，`if (!el) return` 跳过，二维码永不渲染。修复：`import { nextTick } from 'vue'`，`await nextTick()` 后再操作 DOM。模式：`showQR.value = true; await nextTick(); document.getElementById(...)`。
- **SMS is fake** — `sms-send` only logs codes. No real SMS provider configured.
- **Docker build uses Chinese mirrors** — `deploy.sh` has `ensure_image()` that pulls from `docker.1ms.run/library` as proxy. If this goes down, builds fail.
- **Rate limiter is global (60/min/IP)** — no per-route granularity. Could be tightened for SMS/register endpoints.
- **Git working tree has `tmp/` files** — `tmp/*.png` are tracked by git and show as deleted in status. They're artifacts, ignore them.
