// Package client 提供 VibeOps Agent 客户端的公共实现
package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
)

// ChatOptions 对话选项
// 用于配置对话请求的各种参数
type ChatOptions struct {
	Timeout    time.Duration    // 超时时间，0 表示不设置超时
	Variables  map[string]any   // 请求变量
	OnEvent    func(*ChatEvent) // 事件回调
	SimpleMode bool             // 简洁模式，只输出最终文本
}

// Config 应用配置
type Config struct {
	Workspace            string
	Endpoint             string
	Region               string
	AccessKeyID          string
	AccessKeySecret      string
	EmployeeName         string
	RetryConfig          *RetryConfig // 重试配置，nil 时使用默认配置，默认启用重试
	SimulateNetworkError bool         // 模拟网络断连，用于测试重试逻辑
}

// LoadConfigFromEnv 从环境变量加载配置
// AK/SK 通过阿里云默认凭据链获取（凭据链已包含环境变量、配置文件、RAM 角色等来源）
func LoadConfigFromEnv() (*Config, error) {
	// 直接通过阿里云默认凭据链获取 AK/SK
	// 默认凭据链已包含环境变量、配置文件、RAM 角色等来源，无需单独读取环境变量
	akID, akSecret, err := LoadCredentialsFromChain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: 默认凭据链获取凭证失败: %v\n", err)
	}

	cfg := &Config{
		Workspace:       os.Getenv("VIBEOPS_WORKSPACE"),
		Endpoint:        os.Getenv("VIBEOPS_ENDPOINT"),
		Region:          os.Getenv("VIBEOPS_REGION"),
		AccessKeyID:     akID,
		AccessKeySecret: akSecret,
		EmployeeName:    os.Getenv("VIBEOPS_EMPLOYEE_NAME"),
	}

	// 验证必需字段
	var missingVars []string
	if cfg.Endpoint == "" {
		missingVars = append(missingVars, "VIBEOPS_ENDPOINT")
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
		cfg.EmployeeName = "default"
	}
	if cfg.Region == "" {
		cfg.Region = "cn-hangzhou"
	}

	cfg.RetryConfig = LoadRetryConfigFromEnv()

	return cfg, nil
}

// AgentClient Agent客户端
type AgentClient struct {
	client *starops.Client
	config *Config
}

// NewAgentClient 创建Agent客户端
func NewAgentClient(cfg *Config) (*AgentClient, error) {
	openApiConfig := &openapi.Config{}
	openApiConfig.SetAccessKeyId(cfg.AccessKeyID)
	openApiConfig.SetAccessKeySecret(cfg.AccessKeySecret)
	openApiConfig.SetEndpoint(cfg.Endpoint)
	openApiConfig.SetSignatureVersion("v3")

	staropsClient, err := starops.NewClient(openApiConfig)
	if err != nil {
		return nil, fmt.Errorf("创建STAROps客户端失败: %w", err)
	}

	return &AgentClient{
		client: staropsClient,
		config: cfg,
	}, nil
}

// Config 获取配置
func (c *AgentClient) Config() *Config {
	return c.config
}

// CreateThread 创建会话
func (c *AgentClient) CreateThread(ctx context.Context, attributes ...map[string]string) (string, error) {
	req := &starops.CreateThreadRequest{}
	req.SetTitle(fmt.Sprintf("Chat-%d", time.Now().Unix()))

	variables := &starops.CreateThreadRequestVariables{}
	variables.SetWorkspace(c.config.Workspace)
	req.SetVariables(variables)

	if len(attributes) > 0 && attributes[0] != nil {
		attrs := make(map[string]*string)
		for k, v := range attributes[0] {
			attrs[k] = dara.String(v)
		}
		req.SetAttributes(attrs)
	}

	resp, err := c.client.CreateThread(dara.String(c.config.EmployeeName), req)
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}

	if resp.Body == nil || resp.Body.ThreadId == nil {
		return "", fmt.Errorf("无效响应: 缺少ThreadID")
	}

	return dara.StringValue(resp.Body.ThreadId), nil
}

// ChatEvent 聊天事件
type ChatEvent struct {
	Body       *starops.CreateChatResponseBody
	RawJSON    string
	StatusCode int32
	IsDone     bool
	Error      error
	Id         string
	Event      string
}

