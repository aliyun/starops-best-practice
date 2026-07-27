// Package interactive — 交互处理器骨架与对外入口
// 职责：Handler 结构、生命周期方法、事件分发、恢复对话、公共导出函数。
// 不做：不实现具体的交互动作（→actions.go）、不做终端 I/O（→io.go）、不解析 payload 字段（→payload.go）。
// 依赖：pkg/errors、pkg/logger、pkg/types
package interactive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/errors"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/logger"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// ChatClient 定义恢复对话所需的最小客户端接口
type ChatClient interface {
	Interact(ctx context.Context, threadID string, userInteractive string, baseVariables map[string]any) <-chan *types.ChatEvent
}

// Handler 交互事件处理器
type Handler struct {
	chat      ChatClient
	timeout   time.Duration
	reader    io.Reader
	writer    io.Writer
	mockInput string
}

// Response 交互响应
type Response struct {
	CallId       string                `json:"callId"`
	Type         types.InteractionType `json:"type"`
	Response     map[string]any        `json:"response"`
	Source       map[string]any        `json:"source"`
	ModifiedData map[string]any        `json:"modifiedData"`
	FormData     map[string]any        `json:"formData"`
	Decision     string                `json:"decision"`
}

// NewHandler 创建交互处理器
func NewHandler(chatClient ChatClient, timeout time.Duration) *Handler {
	return &Handler{
		chat:    chatClient,
		timeout: timeout,
		reader:  os.Stdin,
		writer:  os.Stdout,
	}
}

// SetIO 设置输入输出流
func (h *Handler) SetIO(reader io.Reader, writer io.Writer) {
	h.reader = reader
	h.writer = writer
}

// SetMockInput 设置 mock 模式下的预设交互输入
func (h *Handler) SetMockInput(input string) {
	h.mockInput = input
}

// HandleEvent 处理交互事件
func (h *Handler) HandleEvent(ctx context.Context, evt *types.ItemEvent, callId string) (*Response, error) {
	if evt == nil {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, "事件为空")
	}

	if evt.Type != types.EventTypeInteractive {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, fmt.Sprintf("不支持的事件类型: %s", evt.Type))
	}

	payload, err := h.parseInteractivePayload(evt.Payload)
	if err != nil {
		return nil, err
	}

	switch payload.InteractiveType {
	case types.InteractionTypeUserAck:
		return h.HandleUserAck(ctx, payload, callId)
	case types.InteractionTypeUserSelect:
		return h.HandleUserSelect(ctx, payload, callId)
	case types.InteractionTypeUserInput:
		return h.HandleUserInput(ctx, payload, callId)
	default:
		return nil, errors.NewSDKError(errors.ErrCodeParseError, fmt.Sprintf("不支持的交互类型: %s", payload.InteractiveType))
	}
}

// ResumeChat 使用交互响应恢复对话
func (h *Handler) ResumeChat(ctx context.Context, threadID string, response *Response, baseVariables map[string]any) <-chan *types.ChatEvent {
	if h.chat == nil {
		events := make(chan *types.ChatEvent, 1)
		events <- &types.ChatEvent{Error: errors.NewSDKError(errors.ErrCodeClientCreate, "客户端未初始化")}
		close(events)
		return events
	}

	if response == nil {
		events := make(chan *types.ChatEvent, 1)
		events <- &types.ChatEvent{Error: errors.NewSDKError(errors.ErrCodeParseError, "交互响应为空")}
		close(events)
		return events
	}

	userInteractive := map[string]any{
		"callId":   response.CallId,
		"source":   response.Source,
		"decision": response.Decision,
	}
	if response.Type == types.InteractionTypeUserInput {
		userInteractive["formData"] = response.FormData
	} else {
		userInteractive["modifiedData"] = response.ModifiedData
	}
	uiJSON, err := json.Marshal(userInteractive)
	if err != nil {
		events := make(chan *types.ChatEvent, 1)
		events <- &types.ChatEvent{Error: errors.NewSDKErrorWithCause(errors.ErrCodeParseError, "序列化 userInteractive 失败", err)}
		close(events)
		return events
	}

	return h.chat.Interact(ctx, threadID, string(uiJSON), baseVariables)
}

// =================================================================================
// 便捷导出方法
// =================================================================================

// IsInteractiveEvent 检查事件是否为交互事件
func IsInteractiveEvent(evt *types.ItemEvent) bool {
	return evt != nil && evt.Type == types.EventTypeInteractive
}

// ExtractInteractiveEvents 从消息项中提取交互事件
func ExtractInteractiveEvents(item *types.MessageItem) []*types.ItemEvent {
	if item == nil || len(item.Events) == 0 {
		return nil
	}

	var interactiveEvents []*types.ItemEvent
	for _, evt := range item.Events {
		if IsInteractiveEvent(evt) {
			interactiveEvents = append(interactiveEvents, evt)
		}
	}
	return interactiveEvents
}

// ExtractResponse 从 ChatEvent 中检测交互事件并调用 handler 收集用户响应
func ExtractResponse(ctx context.Context, evt *types.ChatEvent, handler *Handler) *Response {
	if evt == nil || handler == nil {
		return nil
	}
	if evt.Body == nil || len(evt.Body.Messages) == 0 {
		return nil
	}

	for _, msg := range evt.Body.Messages {
		if msg == nil {
			continue
		}
		for _, itemEvt := range msg.Events {
			if itemEvt == nil || itemEvt.Type != types.EventTypeInteractive {
				continue
			}

			resp, err := handler.HandleEvent(ctx, itemEvt, msg.CallID)
			if err != nil {
				logger.Default().Error("交互处理失败", err, nil)
				return nil
			}
			return resp
		}
	}
	return nil
}
