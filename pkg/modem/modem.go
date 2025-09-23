// Package modem 提供与 ModemManager 调制解调器交互的功能
package modem

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"sim-sms-forward/pkg/logger"
	"sim-sms-forward/pkg/types"
)

// Manager 调制解调器管理器
type Manager struct {
	ModemID string // 调制解调器的ID，用于指定要操作的硬件设备
}

// NewManager 创建一个新的调制解调器管理器
// 参数: modemID - 调制解调器的ID字符串
// 返回: 初始化好的 Manager 指针
func NewManager(modemID string) *Manager {
	return &Manager{
		ModemID: modemID,
	}
}

// CheckMMCLI 检查系统中是否安装了 mmcli 命令行工具
// mmcli 是 ModemManager 提供的命令行接口，用于与调制解调器通信
// 返回: 如果未找到 mmcli 命令则返回错误，否则返回 nil
func (m *Manager) CheckMMCLI() error {
	_, err := exec.LookPath("mmcli")
	if err != nil {
		logger.Errorf("未找到mmcli命令，请确保已安装ModemManager: %v", err)
		return fmt.Errorf("错误: 未找到mmcli命令，请确保已安装ModemManager")
	}
	//logger.Info("mmcli 命令检查通过")
	return nil
}

// CheckModem 验证指定ID的调制解调器是否存在且可访问
// 通过执行 mmcli --modem=<ID> 命令来检查调制解调器状态
// 返回: 如果调制解调器不存在或不可访问则返回错误，否则返回 nil
func (m *Manager) CheckModem() error {
	//logger.Infof("检查调制解调器 ID: %s", m.ModemID)
	cmd := exec.Command("mmcli", "--modem="+m.ModemID)
	err := cmd.Run()
	if err != nil {
		logger.Errorf("未找到调制解调器 ID %s: %v", m.ModemID, err)
		return fmt.Errorf("错误: 未找到ID为 %s 的调制解调器", m.ModemID)
	}
	//logger.Infof("调制解调器 %s 检查通过", m.ModemID)
	return nil
}

// GetSMSList 获取指定调制解调器上所有处于接收状态的短信ID列表
// 执行 mmcli --modem=<ID> --messaging-list-sms 命令获取短信列表
// 使用正则表达式解析输出，提取状态为 "(received)" 的短信ID
// 返回: 短信ID字符串切片和可能的错误
func (m *Manager) GetSMSList() ([]string, error) {
	//logger.Infof("获取调制解调器 %s 的短信列表", m.ModemID)
	cmd := exec.Command("mmcli", "--modem="+m.ModemID, "--messaging-list-sms")
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("获取短信列表失败: %v", err)
		return nil, fmt.Errorf("获取短信列表失败: %v", err)
	}

	// 使用正则表达式提取所有 (received) 状态的短信ID
	// 匹配格式: /org/freedesktop/ModemManager1/SMS/<数字> (received)
	re := regexp.MustCompile(`/org/freedesktop/ModemManager1/SMS/(\d+)\s+\(received\)`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	var smsIDs []string
	for _, match := range matches {
		if len(match) >= 2 {
			smsIDs = append(smsIDs, match[1])
		}
	}

	logger.Infof("找到 %d 条接收状态的短信", len(smsIDs))
	return smsIDs, nil
}

// ExtractSMSInfo 从指定的短信ID提取完整的短信信息
// 执行 mmcli -s <smsID> 命令获取短信详细信息
// 解析命令输出提取发送方号码、时间戳和短信内容
// 参数: smsID - 要提取信息的短信ID
// 返回: SMS结构体指针和可能的错误
func (m *Manager) ExtractSMSInfo(smsID string) (*types.SMS, error) {
	logger.Infof("提取短信 %s 的详细信息", smsID)
	cmd := exec.Command("mmcli", "-s", smsID)
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("获取短信 %s 详情失败: %v", smsID, err)
		return nil, fmt.Errorf("获取短信 %s 详情失败: %v", smsID, err)
	}

	if len(output) == 0 {
		logger.Errorf("未找到短信 ID %s", smsID)
		return nil, fmt.Errorf("警告: 未找到ID为 %s 的短信，跳过处理", smsID)
	}

	sms := &types.SMS{ID: smsID}
	lines := strings.Split(string(output), "\n")

	// 提取发送方号码和时间戳
	for _, line := range lines {
		if strings.Contains(line, "number:") {
			parts := strings.Split(line, "number:")
			if len(parts) >= 2 {
				sms.Sender = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "timestamp:") {
			parts := strings.Split(line, "timestamp:")
			if len(parts) >= 2 {
				sms.Timestamp = strings.TrimSpace(parts[1])
			}
		}
	}

	// 提取短信内容 - 使用更简单的方法
	contentStart := false
	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "text:") {
			contentStart = true
			parts := strings.Split(line, "text:")
			if len(parts) >= 2 {
				contentLines = append(contentLines, strings.TrimSpace(parts[1]))
			}
			continue
		}
		if contentStart {
			if strings.HasPrefix(line, "  -") || strings.TrimSpace(line) == "" {
				break
			}
			// 移除行首的管道符和空格
			cleaned := strings.TrimLeft(line, " |")
			if cleaned != "" {
				contentLines = append(contentLines, cleaned)
			}
		}
	}
	// 将多行内容合并为单个字符串
	rawContent := strings.Join(contentLines, "")
	sms.Content = rawContent

	// 为空字段设置默认值，确保数据完整性
	if sms.Sender == "" {
		sms.Sender = "未知号码"
	}
	if sms.Timestamp == "" {
		sms.Timestamp = "未知时间"
	}
	if sms.Content == "" {
		sms.Content = "无内容"
	}

	logger.Infof("成功提取短信信息 - ID: %s, 发送方: %s, 时间: %s", sms.ID, sms.Sender, sms.Timestamp)
	return sms, nil
}

// DeleteSMS 从调制解调器中删除指定的短信
// 执行 mmcli -m <modemID> --messaging-delete-sms=<smsID> 命令
// 参数: smsID - 要删除的短信ID
// 返回: 删除成功返回 nil，失败返回错误
func (m *Manager) DeleteSMS(smsID string) error {
	logger.Infof("删除短信 %s", smsID)
	cmd := exec.Command("mmcli", "-m", m.ModemID, "--messaging-delete-sms="+smsID)
	err := cmd.Run()
	if err != nil {
		logger.Errorf("删除短信 %s 失败: %v", smsID, err)
		return fmt.Errorf("删除短信 %s 失败: %v", smsID, err)
	}
	logger.Infof("成功删除短信 %s", smsID)
	return nil
}
