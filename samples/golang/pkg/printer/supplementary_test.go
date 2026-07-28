package printer

import (
	"testing"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// ============================================================
// GetFinalText / Reset 测试
// ============================================================

func makeBodyWithArtifacts(role types.MessageRole, text string) *types.ChatEventBody {
	return &types.ChatEventBody{
		Messages: []*types.MessageItem{
			{
				Role: role,
				Artifacts: []map[string]any{
					{
						"parts": []any{
							map[string]any{"kind": "text", "text": text},
						},
					},
				},
			},
		},
	}
}

func TestSimplePrinter_GetFinalText_Accumulates(t *testing.T) {
	p := NewSimplePrinter()

	makeTextEvent := func(text string) *types.ChatEvent {
		return &types.ChatEvent{
			Event: "text",
			Body:  makeBodyWithArtifacts(types.MessageItemRoleSystem, text),
		}
	}

	p.ProcessEvent(makeTextEvent("Hello "))
	p.ProcessEvent(makeTextEvent("World"))

	if got := p.GetFinalText(); got != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", got)
	}
}

func TestSimplePrinter_Reset(t *testing.T) {
	p := NewSimplePrinter()

	evt := &types.ChatEvent{
		Event: "text",
		Body:  makeBodyWithArtifacts(types.MessageItemRoleSystem, "content"),
	}
	p.ProcessEvent(evt)
	p.Reset()

	if got := p.GetFinalText(); got != "" {
		t.Errorf("expected empty after Reset, got %q", got)
	}

	p.ProcessEvent(evt)
	if got := p.GetFinalText(); got != "content" {
		t.Errorf("expected 'content' after Reset+reprocess, got %q", got)
	}
}

func TestSimplePrinter_Deduplication(t *testing.T) {
	p := NewSimplePrinter()

	evt := &types.ChatEvent{
		Event: "text",
		Body:  makeBodyWithArtifacts(types.MessageItemRoleSystem, "重复内容"),
	}

	first := p.ProcessEvent(evt)
	second := p.ProcessEvent(evt)

	if first != "重复内容" {
		t.Errorf("first call should return text, got %q", first)
	}
	if second != "" {
		t.Errorf("duplicate should be filtered, got %q", second)
	}
}

// ============================================================
// extractTextFromArtifacts 测试
// ============================================================

func TestExtractTextFromArtifacts_Normal(t *testing.T) {
	artifacts := []map[string]any{
		{
			"parts": []any{
				map[string]any{"kind": "text", "text": "Hello "},
				map[string]any{"kind": "image", "url": "img.png"},
				map[string]any{"kind": "text", "text": "World"},
			},
		},
	}
	got := extractTextFromArtifacts(artifacts)
	if got != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", got)
	}
}

func TestExtractTextFromArtifacts_MultipleArtifacts(t *testing.T) {
	artifacts := []map[string]any{
		{"parts": []any{map[string]any{"kind": "text", "text": "A"}}},
		{"parts": []any{map[string]any{"kind": "text", "text": "B"}}},
	}
	got := extractTextFromArtifacts(artifacts)
	if got != "AB" {
		t.Errorf("expected 'AB', got %q", got)
	}
}

func TestExtractTextFromArtifacts_Empty(t *testing.T) {
	if got := extractTextFromArtifacts(nil); got != "" {
		t.Errorf("expected empty for nil, got %q", got)
	}
	artifacts := []map[string]any{nil, {"parts": nil}}
	if got := extractTextFromArtifacts(artifacts); got != "" {
		t.Errorf("expected empty for nil parts, got %q", got)
	}
}

func TestExtractTextFromArtifacts_NonTextParts(t *testing.T) {
	artifacts := []map[string]any{
		{"parts": []any{
			map[string]any{"kind": "image", "url": "img.png"},
			map[string]any{"kind": "code", "text": "ignored"},
		}},
	}
	if got := extractTextFromArtifacts(artifacts); got != "" {
		t.Errorf("expected empty for non-text parts, got %q", got)
	}
}

// ============================================================
// PrettyPrintJSON 测试
// ============================================================

func TestPrettyPrintJSON_Valid(t *testing.T) {
	input := `{"name":"test","value":123}`
	got, err := PrettyPrintJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty output")
	}
	if len(got) <= len(input) {
		t.Error("pretty printed should be longer than compact")
	}
}

func TestPrettyPrintJSON_Invalid(t *testing.T) {
	_, err := PrettyPrintJSON("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPrettyPrintJSON_Empty(t *testing.T) {
	_, err := PrettyPrintJSON("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}
