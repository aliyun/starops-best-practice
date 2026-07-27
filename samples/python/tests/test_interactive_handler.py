"""
Tests for InteractiveHandler — 交互事件识别与 user_ack / user_input happy path
参考 golang/pkg/interactive/interactive_test.go 的测试思路，
使用简化事件对象与 mock 客户端，不依赖真实 SDK。
"""

import io
import json

import pytest

from starops_sdk_samples.client.interactive_handler import (
    InteractiveHandler,
    InteractiveResponse,
)
from starops_sdk_samples.types import InteractionType
from starops_sdk_samples.types.events import ChatEvent


class MockChatClient:
    """模拟 ChatClientProtocol，仅记录 interact 调用"""

    def __init__(self):
        self.interact_called = False
        self.last_thread_id = None
        self.last_user_interactive = None

    async def interact(self, thread_id, user_interactive, base_variables=None):
        self.interact_called = True
        self.last_thread_id = thread_id
        self.last_user_interactive = user_interactive
        yield ChatEvent.done()


def _make_handler(input_text: str = "") -> InteractiveHandler:
    return InteractiveHandler(
        client=MockChatClient(),
        reader=io.StringIO(input_text),
        writer=io.StringIO(),
    )


# ---------------------------------------------------------------------------
# is_interactive_event
# ---------------------------------------------------------------------------

class TestIsInteractiveEvent:
    def test_none_event(self):
        assert InteractiveHandler.is_interactive_event(None) is False

    def test_interactive_event(self):
        assert InteractiveHandler.is_interactive_event({"type": "interactive"}) is True

    def test_thinking_event(self):
        assert InteractiveHandler.is_interactive_event({"type": "thinking"}) is False


# ---------------------------------------------------------------------------
# extract_interactive_events
# ---------------------------------------------------------------------------

class TestExtractInteractiveEvents:
    def test_none_message(self):
        assert InteractiveHandler.extract_interactive_events(None) == []

    def test_mixed_events(self):
        message = {
            "events": [
                {"type": "thinking"},
                {"type": "interactive"},
                {"type": "error"},
                {"type": "interactive"},
            ]
        }
        result = InteractiveHandler.extract_interactive_events(message)
        assert len(result) == 2
        assert all(e["type"] == "interactive" for e in result)

    def test_no_interactive(self):
        message = {"events": [{"type": "thinking"}, {"type": "error"}]}
        assert InteractiveHandler.extract_interactive_events(message) == []


# ---------------------------------------------------------------------------
# user_ack happy path
# ---------------------------------------------------------------------------

class TestUserAck:
    def test_handle_user_ack_confirm(self):
        handler = _make_handler("y\n")
        event = {
            "type": "interactive",
            "payload": {
                "type": "user_ack",
                "userAck": {
                    "message": "是否继续执行？",
                    "source": {"app": "test-app"},
                    "data": {"title": "确认请求"},
                },
            },
        }

        resp = handler.handle_event(event, "call-001")
        assert isinstance(resp, InteractiveResponse)
        assert resp.call_id == "call-001"
        assert resp.type == InteractionType.USER_ACK
        assert resp.response == {"confirmed": True}
        assert resp.decision == "yes"
        assert resp.source == {"app": "test-app"}

    def test_handle_user_ack_reject(self):
        handler = _make_handler("n\n")
        event = {
            "type": "interactive",
            "payload": {
                "type": "user_ack",
                "userAck": {"message": "是否继续执行？"},
            },
        }

        resp = handler.handle_event(event, "call-002")
        assert resp.response == {"confirmed": False}
        assert resp.decision == "no"


# ---------------------------------------------------------------------------
# user_input happy path
# ---------------------------------------------------------------------------

class TestUserInput:
    def test_handle_user_input_form(self):
        handler = _make_handler("my-cluster\n")
        event = {
            "type": "interactive",
            "payload": {
                "type": "user_input",
                "userInput": {
                    "title": "填写集群信息",
                    "description": "请输入目标集群",
                    "source": {"app": "test-app"},
                    "formSpec": {
                        "ui_schema": {
                            "elements": [
                                {"field": "cluster", "label": "集群名称", "widget": "input"}
                            ]
                        },
                        "initialValues": {"cluster": "default-cluster"},
                    },
                },
            },
        }

        resp = handler.handle_event(event, "call-003")
        assert resp.type == InteractionType.USER_INPUT
        assert resp.decision == "submit"
        assert resp.form_data == {"cluster": "my-cluster"}
        assert resp.response == {"value": {"cluster": "my-cluster"}}

    def test_handle_user_input_default_value(self):
        handler = _make_handler("\n")  # 空输入使用默认值
        event = {
            "type": "interactive",
            "payload": {
                "type": "user_input",
                "userInput": {
                    "title": "填写集群信息",
                    "formSpec": {
                        "ui_schema": {
                            "elements": [
                                {"field": "cluster", "label": "集群名称", "widget": "input"}
                            ]
                        },
                        "initialValues": {"cluster": "default-cluster"},
                    },
                },
            },
        }

        resp = handler.handle_event(event, "call-004")
        assert resp.form_data == {"cluster": "default-cluster"}


# ---------------------------------------------------------------------------
# handle_event 边界
# ---------------------------------------------------------------------------

class TestHandleEventErrors:
    def test_empty_event(self):
        handler = _make_handler()
        with pytest.raises(Exception, match="事件为空"):
            handler.handle_event(None, "call-x")

    def test_unsupported_event_type(self):
        handler = _make_handler()
        with pytest.raises(Exception, match="不支持的事件类型"):
            handler.handle_event({"type": "thinking"}, "call-x")


# ---------------------------------------------------------------------------
# resume_chat
# ---------------------------------------------------------------------------

class TestResumeChat:
    async def test_resume_chat_valid_response(self):
        client = MockChatClient()
        handler = InteractiveHandler(
            client=client, reader=io.StringIO(), writer=io.StringIO()
        )
        resp = InteractiveResponse(
            call_id="fake-call-id",
            type=InteractionType.USER_ACK,
            response={"confirmed": True},
            source={"app": "test-app"},
            decision="yes",
        )

        events = []
        async for event in handler.resume_chat("fake-thread-id", resp):
            events.append(event)

        assert client.interact_called is True
        assert client.last_thread_id == "fake-thread-id"
        payload = json.loads(client.last_user_interactive)
        assert payload["callId"] == "fake-call-id"
        assert payload["decision"] == "yes"
        assert len(events) == 1
        assert events[0].is_done is True

    async def test_resume_chat_nil_response(self):
        client = MockChatClient()
        handler = InteractiveHandler(
            client=client, reader=io.StringIO(), writer=io.StringIO()
        )

        events = []
        async for event in handler.resume_chat("fake-thread-id", None):
            events.append(event)

        assert len(events) == 1
        assert events[0].error is not None
        assert client.interact_called is False

    async def test_resume_chat_nil_client(self):
        handler = InteractiveHandler(
            client=None, reader=io.StringIO(), writer=io.StringIO()
        )
        resp = InteractiveResponse(
            call_id="fake-call-id",
            type=InteractionType.USER_ACK,
            response={"confirmed": True},
            decision="yes",
        )

        events = []
        async for event in handler.resume_chat("fake-thread-id", resp):
            events.append(event)

        assert len(events) == 1
        assert events[0].error is not None
