// Package main 实现了一个基于 Go 语言的短信转发系统
// 该系统使用 ModemManager 的 mmcli 命令从 SIM 卡调制解调器检索短信
// 并通过 Bark 通知服务转发短信内容
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// SMS 结构体表示一条短信的完整信息
// 包含短信的ID、发送方号码、接收时间戳和短信内容
type SMS struct {
	ID        string // 短信在系统中的唯一标识符
	Sender    string // 发送方的电话号码
	Timestamp string // 短信接收的时间戳
	Content   string // 短信的文本内容
}

// BarkRequest 表示发送到 Bark API 的请求数据结构
// Bark 是一个 iOS 推送通知服务
type BarkRequest struct {
	Body     string `json:"body"`               // 通知的主体内容
	Title    string `json:"title"`              // 通知的标题
	Subtitle string `json:"subtitle,omitempty"` // 可选的副标题
}

// BarkResponse 表示 Bark API 返回的响应数据结构
type BarkResponse struct {
	Code int    `json:"code"` // 响应状态码，200表示成功
	Data string `json:"data"` // 响应的附加数据
}

// SMSProcessor 短信处理器结构体
// 封装了调制解调器ID和Bark API密钥，提供短信处理的核心功能
type SMSProcessor struct {
	ModemID string // 调制解调器的ID，用于指定要操作的硬件设备
	BarkKey string // Bark 服务的 API 密钥，用于身份验证
}

// NewSMSProcessor 创建并返回一个新的短信处理器实例
// 参数:
//   - modemID: 调制解调器的ID字符串
//   - barkKey: Bark API 的密钥字符串
//
// 返回: 初始化好的 SMSProcessor 指针
func NewSMSProcessor(modemID, barkKey string) *SMSProcessor {
	return &SMSProcessor{
		ModemID: modemID,
		BarkKey: barkKey,
	}
}

// checkMMCLI 检查系统中是否安装了 mmcli 命令行工具
// mmcli 是 ModemManager 提供的命令行接口，用于与调制解调器通信
// 返回: 如果未找到 mmcli 命令则返回错误，否则返回 nil
func (sp *SMSProcessor) checkMMCLI() error {
	_, err := exec.LookPath("mmcli")
	if err != nil {
		return fmt.Errorf("错误: 未找到mmcli命令，请确保已安装ModemManager")
	}
	return nil
}

// checkModem 验证指定ID的调制解调器是否存在且可访问
// 通过执行 mmcli --modem=<ID> 命令来检查调制解调器状态
// 返回: 如果调制解调器不存在或不可访问则返回错误，否则返回 nil
func (sp *SMSProcessor) checkModem() error {
	cmd := exec.Command("mmcli", "--modem="+sp.ModemID)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("错误: 未找到ID为 %s 的调制解调器", sp.ModemID)
	}
	return nil
}

// getSMSList 获取指定调制解调器上所有处于接收状态的短信ID列表
// 执行 mmcli --modem=<ID> --messaging-list-sms 命令获取短信列表
// 使用正则表达式解析输出，提取状态为 "(received)" 的短信ID
// 返回: 短信ID字符串切片和可能的错误
func (sp *SMSProcessor) getSMSList() ([]string, error) {
	cmd := exec.Command("mmcli", "--modem="+sp.ModemID, "--messaging-list-sms")
	output, err := cmd.Output()
	if err != nil {
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

	return smsIDs, nil
}

// extractSMSInfo 从指定的短信ID提取完整的短信信息
// 执行 mmcli -s <smsID> 命令获取短信详细信息
// 解析命令输出提取发送方号码、时间戳和短信内容
// 参数: smsID - 要提取信息的短信ID
// 返回: SMS结构体指针和可能的错误
func (sp *SMSProcessor) extractSMSInfo(smsID string) (*SMS, error) {
	cmd := exec.Command("mmcli", "-s", smsID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取短信 %s 详情失败: %v", smsID, err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("警告: 未找到ID为 %s 的短信，跳过处理", smsID)
	}

	sms := &SMS{ID: smsID}
	lines := strings.Split(string(output), "\n")

	// 提取发送方号码
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

	// 处理短信内容：移除回车符并将换行符转换为字面字符串
	// 对应原脚本中的: $(echo -n "$3" | sed 's/\r//g' | sed ':a;N;$!ba;s/\n/\\n/g')
	//sms.Content = processSMSContent(rawContent)
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

	return sms, nil
}

// processSMSContent 处理短信内容，对应原脚本中的 sed 命令
// 移除回车符(\r)并将换行符转换为字面上的 \n 字符串
// 对应原脚本: $(echo -n "$3" | sed 's/\r//g' | sed ':a;N;$!ba;s/\n/\\n/g')
func processSMSContent(content string) string {
	// 移除回车符 (对应 sed 's/\r//g')
	content = strings.ReplaceAll(content, "\r\n", "\\n")
	content = strings.ReplaceAll(content, "\r", "\\n")

	// 将换行符转换为字面上的 \n 字符串 (对应 sed ':a;N;$!ba;s/\n/\\n/g')
	content = strings.ReplaceAll(content, "\n", "\\n")

	return content
}

// deleteSMS 从调制解调器中删除指定的短信
// 执行 mmcli -m <modemID> --messaging-delete-sms=<smsID> 命令
// 参数: smsID - 要删除的短信ID
// 返回: 删除成功返回 nil，失败返回错误
func (sp *SMSProcessor) deleteSMS(smsID string) error {
	cmd := exec.Command("mmcli", "-m", sp.ModemID, "--messaging-delete-sms="+smsID)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("删除短信 %s 失败: %v", smsID, err)
	}
	fmt.Printf("删除短信 %s\n", smsID)
	return nil
}

