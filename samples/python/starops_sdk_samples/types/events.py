"""
Event and message types for STAROps SDK
STAROps SDK 事件和消息类型
"""

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

from .enums import ContentType, EventType, ItemStatus, MessageRole


@dataclass
class ItemContent:
    """消息内容 / Message content"""
    type: ContentType
    value: str = ""
    append: bool = False
    last_chunk: bool = False


@dataclass
class ItemEvent:
    """事件定义 / Event definition"""
    type: EventType
    payload: Optional[Dict[str, Any]] = None


@dataclass
class ItemTool:
    """工具调用详情 / Tool call details"""
    id: str = ""
    name: str = ""
    tool_call_id: str = ""
    arguments_delta: str = ""
    arguments: Optional[Any] = None
    status: Optional[ItemStatus] = None
    contents: List[ItemContent] = field(default_factory=list)


@dataclass
class ItemAgent:
    """子 Agent 调用详情 / Sub agent call details"""
    id: str = ""
    name: str = ""
    call_id: str = ""
    status: Optional[ItemStatus] = None
    inputs: List[ItemContent] = field(default_factory=list)
    results: List[ItemContent] = field(default_factory=list)


@dataclass
class MessageItem:
    """消息条目 / Message item"""
    parent_call_id: str = ""
    call_id: str = ""
    role: Optional[MessageRole] = None
    timestamp: str = ""
    contents: List[ItemContent] = field(default_factory=list)
    tools: List[ItemTool] = field(default_factory=list)
    agents: List[ItemAgent] = field(default_factory=list)
    events: List[ItemEvent] = field(default_factory=list)
    artifacts: List[Dict[str, Any]] = field(default_factory=list)


# ===================== ChatEvent / ThreadInfo / ThreadMessage =====================
# These are the top-level SSE event DTO and thread DTOs, moved from client/agent_client.py


@dataclass
class ChatEvent:
    """聊天事件 / Chat event"""
    body: Optional[Dict[str, Any]] = None
    raw_json: str = ""
    status_code: int = 0
    is_done: bool = False
    error: Optional[Exception] = None
    id: Optional[str] = None
    event: Optional[str] = None

    @classmethod
    def done(cls) -> "ChatEvent":
        return cls(is_done=True)

    @classmethod
    def from_error(cls, error: Exception) -> "ChatEvent":
        return cls(error=error)

    @classmethod
    def from_response(cls, body: Dict[str, Any], raw_json: str, status_code: int) -> "ChatEvent":
        is_done = cls._is_done_message(body)
        return cls(
            body=body,
            raw_json=raw_json,
            status_code=status_code,
            is_done=is_done,
            id=body.get("id") if body else None,
            event=body.get("event") if body else None,
        )

    @staticmethod
    def _is_done_message(body: Optional[Dict[str, Any]]) -> bool:
        if not body:
            return False
        # 优先使用 response 级别的 event 字段
        if body.get("event") == "done":
            return True
        # fallback: 遍历 messages
        messages = body.get("messages", [])
        for msg in messages:
            if isinstance(msg, dict) and msg.get("type") == "done":
                return True
        return False

    def has_error(self) -> bool:
        return self.error is not None


@dataclass
class ThreadInfo:
    """会话信息 / Thread information"""
    thread_id: str
    title: str = ""
    status: str = ""
    create_time: str = ""
    update_time: str = ""


@dataclass
class ThreadMessage:
    """会话消息 / Thread message"""
    role: str
    content: str
    timestamp: str = ""
