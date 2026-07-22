.PHONY: all build test clean dev dev-backend dev-frontend dev-admin docker-build docker-up docker-down deploy

# 变量
APP_NAME   := yiguan
GO         := go
GO_FLAGS   :=
DOCKER     := docker compose

# 默认目标
all: build

# ========== 构建 ==========

build: build-backend build-frontend
	@echo "✅ 构建完成"

build-backend:
	$(GO) build $(GO_FLAGS) -ldflags="-s -w" -o deploy/bin/$(APP_NAME) ./cmd/server/
	@echo "✅ 后端二进制: deploy/bin/$(APP_NAME)"

build-frontend:
	cd web/front && npm install && npm run build
	cd web/admin && npm install && npm run build
	@echo "✅ 前端构建完成"

# ========== 测试 ==========

test:
	$(GO) test ./... -count=1 -v

test-short:
	$(GO) test ./... -count=1 -short

# ========== 开发 ==========

dev: dev-backend

dev-backend:
	$(GO) run ./cmd/server/

dev-frontend:
	cd web/front && npm run dev

dev-admin:
	cd web/admin && npm run dev

# ========== 清理 ==========

clean:
	rm -rf deploy/bin/$(APP_NAME)
	rm -rf web/front/dist web/admin/dist
	@echo "✅ 清理完成"

# ========== Docker ==========

docker-build:
	$(DOCKER) build

docker-up:
	$(DOCKER) up -d

docker-down:
	$(DOCKER) down

docker-logs:
	$(DOCKER) logs -f

# ========== 一键部署 ==========

deploy:
	$(DOCKER) up -d --build
	@echo "🚀 部署完成! http://localhost:$${HTTP_PORT:-80}"

# ========== 微信小程序上传 ==========

# 使用方式：
#   1. 先在微信公众平台生成上传密钥，下载为 private.xxx.key
#   2. 放到 miniapp-native/ 目录下
#   3. 执行: make upload-miniapp VERSION=1.0.1 DESC="版本描述"
#   4. ⚠️ 每次上传版本号必须递增！小程序不允许重复版本号
#   5. 上传后登录 mp.weixin.qq.com → 版本管理 → 提交审核 → 发布
upload-miniapp:
	@if [ -z "$$VERSION" ]; then echo "❌ 请指定版本号: make upload-miniapp VERSION=1.0.1"; exit 1; fi
	@KEY_FILE=$$(ls miniapp-native/private.*.key 2>/dev/null | head -1); \
	if [ -z "$$KEY_FILE" ]; then echo "❌ 未找到上传密钥文件 (private.*.key)，请先在微信公众平台生成"; exit 1; fi; \
	ABS_KEY=$$(cd miniapp-native && pwd)/$$(basename "$$KEY_FILE"); \
	cd miniapp-native && npx miniprogram-ci upload \
		--pp . \
		--pkp "$$ABS_KEY" \
		--appid wx9e87b7216be83619 \
		--uv "$(VERSION)" \
		--ud "$(DESC)"
	@echo "✅ 小程序上传完成，版本: $(VERSION)"

# ========== 远程部署 ==========

deploy-remote:
	git push origin main
	ssh root@<server-ip> 'cd /root/yiguan && git pull && docker compose up -d --build'
	@echo "✅ 已上线 https://zgjz.insightj.cn"
