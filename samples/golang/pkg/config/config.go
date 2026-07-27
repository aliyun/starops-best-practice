// Package config — 应用配置与环境变量加载
// 职责：定义 Config / RetryConfig 结构体、从 .env 与阿里云凭据链加载运行参数。
// 不做：不创建客户端、不处理 SSE、不处理重连（→pkg/client）。
// 依赖：godotenv、阿里云 credentials-go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/logger"
)

const (
	defaultDigitalEmployeeName = "apsara-ops"
	defaultRegion          = "cn-beijing"
	defaultMaxRetries      = 10
	defaultIdleTimeout     = 60 * time.Second
	defaultInitialBackoff = 1 * time.Second
	defaultMaxBackoff      = 30 * time.Second
	defaultBackoffFactor   = 2.0
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int           // 最大重试次数，默认10
	InitialBackoff time.Duration // 初始退避时间，默认1s
	MaxBackoff     time.Duration // 最大退避时间，默认30s
	BackoffFactor  float64       // 退避系数，默认2.0
	IdleTimeout    time.Duration // 空闲超时：超过此时长未收到任何消息视为连接中断，默认 60s
}

// DefaultRetryConfig 返回默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     defaultMaxRetries,
		InitialBackoff: defaultInitialBackoff,
		MaxBackoff:     defaultMaxBackoff,
		BackoffFactor:  defaultBackoffFactor,
		IdleTimeout:    defaultIdleTimeout,
	}
}

// LoadRetryConfigFromEnv 从环境变量加载重试配置
func LoadRetryConfigFromEnv() *RetryConfig {
	cfg := DefaultRetryConfig()

	if maxRetries := os.Getenv("STAROPS_MAX_RETRIES"); maxRetries != "" {
		if n, err := strconv.Atoi(maxRetries); err == nil && n > 0 {
			cfg.MaxRetries = n
		}
	}

	if idleTimeout := os.Getenv("STAROPS_IDLE_TIMEOUT"); idleTimeout != "" {
		if n, err := strconv.Atoi(idleTimeout); err == nil && n > 0 {
			cfg.IdleTimeout = time.Duration(n) * time.Second
		}
	}

	return cfg
}

// Config 应用配置
type Config struct {
	Workspace            string
	Endpoint             string
	Region               string
	AccessKeyID          string
	AccessKeySecret      string
	SecurityToken        string
	EmployeeName         string
	RetryConfig          *RetryConfig // 重试配置，nil 时使用默认配置，默认启用重试
	SimulateNetworkError bool         // 模拟网络断连，用于测试重试逻辑
	MockMode             bool         // Mock 模式：回放预录制的 SSE 事件流
	RecordMode           bool         // 录制模式：旁路捕获 SSE 事件写入文件
	MockFile             string       // Mock/Record 文件路径
	MockInput            string       // Mock 交互输入：自动回复交互事件（如 "yes"/"no"）
}

// LoadConfigFromEnv 从环境变量加载配置
// AK/SK 通过阿里云默认凭据链获取（凭据链已包含环境变量、配置文件、RAM 角色等来源）
func LoadConfigFromEnv() (*Config, error) {
	_ = godotenv.Load()

	// 直接通过阿里云默认凭据链获取 AK/SK
	// 默认凭据链已包含环境变量、配置文件、RAM 角色等来源，无需单独读取环境变量
	credential, err:= LoadCredentialsFromChain()
	if err != nil {
		logger.Default().Warn("默认凭据链获取凭证失败", map[string]any{"error": err.Error()})
	}

	cfg := &Config{
		Workspace:       os.Getenv("STAROPS_WORKSPACE"),
		Endpoint:        os.Getenv("STAROPS_ENDPOINT"),
		Region:          os.Getenv("STAROPS_REGION"),
		AccessKeyID:     credential.AccessKeyID,
		AccessKeySecret: credential.AccessKeySecret,
		SecurityToken:   credential.SecurityToken,
		EmployeeName:    os.Getenv("STAROPS_EMPLOYEE_NAME"),
	}

	// 验证必需字段
	var missingVars []string
	if cfg.Endpoint == "" {
		missingVars = append(missingVars, "STAROPS_ENDPOINT")
	}
	if cfg.AccessKeyID == "" {
		missingVars = append(missingVars, "ALIBABA_CLOUD_ACCESS_KEY_ID")
	}
	if cfg.AccessKeySecret == "" {
		missingVars = append(missingVars, "ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("缺少必需的配置: %s", strings.Join(missingVars, ", "))
	}

	if cfg.EmployeeName == "" {
		cfg.EmployeeName = defaultDigitalEmployeeName
	}
	if cfg.Region == "" {
		cfg.Region = defaultRegion
	}

	cfg.RetryConfig = LoadRetryConfigFromEnv()

	return cfg, nil
}
