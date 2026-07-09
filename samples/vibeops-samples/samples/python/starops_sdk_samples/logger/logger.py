"""
Structured logger for STAROps SDK
STAROps SDK 结构化日志器
"""

import json
import os
import sys
import traceback
from datetime import datetime, timezone
from enum import IntEnum
from typing import Any, Dict, Optional, TextIO


class LogLevel(IntEnum):
    """日志级别 / Log level"""
    DEBUG = 0
    INFO = 1
    WARN = 2
    ERROR = 3

    @classmethod
    def from_string(cls, s: Optional[str]) -> "LogLevel":
        if not s:
            return cls.INFO
        s = s.lower().strip()
        if s == "debug":
            return cls.DEBUG
        elif s in ("warn", "warning"):
            return cls.WARN
        elif s == "error":
            return cls.ERROR
        return cls.INFO


class Logger:
    """结构化日志器 / Structured logger"""

    def __init__(self, level: LogLevel = LogLevel.INFO, output: Optional[TextIO] = None):
        self.level = level
        self.output = output or sys.stdout

    @classmethod
    def from_env(cls) -> "Logger":
        """从环境变量创建日志器 / Create logger from environment variables"""
        level_str = os.getenv("LOG_LEVEL")
        level = LogLevel.from_string(level_str)
        return cls(level=level)

    def set_level(self, level: LogLevel) -> None:
        self.level = level

    def get_level(self) -> LogLevel:
        return self.level

    def set_output(self, output: TextIO) -> None:
        self.output = output

    def _log(
        self,
        level: LogLevel,
        message: str,
        context: Optional[Dict[str, Any]] = None,
        error: Optional[Exception] = None,
        include_stack: bool = False,
    ) -> None:
        if level < self.level:
            return

        entry: Dict[str, Any] = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "level": level.name.lower(),
            "message": message,
        }

        if context:
            entry["context"] = context

        if error:
            entry["error"] = str(error)
            if include_stack:
                entry["stack"] = traceback.format_exc()

        try:
            self.output.write(json.dumps(entry, ensure_ascii=False) + "\n")
            self.output.flush()
        except Exception:
            self.output.write(f"{entry['timestamp']} [{level.name}] {message}\n")
            self.output.flush()

    def debug(self, message: str, context: Optional[Dict[str, Any]] = None) -> None:
        self._log(LogLevel.DEBUG, message, context)

    def info(self, message: str, context: Optional[Dict[str, Any]] = None) -> None:
        self._log(LogLevel.INFO, message, context)

    def warn(self, message: str, context: Optional[Dict[str, Any]] = None) -> None:
        self._log(LogLevel.WARN, message, context)

    def error(
        self,
        message: str,
        error: Optional[Exception] = None,
        context: Optional[Dict[str, Any]] = None,
    ) -> None:
        self._log(LogLevel.ERROR, message, context, error, include_stack=True)

    def log_request(
        self,
        thread_id: str,
        message: str,
        variables: Optional[Dict[str, Any]] = None,
    ) -> None:
        context: Dict[str, Any] = {
            "threadId": thread_id,
            "message": message,
        }
        if variables:
            context["variables"] = variables
        self.debug("发送请求 / Sending request", context)

    def log_response(
        self,
        thread_id: str,
        status_code: int,
        raw_json: str,
        is_done: bool,
        error: Optional[Exception] = None,
    ) -> None:
        context: Dict[str, Any] = {
            "threadId": thread_id,
            "statusCode": status_code,
            "isDone": is_done,
        }

        if raw_json:
            if len(raw_json) > 500:
                context["rawJSON"] = raw_json[:500] + "...(truncated)"
            else:
                context["rawJSON"] = raw_json

        if error:
            self.error("响应错误 / Response error", error, context)
        else:
            self.debug("收到响应 / Received response", context)
