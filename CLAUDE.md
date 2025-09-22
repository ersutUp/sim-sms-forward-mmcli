# CLAUDE.md

这个文件为 Claude Code (claude.ai/code) 在此代码库中工作提供指导。

使用中文与我沟通

## 项目概述

这是一个基于 Go 的短信转发系统，使用 ModemManager 的 `mmcli` 命令从 SIM 卡调制解调器检索短信，并通过 Bark 通知服务转发。系统组成：

- **main.go**: 简单的 Go 启动模板（目前包含示例代码）
- **mmcli_get_sms.sh**: 核心短信处理脚本，从调制解调器检索短信，提取消息详细信息，并调用转发脚本
- **simForward.sh**: 通知转发脚本，将短信内容发送到 Bark API 服务

## 命令

### 构建和运行
```bash
# 构建 Go 应用程序
go build -o sim-sms-forward main.go

# 运行 Go 应用程序
go run main.go

# 运行短信处理脚本（需要调制解调器 ID 参数）
./mmcli_get_sms.sh <modem_id>
```

### 测试和开发
```bash
# 测试 Go 代码格式化
go fmt ./...

# 运行 Go 模块整理
go mod tidy

# 使脚本可执行
chmod +x mmcli_get_sms.sh simForward.sh
```

## 架构

### 短信处理流程
1. **mmcli_get_sms.sh** 查询 ModemManager 获取指定调制解调器上的接收短信
2. 对于每条短信，使用 mmcli 命令提取发送者号码、时间戳和内容
3. 使用提取的短信数据调用 **simForward.sh**
4. **simForward.sh** 将数据格式化为 JSON 并发送到 Bark 通知服务
5. 通知成功后，从调制解调器删除原始短信

### 关键组件
- **ModemManager 集成**: 使用 `mmcli` 命令行工具与蜂窝调制解调器接口
- **Bark API 集成**: 通过 Bark 服务 (api.day.app) 发送推送通知
- **错误处理**: 脚本包含调制解调器存在性、短信处理和 API 响应的验证

### 配置
- Bark API 密钥在 simForward.sh 中硬编码（`bark_key` 变量）
- 调制解调器 ID 必须作为命令行参数提供给 mmcli_get_sms.sh
- 脚本使用中文，带有一些英文注释

### 依赖项
- ModemManager (`mmcli` 命令)
- curl (用于 API 请求)
- 标准 Unix 工具 (bash, awk, sed, grep)