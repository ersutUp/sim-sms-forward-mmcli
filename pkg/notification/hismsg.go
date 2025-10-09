// Package notification 提供通知服务功能，主要是 Bark 推送通知
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"sim-sms-forward/pkg/logger"
	"sim-sms-forward/pkg/types"
)

// HismsgClient Hismsg 通知客户端
type HismsgClient struct {
	userKey  string // Hismsg 服务的 API 密钥，用于身份验证
	APIURL   string // Hismsg API 服务器地址
	DeviceID string // 设备标识，用于标识不同的设备来源
}

// NewHismsgClient 创建一个新的 Hismsg 通知客户端
// 参数:
//   - userKey: Hismsg API 的密钥字符串
//   - apiURL: Hismsg API 服务器地址
//   - deviceID: 设备标识，用于标识不同的设备来源
//
// 返回: 初始化好的 HismsgClient 指针
func NewHismsgClient(userKey, apiURL, deviceID string) *HismsgClient {
	return &HismsgClient{
		userKey:  userKey,
		APIURL:   apiURL,
		DeviceID: deviceID,
	}
}

// SendSMS 将短信内容发送到 Hismsg 通知服务
// Hismsg 是一个 iOS 推送通知服务，可以将通知发送到指定的设备
// 参数: sms - 包含短信信息的 SMS 结构体指针
// 返回: 发送成功返回 nil，失败返回错误
func (bc *HismsgClient) SendSMS(sms *types.SMS) error {
	logger.Infof("开始发送 Hismsg 通知 - 短信 ID: %s, 发送方: %s", sms.ID, sms.Sender)

	// 构建通知内容，格式化标题和正文
	title := fmt.Sprintf("短信转发 %s", sms.Sender)
	body := fmt.Sprintf("%s\n\n发信电话:%s\n时间:%s", sms.Content, sms.Sender, sms.Timestamp)

	HismsgReq := types.HismsgRequest{
		Content: body,
		Title:   title,
		Source:  bc.DeviceID,
		UserKey: bc.userKey,
		Tags:    []string{"短信"},
	}

	// 将请求数据序列化为 JSON 格式
	jsonData, err := json.Marshal(HismsgReq)
	if err != nil {
		logger.Errorf("JSON序列化失败: %v", err)
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 发送 HTTP POST 请求到 Hismsg API
	url := fmt.Sprintf("%s/api/message/push/send", bc.APIURL)
	logger.Infof("发送 Hismsg 请求到: %s", url)

	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("发送 Hismsg 通知失败: %v", err)
		return fmt.Errorf("发送 Hismsg 通知失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析 Hismsg API 的 JSON 响应
	var HismsgResp types.HismsgResponse
	if err := json.NewDecoder(resp.Body).Decode(&HismsgResp); err != nil {
		logger.Errorf("解析 Hismsg 响应失败: %v", err)
		return fmt.Errorf("解析 Hismsg 响应失败: %v", err)
	}

	// 检查 Hismsg API 响应状态码，200 表示成功
	logger.Infof("Hismsg 响应: code=%d", HismsgResp.Code)

	if HismsgResp.Code != 200 {
		logger.Errorf("Hismsg API 返回错误代码: %d", HismsgResp.Code)
		return fmt.Errorf("错误：code非200或格式不正确")
	}

	logger.Info("Hismsg 通知发送成功")
	return nil
}
