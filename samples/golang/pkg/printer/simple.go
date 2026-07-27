// Package printer — 简洁模式输出
// simple.go 职责：从事件流中只提取最终 assistant 文本，过滤工具调用/思考等中间事件。
// 不做：不输出原始事件、不处理交互、不做重连。
// 依赖：pkg/types
package printer

import (
	"strings"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// SimplePrinter 简洁模式打印器
type SimplePrinter struct {
	buffer        strings.Builder
	seenArtifacts map[string]bool
}

// NewSimplePrinter 创建简洁打印器
func NewSimplePrinter() *SimplePrinter {
	return &SimplePrinter{
		seenArtifacts: make(map[string]bool),
	}
}

// ProcessEvent 处理事件，提取文本内容
func (p *SimplePrinter) ProcessEvent(evt *types.ChatEvent) string {
	if evt == nil || evt.Body == nil {
		return ""
	}

	var extracted strings.Builder

	for _, msg := range evt.Body.Messages {
		if msg == nil {
			continue
		}

		if msg.Role == types.MessageItemRoleSystem {
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
func (p *SimplePrinter) GetFinalText() string {
	return p.buffer.String()
}

// Reset 重置缓冲区
func (p *SimplePrinter) Reset() {
	p.buffer.Reset()
	p.seenArtifacts = make(map[string]bool)
}

// extractTextFromArtifacts 从 artifacts 中提取文本（Google A2A 协议格式）
func extractTextFromArtifacts(artifacts []map[string]any) string {
	var result strings.Builder

	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}

		parts, ok := artifact["parts"].([]any)
		if !ok {
			continue
		}

		for _, part := range parts {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}

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
