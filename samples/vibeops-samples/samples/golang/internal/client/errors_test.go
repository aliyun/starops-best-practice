// Package client 提供 VibeOps Agent 客户端的公共实现
// errors_test.go 错误模块单元测试
package client

import (
	"encoding/json"
	"errors"
	"testing"
)

// TestNewSDKError 测试创建基本 SDK 错误
// Tests creating a basic SDK error
func TestNewSDKError(t *testing.T) {
	tests := []struct {
		name    string
		code    ErrorCode
		message string
	}{
		{
			name:    "config missing error",
			code:    ErrCodeConfigMissing,
			message: "missing configuration",
		},
		{
			name:    "timeout error",
			code:    ErrCodeTimeout,
			message: "operation timed out",
		},
		{
			name:    "api error",
			code:    ErrCodeAPIError,
			message: "api returned error",
		},
		{
			name:    "network error",
			code:    ErrCodeNetworkError,
			message: "network failure",
		},
		{
			name:    "parse error",
			code:    ErrCodeParseError,
			message: "failed to parse response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSDKError(tt.code, tt.message)

			if err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, err.Code)
			}
			if err.Message != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, err.Message)
			}
			if err.Cause != nil {
				t.Errorf("expected nil cause, got %v", err.Cause)
			}
		})
	}
}

// TestNewSDKErrorWithCause 测试创建带原因的 SDK 错误
// Tests creating an SDK error with an underlying cause
func TestNewSDKErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewSDKErrorWithCause(ErrCodeClientCreate, "client creation failed", cause)

	if err.Code != ErrCodeClientCreate {
		t.Errorf("expected code %s, got %s", ErrCodeClientCreate, err.Code)
	}
	if err.Message != "client creation failed" {
		t.Errorf("expected message 'client creation failed', got %s", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, err.Cause)
	}
}

