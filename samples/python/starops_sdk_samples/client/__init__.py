"""
Client module for STAROps SDK
STAROps SDK 客户端模块
"""

from .config import Config
from .errors import SDKException, ErrorCode
from .agent_client import AgentClient
from ..types.events import ChatEvent, ThreadInfo, ThreadMessage
from .simple_printer import SimplePrinter
from .event_printer import EventPrinter
from .interactive_handler import InteractiveHandler, InteractiveResponse, ChatClientProtocol
from .retry import RetryConfig, RetryState, ConnectionOutcome
from .models import (
    ModelsConfig,
    Model,
    Provider,
    load_models,
    find_config_path,
    parse_model_flag,
    validate_model,
    build_model_json,
    build_config_with_model,
    display_menu,
    list_model_flags,
    get_model_by_index,
)

__all__ = [
    "Config",
    "SDKException",
    "ErrorCode",
    "AgentClient",
    "ChatEvent",
    "ThreadInfo",
    "ThreadMessage",
    "SimplePrinter",
    "EventPrinter",
    "InteractiveHandler",
    "InteractiveResponse",
    "ChatClientProtocol",
    "RetryConfig",
    "RetryState",
    "ConnectionOutcome",
    "ModelsConfig",
    "Model",
    "Provider",
    "load_models",
    "find_config_path",
    "parse_model_flag",
    "validate_model",
    "build_model_json",
    "build_config_with_model",
    "display_menu",
    "list_model_flags",
    "get_model_by_index",
]
