// event.go — ChatEvent 类型与 SSE 帧解析
// 职责：从 STAROps SDK 响应解析事件、判断 stream_done 语义。
// 不做：不做重连、不做重试、不做输出格式化（→retry.go / pkg/printer）。
package client

import (
	"encoding/json"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// parseChatEvent 从 SDK 响应解析为 ChatEvent
func parseChatEvent(resp *starops.CreateChatResponse) *types.ChatEvent {
	rawJSON := ""
	var body *types.ChatEventBody

	if resp.Body != nil {
		if jsonBytes, err := json.Marshal(resp.Body); err == nil {
			rawJSON = string(jsonBytes)
			body = &types.ChatEventBody{}
			json.Unmarshal(jsonBytes, body)
		}
	}

	return &types.ChatEvent{
		Body:       body,
		RawJSON:    rawJSON,
		StatusCode: dara.Int32Value(resp.StatusCode),
		Id:         dara.StringValue(resp.Id),
		Event:      dara.StringValue(resp.Event),
	}
}

// isStreamDoneEvent 判断事件是否为 stream_done（正常结束标志）
func isStreamDoneEvent(event *types.ChatEvent) bool {
	if event == nil || event.Body == nil {
		return false
	}
	for _, msg := range event.Body.Messages {
		if msg == nil || msg.Events == nil {
			continue
		}
		for _, evt := range msg.Events {
			if evt != nil && evt.Type == types.EventTypeStreamDone {
				return true
			}
		}
	}
	return false
}
