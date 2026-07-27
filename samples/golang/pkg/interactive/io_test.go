package interactive

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// readInputWithTimeout
// ---------------------------------------------------------------------------

func TestReadInputWithTimeout_MockInput(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetMockInput("test")

	input, err := handler.readInputWithTimeout(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input != "test" {
		t.Errorf("expected 'test', got '%s'", input)
	}
}

func TestReadInputWithTimeout_MockInputConsumedOnce(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("fallback\n"), &bytes.Buffer{})
	handler.SetMockInput("first-call")

	// First call returns mock input
	input1, err := handler.readInputWithTimeout(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if input1 != "first-call" {
		t.Errorf("expected 'first-call', got '%s'", input1)
	}

	// Second call reads from reader
	input2, err := handler.readInputWithTimeout(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if input2 != "fallback\n" {
		t.Errorf("expected 'fallback\\n', got '%s'", input2)
	}
}

func TestReadInputWithTimeout_NormalReader(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	handler.SetIO(strings.NewReader("hello\n"), &bytes.Buffer{})

	input, err := handler.readInputWithTimeout(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input != "hello\n" {
		t.Errorf("expected 'hello\\n', got '%s'", input)
	}
}

func TestReadInputWithTimeout_ContextCancelled(t *testing.T) {
	handler := NewHandler(nil, 0) // no timeout, rely on context
	// Use a reader that blocks forever (never provides data)
	r, _ := blockedReader()
	handler.SetIO(r, &bytes.Buffer{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Give the goroutine a moment to notice the cancellation
	_, err := handler.readInputWithTimeout(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// blockedReader returns a reader that blocks forever on Read.
func blockedReader() (*blockingReader, chan struct{}) {
	done := make(chan struct{})
	return &blockingReader{done: done}, done
}

type blockingReader struct {
	done chan struct{}
}

func (r *blockingReader) Read(p []byte) (int, error) {
	<-r.done // block forever until done is closed
	return 0, context.Canceled
}

// ---------------------------------------------------------------------------
// printf
// ---------------------------------------------------------------------------

func TestPrintf_WritesToBuffer(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	buf := &bytes.Buffer{}
	handler.SetIO(strings.NewReader(""), buf)

	handler.printf("Hello %s, count: %d", "test-employee", 42)

	output := buf.String()
	expected := "Hello test-employee, count: 42"
	if output != expected {
		t.Errorf("expected '%s', got '%s'", expected, output)
	}
}

func TestPrintf_EmptyFormat(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	buf := &bytes.Buffer{}
	handler.SetIO(strings.NewReader(""), buf)

	handler.printf("")

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got '%s'", buf.String())
	}
}

func TestPrintf_MultipleCallsAppend(t *testing.T) {
	handler := NewHandler(nil, 30*time.Second)
	buf := &bytes.Buffer{}
	handler.SetIO(strings.NewReader(""), buf)

	handler.printf("line1\n")
	handler.printf("line2\n")

	expected := "line1\nline2\n"
	if buf.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, buf.String())
	}
}
