#!/bin/sh
# ============================================================
# Docker 深度清理脚本 —— 每 30 天由 cron 定期执行
# 仅清理未使用的镜像 / 构建缓存 / 容器 / 系统日志，
# 不影响运行中的容器（docker system prune 不触碰活跃容器）。
#
# cron 行（root）:
#   0 3 1 * * /root/yiguan/deploy/docker-cleanup.sh
# ============================================================
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

echo "===== $(date '+%F %T') 深度清理开始 =====" >> /var/log/docker-cleanup.log

# 清理未使用的镜像、容器、网络、构建缓存（保留运行中容器及其镜像）
docker system prune -af >> /var/log/docker-cleanup.log 2>&1

# 系统日志压缩到 200M 以内
journalctl --vacuum-size=200M >> /var/log/docker-cleanup.log 2>&1

echo "===== $(date '+%F %T') 清理完成 =====" >> /var/log/docker-cleanup.log
