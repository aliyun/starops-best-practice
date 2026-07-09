// Package logger 提供结构化日志功能
// Package logger provides structured logging functionality
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// LevelDebug 调试级别 - 详细的调试信息
	// LevelDebug - Detailed debug information
	LevelDebug LogLevel = iota
	// LevelInfo 信息级别 - 一般运行信息
	// LevelInfo - General operational information
	LevelInfo
	// LevelWarn 警告级别 - 潜在问题警告
	// LevelWarn - Warning about potential issues
	LevelWarn
	// LevelError 错误级别 - 错误信息
	// LevelError - Error information
	LevelError
)

// String 返回日志级别的字符串表示
// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLogLevel 从字符串解析日志级别
// ParseLogLevel parses a log level from a string
func ParseLogLevel(s string) LogLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo // 默认为 info 级别
	}
}

// LogEntry 日志条目
// LogEntry represents a single log entry with structured fields
type LogEntry struct {
	// Timestamp ISO 8601 格式的时间戳
	// Timestamp in ISO 8601 format
	Timestamp string `json:"timestamp"`

	// Level 日志级别
	// Level is the severity level
	Level string `json:"level"`

	// Message 日志消息
	// Message is the log message
	Message string `json:"message"`

	// Context 上下文信息
	// Context contains additional contextual information
	Context map[string]any `json:"context,omitempty"`

	// Error 错误信息
	// Error contains error details if applicable
	Error string `json:"error,omitempty"`

	// Stack 堆栈跟踪
	// Stack contains the stack trace for errors
	Stack string `json:"stack,omitempty"`
}

// Logger 结构化日志器
// Logger provides structured logging with level filtering
type Logger struct {
	level  LogLevel
	output io.Writer
}

// NewLogger 创建日志器
// NewLogger creates a new logger with the specified level and output
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		level:  level,
		output: output,
	}
}

// NewLoggerFromEnv 从环境变量创建日志器
// NewLoggerFromEnv creates a logger configured from the LOG_LEVEL environment variable
// 支持的级别: debug, info, warn, error
// Supported levels: debug, info, warn, error
func NewLoggerFromEnv() *Logger {
	levelStr := os.Getenv("LOG_LEVEL")
	level := ParseLogLevel(levelStr)
	return NewLogger(level, os.Stdout)
}

// SetLevel 设置日志级别
// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel 获取当前日志级别
// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// SetOutput 设置输出目标
// SetOutput sets the output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// log 内部日志方法
// log is the internal logging method
func (l *Logger) log(level LogLevel, msg string, ctx map[string]any, err error, includeStack bool) {
	// 级别过滤
	// Level filtering
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   msg,
		Context:   ctx,
	}

	// 添加错误信息
	// Add error information
	if err != nil {
		entry.Error = err.Error()
	}

	// 添加堆栈跟踪
	// Add stack trace
	if includeStack {
		entry.Stack = getStackTrace(3) // 跳过 log, Error, 和调用者
	}

	// 序列化为 JSON
	// Serialize to JSON
	jsonBytes, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		// 如果 JSON 序列化失败，输出简单格式
		// If JSON serialization fails, output simple format
		fmt.Fprintf(l.output, "%s [%s] %s\n", entry.Timestamp, entry.Level, msg)
		return
	}

	fmt.Fprintln(l.output, string(jsonBytes))
}

// Debug 调试日志
// Debug logs a debug level message
func (l *Logger) Debug(msg string, ctx map[string]any) {
	l.log(LevelDebug, msg, ctx, nil, false)
}

// Info 信息日志
// Info logs an info level message
func (l *Logger) Info(msg string, ctx map[string]any) {
	l.log(LevelInfo, msg, ctx, nil, false)
}

// Warn 警告日志
// Warn logs a warning level message
func (l *Logger) Warn(msg string, ctx map[string]any) {
	l.log(LevelWarn, msg, ctx, nil, false)
}

// Error 错误日志
// Error logs an error level message with stack trace
func (l *Logger) Error(msg string, err error, ctx map[string]any) {
	l.log(LevelError, msg, ctx, err, true)
}

// ChatEvent 用于 LogResponse 的事件接口
// ChatEvent interface for LogResponse
type ChatEvent interface {
	GetRawJSON() string
	GetStatusCode() int32
	IsDoneEvent() bool
	GetError() error
}

// LogRequest 记录请求
// LogRequest logs request parameters at debug level
func (l *Logger) LogRequest(threadID, message string, variables map[string]any) {
	ctx := map[string]any{
		"threadId": threadID,
		"message":  message,
	}
	if variables != nil {
		ctx["variables"] = variables
	}
	l.Debug("发送请求 / Sending request", ctx)
}

// LogResponse 记录响应
// LogResponse logs response summary at debug level
func (l *Logger) LogResponse(threadID string, statusCode int32, rawJSON string, isDone bool, err error) {
	ctx := map[string]any{
		"threadId":   threadID,
		"statusCode": statusCode,
		"isDone":     isDone,
	}

	// 尝试解析 JSON 获取摘要信息
	// Try to parse JSON for summary information
	if rawJSON != "" {
		var summary map[string]any
		if jsonErr := json.Unmarshal([]byte(rawJSON), &summary); jsonErr == nil {
			// 提取关键字段作为摘要
			// Extract key fields as summary
			if messages, ok := summary["messages"]; ok {
				if msgArray, ok := messages.([]any); ok {
					ctx["messageCount"] = len(msgArray)
				}
			}
		}
		// 限制 rawJSON 长度以避免日志过大
		// Limit rawJSON length to avoid oversized logs
		if len(rawJSON) > 500 {
			ctx["rawJSON"] = rawJSON[:500] + "...(truncated)"
		} else {
			ctx["rawJSON"] = rawJSON
		}
	}

	if err != nil {
		l.Error("响应错误 / Response error", err, ctx)
	} else {
		l.Debug("收到响应 / Received response", ctx)
	}
}

// LogResponseEvent 记录响应事件（便捷方法）
// LogResponseEvent logs a response event (convenience method)
func (l *Logger) LogResponseEvent(threadID string, event ChatEvent) {
	if event == nil {
		return
	}
	l.LogResponse(threadID, event.GetStatusCode(), event.GetRawJSON(), event.IsDoneEvent(), event.GetError())
}

// getStackTrace 获取堆栈跟踪
// getStackTrace returns the stack trace
func getStackTrace(skip int) string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var builder strings.Builder

	for {
		frame, more := frames.Next()
		// 跳过 runtime 内部函数
		// Skip runtime internal functions
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}

		fmt.Fprintf(&builder, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)

		if !more {
			break
		}
	}

	return builder.String()
}
