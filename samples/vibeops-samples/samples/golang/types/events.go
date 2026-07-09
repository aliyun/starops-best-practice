package types

// =================================================================================
// 类型定义
// =================================================================================

type MessageRole string
type ContentType string
type ItemStatus string
type InteractionType string

// =================================================================================
// 常量定义
// =================================================================================

// 消息角色 (Role)
const (
	MessageItemRoleUser      MessageRole = "user"      // 用户输入
	MessageItemRoleAssistant MessageRole = "assistant" // Agent 回复或操作
	MessageItemRoleSystem    MessageRole = "system"    // 系统消息
)

// 内容类型 (Content Type)
const (
	MessageItemContentTypeText     ContentType = "text"      // 纯文本
	MessageItemContentTypeSpinText ContentType = "spin_text" // 旋转文本，用于显示工作和思考过程
	MessageItemContentTypeImage    ContentType = "image"     // 图片
)

// 事件类型 (Event Type)
type EventType string

const (
	EventTypeThreadTitleUpdated  EventType = "thread_title_updated" // 会话标题更新
	EventTypeError               EventType = "error"                // 错误事件
	EventTypeThinking            EventType = "thinking"             // 思考事件
	EventTypeInteractive         EventType = "interactive"          // 交互事件
	EventTypeInteractiveResponse EventType = "interactive_response" // 交互响应事件
	EventTypeTaskFinished        EventType = "task_finished"        // 任务完成事件。注意：只有整体任务完成，才会发送，由主调度引擎产生
	EventTypeCancel              EventType = "cancel"
	EventTypeStreamDone          EventType = "stream_done"          // SSE 流正常结束标志
)

// 执行状态 (Status) - 表示执行过程的阶段
const (
	ItemStatusInit      ItemStatus = "init"      // 初始化
	ItemStatusStart     ItemStatus = "start"     // 开始执行
	ItemStatusProgress  ItemStatus = "progress"  // 执行中
	ItemStatusSuspended ItemStatus = "suspended" // 暂停，一般是等待用户追问、用户Ack、额度不够等
	ItemStatusSuccess   ItemStatus = "success"   // 执行完成(成功)
	ItemStatusFail      ItemStatus = "fail"      // 执行完成(失败)
)

// 暂停原因 (Suspend Reason) 暂未使用
type SuspendReason string

const (
	SuspendReasonNone                 SuspendReason = ""
	SuspendReasonAwaitingInput        SuspendReason = "awaiting_input"
	SuspendReasonAwaitingConfirmation SuspendReason = "awaiting_confirmation"
	SuspendReasonAwaitingAnswer       SuspendReason = "awaiting_answer"
)

// 交互类型 (Interaction Type)
const (
	InteractionTypeUserAck    InteractionType = "user_ack"    // 点击确认
	InteractionTypeUserSelect InteractionType = "user_select" // 选择框
	InteractionTypeUserInput  InteractionType = "user_input"  // 用户输入
	InteractionTypeSlsQuery   InteractionType = "sls_query"   // SLS 查询
)

// =================================================================================
// 数据结构定义
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
	// ParentCallID 父调用ID，用于构建调用链。根节点(Thread)为空字符串。
	ParentCallID string `json:"parentCallId"`

	// CallID 当前调用的唯一标识符。
	// 后续的子消息（如 Tool 结果、Sub Agent 内部消息）将使用此 CallID 作为 ParentCallID。
	CallID string `json:"callId"`

	// Role 消息角色，参考 MessageItemRole* 常量。
	Role MessageRole `json:"role"`

	// Timestamp 消息生成的时间戳 (Unix Timestamp)，单位为纳秒。
	Timestamp string `json:"timestamp"`

	// Contents 文本或富媒体内容列表。
	// 当 Role="assistant" 且无 Tools/Agents 时，通常为文本回复。
	Contents []*ItemContent `json:"contents,omitempty"`

	// Tools 工具调用列表。
	// 仅当 Role="assistant" 时出现，表示 Agent 决定调用工具。
	Tools []*ItemTool `json:"tools,omitempty"`

	// Agents 子 Agent 调用列表。
	// 仅当 Role="assistant" 时出现，表示 Agent 决定调用子 Agent。
	Agents []*ItemAgent `json:"agents,omitempty"`

	// Events 事件列表。
	Events []*ItemEvent `json:"events,omitempty"`

	// Artifacts 产物列表。
	// 仅当 Role="assistant" 时出现，表示 Agent 生成的产物。
	// 当 Role="user" 时，Artifacts 表示用户输入的产物。
	// 标准的 google a2a 协议产物
	Artifacts []map[string]any `json:"artifacts,omitempty"`
}

