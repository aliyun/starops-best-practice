// Package interactive — 交互处理器的终端 I/O
// io.go 职责：带超时的输入读取、格式化输出。
// 不做：不解析 payload、不做业务分发（→interactive.go / actions.go）。
package interactive

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/errors"
)

// readInputWithTimeout 带超时的读取用户输入
func (h *Handler) readInputWithTimeout(ctx context.Context) (string, error) {
	// mock 模式：直接返回预设输入
	if h.mockInput != "" {
		input := h.mockInput
		h.mockInput = "" // 一次性消费
		return input, nil
	}

	// 创建输入通道
	inputChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		reader := bufio.NewReader(h.reader)
		input, err := reader.ReadString('\n')
		if err != nil {
			errChan <- err
			return
		}
		inputChan <- input
	}()

	// 根据是否设置超时选择等待方式
	if h.timeout > 0 {
		select {
		case input := <-inputChan:
			return input, nil
		case err := <-errChan:
			return "", errors.NewSDKErrorWithCause(errors.ErrCodeParseError, "读取输入失败", err)
		case <-time.After(h.timeout):
			return "", errors.NewSDKError(errors.ErrCodeInteractiveTimeout, fmt.Sprintf("用户响应超时 (%v)", h.timeout)).
				WithContext("timeout", h.timeout.String())
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return "", errors.NewSDKErrorWithCause(errors.ErrCodeInteractiveTimeout, "上下文超时", ctx.Err())
			}
			return "", errors.NewSDKErrorWithCause(errors.ErrCodeCancelled, "操作已取消", ctx.Err())
		}
	}

	// 无超时，只等待输入或上下文取消
	select {
	case input := <-inputChan:
		return input, nil
	case err := <-errChan:
		return "", errors.NewSDKErrorWithCause(errors.ErrCodeParseError, "读取输入失败", err)
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.NewSDKErrorWithCause(errors.ErrCodeInteractiveTimeout, "上下文超时", ctx.Err())
		}
		return "", errors.NewSDKErrorWithCause(errors.ErrCodeCancelled, "操作已取消", ctx.Err())
	}
}

// printf 格式化输出
func (h *Handler) printf(format string, args ...any) {
	fmt.Fprintf(h.writer, format, args...)
}
