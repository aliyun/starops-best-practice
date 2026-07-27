package interactive

import (
	"testing"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// ---------------------------------------------------------------------------
// parseInteractivePayload
// ---------------------------------------------------------------------------

func TestParseInteractivePayload_Nil(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	_, err := handler.parseInteractivePayload(nil)
	if err == nil {
		t.Error("expected error for nil payload")
	}
}

func TestParseInteractivePayload_DirectPointer(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	original := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserAck,
	}
	result, err := handler.parseInteractivePayload(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != original {
		t.Error("expected same pointer back for direct type assertion")
	}
	if result.InteractiveType != types.InteractionTypeUserAck {
		t.Errorf("expected type user_ack, got '%s'", result.InteractiveType)
	}
}

func TestParseInteractivePayload_MapDeserialization(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	rawPayload := map[string]any{
		"type": "user_ack",
		"meta": map[string]any{
			"title": "Test Title",
		},
	}
	result, err := handler.parseInteractivePayload(rawPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InteractiveType != types.InteractionTypeUserAck {
		t.Errorf("expected type user_ack, got '%s'", result.InteractiveType)
	}
	if result.Meta == nil {
		t.Fatal("expected meta to be parsed")
	}
	if result.Meta["title"] != "Test Title" {
		t.Errorf("expected meta title 'Test Title', got '%v'", result.Meta["title"])
	}
}

// ---------------------------------------------------------------------------
// getTitle
// ---------------------------------------------------------------------------

func TestGetTitle_FromUserAckData(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"data": map[string]any{
			"title": "Ack Title",
		},
	})
	payload := &types.ItemInteractivePayload{
		UserAck: &ackData,
	}
	title := handler.getTitle(payload)
	if title != "Ack Title" {
		t.Errorf("expected 'Ack Title', got '%s'", title)
	}
}

func TestGetTitle_FromMeta(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{
		Meta: map[string]any{
			"title": "Meta Title",
		},
	}
	title := handler.getTitle(payload)
	if title != "Meta Title" {
		t.Errorf("expected 'Meta Title', got '%s'", title)
	}
}

func TestGetTitle_Absent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{}
	title := handler.getTitle(payload)
	if title != "" {
		t.Errorf("expected empty string, got '%s'", title)
	}
}

// ---------------------------------------------------------------------------
// getDescription
// ---------------------------------------------------------------------------

func TestGetDescription_FromUserAckMessage(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"message": "Please confirm this action",
	})
	payload := &types.ItemInteractivePayload{
		UserAck: &ackData,
	}
	desc := handler.getDescription(payload)
	if desc != "Please confirm this action" {
		t.Errorf("expected 'Please confirm this action', got '%s'", desc)
	}
}

func TestGetDescription_FromMeta(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{
		Meta: map[string]any{
			"description": "Meta description",
		},
	}
	desc := handler.getDescription(payload)
	if desc != "Meta description" {
		t.Errorf("expected 'Meta description', got '%s'", desc)
	}
}

func TestGetDescription_Absent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{}
	desc := handler.getDescription(payload)
	if desc != "" {
		t.Errorf("expected empty string, got '%s'", desc)
	}
}

// ---------------------------------------------------------------------------
// getOptions
// ---------------------------------------------------------------------------

func TestGetOptions_FromData(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{
		Data: []map[string]any{
			{"label": "A", "value": "a"},
			{"label": "B", "value": "b"},
		},
	}
	options := handler.getOptions(payload)
	if len(options) != 2 {
		t.Errorf("expected 2 options, got %d", len(options))
	}
}

func TestGetOptions_FromMetaOptions(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{
		Meta: map[string]any{
			"options": []any{
				map[string]any{"label": "X", "value": "x"},
				map[string]any{"label": "Y", "value": "y"},
			},
		},
	}
	options := handler.getOptions(payload)
	if len(options) != 2 {
		t.Errorf("expected 2 options, got %d", len(options))
	}
}

func TestGetOptions_Absent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{}
	options := handler.getOptions(payload)
	if options != nil {
		t.Errorf("expected nil, got %v", options)
	}
}

// ---------------------------------------------------------------------------
// getOptionLabel
// ---------------------------------------------------------------------------

func TestGetOptionLabel_LabelField(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	opt := map[string]any{"label": "My Label"}
	label := handler.getOptionLabel(opt, 0)
	if label != "My Label" {
		t.Errorf("expected 'My Label', got '%s'", label)
	}
}

func TestGetOptionLabel_FallbackToName(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	opt := map[string]any{"name": "My Name"}
	label := handler.getOptionLabel(opt, 0)
	if label != "My Name" {
		t.Errorf("expected 'My Name', got '%s'", label)
	}
}

func TestGetOptionLabel_FallbackToDefault(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	opt := map[string]any{"someKey": "someValue"}
	label := handler.getOptionLabel(opt, 2)
	if label != "选项 3" {
		t.Errorf("expected '选项 3', got '%s'", label)
	}
}

// ---------------------------------------------------------------------------
// getOptionValue
// ---------------------------------------------------------------------------

func TestGetOptionValue_Present(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	opt := map[string]any{"value": "opt_val"}
	val := handler.getOptionValue(opt)
	if val != "opt_val" {
		t.Errorf("expected 'opt_val', got '%s'", val)
	}
}

func TestGetOptionValue_Absent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	opt := map[string]any{"label": "no value"}
	val := handler.getOptionValue(opt)
	if val != "" {
		t.Errorf("expected empty string, got '%s'", val)
	}
}

// ---------------------------------------------------------------------------
// extractSource
// ---------------------------------------------------------------------------

func TestExtractSource_FromUserAck(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"source": map[string]any{
			"app": "test-app",
		},
	})
	payload := &types.ItemInteractivePayload{
		UserAck: &ackData,
	}
	source := handler.extractSource(payload)
	if source == nil {
		t.Fatal("expected source to be non-nil")
	}
	if source["app"] != "test-app" {
		t.Errorf("expected app 'test-app', got '%v'", source["app"])
	}
}

func TestExtractSource_Absent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{}
	source := handler.extractSource(payload)
	if source != nil {
		t.Errorf("expected nil, got %v", source)
	}
}

// ---------------------------------------------------------------------------
// extractData
// ---------------------------------------------------------------------------

func TestExtractData_FromUserAck(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"data": map[string]any{
			"key1": "value1",
		},
	})
	payload := &types.ItemInteractivePayload{
		UserAck: &ackData,
	}
	data := handler.extractData(payload)
	if data == nil {
		t.Fatal("expected data to be non-nil")
	}
	if data["key1"] != "value1" {
		t.Errorf("expected key1 'value1', got '%v'", data["key1"])
	}
}

func TestExtractData_Absent(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	payload := &types.ItemInteractivePayload{}
	data := handler.extractData(payload)
	if data != nil {
		t.Errorf("expected nil, got %v", data)
	}
}
