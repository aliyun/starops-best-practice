package interactive

import (
	"context"
	"testing"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// mockChatClient 模拟 ChatClient 接口
type mockChatClient struct {
	interactCalled bool
	lastThreadID   string
	events         chan *types.ChatEvent
}

func (m *mockChatClient) Interact(ctx context.Context, threadID string, userInteractive string, baseVariables map[string]any) <-chan *types.ChatEvent {
	m.interactCalled = true
	m.lastThreadID = threadID
	return m.events
}

// ---------------------------------------------------------------------------
// IsInteractiveEvent
// ---------------------------------------------------------------------------

func TestIsInteractiveEvent_Nil(t *testing.T) {
	if IsInteractiveEvent(nil) {
		t.Error("expected false for nil event")
	}
}

func TestIsInteractiveEvent_Interactive(t *testing.T) {
	evt := &types.ItemEvent{Type: types.EventTypeInteractive}
	if !IsInteractiveEvent(evt) {
		t.Error("expected true for interactive event")
	}
}

func TestIsInteractiveEvent_Thinking(t *testing.T) {
	evt := &types.ItemEvent{Type: types.EventTypeThinking}
	if IsInteractiveEvent(evt) {
		t.Error("expected false for thinking event")
	}
}

// ---------------------------------------------------------------------------
// ExtractInteractiveEvents
// ---------------------------------------------------------------------------

func TestExtractInteractiveEvents_NilItem(t *testing.T) {
	result := ExtractInteractiveEvents(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestExtractInteractiveEvents_MixedEvents(t *testing.T) {
	item := &types.MessageItem{
		Events: []*types.ItemEvent{
			{Type: types.EventTypeThinking},
			{Type: types.EventTypeInteractive},
			{Type: types.EventTypeError},
			{Type: types.EventTypeInteractive},
		},
	}
	result := ExtractInteractiveEvents(item)
	if len(result) != 2 {
		t.Errorf("expected 2 interactive events, got %d", len(result))
	}
	for _, evt := range result {
		if evt.Type != types.EventTypeInteractive {
			t.Errorf("expected interactive event type, got %s", evt.Type)
		}
	}
}

func TestExtractInteractiveEvents_NoInteractive(t *testing.T) {
	item := &types.MessageItem{
		Events: []*types.ItemEvent{
			{Type: types.EventTypeThinking},
			{Type: types.EventTypeError},
		},
	}
	result := ExtractInteractiveEvents(item)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// ExtractResponse
// ---------------------------------------------------------------------------

func TestExtractResponse_NilEvent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	resp := ExtractResponse(context.Background(), nil, handler)
	if resp != nil {
		t.Errorf("expected nil for nil event, got %+v", resp)
	}
}

func TestExtractResponse_NilHandler(t *testing.T) {
	evt := &types.ChatEvent{}
	resp := ExtractResponse(context.Background(), evt, nil)
	if resp != nil {
		t.Errorf("expected nil for nil handler, got %+v", resp)
	}
}

func TestExtractResponse_NilBody(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	evt := &types.ChatEvent{Body: nil}
	resp := ExtractResponse(context.Background(), evt, handler)
	if resp != nil {
		t.Errorf("expected nil for nil body, got %+v", resp)
	}
}

func TestExtractResponse_EmptyMessages(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	evt := &types.ChatEvent{
		Body: &types.ChatEventBody{
			Messages: []*types.MessageItem{},
		},
	}
	resp := ExtractResponse(context.Background(), evt, handler)
	if resp != nil {
		t.Errorf("expected nil for empty messages, got %+v", resp)
	}
}

// ---------------------------------------------------------------------------
// ResumeChat
// ---------------------------------------------------------------------------

func TestResumeChat_NilClient(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	resp := &Response{
		CallId:   "fake-call-id",
		Type:     types.InteractionTypeUserAck,
		Decision: "yes",
	}
	ch := handler.ResumeChat(context.Background(), "fake-thread-id", resp, nil)
	evt := <-ch
	if evt == nil || evt.Error == nil {
		t.Error("expected error event for nil client")
	}
}

func TestResumeChat_NilResponse(t *testing.T) {
	mock := &mockChatClient{events: make(chan *types.ChatEvent, 1)}
	handler := NewHandler(mock, 30*time.Second)
	ch := handler.ResumeChat(context.Background(), "fake-thread-id", nil, nil)
	evt := <-ch
	if evt == nil || evt.Error == nil {
		t.Error("expected error event for nil response")
	}
}

func TestResumeChat_ValidResponse(t *testing.T) {
	mock := &mockChatClient{events: make(chan *types.ChatEvent, 1)}
	mock.events <- &types.ChatEvent{IsDone: true}
	close(mock.events)

	handler := NewHandler(mock, 30*time.Second)
	resp := &Response{
		CallId:   "fake-call-id",
		Type:     types.InteractionTypeUserAck,
		Decision: "yes",
		Source:   map[string]any{"app": "test-app"},
	}

	ch := handler.ResumeChat(context.Background(), "fake-thread-id", resp, nil)
	evt := <-ch
	if evt == nil {
		t.Fatal("expected event from channel")
	}
	if !mock.interactCalled {
		t.Error("expected Interact to be called on mock client")
	}
	if mock.lastThreadID != "fake-thread-id" {
		t.Errorf("expected thread ID 'fake-thread-id', got '%s'", mock.lastThreadID)
	}
}

// ---------------------------------------------------------------------------
// NewHandler
// ---------------------------------------------------------------------------

func TestNewHandler_ReturnsNonNil(t *testing.T) {
	mock := &mockChatClient{}
	handler := NewHandler(mock, 10*time.Second)
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}
