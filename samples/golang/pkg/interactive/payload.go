// Package interactive — 交互 payload 字段解析辅助函数
// payload.go 职责：解析 ItemInteractivePayload 各字段（title / description / options / source / data / formSpec 等）。
// 不做：不做业务动作（→actions.go）、不做 I/O（→io.go）、不做事件分发（→interactive.go）。
package interactive

import (
	"encoding/json"
	"fmt"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/errors"
	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// parseInteractivePayload 解析交互负载
func (h *Handler) parseInteractivePayload(payload any) (*types.ItemInteractivePayload, error) {
	if payload == nil {
		return nil, errors.NewSDKError(errors.ErrCodeParseError, "交互负载为空")
	}

	// 尝试直接类型断言
	if p, ok := payload.(*types.ItemInteractivePayload); ok {
		return p, nil
	}

	// 尝试 JSON 序列化/反序列化
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.NewSDKErrorWithCause(errors.ErrCodeParseError, "序列化交互负载失败", err)
	}

	var interactivePayload types.ItemInteractivePayload
	if err := json.Unmarshal(payloadJSON, &interactivePayload); err != nil {
		return nil, errors.NewSDKErrorWithCause(errors.ErrCodeParseError, "解析交互负载失败", err)
	}

	return &interactivePayload, nil
}

// getTitle 从负载中获取标题
// 优先从 userAck.data.title / userInput.title 提取，fallback 到 meta.title
func (h *Handler) getTitle(payload *types.ItemInteractivePayload) string {
	if payload.UserAck != nil {
		src := *payload.UserAck
		if data, ok := src["data"].(map[string]any); ok {
			if title, ok := data["title"].(string); ok {
				return title
			}
		}
	}
	if payload.UserInput != nil {
		src := *payload.UserInput
		if title, ok := src["title"].(string); ok {
			return title
		}
	}
	if payload.Meta != nil {
		if title, ok := payload.Meta["title"].(string); ok {
			return title
		}
	}
	return ""
}

// getDescription 从负载中获取描述
// 优先从 userAck.message / userInput.description 提取，fallback 到 meta
func (h *Handler) getDescription(payload *types.ItemInteractivePayload) string {
	if payload.UserAck != nil {
		src := *payload.UserAck
		if message, ok := src["message"].(string); ok {
			return message
		}
	}
	if payload.UserInput != nil {
		src := *payload.UserInput
		if desc, ok := src["description"].(string); ok {
			return desc
		}
	}
	if payload.Meta != nil {
		if desc, ok := payload.Meta["description"].(string); ok {
			return desc
		}
		if desc, ok := payload.Meta["desc"].(string); ok {
			return desc
		}
	}
	return ""
}