// TestSDKError_Error 测试错误格式化
// Tests the Error() method formatting
func TestSDKError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SDKError
		expected string
	}{
		{
			name:     "error without cause",
			err:      NewSDKError(ErrCodeTimeout, "operation timed out"),
			expected: "[TIMEOUT] operation timed out",
		},
		{
			name:     "error with cause",
			err:      NewSDKErrorWithCause(ErrCodeNetworkError, "network failure", errors.New("connection refused")),
			expected: "[NETWORK_ERROR] network failure: connection refused",
		},
		{
			name:     "config missing error",
			err:      NewSDKError(ErrCodeConfigMissing, "missing required config"),
			expected: "[CONFIG_MISSING] missing required config",
		},
		{
			name:     "thread not found error",
			err:      NewSDKError(ErrCodeThreadNotFound, "thread does not exist"),
			expected: "[THREAD_NOT_FOUND] thread does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSDKError_Unwrap 测试错误解包
// Tests the Unwrap() method for error chain support
func TestSDKError_Unwrap(t *testing.T) {
	t.Run("unwrap with cause", func(t *testing.T) {
		cause := errors.New("root cause")
		err := NewSDKErrorWithCause(ErrCodeChatFailed, "chat failed", cause)

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Errorf("expected unwrapped error %v, got %v", cause, unwrapped)
		}
	})

	t.Run("unwrap without cause", func(t *testing.T) {
		err := NewSDKError(ErrCodeCancelled, "operation cancelled")

		unwrapped := err.Unwrap()
		if unwrapped != nil {
			t.Errorf("expected nil unwrapped error, got %v", unwrapped)
		}
	})

	t.Run("errors.Is with wrapped error", func(t *testing.T) {
		cause := errors.New("specific error")
		err := NewSDKErrorWithCause(ErrCodeParseError, "parse failed", cause)

		if !errors.Is(err, cause) {
			t.Error("errors.Is should return true for wrapped cause")
		}
	})

	t.Run("nested error chain", func(t *testing.T) {
		rootCause := errors.New("root cause")
		middleErr := NewSDKErrorWithCause(ErrCodeNetworkError, "network issue", rootCause)
		topErr := NewSDKErrorWithCause(ErrCodeChatFailed, "chat failed", middleErr)

		// Should be able to find root cause through chain
		if !errors.Is(topErr, rootCause) {
			t.Error("errors.Is should find root cause through error chain")
		}
		if !errors.Is(topErr, middleErr) {
			t.Error("errors.Is should find middle error in chain")
		}
	})
}

// TestSDKError_WithContext 测试添加上下文
// Tests adding context to an error
func TestSDKError_WithContext(t *testing.T) {
	err := NewSDKError(ErrCodeThreadNotFound, "thread not found").
		WithContext("threadId", "test-123").
		WithContext("workspace", "default")

	if err.Context == nil {
		t.Fatal("expected context to be initialized")
	}
	if err.Context["threadId"] != "test-123" {
		t.Errorf("expected threadId 'test-123', got %v", err.Context["threadId"])
	}
	if err.Context["workspace"] != "default" {
		t.Errorf("expected workspace 'default', got %v", err.Context["workspace"])
	}
}

// TestSDKError_WithSuggestion 测试添加建议
// Tests adding a suggestion to an error
func TestSDKError_WithSuggestion(t *testing.T) {
	suggestion := "Please check your network connection"
	err := NewSDKError(ErrCodeNetworkError, "network error").
		WithSuggestion(suggestion)

	if err.Suggestion != suggestion {
		t.Errorf("expected suggestion %q, got %q", suggestion, err.Suggestion)
	}
}

// TestSDKError_IsCode 测试错误码检查
// Tests checking error code
func TestSDKError_IsCode(t *testing.T) {
	err := NewSDKError(ErrCodeTimeout, "timeout")

	if !err.IsCode(ErrCodeTimeout) {
		t.Error("IsCode should return true for matching code")
	}
	if err.IsCode(ErrCodeCancelled) {
		t.Error("IsCode should return false for non-matching code")
	}
}

// TestSDKError_MarshalJSON 测试 JSON 序列化
// Tests JSON marshaling of SDK error
func TestSDKError_MarshalJSON(t *testing.T) {
	t.Run("marshal without cause", func(t *testing.T) {
		err := NewSDKError(ErrCodeConfigInvalid, "invalid config").
			WithContext("field", "endpoint").
			WithSuggestion("check endpoint format")

		data, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			t.Fatalf("failed to marshal: %v", marshalErr)
		}

		var result map[string]any
		if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
			t.Fatalf("failed to unmarshal: %v", unmarshalErr)
		}

		if result["code"] != string(ErrCodeConfigInvalid) {
			t.Errorf("expected code %s, got %v", ErrCodeConfigInvalid, result["code"])
		}
		if result["message"] != "invalid config" {
			t.Errorf("expected message 'invalid config', got %v", result["message"])
		}
		if result["cause"] != nil {
			t.Errorf("expected no cause, got %v", result["cause"])
		}
	})

	t.Run("marshal with cause", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := NewSDKErrorWithCause(ErrCodeNetworkError, "network error", cause)

		data, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			t.Fatalf("failed to marshal: %v", marshalErr)
		}

		var result map[string]any
		if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
			t.Fatalf("failed to unmarshal: %v", unmarshalErr)
		}

		if result["cause"] != "connection refused" {
			t.Errorf("expected cause 'connection refused', got %v", result["cause"])
		}
	})
}

