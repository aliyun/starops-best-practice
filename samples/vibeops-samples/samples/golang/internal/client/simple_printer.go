// Package client 提供 VibeOps Agent 客户端的公共实现
package client

import (
	"strings"

	"github.com/vibeops/samples/golang/types"
)

// SimplePrinter 简洁模式打印器
// 只输出最终的 assistant 文本内容，过滤掉工具调用、思考事件等中间消息
// Validates: Requirements 3.1, 3.2, 3.4
type SimplePrinter struct {
	buffer        strings.Builder
	seenArtifacts map[string]bool // 用于去重已处理的 artifacts
}

// NewSimplePrinter 创建简洁打印器
func NewSimplePrinter() *SimplePrinter {
	return &SimplePrinter{
		seenArtifacts: make(map[string]bool),
	}
}

// ProcessEvent 处理事件，提取文本内容
// 只提取最终的文本内容，忽略工具调用、思考事件等
// 返回本次事件提取的文本（如果有）
// Validates: Requirements 3.1, 3.2, 3.4
func (p *SimplePrinter) ProcessEvent(event *ChatEvent) string {
	if event == nil || event.Body == nil {
		return ""
	}
	// 利用 Event 字段快速跳过非文本事件
	if event.Event != "" && event.Event != "text" && event.Event != "task_finished" {
		return ""
	}

	var extracted strings.Builder

	for _, msg := range event.Body.Messages {
		if msg == nil {
			continue
		}

		// 只处理 system 角色的消息（包含最终结果的 artifacts）
		// task_finished 事件的最终结果在 system 角色的 artifacts 中
		// 跳过 assistant 角色的 artifacts 以避免重复
		if msg.Role != nil && *msg.Role == string(types.MessageItemRoleSystem) {
			// 从 artifacts 中提取文本（带去重）
			text := extractTextFromArtifacts(msg.Artifacts)
			if text != "" && !p.seenArtifacts[text] {
				p.seenArtifacts[text] = true
				extracted.WriteString(text)
				p.buffer.WriteString(text)
			}
		}
	}

	return extracted.String()
}

// GetFinalText 获取最终文本
// 返回所有已处理事件中提取的文本内容
func (p *SimplePrinter) GetFinalText() string {
	return p.buffer.String()
}

// Reset 重置缓冲区
func (p *SimplePrinter) Reset() {
	p.buffer.Reset()
	p.seenArtifacts = make(map[string]bool)
}

// hasNonTextEvents 检查是否包含非文本事件（如 thinking、error 等）
func hasNonTextEvents(events []map[string]any) bool {
	for _, event := range events {
		if event == nil {
			continue
		}
		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}
		// 过滤掉 thinking 和 error 事件
		if types.EventType(eventType) == types.EventTypeThinking ||
			types.EventType(eventType) == types.EventTypeError {
			return true
		}
	}
	return false
}

// extractTextFromContents 从内容列表中提取文本
// 按顺序连接所有文本类型的内容
// Validates: Requirement 3.4 - concatenate content items in order
func extractTextFromContents(contents []map[string]any) string {
	var result strings.Builder

	for _, content := range contents {
		if content == nil {
			continue
		}

		// 只提取文本类型的内容
		contentType, ok := content["type"].(string)
		if !ok || contentType != string(types.MessageItemContentTypeText) {
			continue
		}

		value, ok := content["value"].(string)
		if ok && value != "" {
			result.WriteString(value)
		}
	}

	return result.String()
}

// extractTextFromArtifacts 从 artifacts 中提取文本
// artifacts 是 Google A2A 协议的产物格式
// 格式: [{"parts": [{"kind": "text", "text": "..."}]}]
func extractTextFromArtifacts(artifacts []map[string]any) string {
	var result strings.Builder

	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}

		// 获取 parts 数组
		parts, ok := artifact["parts"].([]any)
		if !ok {
			continue
		}

		for _, part := range parts {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}

			// 只提取 kind="text" 的部分
			kind, ok := partMap["kind"].(string)
			if !ok || kind != "text" {
				continue
			}

			text, ok := partMap["text"].(string)
			if ok && text != "" {
				result.WriteString(text)
			}
		}
	}

	return result.String()
}
