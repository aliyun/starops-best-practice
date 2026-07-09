"""
Tests for SSE retry utilities
SSE 重试工具函数测试
"""

import os
from dataclasses import dataclass
from typing import Any, Dict, Optional

from starops_sdk_samples.client.retry import (
    RetryConfig,
    RetryState,
    ConnectionOutcome,
    default_retry_config,
    load_retry_config_from_env,
    is_stream_done_event,
    is_newer_timestamp,
    extract_newest_timestamp,
    calculate_backoff,
    build_reconnect_request,
)


# 轻量事件替身：仅需 body 字段供工具函数读取
# Lightweight event stub: only needs the body attribute
@dataclass
class _FakeEvent:
    body: Optional[Dict[str, Any]] = None


def _event_with_timestamps(*timestamps: str) -> _FakeEvent:
    """构造包含指定 timestamp 消息的事件；空字符串表示不设置 timestamp"""
    messages = []
    for ts in timestamps:
        msg: Dict[str, Any] = {}
        if ts != "":
            msg["timestamp"] = ts
        messages.append(msg)
    return _FakeEvent(body={"messages": messages})


class TestRetryConfig:
    def test_default(self):
        cfg = default_retry_config()
        assert cfg.max_retries == 10
        assert cfg.initial_backoff == 1.0
        assert cfg.max_backoff == 30.0
        assert cfg.backoff_factor == 2.0
        assert cfg.idle_timeout == 60.0

    def test_load_from_env_default(self):
        os.environ.pop("VIBEOPS_MAX_RETRIES", None)
        os.environ.pop("VIBEOPS_IDLE_TIMEOUT", None)
        cfg = load_retry_config_from_env()
        assert cfg.max_retries == 10
        assert cfg.idle_timeout == 60.0

    def test_load_from_env_custom(self):
        os.environ["VIBEOPS_MAX_RETRIES"] = "5"
        os.environ["VIBEOPS_IDLE_TIMEOUT"] = "30"
        try:
            cfg = load_retry_config_from_env()
            assert cfg.max_retries == 5
            assert cfg.idle_timeout == 30.0
        finally:
            os.environ.pop("VIBEOPS_MAX_RETRIES", None)
            os.environ.pop("VIBEOPS_IDLE_TIMEOUT", None)

    def test_load_from_env_invalid(self):
        os.environ["VIBEOPS_MAX_RETRIES"] = "abc"
        try:
            cfg = load_retry_config_from_env()
            assert cfg.max_retries == 10
        finally:
            os.environ.pop("VIBEOPS_MAX_RETRIES", None)

    def test_load_from_env_negative(self):
        os.environ["VIBEOPS_MAX_RETRIES"] = "-1"
        try:
            cfg = load_retry_config_from_env()
            assert cfg.max_retries == 10
        finally:
            os.environ.pop("VIBEOPS_MAX_RETRIES", None)


class TestIsNewerTimestamp:
    def test_cases(self):
        cases = [
            ("", "100", False),      # 空 ts
            ("100", "", True),       # 空 base
            ("", "", False),         # 都为空
            ("200", "100", True),    # 数值更新
            ("50", "100", False),    # 数值更旧
            ("100", "100", False),   # 数值相等
            ("2024-01-02", "2024-01-01", True),   # 字符串更新
            ("2024-01-01", "2024-01-02", False),  # 字符串更旧
            ("abc", "100", False),   # base 数值但 ts 无法解析
            ("200", "abc", False),   # base 无法解析走字符串比较："200" < "abc"
        ]
        for ts, base, expected in cases:
            assert is_newer_timestamp(ts, base) == expected, f"{ts!r} vs {base!r}"


