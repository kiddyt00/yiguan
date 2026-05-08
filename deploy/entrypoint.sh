#!/bin/sh
# 确保挂载目录可写
mkdir -p /data /app/logs

# 同时输出到 Docker stdout（供 compose logs 查看）和日志文件
exec ./yiguan 2>&1 | tee -a /app/logs/yiguan.log
