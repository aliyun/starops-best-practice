package printer

import (
	"testing"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// TestSimplePrinter_TextEventProcessed 测试 text 事件正常处理
func TestSimplePrinter_TextEventProcessed(t *testing.T) {
	p := NewSimplePrinter()

	evt := &types.ChatEvent{
		Event: "text",
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Role: types.MessageItemRoleSystem,
					Artifacts: []map[string]any{
						{
							"parts": []any{
								map[string]any{"kind": "text", "text": "文本内容"},
							},
						},
					},
				},
			},
		},
	}
	result := p.ProcessEvent(evt)
	if result != "文本内容" {
		t.Errorf("event=text 时应返回文本内容，实际返回 %q", result)
	}
}

// TestSimplePrinter_TaskFinishedEventProcessed 测试 task_finished 事件正常处理
func TestSimplePrinter_TaskFinishedEventProcessed(t *testing.T) {
	p := NewSimplePrinter()

	evt := &types.ChatEvent{
		Event: "task_finished",
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Role: types.MessageItemRoleSystem,
					Artifacts: []map[string]any{
						{
							"parts": []any{
								map[string]any{"kind": "text", "text": "任务完成"},
							},
						},
					},
				},
			},
		},
	}
	result := p.ProcessEvent(evt)
	if result != "任务完成" {
		t.Errorf("event=task_finished 时应返回文本内容，实际返回 %q", result)
	}
}

// TestSimplePrinter_EmptyEventFallthrough 测试空 Event 字段走正常处理流程
func TestSimplePrinter_EmptyEventFallthrough(t *testing.T) {
	p := NewSimplePrinter()

	evt := &types.ChatEvent{
		Event: "",
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Role: types.MessageItemRoleSystem,
					Artifacts: []map[string]any{
						{
							"parts": []any{
								map[string]any{"kind": "text", "text": "无事件字段"},
							},
						},
					},
				},
			},
		},
	}
	result := p.ProcessEvent(evt)
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
		evt := &types.ChatEvent{Event: "text", Body: nil}
		result := p.ProcessEvent(evt)
		if result != "" {
			t.Errorf("nil body 时应返回空字符串，实际返回 %q", result)
		}
	})
}
