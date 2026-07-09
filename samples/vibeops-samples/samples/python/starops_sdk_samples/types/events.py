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