// sendToBark 将短信内容发送到 Bark 通知服务
// Bark 是一个 iOS 推送通知服务，可以将通知发送到指定的设备
// 参数: sms - 包含短信信息的 SMS 结构体指针
// 返回: 发送成功返回 nil，失败返回错误
func (sp *SMSProcessor) sendToBark(sms *SMS) error {
	// 构建通知内容，格式化标题和正文
	title := fmt.Sprintf("短信转发 %s", sms.Sender)
	body := fmt.Sprintf("%s\n\n发信电话:%s\n时间:%s", sms.Content, sms.Sender, sms.Timestamp)

	barkReq := BarkRequest{
		Body:  body,
		Title: title,
	}

	// 将请求数据序列化为 JSON 格式
	jsonData, err := json.Marshal(barkReq)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 发送 HTTP POST 请求到 Bark API
	url := fmt.Sprintf("https://api.day.app/%s", sp.BarkKey)
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送 Bark 通知失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析 Bark API 的 JSON 响应
	var barkResp BarkResponse
	if err := json.NewDecoder(resp.Body).Decode(&barkResp); err != nil {
		return fmt.Errorf("解析 Bark 响应失败: %v", err)
	}

	// 检查 Bark API 响应状态码，200 表示成功
	fmt.Printf("Bark 响应: code=%d\n", barkResp.Code)

	if barkResp.Code != 200 {
		return fmt.Errorf("错误：code非200或格式不正确")
	}

	fmt.Println("code为200，正常")
	return nil
}

// processSMS 处理单条短信的完整流程
// 包括：提取短信信息、显示详情、发送 Bark 通知、删除短信
// 参数: smsID - 要处理的短信ID
// 返回: 处理成功返回 nil，失败返回错误
func (sp *SMSProcessor) processSMS(smsID string) error {
	// 从调制解调器提取短信详细信息
	sms, err := sp.extractSMSInfo(smsID)
	if err != nil {
		return err
	}

	// 在控制台显示短信详细信息
	fmt.Println("======================================")
	fmt.Printf("短信 ID: %s\n", sms.ID)
	fmt.Printf("发送方号码: %s\n", sms.Sender)
	fmt.Printf("接收时间: %s\n", sms.Timestamp)
	fmt.Printf("短信内容: %s\n", sms.Content)
	fmt.Println("======================================")
	fmt.Println()

	// 通过 Bark 服务发送推送通知
	if err := sp.sendToBark(sms); err != nil {
		return fmt.Errorf("Bark通知异常: %v", err)
	}

	// 从调制解调器中删除已处理的短信
	if err := sp.deleteSMS(sms.ID); err != nil {
		return fmt.Errorf("删除短信失败: %v", err)
	}

	return nil
}

// ProcessAllSMS 处理指定调制解调器上所有接收状态的短信
// 这是主要的对外接口，封装了完整的短信处理流程
// 包括：环境检查、获取短信列表、逐个处理短信
// 返回: 处理成功返回 nil，失败返回错误
func (sp *SMSProcessor) ProcessAllSMS() error {
	// 检查前置条件：mmcli 命令和调制解调器可用性
	if err := sp.checkMMCLI(); err != nil {
		return err
	}

	if err := sp.checkModem(); err != nil {
		return err
	}

	// 从调制解调器获取所有接收状态的短信ID列表
	smsIDs, err := sp.getSMSList()
	if err != nil {
		return err
	}
	if len(smsIDs) == 0 {
		// 如果没有短信，直接返回
		fmt.Printf("调制解调器 %s 上没有接收状态的短信\n", sp.ModemID)
		return nil
	}

	fmt.Printf("正在读取调制解调器 %s 上所有接收的短信（received状态）...\n", sp.ModemID)
	fmt.Println("--------------------------------------")

	// 逐个处理每条短信，失败时记录错误但继续处理其他短信
	for _, smsID := range smsIDs {
		if err := sp.processSMS(smsID); err != nil {
			log.Printf("处理短信 %s 失败: %v\n", smsID, err)
			continue
		}
	}

	fmt.Printf("调制解调器 %s 上所有接收的短信处理完毕\n", sp.ModemID)
	return nil
}

// main 函数是程序的入口点
// 解析命令行参数，创建短信处理器并开始处理短信
func main() {
	// 检查命令行参数数量，必须提供调制解调器ID和Bark密钥
	if len(os.Args) != 3 {
		fmt.Printf("用法: %s <调制解调器ID> <Bark密钥>\n", os.Args[0])
		fmt.Println("示例: ./sim-sms-forward 0 xxxxx")
		os.Exit(1)
	}

	// 从命令行参数获取调制解调器ID和Bark密钥
	modemID := os.Args[1]
	barkKey := os.Args[2]

	// 验证调制解调器ID是否为数字
	if _, err := strconv.Atoi(modemID); err != nil {
		log.Fatalf("调制解调器ID必须是数字: %v", err)
	}

	// 创建短信处理器实例
	processor := NewSMSProcessor(modemID, barkKey)

	// 开始处理所有短信
	if err := processor.ProcessAllSMS(); err != nil {
		log.Fatalf("处理短信失败: %v", err)
	}
}
