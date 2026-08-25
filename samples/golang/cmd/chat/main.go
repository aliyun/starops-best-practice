// 交互式对话示例
// Interactive Chat Example
//
// 用法 / Usage: go run ./cmd/chat/ [--simulate-error]
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/client"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/config"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/interactive"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/models"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/printer"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
	"github.com/spf13/cobra"
)

// chatOptions 保存 chat 命令的参数
type chatOptions struct {
	simulateError bool
	model         string
	listModels    bool
}

func main() {
	opts := &chatOptions{}

	rootCmd := &cobra.Command{
		Use:   "chat",
		Short: "STAROps 交互式对话示例",
		Long: `STAROps 交互式对话示例

从终端读取用户输入并与 STAROps Agent 进行流式对话，
支持交互事件恢复与网络断连重试。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(opts)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.Flags().BoolVar(&opts.simulateError, "simulate-error", false, "模拟网络断连，测试重试逻辑")
	rootCmd.Flags().StringVar(&opts.model, "model", "", "指定模型 (格式: provider:modelId)")
	rootCmd.Flags().BoolVar(&opts.listModels, "list-models", false, "列出所有可传入 --model 的模型取值后退出")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
}

func runChat(opts *chatOptions) error {
	fmt.Println("🚀 STAROps Chat")
	fmt.Println(strings.Repeat("=", 60))

	// 加载模型配置（不依赖凭据）
	modelsCfg, _ := models.LoadModels(models.FindConfigPath())
	if opts.listModels {
		if modelsCfg == nil {
			fmt.Println("⚠️  模型配置未加载")
			return nil
		}
		fmt.Println(models.ListModelFlags(modelsCfg))
		return nil
	}

	// 加载凭据配置
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		fmt.Println("\n请设置环境变量:")
		fmt.Println("  STAROPS_ENDPOINT")
		fmt.Println("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")
		return fmt.Errorf("配置加载失败: %w", err)
	}

	fmt.Printf("📋 Employee: %s\n\n", cfg.EmployeeName)

	if opts.simulateError {
		cfg.SimulateNetworkError = true
		fmt.Println("⚠️  已启用网络断连模拟，将在收到首个事件后触发重试")
	}

	// 创建客户端
	agentClient, err := client.NewAgentClient(cfg)
	if err != nil {
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	ctx := context.Background()

	// 创建会话
	fmt.Println("📝 创建会话...")
	var threadAttrs map[string]string
	if opts.model != "" {
		provider, modelID, err := models.ParseModelFlag(opts.model)
		if err != nil {
			return fmt.Errorf("解析 --model 失败: %w", err)
		}
		if modelsCfg != nil && !models.ValidateModel(modelsCfg, provider, modelID) {
			return fmt.Errorf("未知模型: %s:%s，可用 --list-models 查看可选值", provider, modelID)
		}
		threadAttrs = map[string]string{"model": models.BuildModelJSON(provider, modelID)}
		fmt.Printf("🤖 已指定模型: %s:%s\n", provider, modelID)
	}
	threadID, err := agentClient.CreateThread(ctx, threadAttrs)
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	fmt.Printf("✅ ThreadID: %s\n\n", threadID)

	// 捕获中断信号，发送 stop 请求
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n⏹️  正在停止对话...")
		// 带超时的 goroutine，确保即使 Stop 卡住也能退出
		done := make(chan struct{})
		go func() {
			if err := agentClient.Stop(context.Background(), threadID, nil); err != nil {
				fmt.Printf("⚠️  stop 请求失败: %v\n", err)
			}
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(6 * time.Second):
			fmt.Println("⚠️  stop 请求超时，强制退出")
		}
		os.Exit(0)
	}()

	// 创建打印器
	simplePrinter := printer.NewSimplePrinter()
	interactiveHandler := interactive.NewHandler(agentClient, 0)

	// 交互循环
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("👤 请输入 (quit 退出): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n 再见!")
				break
			}
			continue
		}

		input = strings.TrimSpace(input)
		switch input {
		case "":
			continue
		case "quit", "exit":
			fmt.Println("👋 再见!")
			return nil
		case "/model":
			if modelsCfg == nil {
				fmt.Println("⚠️  模型配置未加载")
				continue
			}
			fmt.Println(models.DisplayMenu(modelsCfg))
			fmt.Print("请输入序号选择模型: ")
			choice, _ := reader.ReadString('\n')
			idx, err := strconv.Atoi(strings.TrimSpace(choice))
			if err != nil {
				fmt.Println("❌ 无效输入")
				continue
			}
			provider, modelID, err := models.GetModelByIndex(modelsCfg, idx-1)
			if err != nil {
				fmt.Println("❌", err)
				continue
			}
			if err := agentClient.UpdateThread(ctx, threadID, map[string]string{"model": models.BuildModelJSON(provider, modelID)}); err != nil {
				fmt.Printf("❌ 更新失败: %v\n", err)
				continue
			}
			fmt.Println("✅ 更新模型成功")
			continue
		default:
			fmt.Println(strings.Repeat("-", 60))
			simplePrinter.Reset()
			events := agentClient.Chat(ctx, threadID, input)
			processChatEvents(events, simplePrinter, interactiveHandler, ctx, threadID)
			fmt.Println()
			fmt.Println(strings.Repeat("=", 60))
			fmt.Println()
		}
	}

	return nil
}

// processChatEvents 处理 SSE 事件流，支持交互事件检测和流恢复
func processChatEvents(
	events <-chan *types.ChatEvent,
	simplePrinter *printer.SimplePrinter,
	handler *interactive.Handler,
	ctx context.Context,
	threadID string,
) {
	for events != nil {
		event, ok := <-events
		if !ok {
			break
		}

		if event.Error != nil {
			fmt.Printf("❌ 错误: %v\n", event.Error)
			continue
		}

		// 正常输出（先输出，确保交互事件内容可见）
		text := simplePrinter.ProcessEvent(event)
		if text != "" {
			fmt.Print(text)
		}

		// 检测交互事件（在输出之后，确保用户看到交互内容）
		interactiveResp := interactive.ExtractResponse(ctx, event, handler)
		if interactiveResp != nil {
			variables := map[string]any{}
			events = handler.ResumeChat(ctx, threadID, interactiveResp, variables)
			continue
		}
	}
}
