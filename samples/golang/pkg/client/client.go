// Package client — STAROps Agent SDK
// 职责：定义 AgentClient，聚合 Chat / Interact / Thread 管理入口，并处理 SSE 流与重连。
// 不做：不实现终端交互（→pkg/interactive）、不做输出格式化（→pkg/printer）。
// 依赖：starops SDK、pkg/config、pkg/errors、pkg/types
package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/config"
)

// AgentClient Agent 客户端
type AgentClient struct {
	client *starops.Client
	config *config.Config
	mock   mockState
}

// NewAgentClient 创建 Agent 客户端
func NewAgentClient(cfg *config.Config) (*AgentClient, error) {
	openApiConfig := &openapi.Config{}
	openApiConfig.SetAccessKeyId(cfg.AccessKeyID)
	openApiConfig.SetAccessKeySecret(cfg.AccessKeySecret)
	// 仅在 STS 临时凭证场景下设置 SecurityToken，长期 AK/SK 时为空不设置
	if cfg.SecurityToken != "" {
		openApiConfig.SetSecurityToken(cfg.SecurityToken)
	}
	openApiConfig.SetEndpoint(cfg.Endpoint)
	openApiConfig.SetSignatureVersion("v3")

	sdkClient, err := starops.NewClient(openApiConfig)
	if err != nil {
		return nil, fmt.Errorf("创建StarOps客户端失败: %w", err)
	}

	return &AgentClient{
		client: sdkClient,
		config: cfg,
	}, nil
}

// Config 获取配置
func (c *AgentClient) GetConfig() *config.Config {
	return c.config
}

// CreateThread 创建会话
func (c *AgentClient) CreateThread(ctx context.Context, attributes ...map[string]string) (string, error) {
	req := &starops.CreateThreadRequest{}
	req.SetTitle("New Chat")
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
