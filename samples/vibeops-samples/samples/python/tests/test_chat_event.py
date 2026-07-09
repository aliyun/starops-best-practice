"""ChatEvent 测试"""
from starops_sdk_samples.client.agent_client import ChatEvent


class TestChatEvent:
    def test_from_response_with_event_field_done(self):
        """event 字段为 done 时 is_done 应为 True"""
        body = {"event": "done", "messages": []}
        event = ChatEvent.from_response(body, '{"event":"done"}', 200)
        assert event.is_done is True

    def test_from_response_with_messages_done_fallback(self):
        """无 event 字段时 fallback 到 messages[].type"""
        body = {"messages": [{"type": "done"}]}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.is_done is True

    def test_from_response_not_done(self):
        """非 done 事件"""
        body = {"event": "text", "messages": [{"type": "text"}]}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.is_done is False

    def test_from_response_extracts_id_and_event(self):
        """提取 id 和 event 字段"""
        body = {"id": "evt-123", "event": "text", "messages": []}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.id == "evt-123"
        assert event.event == "text"

    def test_from_response_missing_id_event(self):
        """id/event 缺失时为 None"""
        body = {"messages": []}
        event = ChatEvent.from_response(body, '{}', 200)
        assert event.id is None
        assert event.event is None
