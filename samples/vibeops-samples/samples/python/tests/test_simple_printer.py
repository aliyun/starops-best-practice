"""
Tests for SimplePrinter - standalone tests without SDK dependency
"""

import pytest
from typing import Any, Dict, Optional, Set


# Copy of SimplePrinter for testing without SDK dependency
class SimplePrinter:
    """简洁模式打印器 / Simple mode printer"""

    def __init__(self):
        self._buffer: list = []
        self._seen_artifacts: Set[str] = set()

    def process_event(self, event: Optional[Any]) -> str:
        """处理事件，提取文本内容 / Process event and extract text content"""
        if not event:
            return ""
        
        body = getattr(event, 'body', None)
        if not body:
            return ""

        extracted = []
        messages = body.get("messages", [])

        for msg in messages:
            if not isinstance(msg, dict):
                continue

            # Only process system role messages
            if msg.get("role") == "system":
                text = self._extract_text_from_artifacts(msg.get("artifacts", []))
                if text and text not in self._seen_artifacts:
                    self._seen_artifacts.add(text)
                    extracted.append(text)
                    self._buffer.append(text)

        return "".join(extracted)

    def _extract_text_from_artifacts(self, artifacts: list) -> str:
        """从 artifacts 中提取文本 / Extract text from artifacts"""
        if not artifacts:
            return ""

        result = []
        for artifact in artifacts:
            if not isinstance(artifact, dict):
                continue

            parts = artifact.get("parts", [])
            for part in parts:
                if not isinstance(part, dict):
                    continue

                if part.get("kind") == "text" and part.get("text"):
                    result.append(part["text"])

        return "".join(result)

    def get_final_text(self) -> str:
        """获取最终文本 / Get final text"""
        return "".join(self._buffer)

    def reset(self) -> None:
        """重置缓冲区 / Reset buffer"""
        self._buffer.clear()
        self._seen_artifacts.clear()


# Mock ChatEvent for testing
class MockChatEvent:
    def __init__(self, body=None, raw_json="", status_code=200, is_done=False, error=None):
        self.body = body
        self.raw_json = raw_json
        self.status_code = status_code
        self.is_done = is_done
        self.error = error

    def has_error(self):
        return self.error is not None


class TestSimplePrinter:
    def setup_method(self):
        self.printer = SimplePrinter()

    def test_process_none_event(self):
        result = self.printer.process_event(None)
        assert result == ""

    def test_process_event_with_none_body(self):
        event = MockChatEvent()
        result = self.printer.process_event(event)
        assert result == ""

    def test_process_event_with_system_artifacts(self):
        body = {
            "messages": [
                {
                    "role": "system",
                    "artifacts": [
                        {"parts": [{"kind": "text", "text": "Hello World"}]}
                    ]
                }
            ]
        }
        event = MockChatEvent(body=body)
        result = self.printer.process_event(event)

        assert result == "Hello World"
        assert self.printer.get_final_text() == "Hello World"

    def test_process_event_ignores_assistant_role(self):
        body = {
            "messages": [
                {
                    "role": "assistant",
                    "artifacts": [
                        {"parts": [{"kind": "text", "text": "Should be ignored"}]}
                    ]
                }
            ]
        }
        event = MockChatEvent(body=body)
        result = self.printer.process_event(event)

        assert result == ""

    def test_deduplication(self):
        body = {
            "messages": [
                {
                    "role": "system",
                    "artifacts": [
                        {"parts": [{"kind": "text", "text": "Duplicate"}]}
                    ]
                }
            ]
        }
        event = MockChatEvent(body=body)

        # Process same event twice
        result1 = self.printer.process_event(event)
        result2 = self.printer.process_event(event)

        assert result1 == "Duplicate"
        assert result2 == ""  # Should be empty due to deduplication
        assert self.printer.get_final_text() == "Duplicate"

    def test_reset(self):
        body = {
            "messages": [
                {
                    "role": "system",
                    "artifacts": [
                        {"parts": [{"kind": "text", "text": "Test"}]}
                    ]
                }
            ]
        }
        event = MockChatEvent(body=body)
        self.printer.process_event(event)

        assert self.printer.get_final_text() == "Test"

        self.printer.reset()

        assert self.printer.get_final_text() == ""

        # After reset, same content should be processed again
        result = self.printer.process_event(event)
        assert result == "Test"

    def test_multiple_artifacts(self):
        body = {
            "messages": [
                {
                    "role": "system",
                    "artifacts": [
                        {"parts": [{"kind": "text", "text": "Part1"}]},
                        {"parts": [{"kind": "text", "text": "Part2"}]}
                    ]
                }
            ]
        }
        event = MockChatEvent(body=body)
        result = self.printer.process_event(event)

        assert result == "Part1Part2"

    def test_ignores_non_text_kind(self):
        body = {
            "messages": [
                {
                    "role": "system",
                    "artifacts": [
                        {"parts": [{"kind": "image", "url": "http://example.com"}]}
                    ]
                }
            ]
        }
        event = MockChatEvent(body=body)
        result = self.printer.process_event(event)

        assert result == ""


