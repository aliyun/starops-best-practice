// stream_record.go — 录制通路：代理真实 SDK 并旁路写入 JSONL 文件
// 职责：转发 realStream 事件到上层，同时以 JSONL 追加写入 MockFile，写入 `---` 作为流分隔。
// 不做：不做回放、不做重连（→stream_mock.go / retry.go）。
package client

import (
	"context"
	"encoding/json"
	"os"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/logger"
)

// recordStream 代理真实 SDK 并旁路写入 JSONL 文件
func (c *AgentClient) recordStream(ctx context.Context, req *starops.CreateChatRequest) (chan *starops.CreateChatResponse, chan error) {
	realYield, realYieldErr := c.realStream(ctx, req)

	f, err := os.OpenFile(c.config.MockFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Default().Warn("创建录制文件失败，降级为普通模式", map[string]any{
			"file":  c.config.MockFile,
			"error": err.Error(),
		})
		return realYield, realYieldErr
	}

	yield := make(chan *starops.CreateChatResponse, 10)
	yieldErr := make(chan error, 1)

	// 两个独立 goroutine，避免 select 竞争导致事件丢失
	go func() {
		defer close(yield)
		defer f.Close()
		encoder := json.NewEncoder(f)
		for resp := range realYield {
			evt := mockSSEEvent{
				ID:    dara.StringValue(resp.Id),
				Event: dara.StringValue(resp.Event),
			}
			if resp.Body != nil {
				evt.Data = resp.Body
			}
			encoder.Encode(evt)
			select {
			case yield <- resp:
			case <-ctx.Done():
				return
			}
		}
		f.WriteString(streamSeparator + "\n")
	}()

	go func() {
		defer close(yieldErr)
		for err := range realYieldErr {
			select {
			case yieldErr <- err:
			case <-ctx.Done():
				return
			}
		}
	}()

	return yield, yieldErr
}
