package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
)

// ===================== 一、配置 =====================

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int           // 最大重试次数，默认10
	InitialBackoff time.Duration // 初始退避时间，默认1s
	MaxBackoff     time.Duration // 最大退避时间，默认30s
	BackoffFactor  float64       // 退避系数，默认2.0
	IdleTimeout    time.Duration // 空闲超时：超过此时长未收到任何消息视为连接中断，默认 60s
}

// DefaultRetryConfig 返回默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     10,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
		IdleTimeout:    60 * time.Second,
	}
}

// LoadRetryConfigFromEnv 从环境变量加载重试配置
func LoadRetryConfigFromEnv() *RetryConfig {
	cfg := DefaultRetryConfig()

	if maxRetries := os.Getenv("VIBEOPS_MAX_RETRIES"); maxRetries != "" {
		if n, err := strconv.Atoi(maxRetries); err == nil && n > 0 {
			cfg.MaxRetries = n
		}
	}

	if idleTimeout := os.Getenv("VIBEOPS_IDLE_TIMEOUT"); idleTimeout != "" {
		if n, err := strconv.Atoi(idleTimeout); err == nil && n > 0 {
			cfg.IdleTimeout = time.Duration(n) * time.Second
		}
	}

	return cfg
}

// ===================== 二、状态定义 =====================

// retryState 聚合重连过程中的状态
type retryState struct {
	lastTimestamp  string // 最后一条已转发消息的时间戳，用于重连去重
	inDedupeWindow bool   // true=重连后去重窗口，仅转发更新的消息
	retryCount     int    // 当前连续重试次数
}

// connectionOutcome 单次连接的结束原因
type connectionOutcome int

const (
	outcomeDone        connectionOutcome = iota // 收到 stream_done，正常结束
	outcomeInterrupted                          // 连接中断，需重连
	outcomeFatal                                // ctx 取消
)

// ===================== 三、核心编排 =====================

// streamSSE 启动带重试能力的 SSE 流处理（默认启用重试）
// 编排层：外层重连循环，连接中断时自动重连并通过 timestamp 去重
func (c *AgentClient) streamSSE(ctx context.Context, req *starops.CreateChatRequest, events chan *ChatEvent) {
	cfg := c.config.RetryConfig
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}
	// 初始化 重连状态
	state := &retryState{}
	for {
		outcome := c.streamOnce(ctx, req, events, state, cfg)
		switch outcome {
		case outcomeDone: // stream_done,正常结束
			return
		case outcomeFatal: // ctx 取消，已返回
			return
		case outcomeInterrupted: // 需重连
			if !c.prepareReconnect(ctx, events, state, cfg) {
				return // 超过最大重试或 ctx 取消，错误已写入
			}
			req = buildReconnectRequest(req)
		}
	}
}

// ===================== 四、单次连接 =====================

// streamOnce 消费单次连接的事件流，返回本次连接的结束原因
func (c *AgentClient) streamOnce(ctx context.Context, req *starops.CreateChatRequest,
	events chan *ChatEvent, state *retryState, cfg *RetryConfig) connectionOutcome {

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	yield, yieldErr := c.openSSEStream(innerCtx, req)
	idleTimer := time.NewTimer(cfg.IdleTimeout)
	defer idleTimer.Stop()
	start := time.Now()

	for {
		select {
		case resp, ok := <-yield:
			if !ok {
				fmt.Fprintln(os.Stdout, "连接中断，中断原因：通道关闭且未收到 stream_done")
				return outcomeInterrupted // 通道关闭且未收到 stream_done → 连接中断
			}
			// 重置空闲定时器：Stop 后 drain 掉可能已触发的信号，再 Reset
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C: // 如果已被触发，消费掉触发的事件
				default:
				}
			}
			idleTimer.Reset(cfg.IdleTimeout)

			event := parseChatEvent(resp)
			if isStreamDoneEvent(event) { // stream_done 是唯一正常结束标志
				event.IsDone = true
				events <- event
				return outcomeDone
			}
			if c.forwardEvent(event, state, events) {
				state.retryCount = 0
			}

			if c.config.SimulateNetworkError { // 模拟断连（转发后触发）
				if time.Since(start) > 5*time.Second {
					c.config.SimulateNetworkError = false
					fmt.Fprintf(os.Stderr, "模拟网络断连，触发重连...\n")
					return outcomeInterrupted
				}

			}
		// 非 stream_done 的任何错误都视为连接中断，触发重连
		case err, ok := <-yieldErr:
			if !ok {
				yieldErr = nil
				continue
			}
			if err == nil {
				continue
			}
			fmt.Fprintln(os.Stdout, "连接中断，中断原因：SSE连接错误")
			return outcomeInterrupted

		case <-idleTimer.C:
			fmt.Fprintln(os.Stdout, "连接中断，中断原因：空闲超时，未收到消息")
			return outcomeInterrupted // 空闲超时，未收到消息 → 连接中断

		case <-ctx.Done():
			events <- &ChatEvent{Error: ctx.Err()}
			return outcomeFatal
		}
	}
}

// ===================== 五、事件转发 =====================

// forwardEvent 去重转发普通事件，返回是否实际转发了消息
func (c *AgentClient) forwardEvent(event *ChatEvent, state *retryState, events chan *ChatEvent) bool {
	ts := extractNewestTimestamp(event, state.lastTimestamp)

	if state.inDedupeWindow {
		if ts == "" {
			return false // 重复消息，跳过
		}
		state.inDedupeWindow = false // 收到新消息，退出去重窗口
	}

	if ts != "" {
		state.lastTimestamp = ts
	}
	events <- event
	return true
}

