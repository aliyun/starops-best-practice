package client

import (
	"os"
	"testing"
	"time"

	starops "github.com/alibabacloud-go/starops-20260428/client"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 10 {
		t.Errorf("MaxRetries: 期望 10, 实际 %d", cfg.MaxRetries)
	}
	if cfg.InitialBackoff != 1*time.Second {
		t.Errorf("InitialBackoff: 期望 1s, 实际 %v", cfg.InitialBackoff)
	}
	if cfg.MaxBackoff != 30*time.Second {
		t.Errorf("MaxBackoff: 期望 30s, 实际 %v", cfg.MaxBackoff)
	}
	if cfg.BackoffFactor != 2.0 {
		t.Errorf("BackoffFactor: 期望 2.0, 实际 %f", cfg.BackoffFactor)
	}
}

func TestLoadRetryConfigFromEnv(t *testing.T) {
	// 测试默认值
	t.Run("default", func(t *testing.T) {
		os.Unsetenv("VIBEOPS_MAX_RETRIES")
		cfg := LoadRetryConfigFromEnv()
		if cfg.MaxRetries != 10 {
			t.Errorf("默认 MaxRetries: 期望 10, 实际 %d", cfg.MaxRetries)
		}
	})

	// 测试环境变量覆盖
	t.Run("custom_max_retries", func(t *testing.T) {
		os.Setenv("VIBEOPS_MAX_RETRIES", "5")
		defer os.Unsetenv("VIBEOPS_MAX_RETRIES")
		cfg := LoadRetryConfigFromEnv()
		if cfg.MaxRetries != 5 {
			t.Errorf("MaxRetries: 期望 5, 实际 %d", cfg.MaxRetries)
		}
	})

	// 测试无效值
	t.Run("invalid_value", func(t *testing.T) {
		os.Setenv("VIBEOPS_MAX_RETRIES", "abc")
		defer os.Unsetenv("VIBEOPS_MAX_RETRIES")
		cfg := LoadRetryConfigFromEnv()
		if cfg.MaxRetries != 10 {
			t.Errorf("无效值时 MaxRetries: 期望 10, 实际 %d", cfg.MaxRetries)
		}
	})

	// 测试负数值
	t.Run("negative_value", func(t *testing.T) {
		os.Setenv("VIBEOPS_MAX_RETRIES", "-1")
		defer os.Unsetenv("VIBEOPS_MAX_RETRIES")
		cfg := LoadRetryConfigFromEnv()
		if cfg.MaxRetries != 10 {
			t.Errorf("负数值时 MaxRetries: 期望 10, 实际 %d", cfg.MaxRetries)
		}
	})
}

func TestCalculateBackoff(t *testing.T) {
	cfg := DefaultRetryConfig()

	tests := []struct {
		retryCount int
		expected   time.Duration
	}{
		{1, 1 * time.Second},   // 1s * 2^0 = 1s
		{2, 2 * time.Second},   // 1s * 2^1 = 2s
		{3, 4 * time.Second},   // 1s * 2^2 = 4s
		{4, 8 * time.Second},   // 1s * 2^3 = 8s
		{5, 16 * time.Second},  // 1s * 2^4 = 16s
		{6, 30 * time.Second},  // 1s * 2^5 = 32s, 超过 MaxBackoff，取 30s
		{7, 30 * time.Second},  // 超过 MaxBackoff
		{10, 30 * time.Second}, // 超过 MaxBackoff
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := calculateBackoff(tt.retryCount, cfg)
			if result != tt.expected {
				t.Errorf("retryCount=%d: 期望 %v, 实际 %v", tt.retryCount, tt.expected, result)
			}
		})
	}
}

func TestCalculateBackoff_CustomConfig(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		BackoffFactor:  3.0,
	}

	// 500ms * 3^0 = 500ms
	result := calculateBackoff(1, cfg)
	if result != 500*time.Millisecond {
		t.Errorf("retryCount=1: 期望 500ms, 实际 %v", result)
	}

	// 500ms * 3^1 = 1500ms
	result = calculateBackoff(2, cfg)
	if result != 1500*time.Millisecond {
		t.Errorf("retryCount=2: 期望 1500ms, 实际 %v", result)
	}

	// 500ms * 3^2 = 4500ms
	result = calculateBackoff(3, cfg)
	if result != 4500*time.Millisecond {
		t.Errorf("retryCount=3: 期望 4500ms, 实际 %v", result)
	}
}

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
		// base 无法解析为数值时 fallback 字符串比较："200" > "abc" → '2'(0x32) < 'a'(0x61) → false
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

// newEventWithTimestamps 构造一个包含指定 timestamp 消息的 ChatEvent
// 传入空字符串表示该消息不设置 timestamp 字段
func newEventWithTimestamps(timestamps ...string) *ChatEvent {
	msgs := make([]*starops.CreateChatResponseBodyMessages, 0, len(timestamps))
	for _, ts := range timestamps {
		msg := &starops.CreateChatResponseBodyMessages{}
		if ts != "" {
			msg.SetTimestamp(ts)
		}
		msgs = append(msgs, msg)
	}
	body := &starops.CreateChatResponseBody{Messages: msgs}
	return &ChatEvent{Body: body}
}

func TestExtractNewestTimestamp(t *testing.T) {
	t.Run("event 为 nil", func(t *testing.T) {
		if got := extractNewestTimestamp(nil, "100"); got != "" {
			t.Errorf("期望 \"\", 实际 %q", got)
		}
	})

	t.Run("event 无消息", func(t *testing.T) {
		event := &ChatEvent{Body: &starops.CreateChatResponseBody{}}
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
		events := make(chan *ChatEvent, 1)
		state := &retryState{inDedupeWindow: false, lastTimestamp: "100"}
		event := newEventWithTimestamps("200")

		if !c.forwardEvent(event, state, events) {
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
		events := make(chan *ChatEvent, 1)
		state := &retryState{inDedupeWindow: true, lastTimestamp: "200"}
		event := newEventWithTimestamps("150")

		if c.forwardEvent(event, state, events) {
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
		events := make(chan *ChatEvent, 1)
		state := &retryState{inDedupeWindow: true, lastTimestamp: "200"}
		event := newEventWithTimestamps("300")

		if !c.forwardEvent(event, state, events) {
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
		events := make(chan *ChatEvent, 1)
		state := &retryState{inDedupeWindow: false, lastTimestamp: "100"}
		event := newEventWithTimestamps("")

		if !c.forwardEvent(event, state, events) {
			t.Fatal("正常模式无时间戳消息应返回 true")
		}
		if len(events) != 1 {
			t.Errorf("期望转发 1 条消息, 实际 %d", len(events))
		}
		// 无时间戳不更新 lastTimestamp
		if state.lastTimestamp != "100" {
			t.Errorf("lastTimestamp 期望 \"100\", 实际 %q", state.lastTimestamp)
		}
	})
}
