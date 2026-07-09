// Package client 提供 VibeOps Agent 客户端的公共实现
// errors.go 定义 SDK 错误类型和错误码
package client

import (
	"encoding/json"
	"fmt"
)

// ErrorCode 错误码
// ErrorCode represents the type of error that occurred
type ErrorCode string

const (
	// ErrCodeConfigMissing 配置缺失
	// Configuration is missing required fields
	ErrCodeConfigMissing ErrorCode = "CONFIG_MISSING"

	// ErrCodeConfigInvalid 配置无效
	// Configuration contains invalid values
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"

	// ErrCodeClientCreate 客户端创建失败
	// Failed to create the client
	ErrCodeClientCreate ErrorCode = "CLIENT_CREATE"

	// ErrCodeThreadCreate 会话创建失败
	// Failed to create a thread
	ErrCodeThreadCreate ErrorCode = "THREAD_CREATE"

	// ErrCodeThreadNotFound 会话不存在
	// Thread was not found
	ErrCodeThreadNotFound ErrorCode = "THREAD_NOT_FOUND"

	// ErrCodeChatFailed 对话失败
	// Chat operation failed
	ErrCodeChatFailed ErrorCode = "CHAT_FAILED"

	// ErrCodeTimeout 超时
	// Operation timed out
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// ErrCodeCancelled 已取消
	// Operation was cancelled
	ErrCodeCancelled ErrorCode = "CANCELLED"

	// ErrCodeNetworkError 网络错误
	// Network error occurred
	ErrCodeNetworkError ErrorCode = "NETWORK_ERROR"

	// ErrCodeAPIError API 错误
	// API returned an error
	ErrCodeAPIError ErrorCode = "API_ERROR"

	// ErrCodeParseError 解析错误
	// Failed to parse response
	ErrCodeParseError ErrorCode = "PARSE_ERROR"

	// ErrCodeInteractiveTimeout 交互超时
	// Interactive operation timed out waiting for user response
	ErrCodeInteractiveTimeout ErrorCode = "INTERACTIVE_TIMEOUT"
)

// SDKError SDK 错误
// SDKError represents a structured error from the SDK
type SDKError struct {
	// Code 错误码
	// The error code identifying the type of error
	Code ErrorCode `json:"code"`

	// Message 错误消息
	// Human-readable error message
	Message string `json:"message"`

	// Cause 原因错误
	// The underlying error that caused this error (if any)
	Cause error `json:"-"`

	// Context 上下文信息
	// Additional context about the error
	Context map[string]any `json:"context,omitempty"`

	// Suggestion 建议
	// Suggested action to resolve the error
	Suggestion string `json:"suggestion,omitempty"`
}

// Error 实现 error 接口
// Error implements the error interface
func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
// Unwrap implements the errors.Unwrap interface for error chain support
func (e *SDKError) Unwrap() error {
	return e.Cause
}