// getOptions 从负载中获取选项列表
// 优先从 userAck.options 提取，fallback 到 Data → Meta.options
func (h *Handler) getOptions(payload *types.ItemInteractivePayload) []map[string]any {
	// 优先从 userAck.options 获取
	if payload.UserAck != nil {
		src := *payload.UserAck
		if opts, ok := src["options"].([]any); ok {
			result := make([]map[string]any, 0, len(opts))
			for _, opt := range opts {
				if optMap, ok := opt.(map[string]any); ok {
					result = append(result, optMap)
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}

	// fallback: Data 字段
	if len(payload.Data) > 0 {
		return payload.Data
	}

	// fallback: Meta.options
	if payload.Meta != nil {
		if options, ok := payload.Meta["options"].([]any); ok {
			result := make([]map[string]any, 0, len(options))
			for _, opt := range options {
				if optMap, ok := opt.(map[string]any); ok {
					result = append(result, optMap)
				}
			}
			return result
		}
	}

	return nil
}

// getOptionLabel 获取选项的显示标签
func (h *Handler) getOptionLabel(option map[string]any, index int) string {
	// 尝试获取 label 字段
	if label, ok := option["label"].(string); ok {
		return label
	}
	// 尝试获取 name 字段
	if name, ok := option["name"].(string); ok {
		return name
	}
	// 尝试获取 title 字段
	if title, ok := option["title"].(string); ok {
		return title
	}
	// 尝试获取 value 字段
	if value, ok := option["value"].(string); ok {
		return value
	}
	// 默认显示选项编号
	return fmt.Sprintf("选项 %d", index+1)
}

// getOptionValue 获取选项的 value 字段（用于 decision）
func (h *Handler) getOptionValue(option map[string]any) string {
	if value, ok := option["value"].(string); ok {
		return value
	}
	return ""
}

// extractSource 从交互负载中提取 source 信息
// 优先从 userAck.source / userInput.source 提取，fallback 到 meta
func (h *Handler) extractSource(payload *types.ItemInteractivePayload) map[string]any {
	if payload.UserAck != nil {
		src := *payload.UserAck
		if source, ok := src["source"].(map[string]any); ok {
			return source
		}
	}
	if payload.UserInput != nil {
		src := *payload.UserInput
		if source, ok := src["source"].(map[string]any); ok {
			return source
		}
	}
	if payload.Meta != nil {
		if source, ok := payload.Meta["source"].(map[string]any); ok {
			return source
		}
	}
	return nil
}

// extractData 从交互负载中提取 data 信息（用于 modifiedData）
// userAck: userAck.data; userInput: 无(使用 formData 替代)
func (h *Handler) extractData(payload *types.ItemInteractivePayload) map[string]any {
	if payload.UserAck != nil {
		src := *payload.UserAck
		if data, ok := src["data"].(map[string]any); ok {
			return data
		}
	}
	if payload.Meta != nil {
		if data, ok := payload.Meta["data"].(map[string]any); ok {
			return data
		}
	}
	return nil
}

// =================================================================================
// formSpec 辅助方法 (user_input 表单模式)
// =================================================================================

// extractFormSpec 从 userInput 负载中提取 formSpec
func (h *Handler) extractFormSpec(payload *types.ItemInteractivePayload) map[string]any {
	if payload.UserInput != nil {
		src := *payload.UserInput
		if formSpec, ok := src["formSpec"].(map[string]any); ok {
			return formSpec
		}
	}
	return nil
}

// getFormElements 从 formSpec 中提取 ui_schema.elements
func (h *Handler) getFormElements(formSpec map[string]any) []map[string]any {
	if formSpec == nil {
		return nil
	}
	uiSchema, ok := formSpec["ui_schema"].(map[string]any)
	if !ok {
		return nil
	}
	elements, ok := uiSchema["elements"].([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(elements))
	for _, elem := range elements {
		if m, ok := elem.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

// getFormInitialValues 从 formSpec 中提取 initialValues
func (h *Handler) getFormInitialValues(formSpec map[string]any) map[string]any {
	if formSpec == nil {
		return nil
	}
	initialValues, ok := formSpec["initialValues"].(map[string]any)
	if !ok {
		return nil
	}
	return initialValues
}

// getFieldKey 从元素中提取 field 键
func (h *Handler) getFieldKey(elem map[string]any) string {
	if field, ok := elem["field"].(string); ok {
		return field
	}
	return ""
}

// getFieldLabel 从元素中提取 label 作为显示名
func (h *Handler) getFieldLabel(elem map[string]any, field string) string {
	if label, ok := elem["label"].(string); ok {
		return label
	}
	return field
}

// getFieldWidget 从元素中提取 widget 类型
func (h *Handler) getFieldWidget(elem map[string]any) string {
	if widget, ok := elem["widget"].(string); ok {
		return widget
	}
	return "input"
}

// getFieldPlaceholder 从元素中提取 placeholder
func (h *Handler) getFieldPlaceholder(elem map[string]any) string {
	if placeholder, ok := elem["placeholder"].(string); ok {
		return placeholder
	}
	return ""
}

// getInitialValue 从 initialValues 中获取字段的默认值
func (h *Handler) getInitialValue(initialValues map[string]any, field string) string {
	if initialValues == nil {
		return ""
	}
	if val, ok := initialValues[field]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// getFieldEnum 从 formSpec.schema.properties[field].enum 获取枚举选项
func (h *Handler) getFieldEnum(formSpec map[string]any, field string) []string {
	if formSpec == nil {
		return nil
	}
	schema, ok := formSpec["schema"].(map[string]any)
	if !ok {
		return nil
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil
	}
	prop, ok := properties[field].(map[string]any)
	if !ok {
		return nil
	}
	enumVals, ok := prop["enum"].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(enumVals))
	for _, v := range enumVals {
		result = append(result, fmt.Sprintf("%v", v))
	}
	return result
}