class TestExtractNewestTimestamp:
    def test_none_event(self):
        assert extract_newest_timestamp(None, "100") == ""

    def test_no_messages(self):
        assert extract_newest_timestamp(_FakeEvent(body={}), "100") == ""

    def test_single_newer(self):
        assert extract_newest_timestamp(_event_with_timestamps("200"), "100") == "200"

    def test_multiple_pick_newest(self):
        event = _event_with_timestamps("150", "300", "200")
        assert extract_newest_timestamp(event, "100") == "300"

    def test_all_older_or_equal(self):
        event = _event_with_timestamps("50", "80", "100")
        assert extract_newest_timestamp(event, "100") == ""

    def test_message_without_timestamp(self):
        assert extract_newest_timestamp(_event_with_timestamps(""), "100") == ""


class TestCalculateBackoff:
    def test_default_config(self):
        cfg = default_retry_config()
        cases = [
            (1, 1.0),    # 1 * 2^0
            (2, 2.0),    # 1 * 2^1
            (3, 4.0),    # 1 * 2^2
            (4, 8.0),    # 1 * 2^3
            (5, 16.0),   # 1 * 2^4
            (6, 30.0),   # 32 → 上限 30
            (7, 30.0),
            (10, 30.0),
        ]
        for retry_count, expected in cases:
            assert calculate_backoff(retry_count, cfg) == expected

    def test_custom_config(self):
        cfg = RetryConfig(
            max_retries=5,
            initial_backoff=0.5,
            max_backoff=10.0,
            backoff_factor=3.0,
        )
        assert calculate_backoff(1, cfg) == 0.5    # 0.5 * 3^0
        assert calculate_backoff(2, cfg) == 1.5    # 0.5 * 3^1
        assert calculate_backoff(3, cfg) == 4.5    # 0.5 * 3^2


class TestIsStreamDoneEvent:
    def test_none(self):
        assert is_stream_done_event(None) is False

    def test_no_body(self):
        assert is_stream_done_event(_FakeEvent(body=None)) is False

    def test_stream_done(self):
        event = _FakeEvent(body={
            "messages": [{"events": [{"type": "stream_done"}]}]
        })
        assert is_stream_done_event(event) is True

    def test_other_event(self):
        event = _FakeEvent(body={
            "messages": [{"events": [{"type": "thinking"}]}]
        })
        assert is_stream_done_event(event) is False

    def test_no_events(self):
        event = _FakeEvent(body={"messages": [{"timestamp": "100"}]})
        assert is_stream_done_event(event) is False


@dataclass
class _FakeRequest:
    action: str = "create"
    thread_id: str = ""
    digital_employee_name: str = ""
    variables: Optional[Dict[str, Any]] = None
    messages: Optional[Any] = None


class TestBuildReconnectRequest:
    def test_basic(self):
        orig = _FakeRequest(
            action="create",
            thread_id="thread-1",
            digital_employee_name="emp-1",
            variables={"workspace": "ws", "region": "cn-hangzhou"},
            messages=["m1"],
        )
        new_req = build_reconnect_request(orig)
        assert new_req.action == "reconnect"
        assert new_req.thread_id == "thread-1"
        assert new_req.digital_employee_name == "emp-1"
        assert new_req.variables == {"workspace": "ws", "region": "cn-hangzhou"}
        # 重连不携带 messages
        assert getattr(new_req, "messages", None) is None

    def test_variables_copied(self):
        orig_vars = {"k": "v"}
        orig = _FakeRequest(thread_id="t", digital_employee_name="e", variables=orig_vars)
        new_req = build_reconnect_request(orig)
        # 应为拷贝，修改原 variables 不影响新请求
        orig_vars["k"] = "changed"
        assert new_req.variables == {"k": "v"}


class TestRetryState:
    def test_defaults(self):
        state = RetryState()
        assert state.last_timestamp == ""
        assert state.in_dedupe_window is False
        assert state.retry_count == 0


class TestConnectionOutcome:
    def test_members(self):
        assert ConnectionOutcome.DONE != ConnectionOutcome.INTERRUPTED
        assert ConnectionOutcome.FATAL != ConnectionOutcome.DONE
