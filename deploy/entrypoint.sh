#!/bin/sh
# 确保挂载目录可写
mkdir -p /data /app/logs

exec ./yiguan 2>&1 | tee /app/logs/yiguan.log
