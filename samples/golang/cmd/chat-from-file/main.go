// 从文件加载请求 - 批量处理 requests 目录下的请求文件
// Load Requests from File - Batch process request files in requests directory
//
// 用法 / Usage:
//
//	go run ./cmd/chat-from-file/ --file ../sample-requests/entity.json
//	go run ./cmd/chat-from-file/ --dir ../sample-requests/              # 处理目录下所有文件
//	go run ./cmd/chat-from-file/ --file entity.json --simple               # 简洁模式
//
// 功能 / Features:
//   - 从 JSON 文件加载请求参数
//   - 支持批量处理目录下所有 JSON 文件
//   - 自动输出日志到 output 目录
//   - 支持简洁模式输出
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/client"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/config"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/interactive"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/printer"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
	"github.com/spf13/cobra"
)

// RequestFile JSON 请求文件结构
type RequestFile struct {
	Region              string         `json:"region"`
	DigitalEmployeeName string         `json:"digitalEmployeeName"`
	ThreadId            string         `json:"threadId,omitempty"`
	Action              string         `json:"action"`
	Messages            []RequestMsg   `json:"messages"`
	Variables           map[string]any `json:"variables"`
}

// RequestMsg 请求消息
type RequestMsg struct {
	Role     string           `json:"role"`
	Contents []RequestContent `json:"contents"`
}

// RequestContent 消息内容
type RequestContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// cliOptions 保存所有 CLI 参数
type cliOptions struct {
	filePath      string
	dirPath       string
	simpleMode    bool
	outputDir     string
	simulateError bool
	mockMode      bool
	mockFile      string
	recordMode    bool
	mockInput     string
}

func main() {
	opts := &cliOptions{}

	rootCmd := &cobra.Command{
		Use:   "chat-from-file",
		Short: "从 JSON 文件加载并批量处理 STAROps 请求",
		Long: `chat-from-file 从 JSON 请求文件加载参数并调用 STAROps Agent，
支持单文件、目录批量处理，以及 Mock / Record 模式与网络断连模拟。`,
		Example: `  # 处理单个文件
  chat-from-file --file ../sample-requests/entity.json

  # 批量处理目录下所有 JSON 文件
  chat-from-file --dir ../sample-requests/

  # 简洁模式
  chat-from-file --file entity.json --simple

  # Mock 回放
  chat-from-file --file entity.json --mock-file ../.starops-samples/record/data_agent.json.mock --mock

  # 录制 SSE 事件
  chat-from-file --file entity.json --record-file ../.starops-samples/record/data_agent.json.mock --record`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.filePath == "" && opts.dirPath == "" {
				return fmt.Errorf("必须指定 --file 或 --dir")
			}
			return runChatFromFile(opts)
		},
		SilenceUsage:  false,
		SilenceErrors: true,
	}

	flags := rootCmd.Flags()
	flags.StringVarP(&opts.filePath, "file", "f", "", "请求 JSON 文件路径")
	flags.StringVarP(&opts.dirPath, "dir", "d", "", "请求文件目录，处理目录下所有 JSON 文件")
	flags.BoolVarP(&opts.simpleMode, "simple", "s", false, "简洁模式，只输出最终文本")
	flags.StringVarP(&opts.outputDir, "output", "o", "../.starops-samples/logs", "输出目录")
	flags.BoolVar(&opts.simulateError, "simulate-error", false, "模拟网络断连，测试重试逻辑")
	flags.BoolVar(&opts.mockMode, "mock", false, "Mock 模式：回放预录制的 SSE 事件文件")
	flags.StringVar(&opts.mockFile, "mock-file", "", "录制和回放的 SSE 事件文件")
	flags.BoolVar(&opts.recordMode, "record", false, "录制模式：旁路捕获 SSE 事件并写入文件")
	flags.StringVar(&opts.mockInput, "mock-input", "", "Mock 交互输入：自动回复交互事件（如 yes/no）")

	rootCmd.MarkFlagsMutuallyExclusive("file", "dir")
	rootCmd.MarkFlagsMutuallyExclusive("mock", "record")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
}

