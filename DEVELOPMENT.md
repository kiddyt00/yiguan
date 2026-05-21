# 开发工作流

## 标准流程

```bash
# 1. 开发
# 修改代码...

# 2. 本地编译验证
make build-backend  # 后端编译
# 前端构建由 Docker 完成

# 3. 提交 & 推送
git add -A
git commit -m "type: 描述"
git push origin main

# 4. 远程部署（自动拉取 + Docker 构建）
make deploy-remote
# 等价于：
# ssh ubuntu@49.235.108.61 'sudo bash -c "cd /root/yiguan && git pull && bash deploy.sh"'
```

## 远程部署细节

- 目标服务器：`49.235.108.61`
- 代码目录：`/root/yiguan`
- 部署脚本：`deploy.sh`（拉取基础镜像 → `docker compose up -d --build`）
- 前端地址：`http://49.235.108.61:8080`
- 管理后台：`http://49.235.108.61:8080/admin`
- 后端 API：通过前端 Nginx 反向代理 `/api/`

## 环境变量

通过项目根目录的 `.env` 文件配置：

```bash
JWT_SECRET=xxx
LLM_API_KEY=xxx
ADMIN_PHONE=138xxxx
ADMIN_PASSWORD=xxx
WX_APPID=xxx          # 微信小程序（手机端）
WX_SECRET=xxx
WX_OPEN_APPID=xxx     # 微信开放平台（网页扫码登录）
WX_OPEN_SECRET=xxx
SMS_ACCESS_KEY_ID=xxx # 阿里云短信
SMS_ACCESS_KEY_SECRET=xxx
SMS_SIGN_NAME=观己斋
SMS_TEMPLATE_CODE=SMS_xxx
```

## 注意事项

- 前端 `/admin` 是独立 SPA，修改后需重新构建 Docker 镜像
- 数据库文件挂载在 `./data/yiguan.db`，重启不丢失
- 生产环境密钥通过 `.env` 注入，不硬编码在代码中
