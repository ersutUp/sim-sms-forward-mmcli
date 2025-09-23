# CLAUDE.md

这个文件为 Claude Code (claude.ai/code) 在此代码库中工作提供指导。

使用中文与我沟通

尽量多添加注释

## 项目概述

这是一个基于 Go 的短信转发系统，使用 ModemManager 的 `mmcli` 命令从 SIM 卡调制解调器检索短信，并通过 Bark 通知服务转发。

### Go 应用程序结构
- **main.go**: 主程序入口，处理命令行参数/配置文件并启动短信处理器
- **pkg/config**: 配置管理包，处理配置文件的解析、验证和加载
- **pkg/types**: 数据类型定义包，包含 SMS、BarkRequest、BarkResponse 结构体
- **pkg/modem**: 调制解调器操作包，处理 mmcli 命令交互
- **pkg/notification**: 通知服务包，处理 Bark API 通信
- **pkg/processor**: 短信处理器包，协调整个短信处理流程

### 传统脚本（保留）
- **mmcli_get_sms.sh**: 原始短信处理脚本
- **simForward.sh**: 原始通知转发脚本

## 命令

### 构建和运行
```bash
# 构建 Go 应用程序
go build -o sim-sms-forward main.go

# 使用配置文件运行（推荐方式）
./sim-sms-forward config.json

# 兼容原有命令行参数方式
./sim-sms-forward <调制解调器ID> <Bark密钥>

# 示例
./sim-sms-forward config.example.json
./sim-sms-forward 0 your_bark_key_here

# 运行短信处理脚本（需要调制解调器 ID 参数）
./mmcli_get_sms.sh <modem_id>
```

### 测试和开发
```bash
# 测试 Go 代码格式化
go fmt ./...

# 运行 Go 模块整理
go mod tidy

# 编译检查
go build

# 使脚本可执行
chmod +x mmcli_get_sms.sh simForward.sh build.sh
```

### 跨平台打包
```bash
# 使用 Makefile (推荐)
make build-all      # 构建所有支持的平台
make build-main     # 构建主要平台 (Linux, Windows, macOS)
make build          # 构建当前平台
make clean          # 清理构建文件

# 直接运行脚本
./build.sh          # Linux/macOS 打包脚本
build.bat           # Windows 打包脚本

# 其他有用的命令
make help           # 显示所有可用命令
make dev-setup      # 开发环境准备
make release        # 发布准备
```

### 支持的平台
构建脚本支持以下平台和架构：
- Linux: amd64, arm64, arm
- Windows: amd64, arm64
- macOS: amd64, arm64 (Intel & Apple Silicon)
- FreeBSD: amd64
- OpenBSD: amd64

## 架构

### Go 应用程序架构
采用分层架构，各包职责明确：

1. **config 包**: 管理配置文件的解析、验证和加载
2. **types 包**: 定义系统中使用的数据结构
3. **modem 包**: 封装与 ModemManager 的交互逻辑
4. **notification 包**: 封装与 Bark API 的通信逻辑  
5. **processor 包**: 协调各组件，实现完整的短信处理流程
6. **main 包**: 程序入口，参数解析和启动

### 短信处理流程
1. 检查 mmcli 命令可用性和调制解调器状态
2. 获取调制解调器上所有接收状态的短信ID列表
3. 逐个处理每条短信：
   - 提取短信详细信息（发送方、时间戳、内容）
   - 显示短信详情
   - 根据配置决定是否发送 Bark 推送通知
   - 根据配置决定是否从调制解调器删除已处理的短信

### 关键组件
- **ModemManager 集成**: 使用 `mmcli` 命令行工具与蜂窝调制解调器接口
- **Bark API 集成**: 通过 Bark 服务 (api.day.app) 发送推送通知
- **配置系统**: 支持 JSON 配置文件和功能开关控制
- **日志系统**: 自动在可执行文件目录下创建日志文件，使用+8时区24小时制时间格式
- **错误处理**: 包含完整的错误处理和日志记录
- **包结构**: 模块化设计，便于维护和扩展

### 配置

#### 配置文件方式（推荐）
应用程序支持 JSON 格式的配置文件，包含以下配置项：
- `modem_id`: 调制解调器ID
- `bark_key`: Bark API 密钥
- `delete_after_forward`: 删除开关，控制是否在转发后删除短信
- `enable_bark`: Bark开关，控制是否启用 Bark 通知
- `sleep_duration`: 检查间隔时间（秒）
- `log_level`: 日志级别 (DEBUG, INFO, WARN, ERROR)

#### 配置文件示例
```json
{
  "modem_id": "0",
  "bark_key": "your_bark_key_here",
  "delete_after_forward": true,
  "enable_bark": true,
  "sleep_duration": 3,
  "log_level": "INFO"
}
```

#### 兼容性
- 仍然支持通过命令行参数传递调制解调器 ID 和 Bark API 密钥
- 支持中文显示和日志输出

### 依赖项
- Go 1.21+ 
- ModemManager (`mmcli` 命令)
- 网络连接（用于 Bark API 请求）