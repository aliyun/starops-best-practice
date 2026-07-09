"""
Error handling for STAROps SDK
STAROps SDK 错误处理
"""

from enum import Enum
from typing import Any, Dict, List, Optional


class ErrorCode(str, Enum):
    """错误码 / Error code"""
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


class SDKException(Exception):
    """SDK 异常类 / SDK exception class"""

    def __init__(
        self,
        code: ErrorCode,
        message: str,
        cause: Optional[Exception] = None,
    ):
        super().__init__(message)
        self.code = code
        self.message = message
        self.cause = cause
        self.context: Dict[str, Any] = {}
        self.suggestion: Optional[str] = None

    def with_context(self, key: str, value: Any) -> "SDKException":
        """添加上下文信息 / Add context information"""
        self.context[key] = value
        return self

    def with_suggestion(self, suggestion: str) -> "SDKException":
        """添加建议 / Add suggestion"""
        self.suggestion = suggestion
        return self

    def __str__(self) -> str:
        result = f"[{self.code.value}] {self.message}"
        if self.cause:
            result += f": {self.cause}"
        return result

    @classmethod
    def config_missing(cls, missing_vars: List[str]) -> "SDKException":
        """创建配置缺失错误 / Create config missing error"""
        return cls(
            ErrorCode.CONFIG_MISSING,
            f"缺少必需的配置项: {', '.join(missing_vars)}",
        ).with_context("missingVariables", missing_vars).with_suggestion(
            "请检查 .env 文件或环境变量设置"
        )

    @classmethod
    def config_invalid(cls, field: str, reason: str) -> "SDKException":
        """创建配置无效错误 / Create config invalid error"""
        return cls(
            ErrorCode.CONFIG_INVALID,
            f"配置项 {field} 无效: {reason}",
        ).with_context("field", field).with_context("reason", reason).with_suggestion(
            "请检查配置值是否正确"
        )

    @classmethod
    def client_create(cls, cause: Exception) -> "SDKException":
        """创建客户端创建失败错误 / Create client creation error"""
        return cls(
            ErrorCode.CLIENT_CREATE,
            "创建客户端失败",
            cause,
        ).with_suggestion("请检查网络连接和认证信息")

    @classmethod
    def thread_create(cls, cause: Exception) -> "SDKException":
        """创建会话创建失败错误 / Create thread creation error"""
        return cls(
            ErrorCode.THREAD_CREATE,
            "创建会话失败",
            cause,
        ).with_suggestion("请检查 API 权限和配额")

    @classmethod
    def thread_not_found(cls, thread_id: str) -> "SDKException":
        """创建会话不存在错误 / Create thread not found error"""
        return cls(
            ErrorCode.THREAD_NOT_FOUND,
            f"会话不存在: {thread_id}",
        ).with_context("threadId", thread_id).with_suggestion(
            "请检查会话 ID 是否正确，或创建新会话"
        )

    @classmethod
    def chat_failed(cls, cause: Exception) -> "SDKException":
        """创建对话失败错误 / Create chat failed error"""
        return cls(
            ErrorCode.CHAT_FAILED,
            "对话失败",
            cause,
        ).with_suggestion("请稍后重试")

    @classmethod
    def timeout(cls, duration: str) -> "SDKException":
        """创建超时错误 / Create timeout error"""
        return cls(
            ErrorCode.TIMEOUT,
            f"操作超时: {duration}",
        ).with_context("duration", duration).with_suggestion(
            "请增加超时时间或检查网络连接"
        )

    @classmethod
    def cancelled(cls) -> "SDKException":
        """创建已取消错误 / Create cancelled error"""
        return cls(
            ErrorCode.CANCELLED,
            "操作已取消",
        ).with_suggestion("如需继续，请重新发起请求")

    @classmethod
    def network_error(cls, cause: Exception) -> "SDKException":
        """创建网络错误 / Create network error"""
        return cls(
            ErrorCode.NETWORK_ERROR,
            "网络错误",
            cause,
        ).with_suggestion("请检查网络连接")

    @classmethod
    def api_error(cls, api_code: str, api_message: str) -> "SDKException":
        """创建 API 错误 / Create API error"""
        return cls(
            ErrorCode.API_ERROR,
            f"API 错误 [{api_code}]: {api_message}",
        ).with_context("apiCode", api_code).with_context(
            "apiMessage", api_message
        ).with_suggestion("请参考 API 文档检查请求参数")

    @classmethod
    def parse_error(cls, cause: Exception) -> "SDKException":
        """创建解析错误 / Create parse error"""
        return cls(
            ErrorCode.PARSE_ERROR,
            "解析响应失败",
            cause,
        ).with_suggestion("请检查 SDK 版本是否最新")

    @classmethod
    def interactive_timeout(cls, duration: str) -> "SDKException":
        """创建交互超时错误 / Create interactive timeout error"""
        return cls(
            ErrorCode.INTERACTIVE_TIMEOUT,
            f"等待用户响应超时: {duration}",
        ).with_context("duration", duration).with_suggestion(
            "请重新操作并在规定时间内响应"
        )
