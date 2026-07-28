// stream_sse.go — SSE 流编排与事件去重
// 职责：管理 SSE 连接生命周期（streamSSE/streamOnce），事件去重转发，时间戳比较。
// 不做：不打开底层连接（→stream.go）、不解析事件载荷（→event.go）、不计算退避（→retry.go）。
package client

import (
	"context"
	"strconv"
	"time"

	starops "github.com/alibabacloud-go/starops-20260428/client"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/config"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/logger"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// retryState 聚合重连过程中的状态
type retryState struct {
	lastTimestamp  string
	inDedupeWindow bool
	retryCount     int
}

// connectionOutcome 单次连接的结束原因
type connectionOutcome int

const (
	outcomeDone        connectionOutcome = iota
	outcomeInterrupted
	outcomeFatal
)

// isNewerTimestamp 判断 ts 是否比 base 更新
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
		return false
	}
	return ts > base
}

// extractNewestTimestamp 从事件中提取比 base 更新的最大消息 timestamp
func extractNewestTimestamp(evt *types.ChatEvent, base string) string {
	if evt == nil || evt.Body == nil || evt.Body.Messages == nil {
		return ""
	}

	newest := base
	for _, msg := range evt.Body.Messages {
		if msg == nil {
			continue
		}
		ts := msg.Timestamp
		if isNewerTimestamp(ts, newest) {
			newest = ts
		}
	}
	if newest == base {
		return ""
	}
	return newest
}

// forwardEvent 去重转发普通事件，返回是否实际转发了消息
func (c *AgentClient) forwardEvent(ctx context.Context, evt *types.ChatEvent, state *retryState, events chan *types.ChatEvent) bool {
	ts := extractNewestTimestamp(evt, state.lastTimestamp)

	if state.inDedupeWindow {
		if ts == "" {
			return false
		}
		state.inDedupeWindow = false
	}

	if ts != "" {
		state.lastTimestamp = ts
	}
	select {
	case events <- evt:
	case <-ctx.Done():
		return false
	}
	return true
}

// streamSSE 启动带重试能力的 SSE 流处理
func (c *AgentClient) streamSSE(ctx context.Context, req *starops.CreateChatRequest, events chan *types.ChatEvent) {
	cfg := c.config.RetryConfig
	if cfg == nil {
		cfg = config.LoadRetryConfigFromEnv()
	}
	state := &retryState{}
	for {
		outcome := c.streamOnce(ctx, req, events, state, cfg)
		switch outcome {
		case outcomeDone:
			return
		case outcomeFatal:
			return
		case outcomeInterrupted:
			if !c.prepareReconnect(ctx, events, state, cfg) {
				return
			}
			req = buildReconnectRequest(req)
		}
	}
}

// streamOnce 消费单次连接的事件流，返回本次连接的结束原因
func (c *AgentClient) streamOnce(ctx context.Context, req *starops.CreateChatRequest,
	events chan *types.ChatEvent, state *retryState, cfg *config.RetryConfig) connectionOutcome {

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
				logger.Default().Info("连接中断", map[string]any{"reason": "channel_closed_no_stream_done"})
				return outcomeInterrupted
			}
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(cfg.IdleTimeout)

			evt := parseChatEvent(resp)
			if isStreamDoneEvent(evt) {
				evt.IsDone = true
				select {
				case events <- evt:
				case <-ctx.Done():
				}
				return outcomeDone
			}
			if c.forwardEvent(ctx, evt, state, events) {
				state.retryCount = 0
			}

			if c.config.SimulateNetworkError {
				if time.Since(start) > 5*time.Second {
					c.config.SimulateNetworkError = false
					logger.Default().Warn("模拟网络断连，触发重连", nil)
					return outcomeInterrupted
				}
			}

		case err, ok := <-yieldErr:
			if !ok {
				yieldErr = nil
				continue
			}
			if err == nil {
				continue
			}
			logger.Default().Info("连接中断", map[string]any{"reason": "sse_error", "error": err.Error()})
			return outcomeInterrupted

		case <-idleTimer.C:
			logger.Default().Info("连接中断", map[string]any{"reason": "idle_timeout"})
			return outcomeInterrupted

		case <-ctx.Done():
			select {
			case events <- &types.ChatEvent{Error: ctx.Err()}:
			default:
			}
			return outcomeFatal
		}
	}
}
