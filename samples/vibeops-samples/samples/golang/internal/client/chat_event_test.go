package client

import (
	"testing"

	starops "github.com/alibabacloud-go/starops-20260428/client"
	"github.com/alibabacloud-go/tea/dara"
)

func TestIsDoneMessage_EventField(t *testing.T) {
	resp := &starops.CreateChatResponse{}
	resp.SetEvent("done")
	if !isDoneMessage(resp) {
		t.Error("event=done 时应返回 true")
	}
}

func TestIsDoneMessage_MessagesFallback(t *testing.T) {
	msg := &starops.CreateChatResponseBodyMessages{}
	msg.SetType("done")
	body := &starops.CreateChatResponseBody{
		Messages: []*starops.CreateChatResponseBodyMessages{msg},
	}
	resp := &starops.CreateChatResponse{Body: body}
	if !isDoneMessage(resp) {
		t.Error("messages[].type=done 时应返回 true")
	}
}

func TestIsDoneMessage_NotDone(t *testing.T) {
	resp := &starops.CreateChatResponse{}
	resp.SetEvent("text")
	if isDoneMessage(resp) {
		t.Error("event=text 时应返回 false")
	}
}

func TestIsDoneMessage_NilResp(t *testing.T) {
	if isDoneMessage(nil) {
		t.Error("nil resp 时应返回 false")
	}
}

func TestIsDoneMessage_NilBody(t *testing.T) {
	resp := &starops.CreateChatResponse{}
	if isDoneMessage(resp) {
		t.Error("nil body 且无 event 时应返回 false")
	}
}

func TestChatEvent_IdAndEvent(t *testing.T) {
	resp := &starops.CreateChatResponse{}
	resp.SetEvent("text")
	resp.SetId("evt-123")
	// 验证 ChatEvent 结构体包含 Id 和 Event 字段
	event := &ChatEvent{
		Body:  resp.Body,
		Id:    dara.StringValue(resp.Id),
		Event: dara.StringValue(resp.Event),
	}
	if event.Event != "text" {
		t.Errorf("期望 event=text, 实际=%s", event.Event)
	}
	if event.Id != "evt-123" {
		t.Errorf("期望 id=evt-123, 实际=%s", event.Id)
	}
}
