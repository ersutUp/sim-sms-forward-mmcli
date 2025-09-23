#!/bin/bash
# 1. 接收并处理根目录参数（加双引号避免空参数解析问题）
ROOT_DIR="$1"

# 2. 设置默认根目录（若未传入参数）
if [ -z "$ROOT_DIR" ]; then
  ROOT_DIR="/home/sim-sms-forward-mmcli"
fi

# 配置区域 - 根据实际情况修改
PROGRAM_PATH="$ROOT_DIR/sim-sms-forward"  # 程序路径（与前一个脚本保持一致）
LOG_FILE="$ROOT_DIR/logs/watchdog.log"     # 日志文件路径
RUN_SCRIPT="$ROOT_DIR/run.sh"

# 确保日志文件存在并可写
if [ ! -f "$LOG_FILE" ]; then
    touch "$LOG_FILE" || { echo "无法创建日志文件: $LOG_FILE"; exit 1; }
fi

# 日志输出函数（带时间戳）
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"  # 同时输出到控制台
}

# 检查程序是否运行
is_running() {
    # 通过程序路径匹配进程（避免同名进程干扰）
    pgrep -f "$PROGRAM_PATH" > /dev/null 2>&1
    return $?  # 0=运行中，1=未运行
}

# 启动程序
start_program() {
    if [ ! -x "$PROGRAM_PATH" ]; then
        log "错误：程序文件不可执行或不存在 → $PROGRAM_PATH"
        return 1
    fi

    log "启动程序: $PROGRAM_PATH"
    ("$RUN_SCRIPT" "$ROOT_DIR" > /dev/null) &  # 后台启动程序

    # 检查启动是否成功
    sleep 2  # 等待程序初始化
    if is_running; then
        log "程序启动成功"
        return 0
    else
        log "程序启动失败！"
        return 1
    fi
}

if is_running; then
    # 程序运行中（仅记录状态，可注释以减少日志量）
    log "程序运行正常"
else
    # 程序未运行，尝试拉起
    log "程序未运行，准备启动..."
    start_program
fi
