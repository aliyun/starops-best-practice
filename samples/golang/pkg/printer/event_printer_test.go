package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// captureStdout captures stdout output produced by function f.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

// ---------------------------------------------------------------------------
// NewEventPrinter
// ---------------------------------------------------------------------------

func TestNewEventPrinter_BothTrue(t *testing.T) {
	p := NewEventPrinter(true, true)
	if !p.PrintRawBody {
		t.Error("expected PrintRawBody=true")
	}
	if !p.PrintParsed {
		t.Error("expected PrintParsed=true")
	}
	if !p.PrintSeparator {
		t.Error("expected PrintSeparator defaults to true")
	}
}

func TestNewEventPrinter_BothFalse(t *testing.T) {
	p := NewEventPrinter(false, false)
	if p.PrintRawBody {
		t.Error("expected PrintRawBody=false")
	}
	if p.PrintParsed {
		t.Error("expected PrintParsed=false")
	}
}

// ---------------------------------------------------------------------------
// PrintEvent — basic scenarios
// ---------------------------------------------------------------------------

func TestPrintEvent_ErrorEvent(t *testing.T) {
	p := NewEventPrinter(false, false)
	evt := &types.ChatEvent{Error: fmt.Errorf("something went wrong")}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "错误") {
		t.Errorf("expected output to contain '错误', got: %s", out)
	}
}

func TestPrintEvent_DoneNilBody(t *testing.T) {
	p := NewEventPrinter(false, false)
	evt := &types.ChatEvent{IsDone: true, Body: nil}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "对话完成") {
		t.Errorf("expected output to contain '对话完成', got: %s", out)
	}
}

func TestPrintEvent_NilBodyNotDone(t *testing.T) {
	p := NewEventPrinter(false, false)
	evt := &types.ChatEvent{Body: nil}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if out != "" {
		t.Errorf("expected empty output for nil body not done, got: %s", out)
	}
}

func TestPrintEvent_Separator(t *testing.T) {
	p := NewEventPrinter(false, false)
	p.PrintSeparator = true
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 5) })
	if !strings.Contains(out, "事件 #") {
		t.Errorf("expected separator with '事件 #', got: %s", out)
	}
}

func TestPrintEvent_RawBody(t *testing.T) {
	p := NewEventPrinter(true, false)
	evt := &types.ChatEvent{
		Body:    &types.ChatEventBody{},
		RawJSON: `{"hello":"world"}`,
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "原始 Body") {
		t.Errorf("expected '原始 Body' in output, got: %s", out)
	}
}

func TestPrintEvent_ParsedMode(t *testing.T) {
	p := NewEventPrinter(false, true)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{Role: "assistant"},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "解析详情") {
		t.Errorf("expected '解析详情' in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// Parsed output — message fields
// ---------------------------------------------------------------------------

func TestParsed_MessageWithRole(t *testing.T) {
	p := NewEventPrinter(false, true)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{Role: "user"},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "角色:") {
		t.Errorf("expected '角色:' in output, got: %s", out)
	}
}

func TestParsed_MessageWithCallID(t *testing.T) {
	p := NewEventPrinter(false, true)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{CallID: "call-123"},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "CallID:") {
		t.Errorf("expected 'CallID:' in output, got: %s", out)
	}
}

func TestParsed_MessageWithContents(t *testing.T) {
	p := NewEventPrinter(false, true)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Contents: []*types.ItemContent{
						{Type: "text", Value: "hello"},
					},
				},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "内容:") {
		t.Errorf("expected '内容:' in output, got: %s", out)
	}
	if !strings.Contains(out, "类型:") {
		t.Errorf("expected '类型:' in output, got: %s", out)
	}
}

func TestParsed_MessageWithTools(t *testing.T) {
	p := NewEventPrinter(false, true)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Tools: []*types.ItemTool{
						{Name: "search", Status: "success"},
					},
				},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "工具调用:") {
		t.Errorf("expected '工具调用:' in output, got: %s", out)
	}
}

func TestParsed_MessageWithAgents(t *testing.T) {
	p := NewEventPrinter(false, true)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Agents: []*types.ItemAgent{
						{Name: "sub-agent", Status: "start"},
					},
				},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "Agent调用:") {
		t.Errorf("expected 'Agent调用:' in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// Event payloads
// ---------------------------------------------------------------------------

func TestParsed_ThinkingEvent(t *testing.T) {
	p := NewEventPrinter(false, true)
	payload := types.ItemThinkingPayload{ReasoningDelta: "let me think about this"}
	payloadJSON, _ := json.Marshal(payload)
	var rawPayload any
	json.Unmarshal(payloadJSON, &rawPayload)

	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Events: []*types.ItemEvent{
						{Type: types.EventTypeThinking, Payload: rawPayload},
					},
				},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "思考:") {
		t.Errorf("expected '思考:' in output, got: %s", out)
	}
}

func TestParsed_ErrorEventPayload(t *testing.T) {
	p := NewEventPrinter(false, true)
	payload := types.ItemErrorPayload{Code: "ERR_001", Message: "bad request"}
	payloadJSON, _ := json.Marshal(payload)
	var rawPayload any
	json.Unmarshal(payloadJSON, &rawPayload)

	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Events: []*types.ItemEvent{
						{Type: types.EventTypeError, Payload: rawPayload},
					},
				},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "错误码:") {
		t.Errorf("expected '错误码:' in output, got: %s", out)
	}
	if !strings.Contains(out, "消息:") {
		t.Errorf("expected '消息:' in output, got: %s", out)
	}
}

func TestParsed_TaskFinished(t *testing.T) {
	p := NewEventPrinter(false, true)
	payload := types.ItemTaskFinishedPayload{Success: true}
	payloadJSON, _ := json.Marshal(payload)
	var rawPayload any
	json.Unmarshal(payloadJSON, &rawPayload)

	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{
				{
					Events: []*types.ItemEvent{
						{Type: types.EventTypeTaskFinished, Payload: rawPayload},
					},
				},
			},
		},
	}
	out := captureStdout(func() { p.PrintEvent(evt, 0) })
	if !strings.Contains(out, "成功:") {
		t.Errorf("expected '成功:' in output, got: %s", out)
	}
}
