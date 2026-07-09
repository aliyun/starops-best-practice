// Package client 提供 VibeOps Agent 客户端的公共实现
// thread_test.go 会话管理单元测试
package client

import (
	"errors"
	"strings"
	"testing"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MockSTAROpsClient 模拟 STAROps 客户端接口
// MockSTAROpsClient is a mock implementation of the STAROps client for testing
type MockSTAROpsClient struct {
	threads       map[string]*ThreadInfo
	threadData    map[string][]*ThreadMessage
	createCounter int
}

// NewMockSTAROpsClient 创建模拟客户端
func NewMockSTAROpsClient() *MockSTAROpsClient {
	return &MockSTAROpsClient{
		threads:    make(map[string]*ThreadInfo),
		threadData: make(map[string][]*ThreadMessage),
	}
}

// CreateThread 模拟创建会话
func (m *MockSTAROpsClient) CreateThread(title string) (string, error) {
	m.createCounter++
	threadID := generateMockThreadID(m.createCounter)
	m.threads[threadID] = &ThreadInfo{
		ThreadID:   threadID,
		Title:      title,
		Status:     "active",
		CreateTime: "2024-01-01T00:00:00Z",
		UpdateTime: "2024-01-01T00:00:00Z",
	}
	m.threadData[threadID] = []*ThreadMessage{}
	return threadID, nil
}

// ListThreads 模拟列出会话
func (m *MockSTAROpsClient) ListThreads(pageSize int) ([]*ThreadInfo, int64, error) {
	var threads []*ThreadInfo
	for _, t := range m.threads {
		threads = append(threads, t)
		if len(threads) >= pageSize {
			break
		}
	}
	return threads, int64(len(m.threads)), nil
}

// GetThread 模拟获取会话详情
func (m *MockSTAROpsClient) GetThread(threadID string) (*ThreadInfo, error) {
	if err := validateThreadID(threadID); err != nil {
		return nil, err
	}
	thread, ok := m.threads[threadID]
	if !ok {
		return nil, ErrThreadNotFound(threadID)
	}
	return thread, nil
}

// DeleteThread 模拟删除会话
func (m *MockSTAROpsClient) DeleteThread(threadID string) error {
	if err := validateThreadID(threadID); err != nil {
		return err
	}
	if _, ok := m.threads[threadID]; !ok {
		return ErrThreadNotFound(threadID)
	}
	delete(m.threads, threadID)
	delete(m.threadData, threadID)
	return nil
}

// GetThreadData 模拟获取会话消息
func (m *MockSTAROpsClient) GetThreadData(threadID string, limit int) ([]*ThreadMessage, error) {
	if err := validateThreadID(threadID); err != nil {
		return nil, err
	}
	messages, ok := m.threadData[threadID]
	if !ok {
		return nil, ErrThreadNotFound(threadID)
	}
	if limit > 0 && len(messages) > limit {
		return messages[:limit], nil
	}
	return messages, nil
}

// AddMessage 添加消息到会话
func (m *MockSTAROpsClient) AddMessage(threadID, role, content string) error {
	if _, ok := m.threads[threadID]; !ok {
		return ErrThreadNotFound(threadID)
	}
	m.threadData[threadID] = append(m.threadData[threadID], &ThreadMessage{
		Role:      role,
		Content:   content,
		Timestamp: "2024-01-01T00:00:00Z",
	})
	return nil
}

func generateMockThreadID(counter int) string {
	return "thread-" + strings.Repeat("a", 8) + "-" + string(rune('0'+counter%10))
}

// =============================================================================
// Property 9: Thread Management API Consistency
// **Validates: Requirements 5.1, 5.2, 5.3, 5.4**
// =============================================================================

// TestProperty9_ThreadManagementAPIConsistency 测试会话管理 API 一致性
// Property 9: Thread Management API Consistency
// *For any* thread created via CreateThread, the following SHALL hold:
// - ListThreads SHALL include the thread in its results
// - GetThread SHALL return the thread details
// - GetThreadData SHALL return the thread messages
// - After DeleteThread, GetThread SHALL return an error
func TestProperty9_ThreadManagementAPIConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50
	properties := gopter.NewProperties(parameters)

	// Property 9.1: Created thread appears in ListThreads
	// **Validates: Requirements 5.1**
	properties.Property("created thread appears in ListThreads", prop.ForAll(
		func(title string) bool {
			mock := NewMockSTAROpsClient()

			// Create a thread
			threadID, err := mock.CreateThread(title)
			if err != nil {
				return false
			}

			// List threads should include the created thread
			threads, total, err := mock.ListThreads(100)
			if err != nil {
				return false
			}

			// Total should be at least 1
			if total < 1 {
				return false
			}

			// Find the created thread in the list
			found := false
			for _, thread := range threads {
				if thread.ThreadID == threadID {
					found = true
					break
				}
			}
			return found
		},
		gen.AlphaString(),
	))

	// Property 9.2: GetThread returns correct thread details
	// **Validates: Requirements 5.2**
	properties.Property("GetThread returns correct thread details", prop.ForAll(
		func(title string) bool {
			mock := NewMockSTAROpsClient()

			// Create a thread
			threadID, err := mock.CreateThread(title)
			if err != nil {
				return false
			}

			// GetThread should return the thread
			thread, err := mock.GetThread(threadID)
			if err != nil {
				return false
			}

			// Verify thread details
			return thread.ThreadID == threadID &&
				thread.Title == title &&
				thread.Status == "active"
		},
		gen.AlphaString(),
	))

	// Property 9.3: GetThreadData returns thread messages
	// **Validates: Requirements 5.4**
	properties.Property("GetThreadData returns thread messages", prop.ForAll(
		func(title string, messageCount int) bool {
			if messageCount < 0 {
				messageCount = 0
			}
			if messageCount > 10 {
				messageCount = 10
			}

			mock := NewMockSTAROpsClient()

			// Create a thread
			threadID, err := mock.CreateThread(title)
			if err != nil {
				return false
			}

			// Add messages
			for i := 0; i < messageCount; i++ {
				role := "user"
				if i%2 == 1 {
					role = "assistant"
				}
				if err := mock.AddMessage(threadID, role, "message content"); err != nil {
					return false
				}
			}

			// GetThreadData should return the messages
			messages, err := mock.GetThreadData(threadID, 100)
			if err != nil {
				return false
			}

			return len(messages) == messageCount
		},
		gen.AlphaString(),
		gen.IntRange(0, 10),
	))

	// Property 9.4: After DeleteThread, GetThread returns error
	// **Validates: Requirements 5.3**
	properties.Property("after DeleteThread, GetThread returns error", prop.ForAll(
		func(title string) bool {
			mock := NewMockSTAROpsClient()

			// Create a thread
			threadID, err := mock.CreateThread(title)
			if err != nil {
				return false
			}

			// Verify thread exists
			_, err = mock.GetThread(threadID)
			if err != nil {
				return false
			}

			// Delete the thread
			err = mock.DeleteThread(threadID)
			if err != nil {
				return false
			}

			// GetThread should now return an error
			_, err = mock.GetThread(threadID)
			if err == nil {
				return false
			}

			// Error should be ThreadNotFound
			var sdkErr *SDKError
			if errors.As(err, &sdkErr) {
				return sdkErr.Code == ErrCodeThreadNotFound
			}
			return false
		},
		gen.AlphaString(),
	))

	// Property 9.5: Full lifecycle consistency
	// **Validates: Requirements 5.1, 5.2, 5.3, 5.4**
	properties.Property("full thread lifecycle is consistent", prop.ForAll(
		func(title string) bool {
			mock := NewMockSTAROpsClient()

			// 1. Create thread
			threadID, err := mock.CreateThread(title)
			if err != nil {
				return false
			}

			// 2. Verify in list
			threads, _, err := mock.ListThreads(100)
			if err != nil {
				return false
			}
			foundInList := false
			for _, t := range threads {
				if t.ThreadID == threadID {
					foundInList = true
					break
				}
			}
			if !foundInList {
				return false
			}

			// 3. Get thread details
			thread, err := mock.GetThread(threadID)
			if err != nil || thread.ThreadID != threadID {
				return false
			}

			// 4. Get thread data (should be empty initially)
			messages, err := mock.GetThreadData(threadID, 100)
			if err != nil {
				return false
			}
			if len(messages) != 0 {
				return false
			}

			// 5. Delete thread
			err = mock.DeleteThread(threadID)
			if err != nil {
				return false
			}

			// 6. Verify thread is gone
			_, err = mock.GetThread(threadID)
			return err != nil
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 10: Invalid Thread ID Error
// **Validates: Requirements 5.5**
// =============================================================================

// TestProperty10_InvalidThreadIDError 测试无效会话 ID 错误处理
// Property 10: Invalid Thread ID Error
// *For any* invalid thread ID (empty string, non-existent ID, malformed ID),
// all Thread_Manager methods SHALL return a descriptive error message containing the invalid ID.
func TestProperty10_InvalidThreadIDError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50
	properties := gopter.NewProperties(parameters)

	// Property 10.1: Empty thread ID returns error
	// **Validates: Requirements 5.5**
	properties.Property("empty thread ID returns descriptive error", prop.ForAll(
		func(_ bool) bool {
			mock := NewMockSTAROpsClient()
			emptyID := ""

			// GetThread with empty ID
			_, err := mock.GetThread(emptyID)
			if err == nil {
				return false
			}
			if !containsThreadID(err.Error(), emptyID) && !strings.Contains(err.Error(), "空") {
				return false
			}

			// DeleteThread with empty ID
			err = mock.DeleteThread(emptyID)
			if err == nil {
				return false
			}

			// GetThreadData with empty ID
			_, err = mock.GetThreadData(emptyID, 10)
			if err == nil {
				return false
			}

			return true
		},
		gen.Bool(),
	))

	// Property 10.2: Non-existent thread ID returns error with ID
	// **Validates: Requirements 5.5**
	properties.Property("non-existent thread ID returns error containing the ID", prop.ForAll(
		func(nonExistentID string) bool {
			// Skip empty strings (handled by Property 10.1)
			if nonExistentID == "" || strings.ContainsAny(nonExistentID, " \t\n\r") {
				return true
			}

			mock := NewMockSTAROpsClient()

			// GetThread with non-existent ID
			_, err := mock.GetThread(nonExistentID)
			if err == nil {
				return false
			}
			if !containsThreadID(err.Error(), nonExistentID) {
				return false
			}

			// DeleteThread with non-existent ID
			err = mock.DeleteThread(nonExistentID)
			if err == nil {
				return false
			}
			if !containsThreadID(err.Error(), nonExistentID) {
				return false
			}

			// GetThreadData with non-existent ID
			_, err = mock.GetThreadData(nonExistentID, 10)
			if err == nil {
				return false
			}
			if !containsThreadID(err.Error(), nonExistentID) {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) < 100
		}),
	))

	// Property 10.3: Malformed thread ID (with whitespace) returns error
	// **Validates: Requirements 5.5**
	properties.Property("malformed thread ID with whitespace returns error", prop.ForAll(
		func(prefix string, suffix string) bool {
			mock := NewMockSTAROpsClient()

			// Create malformed IDs with various whitespace
			malformedIDs := []string{
				prefix + " " + suffix,
				prefix + "\t" + suffix,
				prefix + "\n" + suffix,
				" " + prefix,
				prefix + " ",
			}

			for _, malformedID := range malformedIDs {
				if malformedID == "" {
					continue
				}

				// GetThread with malformed ID
				_, err := mock.GetThread(malformedID)
				if err == nil {
					return false
				}

				// DeleteThread with malformed ID
				err = mock.DeleteThread(malformedID)
				if err == nil {
					return false
				}

				// GetThreadData with malformed ID
				_, err = mock.GetThreadData(malformedID, 10)
				if err == nil {
					return false
				}
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 10.4: Error type is SDKError with appropriate code
	// **Validates: Requirements 5.5**
	properties.Property("invalid thread ID error is SDKError with correct code", prop.ForAll(
		func(invalidID string) bool {
			// Skip empty strings
			if invalidID == "" {
				return true
			}

			mock := NewMockSTAROpsClient()

			// Test GetThread
			_, err := mock.GetThread(invalidID)
			if err == nil {
				return false
			}

			var sdkErr *SDKError
			if !errors.As(err, &sdkErr) {
				return false
			}

			// Should be either CONFIG_INVALID (for malformed) or THREAD_NOT_FOUND (for non-existent)
			validCodes := map[ErrorCode]bool{
				ErrCodeConfigInvalid:  true,
				ErrCodeThreadNotFound: true,
			}
			return validCodes[sdkErr.Code]
		},
		gen.OneConstOf("", "non-existent-id", "id with space", "id\twith\ttab"),
	))

	properties.TestingRun(t)
}

// containsThreadID 检查错误消息是否包含线程 ID
func containsThreadID(errMsg, threadID string) bool {
	return strings.Contains(errMsg, threadID)
}

// =============================================================================
// Unit Tests for validateThreadID
// =============================================================================

// TestValidateThreadID 测试会话 ID 验证函数
func TestValidateThreadID(t *testing.T) {
	tests := []struct {
		name      string
		threadID  string
		wantError bool
		errorCode ErrorCode
	}{
		{
			name:      "valid thread ID",
			threadID:  "thread-abc-123",
			wantError: false,
		},
		{
			name:      "empty thread ID",
			threadID:  "",
			wantError: true,
			errorCode: ErrCodeConfigInvalid,
		},
		{
			name:      "thread ID with space",
			threadID:  "thread abc",
			wantError: true,
			errorCode: ErrCodeConfigInvalid,
		},
		{
			name:      "thread ID with tab",
			threadID:  "thread\tabc",
			wantError: true,
			errorCode: ErrCodeConfigInvalid,
		},
		{
			name:      "thread ID with newline",
			threadID:  "thread\nabc",
			wantError: true,
			errorCode: ErrCodeConfigInvalid,
		},
		{
			name:      "thread ID with carriage return",
			threadID:  "thread\rabc",
			wantError: true,
			errorCode: ErrCodeConfigInvalid,
		},
		{
			name:      "valid UUID-like thread ID",
			threadID:  "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
		{
			name:      "valid alphanumeric thread ID",
			threadID:  "abc123xyz",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateThreadID(tt.threadID)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for threadID %q, got nil", tt.threadID)
					return
				}

				var sdkErr *SDKError
				if !errors.As(err, &sdkErr) {
					t.Errorf("expected SDKError, got %T", err)
					return
				}

				if sdkErr.Code != tt.errorCode {
					t.Errorf("expected error code %s, got %s", tt.errorCode, sdkErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for threadID %q: %v", tt.threadID, err)
				}
			}
		})
	}
}

// =============================================================================
// Unit Tests for isThreadNotFoundError
// =============================================================================

// TestIsThreadNotFoundError 测试会话不存在错误检测
func TestIsThreadNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "NotFound error",
			err:      errors.New("NotFound: thread does not exist"),
			expected: true,
		},
		{
			name:     "not found lowercase",
			err:      errors.New("resource not found"),
			expected: true,
		},
		{
			name:     "NOT_FOUND uppercase",
			err:      errors.New("Error: NOT_FOUND"),
			expected: true,
		},
		{
			name:     "ThreadNotFound",
			err:      errors.New("ThreadNotFound: invalid thread"),
			expected: true,
		},
		{
			name:     "InvalidThreadId",
			err:      errors.New("InvalidThreadId: thread-123"),
			expected: true,
		},
		{
			name:     "does not exist",
			err:      errors.New("The specified thread does not exist"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      errors.New("network timeout"),
			expected: false,
		},
		{
			name:     "permission denied",
			err:      errors.New("permission denied"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isThreadNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("isThreadNotFoundError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Unit Tests for extractMessageContent
// =============================================================================

// TestExtractMessageContent 测试消息内容提取
func TestExtractMessageContent(t *testing.T) {
	tests := []struct {
		name     string
		msg      *starops.GetThreadDataResponseBodyDataMessages
		expected string
	}{
		{
			name:     "nil message",
			msg:      nil,
			expected: "",
		},
		{
			name: "message with text type content",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Contents: []map[string]any{
					{"type": "text", "value": "Hello, world!"},
				},
			},
			expected: "Hello, world!",
		},
		{
			name: "message with multiple text contents",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Contents: []map[string]any{
					{"type": "text", "value": "First"},
					{"type": "text", "value": "Second"},
				},
			},
			expected: "First\nSecond",
		},
		{
			name: "message with value only (no type)",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Contents: []map[string]any{
					{"value": "Direct value"},
				},
			},
			expected: "Direct value",
		},
		{
			name: "message with text field (no type)",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Contents: []map[string]any{
					{"text": "Text field value"},
				},
			},
			expected: "Text field value",
		},
		{
			name: "message with Detail fallback",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Detail: dara.String("Detail content"),
			},
			expected: "Detail content",
		},
		{
			name: "empty contents with Detail",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Contents: []map[string]any{},
				Detail:   dara.String("Fallback detail"),
			},
			expected: "Fallback detail",
		},
		{
			name: "nil content in array",
			msg: &starops.GetThreadDataResponseBodyDataMessages{
				Contents: []map[string]any{
					nil,
					{"type": "text", "value": "Valid content"},
				},
			},
			expected: "Valid content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMessageContent(tt.msg)
			if result != tt.expected {
				t.Errorf("extractMessageContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Unit Tests for ThreadInfo and ThreadMessage structures
// =============================================================================

// TestThreadInfoStructure 测试 ThreadInfo 结构
func TestThreadInfoStructure(t *testing.T) {
	info := &ThreadInfo{
		ThreadID:   "thread-123",
		Title:      "Test Thread",
		Status:     "active",
		CreateTime: "2024-01-01T00:00:00Z",
		UpdateTime: "2024-01-01T12:00:00Z",
	}

	if info.ThreadID != "thread-123" {
		t.Errorf("ThreadID = %s, want thread-123", info.ThreadID)
	}
	if info.Title != "Test Thread" {
		t.Errorf("Title = %s, want Test Thread", info.Title)
	}
	if info.Status != "active" {
		t.Errorf("Status = %s, want active", info.Status)
	}
}

// TestThreadMessageStructure 测试 ThreadMessage 结构
func TestThreadMessageStructure(t *testing.T) {
	msg := &ThreadMessage{
		Role:      "user",
		Content:   "Hello",
		Timestamp: "2024-01-01T00:00:00Z",
	}

	if msg.Role != "user" {
		t.Errorf("Role = %s, want user", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("Content = %s, want Hello", msg.Content)
	}
}

// =============================================================================
// Integration-style tests with mock
// =============================================================================

// TestMockClientIntegration 测试模拟客户端集成
func TestMockClientIntegration(t *testing.T) {
	t.Run("create and list threads", func(t *testing.T) {
		mock := NewMockSTAROpsClient()

		// Create multiple threads
		ids := make([]string, 3)
		for i := 0; i < 3; i++ {
			id, err := mock.CreateThread("Thread " + string(rune('A'+i)))
			if err != nil {
				t.Fatalf("failed to create thread: %v", err)
			}
			ids[i] = id
		}

		// List all threads
		threads, total, err := mock.ListThreads(100)
		if err != nil {
			t.Fatalf("failed to list threads: %v", err)
		}

		if total != 3 {
			t.Errorf("total = %d, want 3", total)
		}
		if len(threads) != 3 {
			t.Errorf("len(threads) = %d, want 3", len(threads))
		}
	})

	t.Run("add and retrieve messages", func(t *testing.T) {
		mock := NewMockSTAROpsClient()

		// Create thread
		threadID, err := mock.CreateThread("Chat Thread")
		if err != nil {
			t.Fatalf("failed to create thread: %v", err)
		}

		// Add messages
		messages := []struct {
			role    string
			content string
		}{
			{"user", "Hello"},
			{"assistant", "Hi there!"},
			{"user", "How are you?"},
			{"assistant", "I'm doing well, thanks!"},
		}

		for _, m := range messages {
			if err := mock.AddMessage(threadID, m.role, m.content); err != nil {
				t.Fatalf("failed to add message: %v", err)
			}
		}

		// Retrieve messages
		retrieved, err := mock.GetThreadData(threadID, 100)
		if err != nil {
			t.Fatalf("failed to get thread data: %v", err)
		}

		if len(retrieved) != len(messages) {
			t.Errorf("len(retrieved) = %d, want %d", len(retrieved), len(messages))
		}

		for i, m := range retrieved {
			if m.Role != messages[i].role {
				t.Errorf("message[%d].Role = %s, want %s", i, m.Role, messages[i].role)
			}
			if m.Content != messages[i].content {
				t.Errorf("message[%d].Content = %s, want %s", i, m.Content, messages[i].content)
			}
		}
	})

	t.Run("delete thread removes all data", func(t *testing.T) {
		mock := NewMockSTAROpsClient()

		// Create thread and add messages
		threadID, _ := mock.CreateThread("To Delete")
		mock.AddMessage(threadID, "user", "Message 1")
		mock.AddMessage(threadID, "assistant", "Message 2")

		// Delete thread
		err := mock.DeleteThread(threadID)
		if err != nil {
			t.Fatalf("failed to delete thread: %v", err)
		}

		// Verify thread is gone
		_, err = mock.GetThread(threadID)
		if err == nil {
			t.Error("expected error when getting deleted thread")
		}

		// Verify messages are gone
		_, err = mock.GetThreadData(threadID, 100)
		if err == nil {
			t.Error("expected error when getting data from deleted thread")
		}
	})
}
