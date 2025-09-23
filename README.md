# sim-sms-forward-mmcli

基于 Go 语言的短信转发系统，使用 ModemManager 的 mmcli 命令从 SIM 卡调制解调器检索短信，并通过 Bark 通知服务转发短信内容。

> 大部分代码由 Claude 写的

## 项目介绍

该系统可以实时监控连接在设备上的 SIM 卡收到的短信，并将短信内容通过 Bark 服务推送到 iOS 设备上。系统支持配置文件和命令行参数两种启动方式，提供了完整的构建、运行和监控脚本，确保服务稳定运行。

### 主要特性

- 🚀 **实时监控**: 持续监控 SIM 卡短信，及时转发
- 🔧 **灵活配置**: 支持 JSON 配置文件和命令行参数两种方式
- 📱 **Bark 集成**: 通过 Bark 服务推送到 iOS 设备
- 🗂️ **模块化设计**: 采用清晰的包结构，便于维护和扩展
- 🔄 **自动重启**: 内置看门狗脚本，确保服务稳定运行
- 📊 **完整日志**: 自动生成详细日志，便于问题诊断

**未来计划**: 支持更多消息推送平台

## 项目的起源

由于我手机单卡，所以接验证码的SIM卡一直要装在另一个手机里，天天需要多带一个手机，直到看到了[这篇文章](https://mp.weixin.qq.com/s/rxjZuqiw5O4BSa-BwegVyg)，完美的解决了我这个问题。

但是一直不太稳定，[DbusSmsForwardCPlus](https://github.com/lkiuyu/DbusSmsForwardCPlus)不知道为什么偶尔会挂掉，后边写了看门狗脚本，程序是可以保证正常启动了。又发现即使在运行有时候也接收不到转发的短信。

所以写了这个项目，目前转发使用该项目，发送短信依旧是DbusSmsForwardCPlus

## 系统要求

### 运行环境

- **硬件**: 支持的 SIM 卡调制解调器设备
- **网络**: 稳定的互联网连接（用于 Bark 推送）

### 软件依赖

- **ModemManager**: 版本 1.22.0+（经过测试的版本）
- **mmcli**: ModemManager 命令行工具
- **Go**: 1.21+ (仅编译时需要)

### 安装 ModemManager (Ubuntu/Debian)

```bash
# 安装 ModemManager
sudo apt-get update
sudo apt-get install modemmanager

# 验证安装
mmcli --version
mmcli --list-modems
```

## 快速开始

#### 1. 克隆项目

```bash
git clone https://github.com/ersutUp/sim-sms-forward-mmcli.git
cd sim-sms-forward-mmcli
```

#### 2. 编译项目

使用 Makefile (推荐)：

```bash
# 本地构建
make build

# 跨平台构建
make build-all

# 构建主要平台
make build-main

# 查看所有可用命令
make help
```

或使用构建脚本：

```bash
chmod +x build.sh
./build.sh
```

编译后的可执行文件将位于 `dist` 目录下。

#### 3. 配置文件设置

复制示例配置文件并编辑：

```bash
cp conf/config.example.json config.json
nano config.json  # 使用您喜欢的编辑器
```

编辑 `config.json` 文件，配置您的参数：

```json
{
  "modem_id": "0",                    // 调制解调器ID，通过 mmcli --list-modems 查看
  "bark_key": "your_bark_key_here",   // Bark服务密钥
  "enable_bark": true,                // 是否启用Bark通知
  "sleep_duration": 3,                // 检查短信的间隔时间（秒）
}
```

#### 4. 启动

```shell
cp dist/对应平台包 sim-sms-forward
./sim-sms-forward config.json
```

## 使用方法

### 获取调制解调器信息

在配置前，您需要确定调制解调器的 ID：

```bash
# 列出所有调制解调器
mmcli --list-modems

# 查看特定调制解调器详情
mmcli --modem=0

# 查看 SIM 卡状态
mmcli --modem=0 --sim=0
```

### 启动程序

程序支持多种启动方式：

#### 1. 使用配置文件启动 (推荐)

```bash
./sim-sms-forward config.json
```

#### 2. 使用命令行参数启动

```bash
./sim-sms-forward <modem_id> <bark_key>
```

示例：
```bash
./sim-sms-forward 0 your_bark_key_here
```

#### 3. 使用运行脚本

项目提供了 `run.sh` 脚本，方便管理和启动程序：

```bash
chmod +x run.sh
./run.sh [程序根目录]  # 默认：/home/sim-sms-forward-mmcli
```

> 注意：使用运行脚本时，程序名必须设置为 `sim-sms-forward`

#### 4. 使用 Makefile 运行

```bash
# 构建并运行 (需要 config.json)
make run
```

## 部署

### 必备的文件

复制 打包的程序、配置文件、watchdog.sh、run.sh到`/home/sim-sms-forward-mmcli`目录

赋予运行权限

```shell
# 赋予脚本执行权限
chmod +x ./*.sh
# 赋予程序执行权限
chmod +x sim-sms-forward
```

### 看门狗脚本

项目提供了 `watchdog.sh` 脚本，确保程序持续稳定运行：

```bash
# 赋予执行权限
chmod +x watchdog.sh

# 手动运行看门狗检查
./watchdog.sh

# 指定自定义程序目录，默认程序目录 /home/sim-sms-forward-mmcli
./watchdog.sh /path/to/your/program/directory
```

### 配置 Cron 定时任务

设置定时检查，确保服务不间断运行：

```bash
# 编辑 cron 配置
crontab -e

# 添加以下内容（每2分钟检查一次）
*/2 * * * * (/bin/bash /home/sim-sms-forward-mmcli/watchdog.sh > /dev/null)
```

## 配置参数详解

### 完整配置选项

| 配置项 | 类型 | 说明 | 默认值 | 必填 |
|--------|------|------|--------|------|
| `modem_id` | 字符串 | 调制解调器的ID，通过 `mmcli --list-modems` 获取 | `"0"` | ✅ |
| `bark_key` | 字符串 | Bark 服务的 API 密钥，用于推送通知 | 无 | ✅ |
| `enable_bark` | 布尔值 | 是否启用 Bark 推送通知功能 | `true` | ❌ |
| `sleep_duration` | 整数 | 两次检查短信之间的间隔时间（秒） | `3` | ❌ |

### 获取 Bark 密钥

1. 在 iOS 设备上安装 [Bark 应用](https://apps.apple.com/app/bark-customed-notifications/id1403753865)
2. 打开应用，复制显示的密钥
3. 将密钥填入配置文件的 `bark_key` 字段

### 配置示例

```json
{
  "modem_id": "0",
  "bark_key": "aBcDeFgHiJkLmN",
  "enable_bark": true,
  "sleep_duration": 3
}
```

## 日志管理

### 日志文件位置

程序运行时会自动创建日志文件：

```
程序目录/
├── logs/
│   ├── sms-forward-2024-01-15.log    # 主程序日志
│   ├── sms-forward-2024-01-16.log    # 按日期分割
│   └── watchdog.log                  # 看门狗脚本日志
├── sim-sms-forward                   # 可执行文件
└── config.json                       # 配置文件
```

### 日志查看命令

```bash
# 查看今天的日志
tail -f logs/sms-forward-$(date +%Y-%m-%d).log

# 查看最近的错误日志
grep "ERROR" logs/sms-forward-*.log | tail -20
```

### 日志轮转(这里不需要，程序中有了)

建议配置 logrotate 来管理日志文件：

```bash
# 创建 logrotate 配置
sudo nano /etc/logrotate.d/sim-sms-forward
```

配置内容：
```
/path/to/sim-sms-forward/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}

```

## 故障排除

### 常见问题和解决方案

#### 1. 调制解调器相关问题

**问题**: 未找到 mmcli 命令
```bash
# 解决方案
sudo apt-get update
sudo apt-get install modemmanager

# 验证安装
mmcli --version
```

**问题**: 未找到调制解调器
```bash
# 检查调制解调器连接状态
mmcli --list-modems

# 检查 SIM 卡状态
mmcli --modem=0 --sim=0

# 如果调制解调器被禁用，启用它
mmcli --modem=0 --enable

# 重启 ModemManager 服务
sudo systemctl restart ModemManager
```

#### 2. Bark 通知问题

**问题**: Bark 通知发送失败
- ✅ 检查 `bark_key` 是否正确
- ✅ 确保网络连接正常  
- ✅ 检查 Bark 服务器状态（访问 https://api.day.app）
- ✅ 验证 iOS 设备上的 Bark 应用是否正常

**问题**: 收不到推送通知

```bash
# 测试 Bark API 连接
curl -X POST "https://api.day.app/your_bark_key" \
     -H "Content-Type: application/json" \
     -d '{"title":"测试","body":"这是一条测试消息"}'
```

#### 3. 程序运行问题

**问题**: 程序无法启动
```bash
# 检查配置文件语法
cat config.json

# 查看详细启动日志
./sim-sms-forward config.json
```

### 调试技巧

#### 启用详细日志
```json
{
  "log_level": "DEBUG",
  "sleep_duration": 1
}
```

#### 手动测试调制解调器
```bash
# 测试获取短信列表
mmcli -m <modem_id> --messaging-list-sms

# 测试读取特定短信
mmcli --sms=<sms_id>

# 测试删除短信
mmcli -m <modem_id> --messaging-delete-sms=<sms_id>
```

#### 网络连接测试
```bash
# 测试 DNS 解析
nslookup api.day.app

# 测试网络连通性
ping -c 4 api.day.app

# 测试 HTTPS 连接
curl -I https://api.day.app
```

## License

[MIT](LICENSE)