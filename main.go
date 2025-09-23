// Package main 实现了一个基于 Go 语言的短信转发系统
// 该系统使用 ModemManager 的 mmcli 命令从 SIM 卡调制解调器检索短信
// 并通过 Bark 通知服务转发短信内容
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"sim-sms-forward/pkg/config"
	"sim-sms-forward/pkg/logger"
	"sim-sms-forward/pkg/processor"
)

// main 函数是程序的入口点
// 支持两种启动方式：使用配置文件或命令行参数
func main() {
	var cfg *config.Config
	var err error

	// 支持两种启动方式
	// 1. 仅指定配置文件路径: ./sim-sms-forward config.json
	// 2. 兼容原有方式: ./sim-sms-forward <调制解调器ID> <Bark密钥>
	switch len(os.Args) {
	case 2:
		// 使用配置文件
		configPath := os.Args[1]
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("加载配置文件失败: %v\n", err)
			os.Exit(1)
		}
	case 3:
		// 兼容原有命令行参数方式
		modemID := os.Args[1]
		barkKey := os.Args[2]

		// 验证调制解调器ID是否为数字
		if _, err := strconv.Atoi(modemID); err != nil {
			fmt.Printf("调制解调器ID必须是数字: %v\n", err)
			os.Exit(1)
		}

		// 创建配置对象
		cfg = &config.Config{
			ModemID:       modemID,
			BarkKey:       barkKey,
			EnableBark:    true,
			SleepDuration: 3,
		}
	default:
		// 显示用法说明
		fmt.Printf("用法: %s <配置文件路径>\n", os.Args[0])
		fmt.Printf("      %s <调制解调器ID> <Bark密钥>\n", os.Args[0])
		fmt.Println("示例:")
		fmt.Println("  ./sim-sms-forward config.json")
		fmt.Println("  ./sim-sms-forward 0 xxxxx")
		os.Exit(1)
	}

	// 初始化日志系统
	// 获取可执行文件所在目录，并在该目录下创建 logs 子目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取可执行文件路径失败: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)
	logDir := filepath.Join(execDir, "logs")
	if err := logger.Init(logDir); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}

	// 记录程序启动日志
	logger.Info("========================================")
	logger.Info("短信转发系统启动")
	logger.Infof("调制解调器ID: %s", cfg.ModemID)
	logger.Infof("Bark密钥: %s", cfg.MaskBarkKey())
	logger.Infof("Bark开关: %v", cfg.EnableBark)
	logger.Infof("休眠时间: %d秒", cfg.SleepDuration)
	logger.Infof("日志目录: %s", logDir)
	logger.Info("========================================")

	// 创建短信处理器实例，传入配置对象
	smsProcessor := processor.NewSMSProcessorWithConfig(cfg)

	// 开始循环处理短信
	logger.Info("开始循环监控短信...")
	for {
		// 开始处理所有短信
		if err := smsProcessor.ProcessAllSMS(); err != nil {
			logger.Fatalf("处理短信失败: %v", err)
		}

		// 等待指定时间后再次检查
		//logger.Infof("等待%d秒后继续检查...", cfg.SleepDuration)
		time.Sleep(cfg.GetSleepDuration())
	}
}
