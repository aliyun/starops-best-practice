package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vibeops/samples/golang/types"
)

// EventPrinter 事件打印器
type EventPrinter struct {
	PrintRawBody   bool
	PrintParsed    bool
	PrintSeparator bool
}

// NewEventPrinter 创建事件打印器
func NewEventPrinter(printRawBody, printParsed bool) *EventPrinter {
	return &EventPrinter{
		PrintRawBody:   printRawBody,
		PrintParsed:    printParsed,
		PrintSeparator: true,
	}
}

// PrintEvent 打印事件
func (p *EventPrinter) PrintEvent(event *ChatEvent, eventIndex int) {
	if event.Error != nil {
		fmt.Printf("\n❌ 错误: %v\n", event.Error)
		return
	}

	if event.IsDone && event.Body == nil {
		fmt.Println("\n✅ 对话完成")
		return
	}

	if event.Body == nil {
		return
	}

	if p.PrintSeparator {
		fmt.Printf("\n%s 事件 #%d %s\n", strings.Repeat("=", 30), eventIndex, strings.Repeat("=", 30))
	}

	if p.PrintRawBody && event.RawJSON != "" {
		fmt.Println("\n📦 原始 Body:")
		prettyJSON, err := PrettyPrintJSON(event.RawJSON)
		if err == nil {
			fmt.Println(prettyJSON)
		} else {
			fmt.Println(event.RawJSON)
		}
	}

	if p.PrintParsed {
		p.printParsedEvent(event)
	}
}

func (p *EventPrinter) printParsedEvent(event *ChatEvent) {
	if event.Body == nil {
		return
	}

	fmt.Println("\n📋 解析详情:")

	for _, msg := range event.Body.Messages {
		msgJSON, _ := json.Marshal(msg)
		fmt.Printf("  原始消息: %s\n", string(msgJSON))
		var messageItem types.MessageItem
		if err := json.Unmarshal(msgJSON, &messageItem); err != nil {
			fmt.Printf("  解析消息失败: %v\n", err)
			continue
		}
		p.printMessageItem(&messageItem)
	}
}

func (p *EventPrinter) printMessageItem(item *types.MessageItem) {
	if item.Role != "" {
		fmt.Printf("  📌 角色: %s\n", item.Role)
	}
	if item.CallID != "" {
		fmt.Printf("  🔗 CallID: %s\n", item.CallID)
	}
	if item.ParentCallID != "" {
		fmt.Printf("  🔗 ParentCallID: %s\n", item.ParentCallID)
	}

	if len(item.Contents) > 0 {
		fmt.Println("  📝 内容:")
		for i, content := range item.Contents {
			fmt.Printf("    [%d] 类型: %s\n", i, content.Type)
			if content.Value != "" {
				value := content.Value
				if len(value) > 200 {
					value = value[:200] + "..."
				}
				fmt.Printf("        值: %s\n", value)
			}
			if content.Append {
				fmt.Printf("        追加: true\n")
			}
			if content.LastChunk {
				fmt.Printf("        最后块: true\n")
			}
		}
	}

	if len(item.Tools) > 0 {
		fmt.Println("  🔧 工具调用:")
		for i, tool := range item.Tools {
			fmt.Printf("    [%d] 名称: %s, 状态: %s\n", i, tool.Name, tool.Status)
			if tool.ToolCallID != "" {
				fmt.Printf("        ToolCallID: %s\n", tool.ToolCallID)
			}
			if tool.Arguments != nil {
				argsJSON, _ := json.Marshal(tool.Arguments)
				argsStr := string(argsJSON)
				if len(argsStr) > 200 {
					argsStr = argsStr[:200] + "..."
				}
				fmt.Printf("        参数: %s\n", argsStr)
			}
		}
	}

	if len(item.Agents) > 0 {
		fmt.Println("  🤖 Agent调用:")
		for i, agent := range item.Agents {
			fmt.Printf("    [%d] 名称: %s, 状态: %s\n", i, agent.Name, agent.Status)
		}
	}

	if len(item.Events) > 0 {
		fmt.Println("  📢 事件:")
		for i, evt := range item.Events {
			fmt.Printf("    [%d] 类型: %s\n", i, evt.Type)
			if evt.Payload != nil {
				p.printEventPayload(evt)
			}
		}
	}
}

func (p *EventPrinter) printEventPayload(evt *types.ItemEvent) {
	payloadJSON, _ := json.Marshal(evt.Payload)

	switch evt.Type {
	case types.EventTypeThinking:
		var thinking types.ItemThinkingPayload
		if err := json.Unmarshal(payloadJSON, &thinking); err == nil && thinking.ReasoningDelta != "" {
			delta := thinking.ReasoningDelta
			if len(delta) > 100 {
				delta = delta[:100] + "..."
			}
			fmt.Printf("        思考: %s\n", delta)
		}

	case types.EventTypeError:
		var errPayload types.ItemErrorPayload
		if err := json.Unmarshal(payloadJSON, &errPayload); err == nil {
			fmt.Printf("        错误码: %s\n", errPayload.Code)
			fmt.Printf("        消息: %s\n", errPayload.Message)
		}

	case types.EventTypeTaskFinished:
		var finished types.ItemTaskFinishedPayload
		if err := json.Unmarshal(payloadJSON, &finished); err == nil {
			fmt.Printf("        成功: %v\n", finished.Success)
			if finished.Statistics != nil {
				fmt.Printf("        耗时: %dms\n", finished.Statistics.Duration/1000000)
			}
		}

	default:
		payloadStr := string(payloadJSON)
		if len(payloadStr) > 200 {
			payloadStr = payloadStr[:200] + "..."
		}
		fmt.Printf("        负载: %s\n", payloadStr)
	}
}

// PrettyPrintJSON 格式化JSON输出
func PrettyPrintJSON(jsonStr string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	prettyBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(prettyBytes), nil
}
