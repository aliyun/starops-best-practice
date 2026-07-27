// stream.go — SSE 流建立与分发
// 职责：根据配置分发到 real/mock/record 通路。
// 不做：不做上层重连（→stream_sse.go）、不解析事件（→event.go）。
package client

import (
	"context"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
)

// openSSEStream 建立单次 SSE 连接
func (c *AgentClient) openSSEStream(ctx context.Context, req *starops.CreateChatRequest) (chan *starops.CreateChatResponse, chan error) {
	if c.config.MockMode {
		return c.mockStream(ctx, req)
	}
	if c.config.RecordMode {
		return c.recordStream(ctx, req)
	}
	return c.realStream(ctx, req)
}

// realStream 通过原始 SDK 建立 SSE 连接
func (c *AgentClient) realStream(ctx context.Context, req *starops.CreateChatRequest) (chan *starops.CreateChatResponse, chan error) {
	yield := make(chan *starops.CreateChatResponse, 10)
	yieldErr := make(chan error, 1)
	runtime := &dara.RuntimeOptions{}
	runtime.SetConnectTimeout(30000)
	runtime.SetReadTimeout(300000)
	go c.client.CreateChatWithSSECtx(ctx, req, nil, runtime, yield, yieldErr)
	return yield, yieldErr
}