// MarshalJSON 自定义 JSON 序列化
// MarshalJSON implements custom JSON marshaling to include cause as string
func (e *SDKError) MarshalJSON() ([]byte, error) {
	type Alias SDKError
	aux := &struct {
		*Alias
		Cause string `json:"cause,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	if e.Cause != nil {
		aux.Cause = e.Cause.Error()
	}
	return json.Marshal(aux)
}

// NewSDKError 创建新的 SDK 错误
// NewSDKError creates a new SDKError with the given code and message
func NewSDKError(code ErrorCode, message string) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
	}
}

// NewSDKErrorWithCause 创建带原因的 SDK 错误
// NewSDKErrorWithCause creates a new SDKError with an underlying cause
func NewSDKErrorWithCause(code ErrorCode, message string, cause error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithContext 添加上下文信息
// WithContext adds context information to the error
func (e *SDKError) WithContext(key string, value any) *SDKError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithSuggestion 添加建议
// WithSuggestion adds a suggestion for resolving the error
func (e *SDKError) WithSuggestion(suggestion string) *SDKError {
	e.Suggestion = suggestion
	return e
}

// IsCode 检查错误码是否匹配
// IsCode checks if the error has the specified error code
func (e *SDKError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

// --- 便捷错误创建函数 / Convenience error creation functions ---

// ErrConfigMissing 创建配置缺失错误
// ErrConfigMissing creates a configuration missing error
func ErrConfigMissing(missingVars []string) *SDKError {
	return NewSDKError(ErrCodeConfigMissing, fmt.Sprintf("缺少必需的配置项: %v", missingVars)).
		WithContext("missingVariables", missingVars).
		WithSuggestion("请检查 .env 文件或环境变量设置")
}

// ErrConfigInvalid 创建配置无效错误
// ErrConfigInvalid creates a configuration invalid error
func ErrConfigInvalid(field, reason string) *SDKError {
	return NewSDKError(ErrCodeConfigInvalid, fmt.Sprintf("配置项 %s 无效: %s", field, reason)).
		WithContext("field", field).
		WithContext("reason", reason).
		WithSuggestion("请检查配置值是否正确")
}

// ErrClientCreate 创建客户端创建失败错误
// ErrClientCreate creates a client creation error
func ErrClientCreate(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeClientCreate, "创建客户端失败", cause).
		WithSuggestion("请检查网络连接和认证信息")
}

// ErrThreadCreate 创建会话创建失败错误
// ErrThreadCreate creates a thread creation error
func ErrThreadCreate(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeThreadCreate, "创建会话失败", cause).
		WithSuggestion("请检查 API 权限和配额")
}

// ErrThreadNotFound 创建会话不存在错误
// ErrThreadNotFound creates a thread not found error
func ErrThreadNotFound(threadID string) *SDKError {
	return NewSDKError(ErrCodeThreadNotFound, fmt.Sprintf("会话不存在: %s", threadID)).
		WithContext("threadId", threadID).
		WithSuggestion("请检查会话 ID 是否正确，或创建新会话")
}

// ErrChatFailed 创建对话失败错误
// ErrChatFailed creates a chat failed error
func ErrChatFailed(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeChatFailed, "对话失败", cause).
		WithSuggestion("请稍后重试")
}

// ErrTimeout 创建超时错误
// ErrTimeout creates a timeout error
func ErrTimeout(duration string) *SDKError {
	return NewSDKError(ErrCodeTimeout, fmt.Sprintf("操作超时: %s", duration)).
		WithContext("duration", duration).
		WithSuggestion("请增加超时时间或检查网络连接")
}

// ErrCancelled 创建已取消错误
// ErrCancelled creates a cancelled error
func ErrCancelled() *SDKError {
	return NewSDKError(ErrCodeCancelled, "操作已取消").
		WithSuggestion("如需继续，请重新发起请求")
}

// ErrNetworkError 创建网络错误
// ErrNetworkError creates a network error
func ErrNetworkError(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeNetworkError, "网络错误", cause).
		WithSuggestion("请检查网络连接")
}

// ErrAPIError 创建 API 错误
// ErrAPIError creates an API error
func ErrAPIError(code string, message string) *SDKError {
	return NewSDKError(ErrCodeAPIError, fmt.Sprintf("API 错误 [%s]: %s", code, message)).
		WithContext("apiCode", code).
		WithContext("apiMessage", message).
		WithSuggestion("请参考 API 文档检查请求参数")
}

// ErrParseError 创建解析错误
// ErrParseError creates a parse error
func ErrParseError(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeParseError, "解析响应失败", cause).
		WithSuggestion("请检查 SDK 版本是否最新")
}

// ErrInteractiveTimeout 创建交互超时错误
// ErrInteractiveTimeout creates an interactive timeout error
func ErrInteractiveTimeout(duration string) *SDKError {
	return NewSDKError(ErrCodeInteractiveTimeout, fmt.Sprintf("等待用户响应超时: %s", duration)).
		WithContext("duration", duration).
		WithSuggestion("请重新操作并在规定时间内响应")
}
