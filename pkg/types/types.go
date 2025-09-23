// Package types 定义了短信转发系统中使用的数据结构
package types

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
