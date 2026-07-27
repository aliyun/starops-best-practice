// Package errors — SDK 统一错误类型与错误码
// 职责：定义 SDKError 结构体和所有错误码常量，供其他模块统一抛出/捕获。
// 不做：不包含业务逻辑、不做重试判断。
// 依赖：无
package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorCode 错误码
// ErrorCode represents the type of error that occurred
type ErrorCode string

const (
	// ErrCodeConfigMissing 配置缺失
	ErrCodeConfigMissing ErrorCode = "CONFIG_MISSING"

	// ErrCodeConfigInvalid 配置无效
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"

	// ErrCodeClientCreate 客户端创建失败
	ErrCodeClientCreate ErrorCode = "CLIENT_CREATE"

	// ErrCodeThreadCreate 会话创建失败
	ErrCodeThreadCreate ErrorCode = "THREAD_CREATE"

	// ErrCodeThreadNotFound 会话不存在
	ErrCodeThreadNotFound ErrorCode = "THREAD_NOT_FOUND"

	// ErrCodeChatFailed 对话失败
	ErrCodeChatFailed ErrorCode = "CHAT_FAILED"

	// ErrCodeTimeout 超时
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// ErrCodeCancelled 已取消
	ErrCodeCancelled ErrorCode = "CANCELLED"

	// ErrCodeNetworkError 网络错误
	ErrCodeNetworkError ErrorCode = "NETWORK_ERROR"

	// ErrCodeAPIError API 错误
	ErrCodeAPIError ErrorCode = "API_ERROR"

	// ErrCodeParseError 解析错误
	ErrCodeParseError ErrorCode = "PARSE_ERROR"

	// ErrCodeInteractiveTimeout 交互超时
	ErrCodeInteractiveTimeout ErrorCode = "INTERACTIVE_TIMEOUT"
)

// SDKError SDK 结构化错误
type SDKError struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	Cause      error          `json:"-"`
	Context    map[string]any `json:"context,omitempty"`
	Suggestion string         `json:"suggestion,omitempty"`
}

// Error 实现 error 接口
func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *SDKError) Unwrap() error {
	return e.Cause
}

// MarshalJSON 自定义 JSON 序列化
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
func NewSDKError(code ErrorCode, message string) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
	}
}

// NewSDKErrorWithCause 创建带原因的 SDK 错误
func NewSDKErrorWithCause(code ErrorCode, message string, cause error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithContext 添加上下文信息
func (e *SDKError) WithContext(key string, value any) *SDKError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithSuggestion 添加建议
func (e *SDKError) WithSuggestion(suggestion string) *SDKError {
	e.Suggestion = suggestion
	return e
}

// IsCode 检查错误码是否匹配
func (e *SDKError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

// --- 便捷错误创建函数 ---

// ErrConfigMissing 创建配置缺失错误
func ErrConfigMissing(missingVars []string) *SDKError {
	return NewSDKError(ErrCodeConfigMissing, fmt.Sprintf("缺少必需的配置项: %v", missingVars)).
		WithContext("missingVariables", missingVars).
		WithSuggestion("请检查 .env 文件或环境变量设置")
}

// ErrConfigInvalid 创建配置无效错误
func ErrConfigInvalid(field, reason string) *SDKError {
	return NewSDKError(ErrCodeConfigInvalid, fmt.Sprintf("配置项 %s 无效: %s", field, reason)).
		WithContext("field", field).
		WithContext("reason", reason).
		WithSuggestion("请检查配置值是否正确")
}

// ErrClientCreate 创建客户端创建失败错误
func ErrClientCreate(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeClientCreate, "创建客户端失败", cause).
		WithSuggestion("请检查网络连接和认证信息")
}

// ErrThreadCreate 创建会话创建失败错误
func ErrThreadCreate(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeThreadCreate, "创建会话失败", cause).
		WithSuggestion("请检查 API 权限和配额")
}

// ErrThreadNotFound 创建会话不存在错误
func ErrThreadNotFound(threadID string) *SDKError {
	return NewSDKError(ErrCodeThreadNotFound, fmt.Sprintf("会话不存在: %s", threadID)).
		WithContext("threadId", threadID).
		WithSuggestion("请检查会话 ID 是否正确，或创建新会话")
}

// ErrChatFailed 创建对话失败错误
func ErrChatFailed(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeChatFailed, "对话失败", cause).
		WithSuggestion("请稍后重试")
}

// ErrTimeout 创建超时错误
func ErrTimeout(duration string) *SDKError {
	return NewSDKError(ErrCodeTimeout, fmt.Sprintf("操作超时: %s", duration)).
		WithContext("duration", duration).
		WithSuggestion("请增加超时时间或检查网络连接")
}

// ErrCancelled 创建已取消错误
func ErrCancelled() *SDKError {
	return NewSDKError(ErrCodeCancelled, "操作已取消").
		WithSuggestion("如需继续，请重新发起请求")
}

// ErrNetworkError 创建网络错误
func ErrNetworkError(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeNetworkError, "网络错误", cause).
		WithSuggestion("请检查网络连接")
}

// ErrAPIError 创建 API 错误
func ErrAPIError(code string, message string) *SDKError {
	return NewSDKError(ErrCodeAPIError, fmt.Sprintf("API 错误 [%s]: %s", code, message)).
		WithContext("apiCode", code).
		WithContext("apiMessage", message).
		WithSuggestion("请参考 API 文档检查请求参数")
}

// ErrParseError 创建解析错误
func ErrParseError(cause error) *SDKError {
	return NewSDKErrorWithCause(ErrCodeParseError, "解析响应失败", cause).
		WithSuggestion("请检查 SDK 版本是否最新")
}

// ErrInteractiveTimeout 创建交互超时错误
func ErrInteractiveTimeout(duration string) *SDKError {
	return NewSDKError(ErrCodeInteractiveTimeout, fmt.Sprintf("等待用户响应超时: %s", duration)).
		WithContext("duration", duration).
		WithSuggestion("请重新操作并在规定时间内响应")
}
