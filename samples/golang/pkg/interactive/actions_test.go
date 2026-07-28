package interactive

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/starops-best-practice/samples/golang/pkg/types"
)

// ---------------------------------------------------------------------------
// HandleUserAck
// ---------------------------------------------------------------------------

func TestHandleUserAck_NilPayload(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	_, err := handler.HandleUserAck(context.Background(), nil, "fake-call-id")
	if err == nil {
		t.Error("expected error for nil payload")
	}
}

func TestHandleUserAck_ConfirmWithY(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("y\n"), &bytes.Buffer{})

	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"data":    map[string]any{"title": "Confirm deploy"},
		"message": "Deploy to production?",
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserAck,
		UserAck:         &ackData,
	}

	resp, err := handler.HandleUserAck(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Decision != "yes" {
		t.Errorf("expected decision 'yes', got '%s'", resp.Decision)
	}
	if resp.Type != types.InteractionTypeUserAck {
		t.Errorf("expected type InteractionTypeUserAck, got '%s'", resp.Type)
	}
}

func TestHandleUserAck_RejectWithN(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("n\n"), &bytes.Buffer{})

	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"data":    map[string]any{"title": "Confirm deploy"},
		"message": "Deploy to production?",
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserAck,
		UserAck:         &ackData,
	}

	resp, err := handler.HandleUserAck(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Decision != "no" {
		t.Errorf("expected decision 'no', got '%s'", resp.Decision)
	}
}

func TestHandleUserAck_CallIdPreserved(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetMockInput("y")

	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"data": map[string]any{"title": "test"},
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserAck,
		UserAck:         &ackData,
	}

	resp, err := handler.HandleUserAck(context.Background(), payload, "test-call-id-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CallId != "test-call-id-123" {
		t.Errorf("expected callId 'test-call-id-123', got '%s'", resp.CallId)
	}
}

func TestHandleUserAck_ResponseType(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetMockInput("y")

	ackData := types.ItemInteractiveUserAckPayload(map[string]any{
		"data": map[string]any{"title": "test"},
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserAck,
		UserAck:         &ackData,
	}

	resp, err := handler.HandleUserAck(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != types.InteractionTypeUserAck {
		t.Errorf("expected type InteractionTypeUserAck, got '%s'", resp.Type)
	}
}

// ---------------------------------------------------------------------------
// HandleUserSelect
// ---------------------------------------------------------------------------

func TestHandleUserSelect_NilPayload(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	_, err := handler.HandleUserSelect(context.Background(), nil, "fake-call-id")
	if err == nil {
		t.Error("expected error for nil payload")
	}
}

func TestHandleUserSelect_ValidSelection(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("1\n"), &bytes.Buffer{})

	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserSelect,
		Data: []map[string]any{
			{"label": "Option A", "value": "opt_a"},
			{"label": "Option B", "value": "opt_b"},
		},
	}

	resp, err := handler.HandleUserSelect(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Decision != "opt_a" {
		t.Errorf("expected decision 'opt_a', got '%s'", resp.Decision)
	}
	if resp.Type != types.InteractionTypeUserSelect {
		t.Errorf("expected type InteractionTypeUserSelect, got '%s'", resp.Type)
	}
}

func TestHandleUserSelect_InvalidZero(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("0\n"), &bytes.Buffer{})

	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserSelect,
		Data: []map[string]any{
			{"label": "Option A", "value": "opt_a"},
		},
	}

	_, err := handler.HandleUserSelect(context.Background(), payload, "fake-call-id")
	if err == nil {
		t.Error("expected error for selection '0'")
	}
}

func TestHandleUserSelect_OutOfRange(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("5\n"), &bytes.Buffer{})

	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserSelect,
		Data: []map[string]any{
			{"label": "Option A", "value": "opt_a"},
			{"label": "Option B", "value": "opt_b"},
		},
	}

	_, err := handler.HandleUserSelect(context.Background(), payload, "fake-call-id")
	if err == nil {
		t.Error("expected error for out-of-range selection")
	}
}

// ---------------------------------------------------------------------------
// HandleUserInput
// ---------------------------------------------------------------------------

func TestHandleUserInput_NilPayload(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	_, err := handler.HandleUserInput(context.Background(), nil, "fake-call-id")
	if err == nil {
		t.Error("expected error for nil payload")
	}
}

func TestHandleUserInput_SingleTextField(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("test-employee\n"), &bytes.Buffer{})

	userInputData := types.ItemInteractiveUserInputPayload(map[string]any{
		"title":       "Enter details",
		"description": "Fill in the form",
		"formSpec": map[string]any{
			"schema": map[string]any{
				"properties": map[string]any{
					"username": map[string]any{
						"type": "string",
					},
				},
			},
			"ui_schema": map[string]any{
				"elements": []any{
					map[string]any{
						"field":       "username",
						"label":       "Username",
						"widget":      "input",
						"placeholder": "Enter your username",
					},
				},
			},
		},
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserInput,
		UserInput:       &userInputData,
	}

	resp, err := handler.HandleUserInput(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FormData == nil {
		t.Fatal("expected formData to be populated")
	}
	if resp.FormData["username"] != "test-employee" {
		t.Errorf("expected username 'test-employee', got '%v'", resp.FormData["username"])
	}
}

func TestHandleUserInput_ResponseType(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetMockInput("some-value")

	userInputData := types.ItemInteractiveUserInputPayload(map[string]any{
		"title": "Input form",
		"formSpec": map[string]any{
			"schema": map[string]any{
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
			"ui_schema": map[string]any{
				"elements": []any{
					map[string]any{
						"field": "name",
						"label": "Name",
					},
				},
			},
		},
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserInput,
		UserInput:       &userInputData,
	}

	resp, err := handler.HandleUserInput(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != types.InteractionTypeUserInput {
		t.Errorf("expected type InteractionTypeUserInput, got '%s'", resp.Type)
	}
}

func TestHandleUserInput_DecisionIsSubmit(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetMockInput("value")

	userInputData := types.ItemInteractiveUserInputPayload(map[string]any{
		"title": "Form",
		"formSpec": map[string]any{
			"schema": map[string]any{
				"properties": map[string]any{
					"field1": map[string]any{"type": "string"},
				},
			},
			"ui_schema": map[string]any{
				"elements": []any{
					map[string]any{
						"field": "field1",
						"label": "Field 1",
					},
				},
			},
		},
	})
	payload := &types.ItemInteractivePayload{
		InteractiveType: types.InteractionTypeUserInput,
		UserInput:       &userInputData,
	}

	resp, err := handler.HandleUserInput(context.Background(), payload, "fake-call-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Decision != "submit" {
		t.Errorf("expected decision 'submit', got '%s'", resp.Decision)
	}
}