// ===================== 六、重连逻辑 =====================

// prepareReconnect 执行退避并判定是否继续重试；返回 false 表示应终止
func (c *AgentClient) prepareReconnect(ctx context.Context, events chan *ChatEvent, state *retryState, cfg *RetryConfig) bool {
	if state.retryCount >= cfg.MaxRetries {
		events <- &ChatEvent{Error: fmt.Errorf("超过最大重试次数 %d 次，连接中断", cfg.MaxRetries)}
		return false
	}
	state.retryCount++
	backoff := calculateBackoff(state.retryCount, cfg)
	fmt.Fprintf(os.Stderr, "连接中断，%v 后重试 (第 %d/%d 次)\n", backoff, state.retryCount, cfg.MaxRetries)
	select {
	case <-time.After(backoff):
		state.inDedupeWindow = true // 进入去重窗口
		return true
	case <-ctx.Done():
		events <- &ChatEvent{Error: ctx.Err()}
		return false
	}
}

// buildReconnectRequest 构建重连请求
func buildReconnectRequest(origReq *starops.CreateChatRequest) *starops.CreateChatRequest {
	reconnReq := &starops.CreateChatRequest{}
	reconnReq.SetAction("reconnect")
	// threadId 和 digitalEmployeeName 从原请求复制
	// 默认非空
	reconnReq.SetThreadId(*origReq.ThreadId)
	reconnReq.SetDigitalEmployeeName(*origReq.DigitalEmployeeName)

	// 复制原始 variables
	variables := make(map[string]interface{})
	if origReq.Variables != nil {
		for k, v := range origReq.Variables {
			variables[k] = v
		}
	}
	reconnReq.SetVariables(variables)

	// 重连不需要 Messages
	return reconnReq
}

// ===================== 七、工具函数 =====================

// openSSEStream 建立单次 SSE 连接，返回响应与错误通道
func (c *AgentClient) openSSEStream(ctx context.Context, req *starops.CreateChatRequest) (chan *starops.CreateChatResponse, chan error) {
	yield := make(chan *starops.CreateChatResponse, 10)
	yieldErr := make(chan error, 1)
	runtime := &dara.RuntimeOptions{}
	runtime.SetConnectTimeout(30000)
	runtime.SetReadTimeout(300000)
	go c.client.CreateChatWithSSECtx(ctx, req, nil, runtime, yield, yieldErr)
	return yield, yieldErr
}

// parseChatEvent 从 SDK 响应解析为 ChatEvent（保持原解析行为）
// IsDone 统一由 streamOnce 在 isStreamDoneEvent 判定后设置
func parseChatEvent(resp *starops.CreateChatResponse) *ChatEvent {
	rawJSON := ""
	if resp.Body != nil {
		if jsonBytes, err := json.Marshal(resp.Body); err == nil {
			rawJSON = string(jsonBytes)
		}
	}

	return &ChatEvent{
		Body:       resp.Body,
		RawJSON:    rawJSON,
		StatusCode: dara.Int32Value(resp.StatusCode),
		Id:         dara.StringValue(resp.Id),
		Event:      dara.StringValue(resp.Event),
	}
}

// isStreamDoneEvent 判断事件是否为 stream_done（正常结束标志）
// 不检查 event.Event 字段，只判断 messages.events.type == stream_done
func isStreamDoneEvent(event *ChatEvent) bool {
	if event == nil {
		return false
	}
	if event.Body != nil && event.Body.Messages != nil {
		for _, msg := range event.Body.Messages {
			if msg == nil || msg.Events == nil {
				continue
			}
			for _, evt := range msg.Events {
				if evtType, ok := evt["type"]; ok {
					if typeStr, ok := evtType.(string); ok && typeStr == "stream_done" {
						return true
					}
				}
			}
		}
	}
	return false
}

// isNewerTimestamp 判断 ts 是否比 base 更新
// 优先数值比较（Unix 时间戳），无法解析时 fallback 为字符串比较
func isNewerTimestamp(ts, base string) bool {
	if ts == "" {
		return false
	}
	if base == "" {
		return true
	}
	tsVal, tsErr := strconv.ParseInt(ts, 10, 64)
	baseVal, baseErr := strconv.ParseInt(base, 10, 64)
	if tsErr == nil && baseErr == nil {
		return tsVal > baseVal
	}
	if tsErr != nil && baseErr == nil {
		return false // 基准是数值但当前 ts 无法解析，视为不更新
	}
	return ts > base
}

// extractNewestTimestamp 从事件中提取比 base 更新的最大消息 timestamp
// 返回空字符串表示没有比 base 更新的时间戳
func extractNewestTimestamp(event *ChatEvent, base string) string {
	if event == nil || event.Body == nil || event.Body.Messages == nil {
		return ""
	}

	newest := base
	for _, msg := range event.Body.Messages {
		if msg == nil {
			continue
		}
		ts := dara.StringValue(msg.Timestamp)
		if isNewerTimestamp(ts, newest) {
			newest = ts
		}
	}
	if newest == base {
		return ""
	}
	return newest
}

// calculateBackoff 计算退避时间
// default config: 1s * 2.0**(retryCount-1)
func calculateBackoff(retryCount int, config *RetryConfig) time.Duration {
	backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffFactor, float64(retryCount-1))
	if time.Duration(backoff) > config.MaxBackoff {
		return config.MaxBackoff
	}
	return time.Duration(backoff)
}
