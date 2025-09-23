#!/bin/bash

# 功能：启动 sim-sms-forward 程序，支持自定义根目录
# 使用：./run.sh [程序根目录]（默认：/home/sim-sms-forward-mmcli）

# 1. 接收并处理根目录参数（加双引号避免空参数解析问题）
ROOT_DIR="$1"

# 2. 设置默认根目录（若未传入参数）
if [ -z "$ROOT_DIR" ]; then
  ROOT_DIR="/home/sim-sms-forward-mmcli"
fi

# 3. 定义关键文件路径（加双引号确保特殊字符兼容）
PROGRAM_PATH="$ROOT_DIR/sim-sms-forward"
CONFIG_PATH="$ROOT_DIR/config.json"

# 4. 检查程序文件是否存在且可执行
if [ ! -f "$PROGRAM_PATH" ]; then
  echo "错误：程序文件不存在 → $PROGRAM_PATH"
  exit 1  # 非0退出码表示执行失败
fi
if [ ! -x "$PROGRAM_PATH" ]; then
  echo "错误：程序文件不可执行 → $PROGRAM_PATH"
  echo "建议：执行 chmod +x $PROGRAM_PATH 赋予权限"
  exit 1
fi

# 5. 检查配置文件是否存在且可读
if [ ! -f "$CONFIG_PATH" ]; then
  echo "错误：配置文件不存在 → $CONFIG_PATH"
  exit 1
fi
if [ ! -r "$CONFIG_PATH" ]; then
  echo "错误：配置文件不可读 → $CONFIG_PATH"
  echo "建议：检查文件权限或当前用户身份"
  exit 1
fi

# 6. 执行程序（加双引号确保路径正确解析）
echo "启动程序：$PROGRAM_PATH"
echo "使用配置：$CONFIG_PATH"
"$PROGRAM_PATH" "$CONFIG_PATH"