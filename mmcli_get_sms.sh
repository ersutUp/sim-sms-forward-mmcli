#!/bin/bash

# 定义提取单条短信信息的函数
extract_sms_info() {
    local sms_id=$1
    # 执行一次mmcli获取数据
    local sms_data=$(mmcli -s "$sms_id" 2>/dev/null)
    
    # 检查短信是否存在
    if [ -z "$sms_data" ]; then
        echo "警告: 未找到ID为 $sms_id 的短信，跳过处理"
        return 1
    fi

    # 提取信息到变量
    local sender=$(echo "$sms_data" | grep 'number: ' | sed -E 's/^.*number: //; s/^ *//')
    local timestamp=$(echo "$sms_data" | grep 'timestamp: ' | sed -E 's/^.*timestamp: //; s/^ *//')
    local content=$(echo "$sms_data" | awk '
        /text: / {sub(/.*text: /, ""); flag=1}
        flag {
            if (/^  -+/) {flag=0; exit}
            gsub(/^ *\| *|^ +/, "");
            print
        }
    ')

    # 验证变量（为空时填充默认值）
    [ -z "$sender" ] && sender="未知号码"
    [ -z "$timestamp" ] && timestamp="未知时间"
    [ -z "$content" ] && content="无内容"

    # 输出信息（可替换为调用其他脚本）
    echo "======================================"
    echo "短信 ID: $sms_id"
    echo "发送方号码: $sender"
    echo "接收时间: $timestamp"
    echo -e "短信内容:$content"
    echo "======================================"
    echo
    
    #通知
    bash ./simForward.sh $sender $timestamp $content
    if [ $? -eq 0 ]; then
      echo "脚本B正确结束（exit 0）"
      #删除短信
      delete_sms $sms_id
    else
      echo "脚本B异常结束（非0 exit），状态码：$?"
    fi
    
}


# 删除短信
delete_sms() {
    mmcli -m $MODEM_ID --messaging-delete-sms=$1 &>/dev/null && echo "删除短信 $1" || echo "删除 $1 失败"
}

# 检查mmcli是否可用
if ! command -v mmcli &> /dev/null; then
    echo "错误: 未找到mmcli命令，请确保已安装ModemManager"
    exit 1
fi

# 检查是否提供了调制解调器ID参数
if [ $# -ne 1 ]; then
    echo "用法: $0 <调制解调器ID>"
    echo "示例: $0 0"  # 传入调制解调器ID，如0
    exit 1
fi

MODEM_ID=$1

# 检查调制解调器是否存在
if ! mmcli --modem="$MODEM_ID" &> /dev/null; then
    echo "错误: 未找到ID为 $MODEM_ID 的调制解调器"
    exit 1
fi

# 批量处理所有received状态的短信
echo "正在读取调制解调器 $MODEM_ID 上所有接收的短信（received状态）..."
echo "--------------------------------------"

# 获取所有received状态的短信ID并循环处理
mmcli --modem="$MODEM_ID" --messaging-list-sms | grep '(received)' | awk -F'/' '{print $NF}' | sed 's/ .*//' | while read -r sms_id; do
    extract_sms_info "$sms_id"
done

echo "调制解调器 $MODEM_ID 上所有接收的短信处理完毕"
