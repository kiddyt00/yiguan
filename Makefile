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

deploy: build-backend docker-build docker-up
	@echo "🚀 部署完成! http://localhost:$${HTTP_PORT:-80}"
