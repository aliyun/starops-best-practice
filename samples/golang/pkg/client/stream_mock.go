// stream_mock.go — Mock 通路：从 JSONL 文件回放 SSE 事件
// 职责：从预录 JSONL 文件逐行回放事件，遇到分隔行暂停并记录偏移。
// 不做：不发真实请求、不做录制（→stream.go / stream_record.go）。
package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	starops "github.com/alibabacloud-go/starops-20260428/client"
)

// mockState Mock/Record 通路的共享状态
type mockState struct {
	offset int64 // 多流回放的文件偏移量
}

// streamSeparator 多流分隔标记（Mock/Record 通路共享）
const streamSeparator = "---"

// mockSSEEvent JSONL 行格式（Mock/Record 通路共享）
type mockSSEEvent struct {
	ID    string      `json:"id"`
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// mockStream 从 JSONL 文件读取事件并回放
func (c *AgentClient) mockStream(ctx context.Context, req *starops.CreateChatRequest) (chan *starops.CreateChatResponse, chan error) {
	_ = req
	yield := make(chan *starops.CreateChatResponse, 10)
	yieldErr := make(chan error, 1)

	go func() {
		defer close(yield)
		defer close(yieldErr)

		f, err := os.Open(c.config.MockFile)
		if err != nil {
			yieldErr <- fmt.Errorf("读取 mock 文件失败: %w", err)
			return
		}
		defer f.Close()

		if c.mock.offset > 0 {
			f.Seek(c.mock.offset, 0)
		}

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
		var bytesRead int64
		for scanner.Scan() {
			line := scanner.Text()
			bytesRead += int64(len(line)) + 1

			if strings.TrimSpace(line) == streamSeparator {
				c.mock.offset += bytesRead
				return
			}
			if len(line) == 0 {
				continue
			}

			var evt mockSSEEvent
			if err := json.Unmarshal([]byte(line), &evt); err != nil {
				yieldErr <- fmt.Errorf("解析 mock 行失败: %w", err)
				return
			}

			resp := &starops.CreateChatResponse{}
			resp.SetId(evt.ID)
			resp.SetEvent(evt.Event)
			resp.SetStatusCode(200)

			if evt.Data != nil {
				body := &starops.CreateChatResponseBody{}
				dataBytes, _ := json.Marshal(evt.Data)
				json.Unmarshal(dataBytes, body)
				resp.SetBody(body)
			}

			select {
			case yield <- resp:
			case <-ctx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yieldErr <- err
		}
	}()

	return yield, yieldErr
}
