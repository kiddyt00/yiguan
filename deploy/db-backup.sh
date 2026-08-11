#!/bin/sh
# ============================================================
# 数据库每日备份（宿主机 sqlite3 在线备份，一致性快照）
# SQLite 文件锁是 OS 级，宿主机进程可安全访问容器挂载的 db；
# `.backup` 命令是 SQLite 官方在线备份 API，不影响运行中的服务。
#
# cron（root）:
#   30 2 * * * /root/yiguan/deploy/db-backup.sh >> /var/log/db-backup.log 2>&1
# ============================================================
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

DB=/root/yiguan/data/yiguan.db
BACKUP_DIR=/root/yiguan/data/backup
AVATARS=/root/yiguan/data/avatars
KEEP=14

mkdir -p "$BACKUP_DIR"
STAMP=$(date +%Y%m%d_%H%M%S)
OUT="$BACKUP_DIR/yiguan-$STAMP.db"

echo "===== $(date '+%F %T') 备份开始 =====" >> /var/log/db-backup.log

# 1. 在线备份数据库（一致性快照）
if sqlite3 "$DB" ".backup '$OUT'"; then
    # 2. 验证备份完整性
    if sqlite3 "$OUT" "PRAGMA integrity_check;" 2>/dev/null | grep -q ok; then
        echo "OK: $OUT ($(du -h "$OUT" | cut -f1))" >> /var/log/db-backup.log
    else
        echo "❌ 备份完整性校验失败: $OUT" >> /var/log/db-backup.log
        exit 1
    fi
else
    echo "❌ 备份失败: $DB" >> /var/log/db-backup.log
    exit 1
fi

# 3. 顺带打包头像目录（可选，丢了可重新上传）
if [ -d "$AVATARS" ] && [ -n "$(ls -A "$AVATARS" 2>/dev/null)" ]; then
    tar -czf "$BACKUP_DIR/avatars-$STAMP.tar.gz" -C "$(dirname "$AVATARS")" "$(basename "$AVATARS")" \
        && echo "OK: avatars-$STAMP.tar.gz" >> /var/log/db-backup.log
fi

# 4. 保留最近 KEEP 份，清理更旧的
ls -1t "$BACKUP_DIR"/yiguan-*.db 2>/dev/null | tail -n +$((KEEP+1)) | xargs -r rm -f
ls -1t "$BACKUP_DIR"/avatars-*.tar.gz 2>/dev/null | tail -n +$((KEEP+1)) | xargs -r rm -f

echo "===== $(date '+%F %T') 备份完成 =====" >> /var/log/db-backup.log
