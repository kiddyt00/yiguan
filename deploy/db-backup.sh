#!/bin/sh
# ============================================================
# 数据库每日备份（宿主机 sqlite3 在线备份，一致性快照）
# SQLite 文件锁是 OS 级，宿主机进程可安全访问容器挂载的 db；
# `.backup` 命令是 SQLite 官方在线备份 API，不影响运行中的服务。
#
# 备份目标（双份）：
#   1. 本地    /root/yiguan/data/backup/（保留最近 KEEP=14 份）
#   2. 腾讯云 COS 异地容灾（保留 COS_KEEP_DAYS=90 天）
#
# cron（root）:
#   30 2 * * * /root/yiguan/deploy/db-backup.sh >> /var/log/db-backup.log 2>&1
#
# COS 配置（写入 /root/yiguan/.env，勿提交 git）:
#   COS_BUCKET=zgjz-backup-1438787644      # 桶名（含 APPID 后缀）
#   COS_REGION=ap-guangzhou                # 桶所在地域，如 ap-guangzhou / ap-beijing
#   COS_SECRET_ID=xxx                      # 子账号最小权限密钥（该桶 Put/Head/List/Delete）
#   COS_SECRET_KEY=xxx
#
# 前置：服务器安装腾讯云官方 coscli（Go 静态二进制，零依赖）
#   curl -s https://cosbrowser.cloud.tencent.com/software/coscli/coscli-linux-amd64 \
#        -o /usr/local/bin/coscli && chmod +x /usr/local/bin/coscli
#
# 任何变量可被环境变量覆盖（便于本地测试/演练）。
# ============================================================
set -u
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# ---------- 路径与本地保留策略 ----------
DB=${DB:-/root/yiguan/data/yiguan.db}
BACKUP_DIR=${BACKUP_DIR:-/root/yiguan/data/backup}
AVATARS=${AVATARS:-/root/yiguan/data/avatars}
KEEP=${KEEP:-14}
LOG_FILE=${LOG_FILE:-/var/log/db-backup.log}

# ---------- COS 配置（默认空 = 跳过 COS 阶段，保持纯本地行为） ----------
COS_BUCKET=${COS_BUCKET:-}
COS_REGION=${COS_REGION:-}
COS_SECRET_ID=${COS_SECRET_ID:-}
COS_SECRET_KEY=${COS_SECRET_KEY:-}
COSCLI_BIN=${COSCLI_BIN:-/usr/local/bin/coscli}
COS_PREFIX=${COS_PREFIX:-yiguan-backup}
COS_KEEP_DAYS=${COS_KEEP_DAYS:-90}

# 注入生产 .env 中的 COS_*（.env 须为合法 shell 赋值格式，项目一贯如此）
if [ -f /root/yiguan/.env ]; then
    set -a
    . /root/yiguan/.env
    set +a
fi

log() { echo "$(date '+%F %T') $*" >> "$LOG_FILE"; }
COS_ENDPOINT="cos.$COS_REGION.myqcloud.com"

# ---------- COS 上传 + head 校验（COS 未配置时直接跳过） ----------
cos_upload() {
    [ -z "$COS_BUCKET" ] && return 0
    file="$1"; key="$2"
    log "COS 上传: $key ($(du -h "$file" | cut -f1))"
    COS_SECRET_ID="$COS_SECRET_ID" COS_SECRET_KEY="$COS_SECRET_KEY" \
        "$COSCLI_BIN" cp "$file" "cos://$COS_BUCKET/$COS_PREFIX/$key" -e "$COS_ENDPOINT" >/dev/null 2>&1
    if [ $? -ne 0 ]; then
        log "❌ COS 上传失败: $key"
        return 1
    fi
    COS_SECRET_ID="$COS_SECRET_ID" COS_SECRET_KEY="$COS_SECRET_KEY" \
        "$COSCLI_BIN" stat "cos://$COS_BUCKET/$COS_PREFIX/$key" -e "$COS_ENDPOINT" >/dev/null 2>&1 \
        && log "OK: COS $key" || { log "❌ COS 校验失败: $key"; return 1; }
}

