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

// BarkClient Bark 通知客户端
type BarkClient struct {
	APIKey string // Bark 服务的 API 密钥，用于身份验证
	APIURL string // Bark API 服务器地址
}

// NewBarkClient 创建一个新的 Bark 通知客户端
// 参数: 
//   - apiKey: Bark API 的密钥字符串
//   - apiURL: Bark API 服务器地址
// 返回: 初始化好的 BarkClient 指针
func NewBarkClient(apiKey, apiURL string) *BarkClient {
	return &BarkClient{
		APIKey: apiKey,
		APIURL: apiURL,
	}
}

// SendSMS 将短信内容发送到 Bark 通知服务
// Bark 是一个 iOS 推送通知服务，可以将通知发送到指定的设备
// 参数: sms - 包含短信信息的 SMS 结构体指针
// 返回: 发送成功返回 nil，失败返回错误
func (bc *BarkClient) SendSMS(sms *types.SMS) error {
	logger.Infof("开始发送 Bark 通知 - 短信 ID: %s, 发送方: %s", sms.ID, sms.Sender)

	// 构建通知内容，格式化标题和正文
	title := fmt.Sprintf("短信转发 %s", sms.Sender)
	body := fmt.Sprintf("%s\n\n发信电话:%s\n时间:%s", sms.Content, sms.Sender, sms.Timestamp)

	barkReq := types.BarkRequest{
		Body:  body,
		Title: title,
	}

	// 将请求数据序列化为 JSON 格式
	jsonData, err := json.Marshal(barkReq)
	if err != nil {
		logger.Errorf("JSON序列化失败: %v", err)
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 发送 HTTP POST 请求到 Bark API
	url := fmt.Sprintf("%s/%s", bc.APIURL, bc.APIKey)
	logger.Infof("发送 Bark 请求到: %s", url)

	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("发送 Bark 通知失败: %v", err)
		return fmt.Errorf("发送 Bark 通知失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析 Bark API 的 JSON 响应
	var barkResp types.BarkResponse
	if err := json.NewDecoder(resp.Body).Decode(&barkResp); err != nil {
		logger.Errorf("解析 Bark 响应失败: %v", err)
		return fmt.Errorf("解析 Bark 响应失败: %v", err)
	}

	// 检查 Bark API 响应状态码，200 表示成功
	logger.Infof("Bark 响应: code=%d", barkResp.Code)

	if barkResp.Code != 200 {
		logger.Errorf("Bark API 返回错误代码: %d", barkResp.Code)
		return fmt.Errorf("错误：code非200或格式不正确")
	}

	logger.Info("Bark 通知发送成功")
	return nil
}
