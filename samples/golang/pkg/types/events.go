// events.go — SSE 事件类型与载荷定义
// 职责：定义事件类型枚举、事件载荷结构、ChatEvent 领域类型。
// 不做：不包含任何逻辑，纯数据结构。
// 依赖：无
// ⚠️ 同步要求：修改枚举值时必须同步 Java/Java8/Python/TypeScript 的对应文件。
package types

// =================================================================================
// 事件相关类型
// =================================================================================

type EventType string
type InteractionType string
type SuspendReason string

// 事件类型 (Event Type)
const (
	EventTypeThreadTitleUpdated  EventType = "thread_title_updated"
	EventTypeError               EventType = "error"
	EventTypeThinking            EventType = "thinking"
	EventTypeInteractive         EventType = "interactive"
	EventTypeInteractiveResponse EventType = "interactive_response"
	EventTypeTaskFinished        EventType = "task_finished"
	EventTypeCancel              EventType = "cancel"
	EventTypeStreamDone          EventType = "stream_done"
)

// 交互类型 (Interaction Type)
const (
	InteractionTypeUserAck    InteractionType = "user_ack"
	InteractionTypeUserSelect InteractionType = "user_select"
	InteractionTypeUserInput  InteractionType = "user_input"
	InteractionTypeSlsQuery   InteractionType = "sls_query"
)

// 暂停原因 (Suspend Reason)
const (
	SuspendReasonNone                 SuspendReason = ""
	SuspendReasonAwaitingInput        SuspendReason = "awaiting_input"
	SuspendReasonAwaitingConfirmation SuspendReason = "awaiting_confirmation"
	SuspendReasonAwaitingAnswer       SuspendReason = "awaiting_answer"
)

// =================================================================================
// 事件数据结构
// =================================================================================

// ItemEvent 事件定义
type ItemEvent struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}

// ItemThreadTitleUpdatedPayload 会话标题更新事件负载
type ItemThreadTitleUpdatedPayload struct {
	Title string `json:"title"`
}

// ItemErrorPayload 错误事件负载
type ItemErrorPayload struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// ItemSessionStatusUpdatedPayload 会话状态更新事件负载
type ItemSessionStatusUpdatedPayload struct {
	Status        ItemStatus    `json:"status"`
	SuspendReason SuspendReason `json:"suspendReason"`
}

// ItemThinkingPayload 思考事件负载
type ItemThinkingPayload struct {
	ReasoningDelta string `json:"reasoningDelta"`
}

// ItemInteractiveUserAckPayload 用户确认交互负载
type ItemInteractiveUserAckPayload map[string]any

// ItemInteractiveUserInputPayload 用户输入交互负载
type ItemInteractiveUserInputPayload map[string]any

// ItemInteractivePayload 交互事件负载
type ItemInteractivePayload struct {
	InteractiveType InteractionType                `json:"type"`
	Meta            map[string]any                 `json:"meta,omitempty"`
	Data            []map[string]any               `json:"data,omitempty"`
	Queries         []map[string]any               `json:"queries,omitempty"`
	UserAck         *ItemInteractiveUserAckPayload `json:"userAck,omitempty"`
	UserInput       *ItemInteractiveUserInputPayload `json:"userInput,omitempty"`
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	Duration int64 `json:"duration"`
}

// ItemTaskFinishedPayload 任务完成事件负载
type ItemTaskFinishedPayload struct {
	Success    bool              `json:"success"`
	Error      *ItemErrorPayload `json:"error,omitempty"`
	Statistics *TaskStatistics   `json:"statistics,omitempty"`
}

// =================================================================================
// ChatEvent — SSE 事件（领域类型，不依赖外部 SDK）
// =================================================================================

// ChatEventBody SDK 响应体的领域映射
type ChatEventBody struct {
	Messages []*MessageItem `json:"messages,omitempty"`
}

// ChatEvent 单个 SSE 事件
type ChatEvent struct {
	Body       *ChatEventBody
	RawJSON    string
	StatusCode int32
	IsDone     bool
	Error      error
	Id         string
	Event      string
}
