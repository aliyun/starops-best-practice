"""
Tests for SDK exceptions - standalone tests without SDK dependency
"""

import pytest
import sys
import os
from enum import Enum
from typing import Any, Dict, List, Optional


# Copy of ErrorCode enum for testing
class ErrorCode(str, Enum):
    CONFIG_MISSING = "CONFIG_MISSING"
    CONFIG_INVALID = "CONFIG_INVALID"
    CLIENT_CREATE = "CLIENT_CREATE"
    THREAD_CREATE = "THREAD_CREATE"
    THREAD_NOT_FOUND = "THREAD_NOT_FOUND"
    CHAT_FAILED = "CHAT_FAILED"
    TIMEOUT = "TIMEOUT"
    CANCELLED = "CANCELLED"
    NETWORK_ERROR = "NETWORK_ERROR"
    API_ERROR = "API_ERROR"
    PARSE_ERROR = "PARSE_ERROR"
    INTERACTIVE_TIMEOUT = "INTERACTIVE_TIMEOUT"


# Copy of SDKException for testing
class SDKException(Exception):
    def __init__(self, code: ErrorCode, message: str, cause: Optional[Exception] = None):
        super().__init__(message)
        self.code = code
        self.message = message
        self.cause = cause
        self.context: Dict[str, Any] = {}
        self.suggestion: Optional[str] = None

    def with_context(self, key: str, value: Any) -> "SDKException":
        self.context[key] = value
        return self

    def with_suggestion(self, suggestion: str) -> "SDKException":
        self.suggestion = suggestion
        return self

    def __str__(self) -> str:
        result = f"[{self.code.value}] {self.message}"
        if self.cause:
            result += f": {self.cause}"
        return result

    @classmethod
    def config_missing(cls, missing_vars: List[str]) -> "SDKException":
        return cls(
            ErrorCode.CONFIG_MISSING,
            f"缺少必需的配置项: {', '.join(missing_vars)}",
        ).with_context("missingVariables", missing_vars).with_suggestion(
            "请检查 .env 文件或环境变量设置"
        )

    @classmethod
    def config_invalid(cls, field: str, reason: str) -> "SDKException":
        return cls(
            ErrorCode.CONFIG_INVALID,
            f"配置项 {field} 无效: {reason}",
        ).with_context("field", field).with_context("reason", reason).with_suggestion(
            "请检查配置值是否正确"
        )

    @classmethod
    def thread_not_found(cls, thread_id: str) -> "SDKException":
        return cls(
            ErrorCode.THREAD_NOT_FOUND,
            f"会话不存在: {thread_id}",
        ).with_context("threadId", thread_id).with_suggestion(
            "请检查会话 ID 是否正确，或创建新会话"
        )

    @classmethod
    def timeout(cls, duration: str) -> "SDKException":
        return cls(
            ErrorCode.TIMEOUT,
            f"操作超时: {duration}",
        ).with_context("duration", duration).with_suggestion(
            "请增加超时时间或检查网络连接"
        )


class TestSDKException:
    def test_config_missing(self):
        missing_vars = ["VAR1", "VAR2"]
        ex = SDKException.config_missing(missing_vars)

        assert ex.code == ErrorCode.CONFIG_MISSING
        assert "VAR1" in str(ex)
        assert "VAR2" in str(ex)
        assert ex.suggestion is not None
        assert ex.context["missingVariables"] == missing_vars

    def test_config_invalid(self):
        ex = SDKException.config_invalid("endpoint", "invalid URL")

        assert ex.code == ErrorCode.CONFIG_INVALID
        assert "endpoint" in str(ex)
        assert "invalid URL" in str(ex)
        assert ex.context["field"] == "endpoint"

    def test_thread_not_found(self):
        ex = SDKException.thread_not_found("thread-123")

        assert ex.code == ErrorCode.THREAD_NOT_FOUND
        assert "thread-123" in str(ex)
        assert ex.context["threadId"] == "thread-123"

    def test_timeout(self):
        ex = SDKException.timeout("30s")

        assert ex.code == ErrorCode.TIMEOUT
        assert "30s" in str(ex)

    def test_with_context(self):
        ex = SDKException(ErrorCode.API_ERROR, "test error")
        ex.with_context("key1", "value1").with_context("key2", 123)

        assert ex.context["key1"] == "value1"
        assert ex.context["key2"] == 123

    def test_with_suggestion(self):
        ex = SDKException(ErrorCode.API_ERROR, "test error")
        ex.with_suggestion("Try again later")

        assert ex.suggestion == "Try again later"

    def test_str(self):
        ex = SDKException(ErrorCode.API_ERROR, "test error")
        s = str(ex)

        assert "API_ERROR" in s
        assert "test error" in s

    def test_with_cause(self):
        cause = RuntimeError("root cause")
        ex = SDKException(ErrorCode.NETWORK_ERROR, "network failed", cause)

        assert ex.cause == cause
        assert "root cause" in str(ex)
