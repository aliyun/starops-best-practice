// Package client 提供 VibeOps Agent 客户端的公共实现
// thread.go 实现会话管理 API
package client

import (
	"context"
	"fmt"
	"strings"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
)

// ThreadInfo 会话信息
// ThreadInfo represents thread metadata
type ThreadInfo struct {
	// ThreadID 会话 ID
	// Unique identifier for the thread
	ThreadID string `json:"threadId"`

	// Title 会话标题
	// Human-readable title for the thread
	Title string `json:"title"`

	// Status 会话状态
	// Current status of the thread (e.g., "active", "archived")
	Status string `json:"status"`

	// CreateTime 创建时间
	// Timestamp when the thread was created
	CreateTime string `json:"createTime"`

	// UpdateTime 更新时间
	// Timestamp when the thread was last updated
	UpdateTime string `json:"updateTime"`
}

// ThreadMessage 会话消息
// ThreadMessage represents a single message in a thread
type ThreadMessage struct {
	// Role 消息角色
	// The role of the message sender (e.g., "user", "assistant", "system")
	Role string `json:"role"`

	// Content 消息内容
	// The text content of the message
	Content string `json:"content"`

	// Timestamp 消息时间戳
	// Timestamp when the message was sent
	Timestamp string `json:"timestamp"`
}

// ListThreads 列出会话
// ListThreads returns a paginated list of threads
// Parameters:
//   - ctx: context for cancellation and timeout
//   - pageSize: maximum number of threads to return (1-100)
//
// Returns:
//   - threads: list of thread information
//   - total: total number of threads available
//   - error: any error that occurred
func (c *AgentClient) ListThreads(ctx context.Context, pageSize int) ([]*ThreadInfo, int64, error) {
	// 验证参数
	// Validate parameters
	if pageSize <= 0 {
		pageSize = 20 // 默认值 / default value
	}
	if pageSize > 100 {
		pageSize = 100 // 最大值 / maximum value
	}

	// 构建请求
	// Build request
	req := &starops.ListThreadsRequest{}
	req.SetMaxResults(int64(pageSize))

	// 调用 API
	// Call API
	resp, err := c.client.ListThreads(dara.String(c.config.EmployeeName), req)
	if err != nil {
		return nil, 0, NewSDKErrorWithCause(ErrCodeAPIError, "获取会话列表失败", err).
			WithSuggestion("请检查网络连接和 API 权限")
	}

	// 检查响应
	// Check response
	if resp.Body == nil {
		return nil, 0, NewSDKError(ErrCodeParseError, "无效响应: 响应体为空").
			WithSuggestion("请稍后重试")
	}

	// 转换结果
	// Convert results
	var threads []*ThreadInfo
	for _, t := range resp.Body.Threads {
		if t == nil {
			continue
		}
		thread := &ThreadInfo{
			ThreadID:   dara.StringValue(t.ThreadId),
			Title:      dara.StringValue(t.Title),
			Status:     dara.StringValue(t.Status),
			CreateTime: dara.StringValue(t.CreateTime),
			UpdateTime: dara.StringValue(t.UpdateTime),
		}
		threads = append(threads, thread)
	}

	total := dara.Int64Value(resp.Body.Total)

	return threads, total, nil
}

// GetThread 获取会话详情
// GetThread retrieves detailed information about a specific thread
// Parameters:
//   - ctx: context for cancellation and timeout
//   - threadID: the unique identifier of the thread
//
// Returns:
//   - thread: detailed thread information
//   - error: any error that occurred (including ErrCodeThreadNotFound if thread doesn't exist)
func (c *AgentClient) GetThread(ctx context.Context, threadID string) (*ThreadInfo, error) {
	// 验证参数
	// Validate parameters
	if err := validateThreadID(threadID); err != nil {
		return nil, err
	}

	// 调用 API
	// Call API
	getThreadReq := &starops.GetThreadRequest{}
	resp, err := c.client.GetThread(dara.String(c.config.EmployeeName), dara.String(threadID), getThreadReq)
	if err != nil {
		// 检查是否为会话不存在错误
		// Check if it's a thread not found error
		if isThreadNotFoundError(err) {
			return nil, ErrThreadNotFound(threadID)
		}
		return nil, NewSDKErrorWithCause(ErrCodeAPIError, fmt.Sprintf("获取会话详情失败: %s", threadID), err).
			WithContext("threadId", threadID).
			WithSuggestion("请检查会话 ID 是否正确")
	}

	// 检查响应
	// Check response
	if resp.Body == nil {
		return nil, NewSDKError(ErrCodeParseError, "无效响应: 响应体为空").
			WithContext("threadId", threadID).
			WithSuggestion("请稍后重试")
	}

	// 转换结果
	// Convert result
	thread := &ThreadInfo{
		ThreadID:   dara.StringValue(resp.Body.ThreadId),
		Title:      dara.StringValue(resp.Body.Title),
		Status:     dara.StringValue(resp.Body.Status),
		CreateTime: dara.StringValue(resp.Body.CreateTime),
		UpdateTime: dara.StringValue(resp.Body.UpdateTime),
	}

	return thread, nil
}

