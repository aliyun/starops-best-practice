// chat.go — 对话与 SSE 请求入口
// 职责：实现 Chat / ChatWithVariables / Interact / Stop 系列对话入口。
// 不做：不处理事件解析（→event.go）、不处理重连（→stream_sse.go）、不做终端交互（→pkg/interactive）。
package client

import (
	"context"
	"fmt"
	"time"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

const (
	defaultTimeZone = "Asia/Shanghai"
	defaultLanguage = "zh"
)

// Chat 开始 SSE 对话（基础版本）
func (c *AgentClient) Chat(ctx context.Context, threadID, message string) <-chan *types.ChatEvent {
	now := time.Now().Unix()
	variables := map[string]any{
		"workspace": c.config.Workspace,
		"region":    c.config.Region,
		"language":  defaultLanguage,
		"timeZone":  defaultTimeZone,
		"timeStamp": fmt.Sprintf("%d", now),
		"startTime": fmt.Sprintf("%d", now-15*60),
		"endTime":   fmt.Sprintf("%d", now),
	}
	return c.ChatWithVariables(ctx, threadID, message, variables)
}

// ensureDefaultVariables 确保变量中包含必要字段的默认值
func (c *AgentClient) ensureDefaultVariables(variables map[string]any) {
	if variables == nil {
		return
	}
	now := time.Now().Unix()
	if _, ok := variables["workspace"]; !ok {
		variables["workspace"] = c.config.Workspace
	}
	if _, ok := variables["region"]; !ok {
		variables["region"] = c.config.Region
	}
	if _, ok := variables["language"]; !ok {
		variables["language"] = defaultLanguage
	}
	if _, ok := variables["timeZone"]; !ok {
		variables["timeZone"] = defaultTimeZone
	}
	if _, ok := variables["timeStamp"]; !ok {
		variables["timeStamp"] = fmt.Sprintf("%d", now)
	}
	if _, ok := variables["startTime"]; !ok {
		variables["startTime"] = fmt.Sprintf("%d", now-15*60)
	}
	if _, ok := variables["endTime"]; !ok {
		variables["endTime"] = fmt.Sprintf("%d", now)
	}
}

// ChatWithVariables 开始 SSE 对话（支持自定义 variables）
func (c *AgentClient) ChatWithVariables(ctx context.Context, threadID, message string, variables map[string]any) <-chan *types.ChatEvent {
	events := make(chan *types.ChatEvent, 10)

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
func (c *AgentClient) Interact(ctx context.Context, threadID string, userInteractive string, baseVariables map[string]any) <-chan *types.ChatEvent {
	events := make(chan *types.ChatEvent, 10)

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
		req.SetVariables(variables)

		c.streamSSE(ctx, req, events)
	}()

	return events
}

// Stop 发送停止请求，中断正在进行的对话
// 使用带超时的同步调用，避免服务端不响应时阻塞退出流程
func (c *AgentClient) Stop(ctx context.Context, threadID string, variables map[string]any) error {
	if c.config.MockMode {
		return nil
	}
	if variables == nil {
		variables = make(map[string]any)
	}
	c.ensureDefaultVariables(variables)

	req := &starops.CreateChatRequest{}
	req.SetAction("stop")
	req.SetThreadId(threadID)
	req.SetDigitalEmployeeName(c.config.EmployeeName)
	req.SetVariables(variables)

	// 设置超时，避免服务端不响应导致退出流程卡住
	runtime := &dara.RuntimeOptions{}
	runtime.SetConnectTimeout(3000)
	runtime.SetReadTimeout(5000)

	_, err := c.client.CreateChatWithOptions(req, make(map[string]*string), runtime)
	return err
}
