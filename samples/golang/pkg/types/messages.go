// messages.go — 消息结构与内容模型
// 职责：定义 SSE 消息体中的数据结构（MessageItem、ItemContent、ItemTool 等）。
// 不做：不包含任何逻辑，纯数据结构。
// 依赖：无
// ⚠️ 同步要求：修改枚举值时必须同步 Java/Java8/Python/TypeScript 的对应文件。
package types

// =================================================================================
// 消息相关类型
// =================================================================================

type MessageRole string
type ContentType string
type ItemStatus string

// 消息角色 (Role)
const (
	MessageItemRoleUser      MessageRole = "user"
	MessageItemRoleAssistant MessageRole = "assistant"
	MessageItemRoleSystem    MessageRole = "system"
)

// 内容类型 (Content Type)
const (
	MessageItemContentTypeText     ContentType = "text"
	MessageItemContentTypeSpinText ContentType = "spin_text"
	MessageItemContentTypeImage    ContentType = "image"
)

// 执行状态 (Status)
const (
	ItemStatusInit      ItemStatus = "init"
	ItemStatusStart     ItemStatus = "start"
	ItemStatusProgress  ItemStatus = "progress"
	ItemStatusSuspended ItemStatus = "suspended"
	ItemStatusSuccess   ItemStatus = "success"
	ItemStatusFail      ItemStatus = "fail"
)

// =================================================================================
// 消息数据结构
// =================================================================================

// MessageItem 消息条目
// 对应 SSE 协议中 messages 数组的一个元素，代表一条独立的交互记录。
//
// 结构设计支持树形调用链：
//   - Thread (parent_call_id="")
//     |-- Tool Call
//     |-- Sub Agent Call
//     |   |-- Inner Tool Call
//     |   |-- Inner Text Reply
type MessageItem struct {
	ParentCallID string         `json:"parentCallId"`
	CallID       string         `json:"callId"`
	Role         MessageRole    `json:"role"`
	Timestamp    string         `json:"timestamp"`
	Contents     []*ItemContent `json:"contents,omitempty"`
	Tools        []*ItemTool    `json:"tools,omitempty"`
	Agents       []*ItemAgent   `json:"agents,omitempty"`
	Events       []*ItemEvent   `json:"events,omitempty"`
	Artifacts    []map[string]any `json:"artifacts,omitempty"`
}

// ItemArtifact 标准 Google A2A 协议产物
type ItemArtifact map[string]any

// ItemContent 消息内容
type ItemContent struct {
	Type      ContentType `json:"type"`
	Value     string      `json:"value"`
	Append    bool        `json:"append"`
	LastChunk bool        `json:"lastChunk"`
}

// ItemTool 工具调用详情
type ItemTool struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	ToolCallID     string         `json:"toolCallId"`
	ArgumentsDelta string         `json:"argumentsDelta,omitempty"`
	Arguments      any            `json:"arguments,omitempty"`
	Status         ItemStatus     `json:"status"`
	Contents       []*ItemContent `json:"contents,omitempty"`
}

// ItemAgent 子 Agent 调用详情
type ItemAgent struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	CallID  string         `json:"callId"`
	Status  ItemStatus     `json:"status"`
	Inputs  []*ItemContent `json:"inputs,omitempty"`
	Results []*ItemContent `json:"results,omitempty"`
}

// ItemInteraction 用户交互定义
type ItemInteraction struct {
	Type    InteractionType `json:"type"`
	ID      string          `json:"id"`
	Title   string          `json:"title"`
	Payload any             `json:"payload"`
}
