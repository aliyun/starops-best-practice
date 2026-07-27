// Package interactive — 三类交互动作实现
// actions.go 职责：HandleUserAck / HandleUserSelect / HandleUserInput 三个交互流程的终端展示与响应收集。
// 不做：不做事件分发（→interactive.go）、不做原始 payload 字段解析（→payload.go）、不做底层 I/O（→io.go）。
package interactive

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/errors"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// HandleUserAck 处理用户确认
// 显示确认提示并等待用户响应
// callId: 来自外层 MessageItem.CallID
// 字段映射: title ← userAck.data.title, message ← userAck.message
func (h *Handler) HandleUserAck(ctx context.Context, payload *types.ItemInteractivePayload, callId string) (*Response, error) {
	if payload == nil {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, "交互负载为空")
	}

	title := h.getTitle(payload)
	message := h.getDescription(payload)
	options := h.getOptions(payload)
	modifiedData := h.extractData(payload)

	h.printf("\n🔔 确认请求\n")
	h.printf("--------------\n")
	if title != "" {
		h.printf("%s\n", title)
	}
	if message != "" {
		h.printf("\n%s\n", message)
	}

	// 展示 data 详情字段
	if modifiedData != nil {
		h.printf("\n")
		for key, value := range modifiedData {
			// 跳过 title 和 message（已在上方展示）
			if key == "title" || key == "message" {
				continue
			}
			h.printf("%s: %v\n", key, value)
		}
	}
	h.printf("--------------\n")

	// 构建选项提示
	if len(options) > 0 {
		h.printf("请输入 ")
		for i, opt := range options {
			if i > 0 {
				h.printf(", ")
			}
			value := h.getOptionValue(opt)
			label := h.getOptionLabel(opt, i)
			h.printf("[%s] %s", value, label)
		}
		h.printf(": ")
	} else {
		h.printf("请输入 [y/yes] 确认，[n/no] 取消: ")
	}

	// 读取用户输入
	input, err := h.readInputWithTimeout(ctx)
	if err != nil {
		return nil, err
	}

	// 解析用户响应
	input = strings.TrimSpace(strings.ToLower(input))
	confirmed := input == "y" || input == "yes" || input == "是" || input == ""

	decision := "no"
	if confirmed {
		decision = "yes"
	}

	// 如果用户输入匹配某个选项的 value，使用该 value 作为 decision
	for _, opt := range options {
		if input == strings.ToLower(h.getOptionValue(opt)) {
			decision = h.getOptionValue(opt)
			confirmed = true
			break
		}
	}

	source := h.extractSource(payload)

	return &Response{
		CallId: callId,
		Type:   types.InteractionTypeUserAck,
		Response: map[string]any{
			"confirmed": confirmed,
		},
		Source:       source,
		ModifiedData: modifiedData,
		Decision:     decision,
	}, nil
}

// HandleUserSelect 处理用户选择
// 显示选项列表并捕获用户选择
func (h *Handler) HandleUserSelect(ctx context.Context, payload *types.ItemInteractivePayload, callId string) (*Response, error) {
	if payload == nil {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, "交互负载为空")
	}

	// 显示选择提示
	title := h.getTitle(payload)
	h.printf("\n📋 请选择\n")
	if title != "" {
		h.printf("   标题: %s\n", title)
	}

	// 获取选项列表
	options := h.getOptions(payload)
	if len(options) == 0 {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, "没有可选项")
	}

	// 显示选项
	h.printf("   选项:\n")
	for i, opt := range options {
		label := h.getOptionLabel(opt, i)
		h.printf("   [%d] %s\n", i+1, label)
	}
	h.printf("   请输入选项编号 (1-%d): ", len(options))

	// 读取用户输入
	input, err := h.readInputWithTimeout(ctx)
	if err != nil {
		return nil, err
	}

	// 解析用户选择
	input = strings.TrimSpace(input)
	selectedIndex, err := strconv.Atoi(input)
	if err != nil || selectedIndex < 1 || selectedIndex > len(options) {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, fmt.Sprintf("无效的选择: %s，请输入 1-%d 之间的数字", input, len(options)))
	}

	// 获取选中的选项
	selectedOption := options[selectedIndex-1]

	// 提取决策值
	decision := h.getOptionValue(selectedOption)

	return &Response{
		CallId: callId,
		Type:   types.InteractionTypeUserSelect,
		Response: map[string]any{
			"selectedIndex": selectedIndex - 1, // 0-based index
			"selectedValue": selectedOption,
		},
		Source:       h.extractSource(payload),
		ModifiedData: h.extractData(payload),
		Decision:     decision,
	}, nil
}

