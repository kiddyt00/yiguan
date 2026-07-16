# 开发工作流

## 标准流程

```bash
# 1. 开发
# 修改代码...

# 2. 提交 & 推送
git add -A
git commit -m "type: 描述"
git push origin main

# 3. 远程部署（自动拉取 + Docker 构建）
ssh root@<server-ip>
cd /root/yiguan
git pull
docker compose up -d --build
```

## 远程部署细节

| 项目 | 值 |
|---|---|
| **目标服务器** | 腾讯云 |
| **SSH** | 凭据在 1Password |
| **代码目录** | `/root/yiguan` |
| **生产域名** | `https://zgjz.insightj.cn` |
| **管理后台** | `https://zgjz.insightj.cn/admin` |
| **部署方式** | `docker compose up -d --build` |
| **数据库** | `./data/yiguan.db`（挂载卷，重启不丢失） |

## 本地开发

```bash
make dev-backend   # 后端（默认 8080 端口）
make dev-frontend  # 前端 SPA（另一个终端）
make dev-admin     # 管理后台（可选）
```

## 环境变量

通过项目根目录的 `.env` 文件配置（生产环境通过 `/root/yiguan/.env`）：

```bash
# 必需
JWT_SECRET=xxx
LLM_API_KEY=xxx

# 管理员
ADMIN_PHONE=<your-phone>
ADMIN_PASSWORD=<your-password>

# 微信开放平台（网站扫码登录）
WX_OPEN_APPID=<your-appid>
WX_OPEN_SECRET=<your-secret>

# 微信小程序（暂未配置）
WX_APPID=
WX_SECRET=

# 阿里云短信（暂未接入）
SMS_ACCESS_KEY_ID=
SMS_ACCESS_KEY_SECRET=
SMS_SIGN_NAME=观己斋
SMS_TEMPLATE_CODE=SMS_xxx

# 前端暴露端口
HTTP_PORT=8080
```

## 注意事项

- 前端 `/admin` 是独立 SPA，修改后需重新构建 Docker 镜像
- 数据库文件挂载在 `./data/yiguan.db`，重启不丢失
- 生产环境密钥通过 `.env` 注入，不硬编码在代码中
- 构建时国内网络需要代理，参考 `deploy.sh` 中的 `ensure_image` 函数
- 完整开发记录见 `docs/PROGRESS.md`