func runChatFromFile(opts *cliOptions) error {
	fmt.Println("🚀 STAROps Chat - 从文件加载请求")
	fmt.Println(strings.Repeat("=", 60))

	// 加载配置
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}
	if opts.simulateError {
		cfg.SimulateNetworkError = true
		fmt.Println("⚠️  已启用网络断连模拟，将在收到首个事件后触发重试")
	}
	// 是否启用 Mock 模式
	if opts.mockMode {
		cfg.MockMode = true
		cfg.MockFile = opts.mockFile
		cfg.MockInput = opts.mockInput
		fmt.Printf("🎭 Mock 模式，回放文件: %s\n", cfg.MockFile)
		if opts.mockInput != "" {
			fmt.Printf("🤖 Mock 交互输入: %s\n", opts.mockInput)
		}
	}
	// 是否启用录制模式
	if opts.recordMode {
		cfg.RecordMode = true
		cfg.MockFile = opts.mockFile
		cfg.MockInput = opts.mockInput
		fmt.Printf("📝 录制模式，输出文件: %s\n", cfg.MockFile)
		if opts.mockInput != "" {
			fmt.Printf("🤖 录制交互输入: %s\n", opts.mockInput)
		}
	}

	// 创建客户端
	agentClient, err := client.NewAgentClient(cfg)
	if err != nil {
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(opts.outputDir, 0755); err != nil {
		fmt.Printf("⚠️ 创建输出目录失败: %v\n", err)
	}

	// 处理文件
	if opts.dirPath != "" {
		processDirectory(agentClient, opts.dirPath, opts)
	} else {
		processFile(agentClient, opts.filePath, opts)
	}
	return nil
}

func processDirectory(agentClient *client.AgentClient, dir string, opts *cliOptions) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		fmt.Printf("❌ 读取目录失败: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("⚠️ 目录中没有 JSON 文件: %s\n", dir)
		return
	}

	fmt.Printf("📁 找到 %d 个请求文件\n\n", len(files))

	for i, file := range files {
		fmt.Printf("━━━ [%d/%d] %s ━━━\n", i+1, len(files), filepath.Base(file))
		processFile(agentClient, file, opts)
		fmt.Println()
	}

	fmt.Printf("✅ 处理完成，共 %d 个文件\n", len(files))
}