// 标准的 google a2a 协议产物
type ItemArtifact map[string]any

// ItemContent 消息内容
type ItemContent struct {
	// Type 内容类型，参考 MessageItemContentType* 常量。
	Type ContentType `json:"type"`

	// Value 内容值，可能是纯文本、JSON 字符串或 URL。
	Value string `json:"value"`

	// Append 是否是追加内容，Stream模式下默认都为true
	Append bool `json:"append"`

	// LastChunk 是否是最后一个 chunk，为 true 时代表最后一个
	LastChunk bool `json:"lastChunk"`
}

// ItemTool 工具调用详情
type ItemTool struct {
	// ID 工具调用的唯一标识
	ID string `json:"id"`

	// Name 工具名称。
	Name string `json:"name"`

	// ToolCallID 本次调用的上下文 ID。
	ToolCallID string `json:"toolCallId"`

	// ArgumentsDelta 工具调用参数增量
	ArgumentsDelta string `json:"argumentsDelta,omitempty"`

	// Arguments 工具调用参数。
	Arguments any `json:"arguments,omitempty"`

	// Status 执行阶段状态 (start/success/fail)。
	Status ItemStatus `json:"status"`

	// Contents 工具执行结果输出。
	Contents []*ItemContent `json:"contents,omitempty"`
}

// ItemAgent 子 Agent 调用详情
type ItemAgent struct {
	// ID Agent 唯一标识。
	ID string `json:"id"`

	// Name Agent 显示名称。
	Name string `json:"name"`

	// CallID 本次调用的上下文 ID。
	CallID string `json:"callId"`

	// Status 执行阶段状态 (start/success/fail)。
	Status ItemStatus `json:"status"`

	// Inputs Agent 调用的输入内容，可能有多个，例如 文本 + 图片等
	Inputs []*ItemContent `json:"inputs,omitempty"`

	// Results Agent 的输出内容概要（详细内容通过独立消息返回）
	Results []*ItemContent `json:"results,omitempty"`
}

// ItemInteraction 用户交互定义
type ItemInteraction struct {
	// Type 交互类型 (confirm/input)。
	Type InteractionType `json:"type"`
	// ID 交互的唯一标识。
	ID string `json:"id"`
	// Title 交互的标题。
	Title string `json:"title"`
	// Payload 交互的负载。
	Payload any `json:"payload"`
}

// ItemEvent 事件定义
type ItemEvent struct {
	// Type 事件类型，参考 MessageItemEventType* 常量。
	Type EventType `json:"type"`

	// Payload 事件内容。
	Payload any `json:"payload"`
}

// 会话标题更新事件负载
type ItemThreadTitleUpdatedPayload struct {
	// Title 会话标题。
	Title string `json:"title"`
}

// 错误事件负载
type ItemErrorPayload struct {
	// Code 错误码。
	Code string `json:"code"`

	// Message 错误消息。
	Message string `json:"message"`

	// Suggestion 错误建议。
	Suggestion string `json:"suggestion"`
}

type ItemSessionStatusUpdatedPayload struct {
	// Status 会话状态。
	Status ItemStatus `json:"status"`

	SuspendReason SuspendReason `json:"suspendReason"`
}

type ItemThinkingPayload struct {
	ReasoningDelta string `json:"reasoningDelta"`
}

type ItemInteractiveUserAckPayload map[string]any
type ItemInteractiveUserInputPayload map[string]any
type ItemInteractivePayload struct {
	InteractiveType InteractionType                  `json:"type"`
	Meta            map[string]any           `json:"meta,omitempty"`
	Data            []map[string]any         `json:"data,omitempty"`
	Queries         []map[string]any         `json:"queries,omitempty"`
	UserAck         *ItemInteractiveUserAckPayload   `json:"userAck,omitempty"`
	UserInput       *ItemInteractiveUserInputPayload `json:"userInput,omitempty"`
}

type TaskStatistics struct {
	Duration int64 `json:"duration"` // 任务整体延迟，单位纳秒
}

type ItemTaskFinishedPayload struct {
	Success    bool              `json:"success"`              // 任务是否成功
	Error      *ItemErrorPayload `json:"error,omitempty"`      // 错误信息，只有失败时才有
	Statistics *TaskStatistics   `json:"statistics,omitempty"` // 任务统计信息
}