# ---------------------------------------------------------------------------
# 以下测试针对真实 SimplePrinter 的 event 字段快速跳过逻辑
# ---------------------------------------------------------------------------
from types import SimpleNamespace
import importlib.util
import os

# 直接加载 simple_printer 模块，绕过包 __init__.py 中对 SDK 的依赖
_sp_path = os.path.join(
    os.path.dirname(__file__), os.pardir,
    "starops_sdk_samples", "client", "simple_printer.py",
)
_spec = importlib.util.spec_from_file_location("simple_printer", _sp_path)
_mod = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_mod)
RealSimplePrinter = _mod.SimplePrinter


def _make_body_with_text(text: str) -> dict:
    """构造包含可提取文本的 body"""
    return {
        "messages": [
            {
                "role": "system",
                "artifacts": [
                    {"parts": [{"kind": "text", "text": text}]}
                ],
            }
        ]
    }


class TestSimplePrinterEventFieldFilter:
    """测试 event 字段快速跳过逻辑（lines 27-29）"""

    def setup_method(self):
        self.printer = RealSimplePrinter()

    def test_non_text_event_type_returns_empty(self):
        """event='interaction' 等非 text/task_finished 事件应直接返回空字符串"""
        event = SimpleNamespace(
            event="interaction",
            body=_make_body_with_text("应该被跳过的内容"),
        )
        result = self.printer.process_event(event)
        assert result == ""
        assert self.printer.get_final_text() == ""

    def test_text_event_type_processes_normally(self):
        """event='text' 的事件应正常提取内容"""
        event = SimpleNamespace(
            event="text",
            body=_make_body_with_text("正常文本"),
        )
        result = self.printer.process_event(event)
        assert result == "正常文本"
        assert self.printer.get_final_text() == "正常文本"

    def test_task_finished_event_type_processes_normally(self):
        """event='task_finished' 的事件应正常提取内容"""
        event = SimpleNamespace(
            event="task_finished",
            body=_make_body_with_text("任务完成"),
        )
        result = self.printer.process_event(event)
        assert result == "任务完成"

    def test_no_event_attribute_processes_normally(self):
        """没有 event 属性的事件应正常处理（向后兼容）"""
        event = SimpleNamespace(
            body=_make_body_with_text("无 event 字段"),
        )
        result = self.printer.process_event(event)
        assert result == "无 event 字段"

    def test_various_skipped_event_types(self):
        """多种非文本事件类型均应被跳过"""
        for event_type in ("interaction", "status", "error", "progress", "unknown"):
            printer = RealSimplePrinter()
            event = SimpleNamespace(
                event=event_type,
                body=_make_body_with_text(f"内容-{event_type}"),
            )
            result = printer.process_event(event)
            assert result == "", f"event='{event_type}' 应返回空字符串"