func processFile(agentClient *client.AgentClient, file string, opts *cliOptions) {
	// 加载请求文件
	reqFile, err := loadRequestFile(file)
	if err != nil {
		fmt.Printf("❌ 加载文件失败: %v\n", err)
		return
	}

	// 获取消息
	message := extractMessage(reqFile)
	if message == "" {
		fmt.Printf("⚠️ 文件中没有消息内容\n")
		return
	}

	fmt.Printf("📄 文件: %s\n", filepath.Base(file))
	fmt.Printf("💬 消息: %s\n", truncate(message, 60))

	ctx := context.Background()

	// 创建会话
	threadID, err := agentClient.CreateThread(ctx)
	if err != nil {
		fmt.Printf("❌ 创建会话失败: %v\n", err)
		return
	}

	// 捕获中断信号，发送 stop 请求
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n⏹️  正在停止对话...")
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

	// 创建输出文件
	outputFile := createOutputFile(file, opts.outputDir)
	defer outputFile.Close()

	// 写入请求信息
	writeOutput(outputFile, fmt.Sprintf("# Request: %s\n", filepath.Base(file)))
	writeOutput(outputFile, fmt.Sprintf("# Time: %s\n", time.Now().Format(time.RFC3339)))
	writeOutput(outputFile, fmt.Sprintf("# ThreadID: %s\n", threadID))
	writeOutput(outputFile, fmt.Sprintf("# Message: %s\n\n", message))

	fmt.Println(strings.Repeat("-", 60))

	// 发送请求
	startTime := time.Now()

	// 处理响应
	var simplePrinter *printer.SimplePrinter
	var eventPrinter *printer.EventPrinter
	if opts.simpleMode {
		simplePrinter = printer.NewSimplePrinter()
	} else {
		eventPrinter = printer.NewEventPrinter(false, true)
	}

	// 初始化交互处理器
	interactiveHandler := interactive.NewHandler(agentClient, 0)
	if agentClient.GetConfig().MockInput != "" {
		interactiveHandler.SetMockInput(agentClient.GetConfig().MockInput)
	}

	events := agentClient.ChatWithVariables(ctx, threadID, message, reqFile.Variables)
	processEvents(events, simplePrinter, eventPrinter, interactiveHandler, outputFile, ctx, threadID, reqFile.Variables)

	elapsed := time.Since(startTime)
	fmt.Println()

	// 写入最终结果
	if opts.simpleMode && simplePrinter != nil {
		finalText := simplePrinter.GetFinalText()
		writeOutput(outputFile, fmt.Sprintf("\n# Final Result:\n%s\n", finalText))
		fmt.Printf("📄 最终文本:\n%s\n", finalText)
	}

	writeOutput(outputFile, fmt.Sprintf("\n# Duration: %v\n", elapsed))
	fmt.Printf("⏱️  耗时: %v\n", elapsed)
	fmt.Printf("📁 输出: %s\n", outputFile.Name())
}

// processEvents 处理 SSE 事件流，支持交互事件检测和流恢复
func processEvents(
	events <-chan *types.ChatEvent,
	simplePrinter *printer.SimplePrinter,
	eventPrinter *printer.EventPrinter,
	handler *interactive.Handler,
	outputFile *os.File,
	ctx context.Context,
	threadID string,
	variables map[string]any,
) {
	eventIndex := 0
	for events != nil {
		event, ok := <-events
		if !ok {
			break
		}
		eventIndex++

		if event.Error != nil {
			fmt.Printf("❌ 错误: %v\n", event.Error)
			writeOutput(outputFile, fmt.Sprintf("[ERROR] %v\n", event.Error))
			continue
		}

		// 写入原始事件
		if event.RawJSON != "" {
			writeOutput(outputFile, fmt.Sprintf("[EVENT %d]\n%s\n\n", eventIndex, event.RawJSON))
		}

		// 正常输出（先输出，确保交互事件内容可见）
		if simplePrinter != nil {
			text := simplePrinter.ProcessEvent(event)
			if text != "" {
				fmt.Print(text)
			}
		} else {
			eventPrinter.PrintEvent(event, eventIndex)
		}

		// 检测交互事件（在输出之后，确保用户看到交互内容）
		interactiveResp := interactive.ExtractResponse(ctx, event, handler)
		if interactiveResp != nil {
			fmt.Printf("\n🔄 检测到交互事件，用户已响应...\n")
			events = handler.ResumeChat(ctx, threadID, interactiveResp, variables)
			eventIndex = 0
			continue
		}

		if event.IsDone {
			break
		}
	}
}

func loadRequestFile(filePath string) (*RequestFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var req RequestFile
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &req, nil
}

func extractMessage(req *RequestFile) string {
	if len(req.Messages) > 0 && len(req.Messages[0].Contents) > 0 {
		return req.Messages[0].Contents[0].Value
	}
	return ""
}

func createOutputFile(inputFile, outputDir string) *os.File {
	baseName := strings.TrimSuffix(filepath.Base(inputFile), ".json")
	timestamp := time.Now().Format("20060102-150405")
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s.log", baseName, timestamp))

	f, err := os.Create(outputPath)
	if err != nil {
		// 如果创建失败，返回一个空的文件句柄
		return nil
	}
	return f
}

func writeOutput(f *os.File, content string) {
	if f != nil {
		f.WriteString(content)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
