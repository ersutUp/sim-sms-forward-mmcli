// Package config 提供配置文件的解析和管理功能
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config 定义应用程序的配置结构
type Config struct {
	// ModemID 调制解调器ID
	ModemID string `json:"modem_id"`
	
	// BarkKey Bark API密钥
	BarkKey string `json:"bark_key"`
	
	// BarkAPIURL Bark API服务器地址
	BarkAPIURL string `json:"bark_api_url"`
	
	// EnableBark 是否启用Bark通知
	EnableBark bool `json:"enable_bark"`
	
	// HismsgKey hismsg 密钥
	HismsgKey string `json:"hismsg_key"`
	
	// HismsgAPIURL Hismsg API服务器地址
	HismsgAPIURL string `json:"hismsg_api_url"`
	
	// EnableHismsg 是否启用hismsg通知
	EnableHismsg bool `json:"enable_hismsg"`
	
	// SleepDuration 检查间隔时间（秒）
	SleepDuration int `json:"sleep_duration"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ModemID:       "0",
		BarkKey:       "",
		BarkAPIURL:    "https://api.day.app",
		EnableBark:    true,
		HismsgKey:     "",
		HismsgAPIURL:  "https://hismsg.com/api/send",
		EnableHismsg:  false,
		SleepDuration: 3,
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON配置
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	// 将配置转换为JSON格式
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证ModemID不为空
	if c.ModemID == "" {
		return fmt.Errorf("调制解调器ID不能为空")
	}

	// 如果启用Bark，验证BarkKey和BarkAPIURL不为空
	if c.EnableBark {
		if c.BarkKey == "" {
			return fmt.Errorf("启用Bark通知时，Bark密钥不能为空")
		}
		if c.BarkAPIURL == "" {
			return fmt.Errorf("启用Bark通知时，Bark API URL不能为空")
		}
	}

	// 如果启用Hismsg，验证HismsgKey和HismsgAPIURL不为空
	if c.EnableHismsg {
		if c.HismsgKey == "" {
			return fmt.Errorf("启用Hismsg通知时，Hismsg密钥不能为空")
		}
		if c.HismsgAPIURL == "" {
			return fmt.Errorf("启用Hismsg通知时，Hismsg API URL不能为空")
		}
	}

	// 验证休眠时间大于0
	if c.SleepDuration <= 0 {
		return fmt.Errorf("休眠时间必须大于0秒")
	}

	return nil
}

// GetSleepDuration 返回休眠时间的Duration对象
func (c *Config) GetSleepDuration() time.Duration {
	return time.Duration(c.SleepDuration) * time.Second
}

// MaskBarkKey 对Bark密钥进行脱敏处理
func (c *Config) MaskBarkKey() string {
	if c.BarkKey == "" {
		return ""
	}
	if len(c.BarkKey) <= 8 {
		return "****"
	}
	return c.BarkKey[:4] + "****" + c.BarkKey[len(c.BarkKey)-4:]
}

// MaskHismsgKey 对Hismsg密钥进行脱敏处理
func (c *Config) MaskHismsgKey() string {
	if c.HismsgKey == "" {
		return ""
	}
	if len(c.HismsgKey) <= 8 {
		return "****"
	}
	return c.HismsgKey[:4] + "****" + c.HismsgKey[len(c.HismsgKey)-4:]
}
