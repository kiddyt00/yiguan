#!/bin/bash
set -e

# 国内代理拉取基础镜像（不存在时才拉取，已存在则跳过）
ensure_image() {
    local proxy="docker.1ms.run/library"
    local image="$1"
    if docker image inspect "$image" &>/dev/null; then
        echo "✅ $image 已存在，跳过"
        return
    fi
    echo "📦 拉取 $image (via $proxy)..."
    docker pull "$proxy/$image"
    docker tag "$proxy/$image" "$image"
}

ensure_image golang:1.24-alpine
ensure_image alpine:3.21
ensure_image node:20-alpine
ensure_image nginx:alpine

echo "🔨 构建并启动容器..."
docker compose up -d --build

echo ""
echo "✅ 部署完成! http://49.235.108.61:8080"
