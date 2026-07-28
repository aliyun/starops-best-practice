package client

import (
	"context"
	"testing"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

func TestIsNewerTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		ts       string
		base     string
		expected bool
	}{
		{"空 ts", "", "100", false},
		{"空 base", "100", "", true},
		{"都为空", "", "", false},
		{"数值更新", "200", "100", true},
		{"数值更旧", "50", "100", false},
		{"数值相等", "100", "100", false},
		{"字符串更新", "2024-01-02", "2024-01-01", true},
		{"字符串更旧", "2024-01-01", "2024-01-02", false},
		{"base 数值但 ts 无法解析", "abc", "100", false},
		{"base 无法解析走字符串比较", "200", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNewerTimestamp(tt.ts, tt.base)
			if result != tt.expected {
				t.Errorf("isNewerTimestamp(%q, %q): 期望 %v, 实际 %v", tt.ts, tt.base, tt.expected, result)
			}
		})
	}
}

func newEventWithTimestamps(timestamps ...string) *types.ChatEvent {
	msgs := make([]*types.MessageItem, 0, len(timestamps))
	for _, ts := range timestamps {
		msg := &types.MessageItem{}
		if ts != "" {
			msg.Timestamp = ts
		}
		msgs = append(msgs, msg)
	}
	return &types.ChatEvent{Body: &types.ChatEventBody{Messages: msgs}}
}

func TestExtractNewestTimestamp(t *testing.T) {
	t.Run("event 为 nil", func(t *testing.T) {
		if got := extractNewestTimestamp(nil, "100"); got != "" {
			t.Errorf("期望 \"\", 实际 %q", got)
		}
	})

	t.Run("event 无消息", func(t *testing.T) {
		event := &types.ChatEvent{Body: &types.ChatEventBody{}}
		if got := extractNewestTimestamp(event, "100"); got != "" {
			t.Errorf("期望 \"\", 实际 %q", got)
		}
	})

	t.Run("单消息且 timestamp > base", func(t *testing.T) {
		event := newEventWithTimestamps("200")
		if got := extractNewestTimestamp(event, "100"); got != "200" {
			t.Errorf("期望 \"200\", 实际 %q", got)
		}
	})

	t.Run("多消息取最新", func(t *testing.T) {
		event := newEventWithTimestamps("150", "300", "200")
		if got := extractNewestTimestamp(event, "100"); got != "300" {
			t.Errorf("期望 \"300\", 实际 %q", got)
		}
	})

	t.Run("所有消息 timestamp <= base", func(t *testing.T) {
		event := newEventWithTimestamps("50", "80", "100")
		if got := extractNewestTimestamp(event, "100"); got != "" {
			t.Errorf("期望 \"\", 实际 %q", got)
		}
	})

	t.Run("消息无 timestamp 字段", func(t *testing.T) {
		event := newEventWithTimestamps("")
		if got := extractNewestTimestamp(event, "100"); got != "" {
			t.Errorf("期望 \"\", 实际 %q", got)
		}
	})
}

func TestForwardEvent(t *testing.T) {
	c := &AgentClient{}

	t.Run("正常模式转发消息", func(t *testing.T) {
		events := make(chan *types.ChatEvent, 1)
		state := &retryState{inDedupeWindow: false, lastTimestamp: "100"}
		event := newEventWithTimestamps("200")

		if !c.forwardEvent(context.Background(), event, state, events) {
			t.Fatal("正常模式应返回 true")
		}
		if len(events) != 1 {
			t.Errorf("期望转发 1 条消息, 实际 %d", len(events))
		}
		if state.lastTimestamp != "200" {
			t.Errorf("lastTimestamp 期望 \"200\", 实际 %q", state.lastTimestamp)
		}
	})

	t.Run("去重窗口内旧消息不转发", func(t *testing.T) {
		events := make(chan *types.ChatEvent, 1)
		state := &retryState{inDedupeWindow: true, lastTimestamp: "200"}
		event := newEventWithTimestamps("150")

		if c.forwardEvent(context.Background(), event, state, events) {
			t.Fatal("去重窗口内旧消息应返回 false")
		}
		if len(events) != 0 {
			t.Errorf("期望不转发, 实际转发 %d 条", len(events))
		}
		if !state.inDedupeWindow {
			t.Error("旧消息不应退出去重窗口")
		}
	})

	t.Run("去重窗口内新消息转发并退出窗口", func(t *testing.T) {
		events := make(chan *types.ChatEvent, 1)
		state := &retryState{inDedupeWindow: true, lastTimestamp: "200"}
		event := newEventWithTimestamps("300")

		if !c.forwardEvent(context.Background(), event, state, events) {
			t.Fatal("去重窗口内新消息应返回 true")
		}
		if len(events) != 1 {
			t.Errorf("期望转发 1 条消息, 实际 %d", len(events))
		}
		if state.inDedupeWindow {
			t.Error("新消息应退出去重窗口")
		}
		if state.lastTimestamp != "300" {
			t.Errorf("lastTimestamp 期望 \"300\", 实际 %q", state.lastTimestamp)
		}
	})

	t.Run("无时间戳消息在正常模式仍转发", func(t *testing.T) {
		events := make(chan *types.ChatEvent, 1)
		state := &retryState{inDedupeWindow: false, lastTimestamp: "100"}
		event := newEventWithTimestamps("")

		if !c.forwardEvent(context.Background(), event, state, events) {
			t.Fatal("正常模式无时间戳消息应返回 true")
		}
		if len(events) != 1 {
			t.Errorf("期望转发 1 条消息, 实际 %d", len(events))
		}
		if state.lastTimestamp != "100" {
			t.Errorf("lastTimestamp 期望 \"100\", 实际 %q", state.lastTimestamp)
		}
	})
}
