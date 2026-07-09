package client

import (
	"testing"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/vibeops/samples/golang/types"
)

// TestSimplePrinter_EventFieldSkip 测试非文本事件被快速跳过
func TestSimplePrinter_EventFieldSkip(t *testing.T) {
	p := NewSimplePrinter()

	// 构造一个带 system 角色 artifacts 的消息，正常情况下会产出文本
	role := string(types.MessageItemRoleSystem)
	msg := &starops.CreateChatResponseBodyMessages{
		Role: &role,
		Artifacts: []map[string]any{
			{
				"parts": []any{
					map[string]any{"kind": "text", "text": "应该被跳过的内容"},
				},
			},
		},
	}
	body := &starops.CreateChatResponseBody{
		Messages: []*starops.CreateChatResponseBodyMessages{msg},
	}

	skipEvents := []string{"interaction", "thinking", "tool_call", "error"}
	for _, eventName := range skipEvents {
		t.Run("skip_"+eventName, func(t *testing.T) {
			event := &ChatEvent{
				Event: eventName,
				Body:  body,
			}
			result := p.ProcessEvent(event)
			if result != "" {
				t.Errorf("event=%s 时应返回空字符串，实际返回 %q", eventName, result)
			}
		})
	}
}

// TestSimplePrinter_TextEventProcessed 测试 text 事件正常处理
func TestSimplePrinter_TextEventProcessed(t *testing.T) {
	p := NewSimplePrinter()

	role := string(types.MessageItemRoleSystem)
	msg := &starops.CreateChatResponseBodyMessages{
		Role: &role,
		Artifacts: []map[string]any{
			{
				"parts": []any{
					map[string]any{"kind": "text", "text": "文本内容"},
				},
			},
		},
	}
	body := &starops.CreateChatResponseBody{
		Messages: []*starops.CreateChatResponseBodyMessages{msg},
	}

	event := &ChatEvent{
		Event: "text",
		Body:  body,
	}
	result := p.ProcessEvent(event)
	if result != "文本内容" {
		t.Errorf("event=text 时应返回文本内容，实际返回 %q", result)
	}
}

// TestSimplePrinter_TaskFinishedEventProcessed 测试 task_finished 事件正常处理
func TestSimplePrinter_TaskFinishedEventProcessed(t *testing.T) {
	p := NewSimplePrinter()

	role := string(types.MessageItemRoleSystem)
	msg := &starops.CreateChatResponseBodyMessages{
		Role: &role,
		Artifacts: []map[string]any{
			{
				"parts": []any{
					map[string]any{"kind": "text", "text": "任务完成"},
				},
			},
		},
	}
	body := &starops.CreateChatResponseBody{
		Messages: []*starops.CreateChatResponseBodyMessages{msg},
	}

	event := &ChatEvent{
		Event: "task_finished",
		Body:  body,
	}
	result := p.ProcessEvent(event)
	if result != "任务完成" {
		t.Errorf("event=task_finished 时应返回文本内容，实际返回 %q", result)
	}
}

// TestSimplePrinter_EmptyEventFallthrough 测试空 Event 字段走正常处理流程
func TestSimplePrinter_EmptyEventFallthrough(t *testing.T) {
	p := NewSimplePrinter()

	role := string(types.MessageItemRoleSystem)
	msg := &starops.CreateChatResponseBodyMessages{
		Role: &role,
		Artifacts: []map[string]any{
			{
				"parts": []any{
					map[string]any{"kind": "text", "text": "无事件字段"},
				},
			},
		},
	}
	body := &starops.CreateChatResponseBody{
		Messages: []*starops.CreateChatResponseBodyMessages{msg},
	}

	event := &ChatEvent{
		Event: "",
		Body:  body,
	}
	result := p.ProcessEvent(event)
	if result != "无事件字段" {
		t.Errorf("空 Event 字段时应正常处理，实际返回 %q", result)
	}
}

// TestSimplePrinter_NilEventAndBody 测试 nil 防护
func TestSimplePrinter_NilEventAndBody(t *testing.T) {
	p := NewSimplePrinter()

	t.Run("nil_event", func(t *testing.T) {
		result := p.ProcessEvent(nil)
		if result != "" {
			t.Errorf("nil event 时应返回空字符串，实际返回 %q", result)
		}
	})

	t.Run("nil_body", func(t *testing.T) {
		event := &ChatEvent{Event: "text", Body: nil}
		result := p.ProcessEvent(event)
		if result != "" {
			t.Errorf("nil body 时应返回空字符串，实际返回 %q", result)
		}
	})
}
