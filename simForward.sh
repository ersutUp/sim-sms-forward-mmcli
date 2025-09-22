#!/bin/bash
# 接受命令行参数
# 发信电话号码
telnum=$1
# 短信收信日期
smsdate=$2
# 短信内容
smscontent=$(echo -n "$3" | sed 's/\r//g' | sed ':a;N;$!ba;s/\n/\\n/g')
echo "smscontent: $smscontent"

smscode=$4
smscodefrom=$5
devicename=$6

bark_key='xxxx'

pushtitle='短信转发 '$telnum
pushcontent="短信内容:$smscontent\n发信电话:$telnum\n时间:$smsdate\n转发设备:$devicename"

# 初始化JSON基础部分（必选字段）
json_data="{\"body\": \"$pushcontent\", \"title\": \"$pushtitle\"}"

# 如果smscode存在，则添加subtitle字段
if [ -n "$smscode" ]; then
  pushsubtitle="$smscodefrom$smscode"
  # 拼接subtitle（注意添加逗号分隔）
  json_data="{\"body\": \"$pushcontent\", \"title\": \"$pushtitle\", \"subtitle\": \"$pushsubtitle\"}"
fi

# 发送请求
res=$(curl -X "POST" "https://api.day.app/$bark_key"  -H 'Content-Type: application/json; charset=utf-8' -d "$json_data" 2>/dev/null)

echo "res: $res"
# 正则匹配 "code":200（允许键值间有空格）
if [[ "$res" =~ "\"code\":200" ]]; then
    echo "code为200，正常"
    exit 0  # 正常退出
else
    echo "错误：code非200或格式不正确"
    exit 1  # 异常退出
fi