// Chat 开始SSE对话（基础版本）
func (c *AgentClient) Chat(ctx context.Context, threadID, message string) <-chan *ChatEvent {
	variables := map[string]any{
		"workspace": c.config.Workspace,
		"region":    c.config.Region,
		"language":  "zh",
		"timeZone":  "Asia/Shanghai",
		"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	return c.ChatWithVariables(ctx, threadID, message, variables)
}

// ensureDefaultVariables 确保变量中包含必要字段的默认值
func (c *AgentClient) ensureDefaultVariables(variables map[string]any) {
	if variables == nil {
		return
	}
	if _, ok := variables["workspace"]; !ok {
		variables["workspace"] = c.config.Workspace
	}
	if _, ok := variables["region"]; !ok {
		variables["region"] = c.config.Region
	}
	if _, ok := variables["language"]; !ok {
		variables["language"] = "zh"
	}
	if _, ok := variables["timeZone"]; !ok {
		variables["timeZone"] = "Asia/Shanghai"
	}
	if _, ok := variables["timeStamp"]; !ok {
		variables["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	}
}

// ChatWithVariables 开始SSE对话（支持自定义 variables）
func (c *AgentClient) ChatWithVariables(ctx context.Context, threadID, message string, variables map[string]any) <-chan *ChatEvent {
	events := make(chan *ChatEvent, 10)

	go func() {
		defer close(events)

		content := &starops.CreateChatRequestMessagesContents{}
		content.SetType("text")
		content.SetValue(message)

		msg := &starops.CreateChatRequestMessages{}
		msg.SetRole("user")
		msg.SetContents([]*starops.CreateChatRequestMessagesContents{content})

		if variables == nil {
			variables = make(map[string]any)
		}
		c.ensureDefaultVariables(variables)

		req := &starops.CreateChatRequest{}
		req.SetAction("create")
		req.SetThreadId(threadID)
		req.SetDigitalEmployeeName(c.config.EmployeeName)
		req.SetMessages([]*starops.CreateChatRequestMessages{msg})
		req.SetVariables(variables)

		c.streamSSE(ctx, req, events)
	}()

	return events
}

// Interact 发送交互响应并恢复 SSE 对话
// 使用 action="interact"，无 messages 字段，交互数据通过 variables.userInteractive 传递
func (c *AgentClient) Interact(ctx context.Context, threadID string, userInteractive string, baseVariables map[string]any) <-chan *ChatEvent {
	events := make(chan *ChatEvent, 10)

	go func() {
		defer close(events)

		variables := make(map[string]any)
		for k, v := range baseVariables {
			variables[k] = v
		}
		variables["userInteractive"] = userInteractive
		c.ensureDefaultVariables(variables)

		req := &starops.CreateChatRequest{}
		req.SetAction("interact")
		req.SetThreadId(threadID)
		req.SetDigitalEmployeeName(c.config.EmployeeName)
		// ponytail: Interact 不设置 Messages，与 create 行为不同
		req.SetVariables(variables)

		c.streamSSE(ctx, req, events)
	}()

	return events
}

// ChatWithOptions 开始SSE对话（支持选项）
// 支持超时控制、自定义变量、事件回调和简洁模式
func (c *AgentClient) ChatWithOptions(ctx context.Context, threadID, message string, opts *ChatOptions) <-chan *ChatEvent {
	// 如果没有提供选项，使用默认值
	if opts == nil {
		opts = &ChatOptions{}
	}

	// 如果设置了超时，创建带超时的 context
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}

	// 使用提供的变量或默认变量
	variables := opts.Variables
	if variables == nil {
		variables = map[string]any{
			"workspace": c.config.Workspace,
			"region":    c.config.Region,
			"language":  "zh",
			"timeZone":  "Asia/Shanghai",
			"timeStamp": fmt.Sprintf("%d", time.Now().Unix()),
		}
	}

	// 获取基础事件通道
	baseEvents := c.ChatWithVariables(ctx, threadID, message, variables)

	// 创建输出通道
	events := make(chan *ChatEvent, 10)

	go func() {
		defer close(events)
		if cancel != nil {
			defer cancel()
		}

		startTime := time.Now()

		for {
			select {
			case event, ok := <-baseEvents:
				if !ok {
					return
				}

				// 如果是 context 超时错误，包装为超时错误
				if event.Error != nil && ctx.Err() == context.DeadlineExceeded {
					elapsed := time.Since(startTime)
					event.Error = &TimeoutError{
						Duration: opts.Timeout,
						Elapsed:  elapsed,
					}
				}

				// 调用事件回调
				if opts.OnEvent != nil {
					opts.OnEvent(event)
				}

				events <- event

				if event.IsDone || event.Error != nil {
					return
				}

			case <-ctx.Done():
				// 处理超时或取消
				elapsed := time.Since(startTime)
				var err error
				if ctx.Err() == context.DeadlineExceeded {
					err = &TimeoutError{
						Duration: opts.Timeout,
						Elapsed:  elapsed,
					}
				} else {
					err = ctx.Err()
				}
				event := &ChatEvent{Error: err}
				if opts.OnEvent != nil {
					opts.OnEvent(event)
				}
				events <- event
				return
			}
		}
	}()

	return events
}

// ChatWithTimeout 带超时的对话
// 这是 ChatWithOptions 的便捷方法，只设置超时参数
func (c *AgentClient) ChatWithTimeout(ctx context.Context, threadID, message string, timeout time.Duration) <-chan *ChatEvent {
	return c.ChatWithOptions(ctx, threadID, message, &ChatOptions{
		Timeout: timeout,
	})
}

// TimeoutError 超时错误
// 包含配置的超时时间和实际经过的时间
type TimeoutError struct {
	Duration time.Duration // 配置的超时时间
	Elapsed  time.Duration // 实际经过的时间
}

// Error 实现 error 接口
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("请求超时: 配置超时 %v, 已经过 %v", e.Duration, e.Elapsed)
}

// IsTimeout 检查错误是否为超时错误
func IsTimeout(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

func isDoneMessage(resp *starops.CreateChatResponse) bool {
	if resp == nil {
		return false
	}
	// 优先使用 response 级别的 Event 字段（SSE 事件类型）
	if resp.Event != nil && *resp.Event == "done" {
		return true
	}
	// fallback: 遍历 Messages
	if resp.Body != nil {
		for _, msg := range resp.Body.Messages {
			if msg.Type != nil && *msg.Type == "done" {
				return true
			}
		}
	}
	return false
}
