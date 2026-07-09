"""
Enumeration types for STAROps SDK
STAROps SDK 枚举类型
"""

from enum import Enum


class EventType(str, Enum):
    """事件类型 / Event type"""
    THREAD_TITLE_UPDATED = "thread_title_updated"
    ERROR = "error"
    THINKING = "thinking"
    INTERACTIVE = "interactive"
    INTERACTIVE_RESPONSE = "interactive_response"
    TASK_FINISHED = "task_finished"
    CANCEL = "cancel"


class MessageRole(str, Enum):
    """消息角色 / Message role"""
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


class ContentType(str, Enum):
    """内容类型 / Content type"""
    TEXT = "text"
    SPIN_TEXT = "spin_text"
    IMAGE = "image"


class ItemStatus(str, Enum):
    """执行状态 / Item status"""
    INIT = "init"
    START = "start"
    PROGRESS = "progress"
    SUSPENDED = "suspended"
    SUCCESS = "success"
    FAIL = "fail"


class InteractionType(str, Enum):
    """交互类型 / Interaction type"""
    USER_ACK = "user_ack"
    USER_SELECT = "user_select"
    USER_INPUT = "user_input"
    SLS_QUERY = "sls_query"
