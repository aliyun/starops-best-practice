"""
SSE 重试逻辑：配置、状态与工具函数
SSE retry logic: config, state and utility functions

跨语言设计规格（与 Go 参考实现保持一致）：
- 重试触发：所有非 stream_done 的中断（流结束、SSE 错误、空闲超时）均触发重连
- 唯一结束：stream_done 是正常结束的唯一标志
- 不区分错误：不判断错误类型，一律重试
- 去重机制：in_dedupe_window + last_timestamp，重连后仅转发比上次更新的消息
- 退避策略：min(initial_backoff * factor**(retry_count-1), max_backoff)，最多 max_retries 次
- 空闲超时：默认 60s 未收消息视为中断（环境变量 VIBEOPS_IDLE_TIMEOUT 覆盖）
- 重连请求：action="reconnect"，复制 thread_id/digital_employee_name/variables
"""

import os
from dataclasses import dataclass
from enum import Enum
from typing import Any, Optional

from alibabacloud_starops20260428 import models as starops_models


# ===================== 一、配置 =====================


@dataclass
class RetryConfig:
    """重试配置 / Retry configuration"""
    max_retries: int = 10          # 最大重试次数，默认 10
    initial_backoff: float = 1.0   # 初始退避时间（秒），默认 1s
    max_backoff: float = 30.0      # 最大退避时间（秒），默认 30s
    backoff_factor: float = 2.0    # 退避系数，默认 2.0
    idle_timeout: float = 60.0     # 空闲超时（秒）：超过此时长未收到任何消息视为连接中断，默认 60s


def default_retry_config() -> RetryConfig:
    """返回默认重试配置 / Return default retry config"""
    return RetryConfig()


def load_retry_config_from_env() -> RetryConfig:
    """从环境变量加载重试配置 / Load retry config from environment variables"""
    cfg = default_retry_config()

    max_retries = os.getenv("VIBEOPS_MAX_RETRIES")
    if max_retries:
        try:
            n = int(max_retries)
            if n > 0:
                cfg.max_retries = n
        except ValueError:
            pass

    idle_timeout = os.getenv("VIBEOPS_IDLE_TIMEOUT")
    if idle_timeout:
        try:
            n = int(idle_timeout)
            if n > 0:
                cfg.idle_timeout = float(n)
        except ValueError:
            pass

    return cfg


# ===================== 二、状态定义 =====================


@dataclass
class RetryState:
    """聚合重连过程中的状态 / Aggregated state during reconnection"""
    last_timestamp: str = ""       # 最后一条已转发消息的时间戳，用于重连去重
    in_dedupe_window: bool = False  # True=重连后去重窗口，仅转发更新的消息
    retry_count: int = 0           # 当前连续重试次数


class ConnectionOutcome(Enum):
    """单次连接的结束原因 / Outcome of a single connection"""
    DONE = "done"                  # 收到 stream_done，正常结束
    INTERRUPTED = "interrupted"    # 连接中断，需重连
    FATAL = "fatal"                # 致命错误（如取消），终止


# ===================== 三、工具函数 =====================


def is_stream_done_event(event: Any) -> bool:
    """判断事件是否为 stream_done（正常结束标志）

    不检查 response 级别的 event 字段，只判断 messages[].events[].type == stream_done
    """
    if event is None or getattr(event, "body", None) is None:
        return False
    messages = event.body.get("messages") or []
    for msg in messages:
        if not isinstance(msg, dict):
            continue
        events = msg.get("events") or []
        for evt in events:
            if isinstance(evt, dict) and evt.get("type") == "stream_done":
                return True
    return False


def _parse_int(value: str) -> Optional[int]:
    """尝试将字符串解析为十进制整数，失败返回 None"""
    try:
        return int(value)
    except (ValueError, TypeError):
        return None


def is_newer_timestamp(ts: str, base: str) -> bool:
    """判断 ts 是否比 base 更新

    优先数值比较（Unix 时间戳），无法解析时 fallback 为字符串比较
    """
    if not ts:
        return False
    if not base:
        return True
    ts_val = _parse_int(ts)
    base_val = _parse_int(base)
    if ts_val is not None and base_val is not None:
        return ts_val > base_val
    if ts_val is None and base_val is not None:
        return False  # 基准是数值但当前 ts 无法解析，视为不更新
    return ts > base


def extract_newest_timestamp(event: Any, base: str) -> str:
    """从事件中提取比 base 更新的最大消息 timestamp

    返回空字符串表示没有比 base 更新的时间戳
    """
    if event is None or getattr(event, "body", None) is None:
        return ""
    messages = event.body.get("messages")
    if not messages:
        return ""

    newest = base
    for msg in messages:
        if not isinstance(msg, dict):
            continue
        ts = msg.get("timestamp") or ""
        if is_newer_timestamp(ts, newest):
            newest = ts
    if newest == base:
        return ""
    return newest


def calculate_backoff(retry_count: int, config: RetryConfig) -> float:
    """计算退避时间（秒）

    default config: 1s * 2.0**(retry_count-1)，上限 max_backoff
    """
    backoff = config.initial_backoff * (config.backoff_factor ** (retry_count - 1))
    if backoff > config.max_backoff:
        return config.max_backoff
    return backoff


def build_reconnect_request(orig_req: Any) -> Any:
    """构建重连请求

    action="reconnect"，复制 thread_id/digital_employee_name/variables，
    不携带 messages 和 lastEventTimestamp
    """
    variables = {}
    orig_variables = getattr(orig_req, "variables", None)
    if orig_variables:
        variables = dict(orig_variables)

    return starops_models.CreateChatRequest(
        action="reconnect",
        thread_id=getattr(orig_req, "thread_id", None),
        digital_employee_name=getattr(orig_req, "digital_employee_name", None),
        variables=variables,
    )
