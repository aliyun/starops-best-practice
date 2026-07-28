package client

import (
	"encoding/json"
	"testing"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// ---------------------------------------------------------------------------
// TestChatEvent_IdAndEvent — 保留已有测试
// ---------------------------------------------------------------------------

func TestChatEvent_IdAndEvent(t *testing.T) {
	event := &types.ChatEvent{
		Id:    "evt-123",
		Event: "text",
	}
	if event.Event != "text" {
		t.Errorf("期望 event=text, 实际=%s", event.Event)
	}
	if event.Id != "evt-123" {
		t.Errorf("期望 id=evt-123, 实际=%s", event.Id)
	}
}

// ---------------------------------------------------------------------------
// TestParseChatEvent
// ---------------------------------------------------------------------------

func TestParseChatEvent(t *testing.T) {
	t.Run("nil body → Body==nil, RawJSON empty", func(t *testing.T) {
		resp := &starops.CreateChatResponse{
			StatusCode: dara.Int32(200),
			Id:         dara.String("fake-id-1"),
			Event:      dara.String("text"),
		}
		ev := parseChatEvent(resp)
		if ev.Body != nil {
			t.Errorf("期望 Body==nil, 实际 %+v", ev.Body)
		}
		if ev.RawJSON != "" {
			t.Errorf("期望 RawJSON==\"\", 实际 %q", ev.RawJSON)
		}
	})

	t.Run("body with messages → Body.Messages populated, RawJSON valid", func(t *testing.T) {
		bodyModel := &starops.CreateChatResponseBody{}
		msg := &starops.CreateChatResponseBodyMessages{}
		msg.SetRole("assistant")
		msg.SetCallId("call-1")
		bodyModel.SetMessages([]*starops.CreateChatResponseBodyMessages{msg})

		resp := &starops.CreateChatResponse{
			Body:       bodyModel,
			StatusCode: dara.Int32(200),
			Id:         dara.String("fake-id-2"),
			Event:      dara.String("message"),
		}
		ev := parseChatEvent(resp)

		if ev.Body == nil {
			t.Fatal("期望 Body!=nil")
		}
		if len(ev.Body.Messages) == 0 {
			t.Fatal("期望 Body.Messages 非空")
		}
		if ev.RawJSON == "" {
			t.Fatal("期望 RawJSON 非空")
		}
		// RawJSON should be valid JSON
		if !json.Valid([]byte(ev.RawJSON)) {
			t.Errorf("RawJSON 不是合法 JSON: %s", ev.RawJSON)
		}
	})

	t.Run("status code mapping (200)", func(t *testing.T) {
		resp := &starops.CreateChatResponse{
			StatusCode: dara.Int32(200),
			Id:         dara.String("fake-id-3"),
			Event:      dara.String("ok"),
		}
		ev := parseChatEvent(resp)
		if ev.StatusCode != 200 {
			t.Errorf("期望 StatusCode=200, 实际=%d", ev.StatusCode)
		}
	})

	t.Run("id and event fields mapped correctly", func(t *testing.T) {
		resp := &starops.CreateChatResponse{
			StatusCode: dara.Int32(0),
			Id:         dara.String("fake-event-id"),
			Event:      dara.String("thinking"),
		}
		ev := parseChatEvent(resp)
		if ev.Id != "fake-event-id" {
			t.Errorf("期望 Id=\"fake-event-id\", 实际=%q", ev.Id)
		}
		if ev.Event != "thinking" {
			t.Errorf("期望 Event=\"thinking\", 实际=%q", ev.Event)
		}
	})

	t.Run("nil fields default to zero values", func(t *testing.T) {
		resp := &starops.CreateChatResponse{}
		ev := parseChatEvent(resp)
		if ev.StatusCode != 0 {
			t.Errorf("期望 StatusCode=0, 实际=%d", ev.StatusCode)
		}
		if ev.Id != "" {
			t.Errorf("期望 Id=\"\", 实际=%q", ev.Id)
		}
		if ev.Event != "" {
			t.Errorf("期望 Event=\"\", 实际=%q", ev.Event)
		}
	})
}

// ---------------------------------------------------------------------------
// TestIsStreamDoneEvent
// ---------------------------------------------------------------------------

func TestIsStreamDoneEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    *types.ChatEvent
		expected bool
	}{
		{
			name:     "nil event → false",
			event:    nil,
			expected: false,
		},
		{
			name:     "nil body → false",
			event:    &types.ChatEvent{Body: nil},
			expected: false,
		},
		{
			name:     "no messages → false",
			event:    &types.ChatEvent{Body: &types.ChatEventBody{Messages: nil}},
			expected: false,
		},
		{
			name: "nil message in slice → false",
			event: &types.ChatEvent{Body: &types.ChatEventBody{
				Messages: []*types.MessageItem{nil},
			}},
			expected: false,
		},
		{
			name: "non-stream_done event → false",
			event: &types.ChatEvent{Body: &types.ChatEventBody{
				Messages: []*types.MessageItem{
					{
						Events: []*types.ItemEvent{
							{Type: types.EventTypeThinking},
						},
					},
				},
			}},
			expected: false,
		},
		{
			name: "single stream_done event → true",
			event: &types.ChatEvent{Body: &types.ChatEventBody{
				Messages: []*types.MessageItem{
					{
						Events: []*types.ItemEvent{
							{Type: types.EventTypeStreamDone},
						},
					},
				},
			}},
			expected: true,
		},
		{
			name: "stream_done in second message → true",
			event: &types.ChatEvent{Body: &types.ChatEventBody{
				Messages: []*types.MessageItem{
					{
						Events: []*types.ItemEvent{
							{Type: types.EventTypeThinking},
						},
					},
					{
						Events: []*types.ItemEvent{
							{Type: types.EventTypeStreamDone},
						},
					},
				},
			}},
			expected: true,
		},
		{
			name: "stream_done among other event types → true",
			event: &types.ChatEvent{Body: &types.ChatEventBody{
				Messages: []*types.MessageItem{
					{
						Events: []*types.ItemEvent{
							{Type: types.EventTypeThinking},
							{Type: types.EventTypeStreamDone},
							{Type: types.EventTypeError},
						},
					},
				},
			}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStreamDoneEvent(tt.event)
			if result != tt.expected {
				t.Errorf("isStreamDoneEvent(): 期望 %v, 实际 %v", tt.expected, result)
			}
		})
	}
}