// TestConvenienceErrorFunctions 测试便捷错误创建函数
// Tests convenience error creation functions
func TestConvenienceErrorFunctions(t *testing.T) {
	t.Run("ErrConfigMissing", func(t *testing.T) {
		missingVars := []string{"ENDPOINT", "ACCESS_KEY"}
		err := ErrConfigMissing(missingVars)

		if err.Code != ErrCodeConfigMissing {
			t.Errorf("expected code %s, got %s", ErrCodeConfigMissing, err.Code)
		}
		if err.Context["missingVariables"] == nil {
			t.Error("expected missingVariables in context")
		}
		if err.Suggestion == "" {
			t.Error("expected suggestion to be set")
		}
	})

	t.Run("ErrConfigInvalid", func(t *testing.T) {
		err := ErrConfigInvalid("endpoint", "invalid URL format")

		if err.Code != ErrCodeConfigInvalid {
			t.Errorf("expected code %s, got %s", ErrCodeConfigInvalid, err.Code)
		}
		if err.Context["field"] != "endpoint" {
			t.Errorf("expected field 'endpoint', got %v", err.Context["field"])
		}
	})

	t.Run("ErrClientCreate", func(t *testing.T) {
		cause := errors.New("auth failed")
		err := ErrClientCreate(cause)

		if err.Code != ErrCodeClientCreate {
			t.Errorf("expected code %s, got %s", ErrCodeClientCreate, err.Code)
		}
		if err.Cause != cause {
			t.Errorf("expected cause %v, got %v", cause, err.Cause)
		}
	})

	t.Run("ErrThreadNotFound", func(t *testing.T) {
		err := ErrThreadNotFound("thread-abc-123")

		if err.Code != ErrCodeThreadNotFound {
			t.Errorf("expected code %s, got %s", ErrCodeThreadNotFound, err.Code)
		}
		if err.Context["threadId"] != "thread-abc-123" {
			t.Errorf("expected threadId 'thread-abc-123', got %v", err.Context["threadId"])
		}
	})

	t.Run("ErrTimeout", func(t *testing.T) {
		err := ErrTimeout("30s")

		if err.Code != ErrCodeTimeout {
			t.Errorf("expected code %s, got %s", ErrCodeTimeout, err.Code)
		}
		if err.Context["duration"] != "30s" {
			t.Errorf("expected duration '30s', got %v", err.Context["duration"])
		}
	})

	t.Run("ErrCancelled", func(t *testing.T) {
		err := ErrCancelled()

		if err.Code != ErrCodeCancelled {
			t.Errorf("expected code %s, got %s", ErrCodeCancelled, err.Code)
		}
	})

	t.Run("ErrNetworkError", func(t *testing.T) {
		cause := errors.New("connection timeout")
		err := ErrNetworkError(cause)

		if err.Code != ErrCodeNetworkError {
			t.Errorf("expected code %s, got %s", ErrCodeNetworkError, err.Code)
		}
		if err.Cause != cause {
			t.Errorf("expected cause %v, got %v", cause, err.Cause)
		}
	})

	t.Run("ErrAPIError", func(t *testing.T) {
		err := ErrAPIError("InvalidParameter", "parameter X is invalid")

		if err.Code != ErrCodeAPIError {
			t.Errorf("expected code %s, got %s", ErrCodeAPIError, err.Code)
		}
		if err.Context["apiCode"] != "InvalidParameter" {
			t.Errorf("expected apiCode 'InvalidParameter', got %v", err.Context["apiCode"])
		}
	})

	t.Run("ErrParseError", func(t *testing.T) {
		cause := errors.New("invalid JSON")
		err := ErrParseError(cause)

		if err.Code != ErrCodeParseError {
			t.Errorf("expected code %s, got %s", ErrCodeParseError, err.Code)
		}
		if err.Cause != cause {
			t.Errorf("expected cause %v, got %v", cause, err.Cause)
		}
	})

	t.Run("ErrInteractiveTimeout", func(t *testing.T) {
		err := ErrInteractiveTimeout("60s")

		if err.Code != ErrCodeInteractiveTimeout {
			t.Errorf("expected code %s, got %s", ErrCodeInteractiveTimeout, err.Code)
		}
		if err.Context["duration"] != "60s" {
			t.Errorf("expected duration '60s', got %v", err.Context["duration"])
		}
	})
}

// TestAllErrorCodes 测试所有错误码
// Tests that all error codes are distinct and properly defined
func TestAllErrorCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeConfigMissing,
		ErrCodeConfigInvalid,
		ErrCodeClientCreate,
		ErrCodeThreadCreate,
		ErrCodeThreadNotFound,
		ErrCodeChatFailed,
		ErrCodeTimeout,
		ErrCodeCancelled,
		ErrCodeNetworkError,
		ErrCodeAPIError,
		ErrCodeParseError,
		ErrCodeInteractiveTimeout,
	}

	// Check all codes are non-empty
	for _, code := range codes {
		if code == "" {
			t.Error("error code should not be empty")
		}
	}

	// Check all codes are unique
	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("duplicate error code: %s", code)
		}
		seen[code] = true
	}
}

// TestSDKError_ChainedMethods 测试链式方法调用
// Tests chained method calls
func TestSDKError_ChainedMethods(t *testing.T) {
	err := NewSDKError(ErrCodeAPIError, "api error").
		WithContext("requestId", "req-123").
		WithContext("statusCode", 400).
		WithSuggestion("check request parameters")

	if err.Code != ErrCodeAPIError {
		t.Errorf("expected code %s, got %s", ErrCodeAPIError, err.Code)
	}
	if err.Context["requestId"] != "req-123" {
		t.Errorf("expected requestId 'req-123', got %v", err.Context["requestId"])
	}
	if err.Context["statusCode"] != 400 {
		t.Errorf("expected statusCode 400, got %v", err.Context["statusCode"])
	}
	if err.Suggestion != "check request parameters" {
		t.Errorf("expected suggestion 'check request parameters', got %s", err.Suggestion)
	}
}
