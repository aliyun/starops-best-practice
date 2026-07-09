"""
Simple printer for STAROps SDK
STAROps SDK 简洁打印器
"""

from typing import Any, Dict, Optional, Set, Protocol


class ChatEventProtocol(Protocol):
    """Protocol for ChatEvent to avoid circular imports"""
    body: Optional[Dict[str, Any]]


class SimplePrinter:
    """简洁模式打印器 / Simple mode printer"""

    def __init__(self):
        self._buffer: list[str] = []
        self._seen_artifacts: Set[str] = set()

    def process_event(self, event: Optional[Any]) -> str:
        """处理事件，提取文本内容 / Process event and extract text content"""
        if not event:
            return ""

        # 利用 event 字段快速跳过非文本事件
        event_type = getattr(event, 'event', None)
        if event_type and event_type not in ('text', 'task_finished'):
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
