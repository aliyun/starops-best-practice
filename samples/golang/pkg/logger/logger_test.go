package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// LogLevel.String
// ---------------------------------------------------------------------------

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
		{LogLevel(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("LogLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseLogLevel
// ---------------------------------------------------------------------------

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  LogLevel
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"", LevelInfo},
		{"garbage", LevelInfo},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			got := ParseLogLevel(tt.input)
			if got != tt.want {
				t.Errorf("ParseLogLevel(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Level filtering
// ---------------------------------------------------------------------------

func TestLevelFiltering_DebugAtDebug(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelDebug, &buf)
	l.Debug("test debug", nil)
	if buf.Len() == 0 {
		t.Error("expected debug message output at LevelDebug")
	}
}

func TestLevelFiltering_DebugAtInfo(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelInfo, &buf)
	l.Debug("test debug", nil)
	if buf.Len() != 0 {
		t.Error("expected debug message filtered at LevelInfo")
	}
}

func TestLevelFiltering_InfoAtInfo(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelInfo, &buf)
	l.Info("test info", nil)
	if buf.Len() == 0 {
		t.Error("expected info message output at LevelInfo")
	}
}

func TestLevelFiltering_WarnAtError(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelError, &buf)
	l.Warn("test warn", nil)
	if buf.Len() != 0 {
		t.Error("expected warn message filtered at LevelError")
	}
}

func TestLevelFiltering_ErrorAtError(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelError, &buf)
	l.Error("test error", nil, nil)
	if buf.Len() == 0 {
		t.Error("expected error message output at LevelError")
	}
}

// ---------------------------------------------------------------------------
// Structured output
// ---------------------------------------------------------------------------

func TestStructuredOutput_InfoBasic(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelInfo, &buf)
	l.Info("test msg", nil)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if entry.Level != "info" {
		t.Errorf("level = %q, want %q", entry.Level, "info")
	}
	if entry.Message != "test msg" {
		t.Errorf("message = %q, want %q", entry.Message, "test msg")
	}
	// Verify timestamp is RFC3339
	if _, err := time.Parse(time.RFC3339, entry.Timestamp); err != nil {
		t.Errorf("timestamp %q is not valid RFC3339: %v", entry.Timestamp, err)
	}
}

func TestStructuredOutput_InfoWithContext(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelInfo, &buf)
	ctx := map[string]any{"key1": "value1", "key2": float64(42)}
	l.Info("with context", ctx)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if entry.Context == nil {
		t.Fatal("expected non-nil context")
	}
	if v, ok := entry.Context["key1"]; !ok || v != "value1" {
		t.Errorf("context[key1] = %v, want %q", v, "value1")
	}
	if v, ok := entry.Context["key2"]; !ok || v != float64(42) {
		t.Errorf("context[key2] = %v, want 42", v)
	}
}

func TestStructuredOutput_ErrorWithCause(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelError, &buf)
	l.Error("something failed", fmt.Errorf("root cause"), nil)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if entry.Error == "" {
		t.Error("expected non-empty error field")
	}
	if !strings.Contains(entry.Error, "root cause") {
		t.Errorf("error field = %q, want to contain 'root cause'", entry.Error)
	}
}

func TestStructuredOutput_ErrorIncludesStack(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelError, &buf)
	l.Error("stack test", fmt.Errorf("err"), nil)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if entry.Stack == "" {
		t.Error("expected non-empty stack field for error log")
	}
}

// ---------------------------------------------------------------------------
// SetLevel / GetLevel
// ---------------------------------------------------------------------------

func TestSetLevel_GetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LevelInfo, &buf)
	l.SetLevel(LevelWarn)
	if l.GetLevel() != LevelWarn {
		t.Errorf("GetLevel() = %v, want LevelWarn", l.GetLevel())
	}
}

// ---------------------------------------------------------------------------
// Default / SetDefault
// ---------------------------------------------------------------------------

func TestDefault_ReturnsNonNil(t *testing.T) {
	// Save and restore the package-level defaultLogger
	saved := defaultLogger
	t.Cleanup(func() { defaultLogger = saved })

	// Reset so Default() re-initializes
	defaultLogger = nil
	d := Default()
	if d == nil {
		t.Error("Default() returned nil")
	}
}

func TestSetDefault_OverridesDefault(t *testing.T) {
	saved := defaultLogger
	t.Cleanup(func() { defaultLogger = saved })

	var buf bytes.Buffer
	custom := NewLogger(LevelWarn, &buf)
	SetDefault(custom)

	got := Default()
	if got != custom {
		t.Error("Default() did not return the custom logger set via SetDefault")
	}
}

// ---------------------------------------------------------------------------
// NewLoggerFromEnv
// ---------------------------------------------------------------------------

func TestNewLoggerFromEnv_DebugLevel(t *testing.T) {
	t.Setenv("STAROPS_LOG_LEVEL", "debug")
	l := NewLoggerFromEnv()
	if l.GetLevel() != LevelDebug {
		t.Errorf("GetLevel() = %v, want LevelDebug", l.GetLevel())
	}
}

func TestNewLoggerFromEnv_EmptyDefault(t *testing.T) {
	t.Setenv("STAROPS_LOG_LEVEL", "")
	l := NewLoggerFromEnv()
	if l.GetLevel() != LevelInfo {
		t.Errorf("GetLevel() = %v, want LevelInfo (default)", l.GetLevel())
	}
}