// HandleUserInput 处理用户输入（表单模式）
// 根据 formSpec 中的 ui_schema 逐字段提示用户输入
// userInput 结构: {title, description, formSpec: {schema, ui_schema, initialValues}, source}
func (h *Handler) HandleUserInput(ctx context.Context, payload *types.ItemInteractivePayload, callId string) (*Response, error) {
	if payload == nil {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, "交互负载为空")
	}

	title := h.getTitle(payload)
	description := h.getDescription(payload)
	source := h.extractSource(payload)

	// 提取 formSpec
	formSpec := h.extractFormSpec(payload)
	elements := h.getFormElements(formSpec)
	initialValues := h.getFormInitialValues(formSpec)

	h.printf("\n✏️  %s\n", title)
	if description != "" {
		h.printf("    %s\n", description)
	}
	h.printf("    %s\n", strings.Repeat("-", 40))

	formData := make(map[string]any)

	// 逐个字段收集输入
	for _, elem := range elements {
		field := h.getFieldKey(elem)
		label := h.getFieldLabel(elem, field)
		widget := h.getFieldWidget(elem)
		placeholder := h.getFieldPlaceholder(elem)
		defaultValue := h.getInitialValue(initialValues, field)

		switch widget {
		case "radio", "segmented":
			// 枚举选择: 从 schema.properties[field].enum 获取选项
			enumOpts := h.getFieldEnum(formSpec, field)
			if len(enumOpts) > 0 {
				h.printf("    %s:\n", label)
				for i, opt := range enumOpts {
					marker := " "
					if defaultValue == opt {
						marker = "*"
					}
					h.printf("      [%d]%s %s\n", i+1, marker, opt)
				}
				h.printf("    请选择 (1-%d)", len(enumOpts))
				if defaultValue != "" {
					h.printf(" [默认: %s]", defaultValue)
				}
				h.printf(": ")

				input, err := h.readInputWithTimeout(ctx)
				if err != nil {
					return nil, err
				}
				input = strings.TrimSpace(input)
				if input == "" && defaultValue != "" {
					formData[field] = defaultValue
				} else {
					idx, err := strconv.Atoi(input)
					if err != nil || idx < 1 || idx > len(enumOpts) {
						formData[field] = defaultValue
					} else {
						formData[field] = enumOpts[idx-1]
					}
				}
			}

		case "textarea":
			fallthrough
		default:
			// 文本输入
			h.printf("    %s", label)
			if placeholder != "" {
				h.printf(" (%s)", placeholder)
			}
			if defaultValue != "" {
				h.printf(" [默认: %s]", defaultValue)
			}
			h.printf(": ")

			input, err := h.readInputWithTimeout(ctx)
			if err != nil {
				return nil, err
			}
			input = strings.TrimSpace(input)
			if input == "" && defaultValue != "" {
				formData[field] = defaultValue
			} else {
				formData[field] = input
			}
		}
	}

	h.printf("    %s\n", strings.Repeat("-", 40))

	return &Response{
		CallId: callId,
		Type:   types.InteractionTypeUserInput,
		Response: map[string]any{
			"value": formData,
		},
		Source:   source,
		FormData: formData,
		Decision: "submit",
	}, nil
}
