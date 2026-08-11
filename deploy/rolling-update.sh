#!/bin/sh
# ============================================================
# 滚动升级脚本（零成本双实例方案）
# 一次只替换一个容器，每层始终有存活实例承接流量，全程无 502。
#
# 用法:
#   make deploy-remote
#   或: ssh root@124.223.16.159 'cd /root/yiguan && ./deploy/rolling-update.sh'
#
# 拓扑:
#   宿主 nginx (frontend_pool :8081/:8082)
#     → frontend-a / frontend-b 容器 nginx (backend_pool)
#       → backend-a / backend-b 容器 → SQLite 共享卷 (/data)
# ============================================================
set -e

cd "$(dirname "$0")/.."   # 回到仓库根目录

echo "==> 1/7 拉取最新代码"
git pull --ff-only

echo "==> 2/7 构建镜像（不影响运行中的容器）"
docker compose build

# 构建后温和清理：只清未使用的旧缓存层，保留近期层（不影响下次构建速度）
docker builder prune -f
docker image prune -f

echo "==> 3/7 滚动替换 backend-a（backend-b 全程承接流量）"
docker compose up -d --no-deps --wait backend-a

echo "==> 4/7 滚动替换 backend-b（backend-a 已就绪接管）"
docker compose up -d --no-deps --wait backend-b

echo "==> 5/7 滚动替换 frontend-a（frontend-b 承接）"
docker compose up -d --no-deps --wait frontend-a

echo "==> 6/7 滚动替换 frontend-b（frontend-a 已就绪接管）"
docker compose up -d --no-deps --wait frontend-b

echo "==> 7/7 验证（循环请求应全部 200）"
check() {
    ok=0
    for i in 1 2 3 4 5; do
        code=$(curl -s -o /dev/null -w "%{http_code}" "$1" || echo 000)
        echo "  $1 -> $code"
        if [ "$code" = "200" ]; then ok=1; break; fi
        sleep 1
    done
    [ "$ok" = "1" ] || { echo "  ❌ $1 验证失败"; exit 1; }
}
check http://localhost:8081/
check http://localhost:8082/
check http://localhost:8081/api/ads/active
check http://localhost:8082/api/ads/active

echo "✅ 滚动升级完成，四容器均已更新，无 502"
