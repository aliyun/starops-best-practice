package client

import (
	"context"
	"testing"
	"time"

	starops "github.com/alibabacloud-go/starops-20260428/client"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/config"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// ---------------------------------------------------------------------------
// TestCalculateBackoff
// ---------------------------------------------------------------------------

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name     string
		retry    int
		cfg      *config.RetryConfig
		expected time.Duration
	}{
		{
			name:  "first retry → InitialBackoff",
			retry: 1,
			cfg: &config.RetryConfig{
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     5 * time.Millisecond,
				BackoffFactor:  2.0,
			},
			expected: 1 * time.Millisecond,
		},
		{
			name:  "second retry → InitialBackoff * BackoffFactor",
			retry: 2,
			cfg: &config.RetryConfig{
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     5 * time.Millisecond,
				BackoffFactor:  2.0,
			},
			expected: 2 * time.Millisecond,
		},
		{
			name:  "capped at MaxBackoff",
			retry: 10,
			cfg: &config.RetryConfig{
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     5 * time.Millisecond,
				BackoffFactor:  2.0,
			},
			expected: 5 * time.Millisecond,
		},
		{
			name:  "constant backoff (factor=1.0)",
			retry: 5,
			cfg: &config.RetryConfig{
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     5 * time.Millisecond,
				BackoffFactor:  1.0,
			},
			expected: 1 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBackoff(tt.retry, tt.cfg)
			if got != tt.expected {
				t.Errorf("calculateBackoff(%d): 期望 %v, 实际 %v", tt.retry, tt.expected, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestBuildReconnectRequest
// ---------------------------------------------------------------------------

func TestBuildReconnectRequest(t *testing.T) {
	t.Run("action is reconnect", func(t *testing.T) {
		orig := &starops.CreateChatRequest{}
		orig.SetThreadId("fake-thread-id")
		orig.SetDigitalEmployeeName("test-employee")
		orig.SetVariables(map[string]interface{}{"key": "val"})

		req := buildReconnectRequest(orig)
		if *req.Action != "reconnect" {
			t.Errorf("期望 Action=\"reconnect\", 实际=%q", *req.Action)
		}
	})

	t.Run("threadId and employeeName copied", func(t *testing.T) {
		orig := &starops.CreateChatRequest{}
		orig.SetThreadId("fake-thread-id")
		orig.SetDigitalEmployeeName("test-employee")
		orig.SetVariables(map[string]interface{}{})

		req := buildReconnectRequest(orig)
		if *req.ThreadId != "fake-thread-id" {
			t.Errorf("期望 ThreadId=\"fake-thread-id\", 实际=%q", *req.ThreadId)
		}
		if *req.DigitalEmployeeName != "test-employee" {
			t.Errorf("期望 DigitalEmployeeName=\"test-employee\", 实际=%q", *req.DigitalEmployeeName)
		}
	})

	t.Run("variables deep copied", func(t *testing.T) {
		orig := &starops.CreateChatRequest{}
		orig.SetThreadId("fake-thread-id")
		orig.SetDigitalEmployeeName("test-employee")
		origVars := map[string]interface{}{"workspace": "test-ws", "extra": "data"}
		orig.SetVariables(origVars)

		req := buildReconnectRequest(orig)

		// Values must match
		if req.Variables["workspace"] != "test-ws" {
			t.Errorf("期望 Variables[workspace]=\"test-ws\", 实际=%v", req.Variables["workspace"])
		}
		if req.Variables["extra"] != "data" {
			t.Errorf("期望 Variables[extra]=\"data\", 实际=%v", req.Variables["extra"])
		}

		// Mutation of original must not affect the copy
		origVars["workspace"] = "mutated"
		if req.Variables["workspace"] == "mutated" {
			t.Error("Variables 应该是深拷贝，修改原始 map 不应影响副本")
		}
	})

	t.Run("nil variables → empty map", func(t *testing.T) {
		orig := &starops.CreateChatRequest{}
		orig.SetThreadId("fake-thread-id")
		orig.SetDigitalEmployeeName("test-employee")
		// Variables not set (nil)

		req := buildReconnectRequest(orig)
		if req.Variables == nil {
			t.Fatal("期望 Variables 为空 map，不为 nil")
		}
		if len(req.Variables) != 0 {
			t.Errorf("期望 Variables 长度为 0, 实际=%d", len(req.Variables))
		}
	})
}

// ---------------------------------------------------------------------------
// TestPrepareReconnect
// ---------------------------------------------------------------------------

func TestPrepareReconnect(t *testing.T) {
	t.Run("max retries exceeded → false, error event", func(t *testing.T) {
		c := &AgentClient{}
		events := make(chan *types.ChatEvent, 1)
		cfg := &config.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     5 * time.Millisecond,
			BackoffFactor:  2.0,
		}
		state := &retryState{retryCount: 3} // already at max

		ctx := context.Background()
		result := c.prepareReconnect(ctx, events, state, cfg)
		if result {
			t.Fatal("期望返回 false（已超过最大重试次数）")
		}
		select {
		case ev := <-events:
			if ev.Error == nil {
				t.Fatal("期望收到 error 事件")
			}
		default:
			t.Fatal("期望 events 通道有一条错误消息")
		}
	})

	t.Run("successful retry → true, state updated", func(t *testing.T) {
		c := &AgentClient{}
		events := make(chan *types.ChatEvent, 1)
		cfg := &config.RetryConfig{
			MaxRetries:     5,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     5 * time.Millisecond,
			BackoffFactor:  2.0,
		}
		state := &retryState{retryCount: 0, inDedupeWindow: false}

		ctx := context.Background()
		result := c.prepareReconnect(ctx, events, state, cfg)
		if !result {
			t.Fatal("期望返回 true（重试成功）")
		}
		if state.retryCount != 1 {
			t.Errorf("期望 retryCount=1, 实际=%d", state.retryCount)
		}
		if !state.inDedupeWindow {
			t.Error("期望 inDedupeWindow=true")
		}
	})

	t.Run("context cancelled → false, error event", func(t *testing.T) {
		c := &AgentClient{}
		events := make(chan *types.ChatEvent, 1)
		cfg := &config.RetryConfig{
			MaxRetries:     5,
			InitialBackoff: 5 * time.Second, // long backoff so cancel fires first
			MaxBackoff:     5 * time.Second,
			BackoffFactor:  1.0,
		}
		state := &retryState{retryCount: 0}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		result := c.prepareReconnect(ctx, events, state, cfg)
		if result {
			t.Fatal("期望返回 false（context 已取消）")
		}
		select {
		case ev := <-events:
			if ev.Error == nil {
				t.Fatal("期望收到 error 事件")
			}
		default:
			t.Fatal("期望 events 通道有一条错误消息")
		}
	})
}