// DeleteThread 删除会话
// DeleteThread removes a thread and all its associated data
// Parameters:
//   - ctx: context for cancellation and timeout
//   - threadID: the unique identifier of the thread to delete
//
// Returns:
//   - error: any error that occurred (including ErrCodeThreadNotFound if thread doesn't exist)
func (c *AgentClient) DeleteThread(ctx context.Context, threadID string) error {
	// 验证参数
	// Validate parameters
	if err := validateThreadID(threadID); err != nil {
		return err
	}

	// 调用 API
	// Call API
	deleteThreadReq := &starops.DeleteThreadRequest{}
	_, err := c.client.DeleteThread(dara.String(c.config.EmployeeName), dara.String(threadID), deleteThreadReq)
	if err != nil {
		// 检查是否为会话不存在错误
		// Check if it's a thread not found error
		if isThreadNotFoundError(err) {
			return ErrThreadNotFound(threadID)
		}
		return NewSDKErrorWithCause(ErrCodeAPIError, fmt.Sprintf("删除会话失败: %s", threadID), err).
			WithContext("threadId", threadID).
			WithSuggestion("请检查会话 ID 是否正确")
	}

	return nil
}

// GetThreadData 获取会话消息
// GetThreadData retrieves the messages in a thread
// Parameters:
//   - ctx: context for cancellation and timeout
//   - threadID: the unique identifier of the thread
//   - limit: maximum number of messages to return (1-100)
//
// Returns:
//   - messages: list of messages in the thread
//   - error: any error that occurred (including ErrCodeThreadNotFound if thread doesn't exist)
func (c *AgentClient) GetThreadData(ctx context.Context, threadID string, limit int) ([]*ThreadMessage, error) {
	// 验证参数
	// Validate parameters
	if err := validateThreadID(threadID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50 // 默认值 / default value
	}
	if limit > 100 {
		limit = 100 // 最大值 / maximum value
	}

	// 构建请求
	// Build request
	req := &starops.GetThreadDataRequest{}
	req.SetMaxResults(int64(limit))

	// 调用 API
	// Call API
	resp, err := c.client.GetThreadData(dara.String(c.config.EmployeeName), dara.String(threadID), req)
	if err != nil {
		// 检查是否为会话不存在错误
		// Check if it's a thread not found error
		if isThreadNotFoundError(err) {
			return nil, ErrThreadNotFound(threadID)
		}
		return nil, NewSDKErrorWithCause(ErrCodeAPIError, fmt.Sprintf("获取会话消息失败: %s", threadID), err).
			WithContext("threadId", threadID).
			WithSuggestion("请检查会话 ID 是否正确")
	}

	// 检查响应
	// Check response
	if resp.Body == nil {
		return nil, NewSDKError(ErrCodeParseError, "无效响应: 响应体为空").
			WithContext("threadId", threadID).
			WithSuggestion("请稍后重试")
	}

	// 策略：优先使用 system Result，而不是 assistant 流式消息
	// Strategy: prefer system Result over assistant streaming messages
	messageMap := make(map[string]*ThreadMessage)
	messageOrder := []string{} // 保持顺序 / maintain order

	// 检查是否存在 system Result
	// Check if any system Result exists
	hasSystemResult := false
	for _, data := range resp.Body.Data {
		if data == nil {
			continue
		}
		for _, msg := range data.Messages {
			if msg == nil {
				continue
			}
			if dara.StringValue(msg.Role) == "system" {
				for _, artifact := range msg.Artifacts {
					if artifact == nil {
						continue
					}
					if nameVal, ok := artifact["name"]; ok {
						if nameStr, ok := nameVal.(string); ok && nameStr == "Result" {
							hasSystemResult = true
							break
						}
					}
				}
			}
			if hasSystemResult {
				break
			}
		}
		if hasSystemResult {
			break
		}
	}

	// 处理消息
	// Process messages
	for _, data := range resp.Body.Data {
		if data == nil {
			continue
		}
		for _, msg := range data.Messages {
			if msg == nil {
				continue
			}

			role := dara.StringValue(msg.Role)
			timestamp := dara.StringValue(msg.Timestamp)

			// 如果存在 system Result，跳过 assistant 流式消息
			// Skip assistant streaming messages if system Result exists
			if role == "assistant" && hasSystemResult {
				continue
			}

			// 根据角色使用不同的键策略
			// Use different key strategy based on role
			var key string
			if role == "user" {
				key = "user-" + timestamp
			} else if role == "system" {
				key = "system-" + timestamp
			} else {
				callID := dara.StringValue(msg.CallId)
				key = "assistant-" + callID
			}

			// 提取消息内容
			// Extract message content
			content := extractMessageContent(msg)
			if content == "" {
				continue
			}

			if existing, ok := messageMap[key]; ok {
				// 追加内容到现有消息
				// Append content to existing message
				existing.Content += content
			} else {
				// 对于 system 消息，显示为 assistant 角色
				// For system messages, display as assistant role
				displayRole := role
				if role == "system" {
					displayRole = "assistant"
				}
				messageMap[key] = &ThreadMessage{
					Role:      displayRole,
					Content:   content,
					Timestamp: timestamp,
				}
				messageOrder = append(messageOrder, key)
			}
		}
	}

	// 按顺序返回消息
	// Return messages in order
	var messages []*ThreadMessage
	for _, key := range messageOrder {
		messages = append(messages, messageMap[key])
	}

	return messages, nil
}

