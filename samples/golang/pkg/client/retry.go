// retry.go — 指数退避与重连请求构建
// 职责：退避时间计算、重试判定、重连请求构造。
// 不做：不消费事件流（→stream_sse.go）、不打开连接（→stream.go）。
package client

import (
	"context"
	"fmt"
	"math"
	"time"

	starops "github.com/alibabacloud-go/starops-20260428/client"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/config"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/logger"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// calculateBackoff 计算退避时间
func calculateBackoff(retryCount int, cfg *config.RetryConfig) time.Duration {
	backoff := float64(cfg.InitialBackoff) * math.Pow(cfg.BackoffFactor, float64(retryCount-1))
	if time.Duration(backoff) > cfg.MaxBackoff {
		return cfg.MaxBackoff
	}
	return time.Duration(backoff)
}

// prepareReconnect 执行退避并判定是否继续重试
func (c *AgentClient) prepareReconnect(ctx context.Context, events chan *types.ChatEvent, state *retryState, cfg *config.RetryConfig) bool {
	if state.retryCount >= cfg.MaxRetries {
		select {
		case events <- &types.ChatEvent{Error: fmt.Errorf("超过最大重试次数 %d 次，连接中断", cfg.MaxRetries)}:
		case <-ctx.Done():
		}
		return false
	}
	state.retryCount++
	backoff := calculateBackoff(state.retryCount, cfg)
	logger.Default().Warn("连接中断，准备重试", map[string]any{
		"backoff":    backoff.String(),
		"retryCount": state.retryCount,
		"maxRetries": cfg.MaxRetries,
	})
	select {
	case <-time.After(backoff):
		state.inDedupeWindow = true
		return true
	case <-ctx.Done():
		select {
		case events <- &types.ChatEvent{Error: ctx.Err()}:
		default:
		}
		return false
	}
}

// buildReconnectRequest 构建重连请求
func buildReconnectRequest(origReq *starops.CreateChatRequest) *starops.CreateChatRequest {
	reconnReq := &starops.CreateChatRequest{}
	reconnReq.SetAction("reconnect")
	if origReq.ThreadId != nil {
		reconnReq.SetThreadId(*origReq.ThreadId)
	}
	if origReq.DigitalEmployeeName != nil {
		reconnReq.SetDigitalEmployeeName(*origReq.DigitalEmployeeName)
	}

	variables := make(map[string]interface{})
	if origReq.Variables != nil {
		for k, v := range origReq.Variables {
			variables[k] = v
		}
	}
	reconnReq.SetVariables(variables)

	return reconnReq
}