# ---------- COS 过期清理（按对象名中的 YYYYMMDD 判断，控制台可另配生命周期兜底） ----------
cos_prune() {
    [ -z "$COS_BUCKET" ] && return 0
    cutoff=$(date -d "-${COS_KEEP_DAYS} days" +%Y%m%d)
    COS_SECRET_ID="$COS_SECRET_ID" COS_SECRET_KEY="$COS_SECRET_KEY" \
        "$COSCLI_BIN" ls -r "cos://$COS_BUCKET/$COS_PREFIX/" -e "$COS_ENDPOINT" 2>/dev/null \
        | awk '$1 ~ /^cos:\/\// {print $1}' \
        | while read -r obj; do
            name=${obj##*/}
            d=$(echo "$name" | grep -oE '[0-9]{8}' | head -n1)
            [ -z "$d" ] && continue
            if [ "$d" -lt "$cutoff" ]; then
                COS_SECRET_ID="$COS_SECRET_ID" COS_SECRET_KEY="$COS_SECRET_KEY" \
                    "$COSCLI_BIN" rm "$obj" -e "$COS_ENDPOINT" >/dev/null 2>&1 \
                    && log "COS 清理: $obj"
            fi
        done
}

mkdir -p "$BACKUP_DIR"
STAMP=$(date +%Y%m%d_%H%M%S)
OUT="$BACKUP_DIR/yiguan-$STAMP.db"

log "===== $(date '+%F %T') 备份开始 ====="

# 1. 在线备份数据库（一致性快照）
if sqlite3 "$DB" ".backup '$OUT'"; then
    # 2. 验证备份完整性
    if sqlite3 "$OUT" "PRAGMA integrity_check;" 2>/dev/null | grep -q ok; then
        log "OK: $OUT ($(du -h "$OUT" | cut -f1))"
    else
        log "❌ 备份完整性校验失败: $OUT"
        exit 1
    fi
else
    log "❌ 备份失败: $DB"
    exit 1
fi

# 3. 顺带打包头像目录（可选，丢了可重新上传）
AVATAR_TAR=""
if [ -d "$AVATARS" ] && [ -n "$(ls -A "$AVATARS" 2>/dev/null)" ]; then
    AVATAR_TAR="avatars-$STAMP.tar.gz"
    tar -czf "$BACKUP_DIR/$AVATAR_TAR" -C "$(dirname "$AVATARS")" "$(basename "$AVATARS")" \
        && log "OK: $AVATAR_TAR" || AVATAR_TAR=""
fi

# 4. 保留最近 KEEP 份，清理本地更旧的
ls -1t "$BACKUP_DIR"/yiguan-*.db 2>/dev/null | tail -n +$((KEEP+1)) | xargs -r rm -f
ls -1t "$BACKUP_DIR"/avatars-*.tar.gz 2>/dev/null | tail -n +$((KEEP+1)) | xargs -r rm -f

# 5. COS 异地备份（可选阶段：未配置则跳过；配置了但缺失 region 视为错误）
COS_FAIL=0
if [ -n "$COS_BUCKET" ]; then
    if [ -z "$COS_REGION" ]; then
        log "❌ COS_BUCKET 已配置但缺少 COS_REGION"
        COS_FAIL=1
    elif [ ! -x "$COSCLI_BIN" ]; then
        log "❌ coscli 未安装: $COSCLI_BIN"
        COS_FAIL=1
    else
        cos_upload "$OUT" "yiguan-$STAMP.db" || COS_FAIL=1
        if [ -n "$AVATAR_TAR" ]; then
            cos_upload "$BACKUP_DIR/$AVATAR_TAR" "$AVATAR_TAR" || COS_FAIL=1
        fi
        cos_prune
    fi
fi

log "===== $(date '+%F %T') 备份完成 ====="
exit $COS_FAIL