// validateThreadID 验证会话 ID
// validateThreadID validates the thread ID format
func validateThreadID(threadID string) error {
	if threadID == "" {
		return NewSDKError(ErrCodeConfigInvalid, "会话 ID 不能为空").
			WithContext("threadId", threadID).
			WithSuggestion("请提供有效的会话 ID")
	}

	// 检查是否包含非法字符
	// Check for invalid characters
	if strings.ContainsAny(threadID, " \t\n\r") {
		return NewSDKError(ErrCodeConfigInvalid, fmt.Sprintf("会话 ID 包含非法字符: %s", threadID)).
			WithContext("threadId", threadID).
			WithSuggestion("会话 ID 不能包含空白字符")
	}

	return nil
}

// isThreadNotFoundError 检查是否为会话不存在错误
// isThreadNotFoundError checks if the error indicates a thread was not found
func isThreadNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// 检查常见的 "not found" 错误模式
	// Check common "not found" error patterns
	notFoundPatterns := []string{
		"NotFound",
		"not found",
		"NOT_FOUND",
		"ThreadNotFound",
		"InvalidThreadId",
		"does not exist",
	}

	for _, pattern := range notFoundPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// extractMessageContent 从消息中提取文本内容
// extractMessageContent extracts text content from a message
func extractMessageContent(msg *starops.GetThreadDataResponseBodyDataMessages) string {
	if msg == nil {
		return ""
	}

	// 1. 尝试从 Contents 字段提取文本（流式文本块）
	// Try to extract text from Contents field (streaming text chunks)
	var contentParts []string
	for _, content := range msg.Contents {
		if content == nil {
			continue
		}

		// 检查 type 是否为 text
		// Check if type is text
		if typeVal, ok := content["type"]; ok {
			if typeStr, ok := typeVal.(string); ok && typeStr == "text" {
				if valueVal, ok := content["value"]; ok {
					if valueStr, ok := valueVal.(string); ok && valueStr != "" {
						contentParts = append(contentParts, valueStr)
					}
				}
			}
		} else {
			// 兜底：无 type 字段时，尝试 value 或 text 字段
			if valueVal, ok := content["value"]; ok {
				if valueStr, ok := valueVal.(string); ok && valueStr != "" {
					contentParts = append(contentParts, valueStr)
				}
			} else if textVal, ok := content["text"]; ok {
				if textStr, ok := textVal.(string); ok && textStr != "" {
					contentParts = append(contentParts, textStr)
				}
			}
		}
	}

	if len(contentParts) > 0 {
		return strings.Join(contentParts, "\n")
	}

	// 2. 尝试从 Artifacts 字段提取文本（最终结果）
	// Try to extract text from Artifacts field (final result)
	for _, artifact := range msg.Artifacts {
		if artifact == nil {
			continue
		}

		// 查找名为 Result 的 artifact
		// Look for artifact named Result
		if nameVal, ok := artifact["name"]; ok {
			if nameStr, ok := nameVal.(string); ok && nameStr == "Result" {
				if partsVal, ok := artifact["parts"]; ok {
					if parts, ok := partsVal.([]any); ok {
						var textParts []string
						for _, partVal := range parts {
							if part, ok := partVal.(map[string]any); ok {
								if kindVal, ok := part["kind"]; ok {
									if kindStr, ok := kindVal.(string); ok && kindStr == "text" {
										if textVal, ok := part["text"]; ok {
											if textStr, ok := textVal.(string); ok && textStr != "" {
												textParts = append(textParts, textStr)
											}
										}
									}
								}
							}
						}
						if len(textParts) > 0 {
							return strings.Join(textParts, "")
						}
					}
				}
			}
		}
	}

	// 如果 Contents 和 Artifacts 都为空，尝试使用 Detail 字段
	// If Contents and Artifacts are empty, try using Detail field
	if msg.Detail != nil && *msg.Detail != "" {
		return *msg.Detail
	}

	return ""
}
