// Package processor 提供短信处理的主要业务逻辑
package processor

import (
	"fmt"

	"sim-sms-forward/pkg/config"
	"sim-sms-forward/pkg/logger"
	"sim-sms-forward/pkg/modem"
	"sim-sms-forward/pkg/notification"
)

// SMSProcessor 短信处理器
// 封装了调制解调器管理器和通知客户端，提供短信处理的核心功能
type SMSProcessor struct {
	Config       *config.Config             // 配置对象
	ModemManager *modem.Manager             // 调制解调器管理器
	BarkClient   *notification.BarkClient   // Bark 通知客户端
	HismsgClient *notification.HismsgClient // hismsg 通知客户端
}

// NewSMSProcessorWithConfig 创建并返回一个使用配置对象的新短信处理器实例
// 参数:
//   - cfg: 配置对象指针
//
// 返回: 初始化好的 SMSProcessor 指针
func NewSMSProcessorWithConfig(cfg *config.Config) *SMSProcessor {
	return &SMSProcessor{
		Config:       cfg,
		ModemManager: modem.NewManager(cfg.ModemID),
		BarkClient:   notification.NewBarkClient(cfg.BarkKey, cfg.BarkAPIURL),
		HismsgClient: notification.NewHismsgClient(cfg.HismsgKey, cfg.HismsgAPIURL),
	}
}

// NewSMSProcessor 创建并返回一个新的短信处理器实例（兼容旧接口）
// 参数:
//   - modemID: 调制解调器的ID字符串
//   - barkKey: Bark API 的密钥字符串
//
// 返回: 初始化好的 SMSProcessor 指针
func NewSMSProcessor(modemID, barkKey string) *SMSProcessor {
	cfg := &config.Config{
		ModemID:       modemID,
		BarkKey:       barkKey,
		BarkAPIURL:    "https://api.day.app",
		EnableBark:    true,
		HismsgKey:     "",
		HismsgAPIURL:  "https://hismsg.com/api/send",
		EnableHismsg:  false,
		SleepDuration: 3,
	}
	return NewSMSProcessorWithConfig(cfg)
}

// processSMS 处理单条短信的完整流程
// 包括：提取短信信息、显示详情、发送 Bark 通知、删除短信
// 参数: smsID - 要处理的短信ID
// 返回: 处理成功返回 nil，失败返回错误
func (sp *SMSProcessor) processSMS(smsID string) error {
	logger.Infof("开始处理短信 ID: %s", smsID)

	// 从调制解调器提取短信详细信息
	sms, err := sp.ModemManager.ExtractSMSInfo(smsID)
	if err != nil {
		return err
	}

	// 在控制台和日志中显示短信详细信息
	logger.Info("======================================")
	logger.Infof("短信 ID: %s", sms.ID)
	logger.Infof("发送方号码: %s", sms.Sender)
	logger.Infof("接收时间: %s", sms.Timestamp)
	logger.Infof("短信内容: %s", sms.Content)
	logger.Info("======================================")

	// 通过 Bark 服务发送推送通知（根据配置开关决定是否发送）
	if sp.Config.EnableBark {
		if err := sp.BarkClient.SendSMS(sms); err != nil {
			return fmt.Errorf("Bark通知异常: %v", err)
		}
		logger.Info("Bark 通知发送成功")
	} else {
		//logger.Info("Bark 通知已禁用，跳过发送")
	}

	// 通过 Hismsg 服务发送推送通知（根据配置开关决定是否发送）
	if sp.Config.EnableHismsg {
		if err := sp.HismsgClient.SendSMS(sms); err != nil {
			return fmt.Errorf("Hismsg通知异常: %v", err)
		}
		logger.Info("Hismsg 通知发送成功")
	} else {
		//logger.Info("Hismsg 通知已禁用，跳过发送")
	}

	// 从调制解调器中删除已处理的短信
	if err := sp.ModemManager.DeleteSMS(sms.ID); err != nil {
		return fmt.Errorf("删除短信失败: %v", err)
	}
	logger.Infof("短信 %s 已从调制解调器删除", sms.ID)

	logger.Infof("短信 %s 处理完成", smsID)
	return nil
}

// ProcessAllSMS 处理指定调制解调器上所有接收状态的短信
// 这是主要的对外接口，封装了完整的短信处理流程
// 包括：环境检查、获取短信列表、逐个处理短信
// 返回: 处理成功返回 nil，失败返回错误
func (sp *SMSProcessor) ProcessAllSMS() error {
	logger.Infof("开始处理调制解调器 %s 上的所有短信", sp.ModemManager.ModemID)

	// 检查前置条件：mmcli 命令和调制解调器可用性
	if err := sp.ModemManager.CheckMMCLI(); err != nil {
		return err
	}

	if err := sp.ModemManager.CheckModem(); err != nil {
		return err
	}

	// 从调制解调器获取所有接收状态的短信ID列表
	smsIDs, err := sp.ModemManager.GetSMSList()
	if err != nil {
		return err
	}
	if len(smsIDs) == 0 {
		// 如果没有短信，直接返回
		logger.Infof("调制解调器 %s 上没有接收状态的短信", sp.ModemManager.ModemID)
		return nil
	}

	logger.Infof("正在读取调制解调器 %s 上所有接收的短信（received状态）...", sp.ModemManager.ModemID)
	logger.Info("--------------------------------------")

	// 逐个处理每条短信，失败时记录错误但继续处理其他短信
	successCount := 0
	for _, smsID := range smsIDs {
		if err := sp.processSMS(smsID); err != nil {
			logger.Errorf("处理短信 %s 失败: %v", smsID, err)
			continue
		}
		successCount++
	}

	logger.Infof("调制解调器 %s 上短信处理完毕，成功处理 %d/%d 条短信",
		sp.ModemManager.ModemID, successCount, len(smsIDs))
	return nil
}
