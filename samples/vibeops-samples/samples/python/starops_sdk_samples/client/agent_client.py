"""
Agent client for STAROps SDK
STAROps SDK Agent 客户端
"""

import asyncio
import json
import sys
import time
from dataclasses import dataclass
from typing import Any, AsyncIterator, Dict, List, Optional

from alibabacloud_starops20260428.client import Client
from alibabacloud_starops20260428 import models as starops_models
from alibabacloud_tea_openapi import models as openapi_models
from alibabacloud_tea_util import models as util_models

from .config import Config
from .errors import SDKException, ErrorCode
from .retry import (
    RetryConfig,
    RetryState,
    ConnectionOutcome,
    load_retry_config_from_env,
    is_stream_done_event,
    extract_newest_timestamp,
    calculate_backoff,
    build_reconnect_request,
)


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


class AgentClient:
    """Agent 客户端 / Agent client"""

    def __init__(self, config: Config):
        self.config = config
        try:
            openapi_config = openapi_models.Config(
                access_key_id=config.access_key_id,
                access_key_secret=config.access_key_secret,
                endpoint=config.endpoint,
                signature_version="v3",
            )
            self._client = Client(openapi_config)
        except Exception as e:
            raise SDKException.client_create(e)

    def create_thread(self) -> str:
        """创建会话 / Create thread"""
        try:
            variables = starops_models.CreateThreadRequestVariables(
                workspace=self.config.workspace
            )
            request = starops_models.CreateThreadRequest(
                title=f"Chat-{int(time.time())}",
                variables=variables,
            )
            response = self._client.create_thread(self.config.employee_name, request)

            if not response.body or not response.body.thread_id:
                raise SDKException(ErrorCode.THREAD_CREATE, "无效响应: 缺少ThreadID")

            return response.body.thread_id
        except SDKException:
            raise
        except Exception as e:
            raise SDKException.thread_create(e)

    async def chat(self, thread_id: str, message: str) -> AsyncIterator[ChatEvent]:
        """开始 SSE 对话 / Start SSE chat"""
        variables = {
            "workspace": self.config.workspace,
            "region": self.config.region,
            "language": "zh",
            "timeZone": "Asia/Shanghai",
            "timeStamp": str(int(time.time())),
        }
        async for event in self.chat_with_variables(thread_id, message, variables):
            yield event


    async def chat_with_variables(
        self, thread_id: str, message: str, variables: Optional[Dict[str, Any]] = None
    ) -> AsyncIterator[ChatEvent]:
        """开始 SSE 对话（支持自定义 variables）/ Start SSE chat with custom variables"""
        try:
            # Build request
            content = starops_models.CreateChatRequestMessagesContents(
                type="text",
                value=message,
            )
            msg = starops_models.CreateChatRequestMessages(
                role="user",
                contents=[content],
            )

            # Ensure required fields
            if variables is None:
                variables = {}
            variables.setdefault("workspace", self.config.workspace)
            variables.setdefault("region", self.config.region)
            variables.setdefault("language", "zh")
            variables.setdefault("timeZone", "Asia/Shanghai")
            variables.setdefault("timeStamp", str(int(time.time())))

            request = starops_models.CreateChatRequest(
                action="create",
                thread_id=thread_id,
                digital_employee_name=self.config.employee_name,
                messages=[msg],
                variables=variables,
            )

            # 使用带重试能力的 SSE 流处理 / Use SSE streaming with retry
            config = load_retry_config_from_env()
            async for event in self._stream_sse(request, config):
                yield event
        except Exception as e:
            yield ChatEvent.from_error(SDKException.chat_failed(e))

    # ===================== SSE 重试编排 =====================

    async def _stream_sse(
        self, request: Any, config: RetryConfig
    ) -> AsyncIterator[ChatEvent]:
        """启动带重试能力的 SSE 流处理（外层重连编排）

        连接中断时自动重连并通过 timestamp 去重；stream_done 是唯一正常结束标志。
        """
        state = RetryState()
        while True:
            outcome = ConnectionOutcome.INTERRUPTED
            terminal_event: Optional[ChatEvent] = None
            async for out, ev in self._stream_once(request, state, config):
                if out is None:
                    yield ev
                    continue
                outcome = out
                terminal_event = ev
                break

            if outcome == ConnectionOutcome.DONE:
                if terminal_event is not None:
                    yield terminal_event
                return
            if outcome == ConnectionOutcome.FATAL:
                if terminal_event is not None:
                    yield terminal_event
                return

            # INTERRUPTED：执行退避并判定是否继续重试
            if not await self._prepare_reconnect(state, config):
                yield ChatEvent.from_error(
                    SDKException(
                        ErrorCode.NETWORK_ERROR,
                        f"超过最大重试次数 {config.max_retries} 次，连接中断",
                    )
                )
                return
            request = build_reconnect_request(request)

    async def _stream_once(
        self, request: Any, state: RetryState, config: RetryConfig
    ) -> AsyncIterator[tuple]:
        """消费单次连接的事件流

        以 (outcome, event) 元组产出：
        - (None, event)：普通事件，需转发
        - (ConnectionOutcome.DONE, event)：收到 stream_done，正常结束
        - (ConnectionOutcome.INTERRUPTED, None)：连接中断，需重连
        - (ConnectionOutcome.FATAL, error_event)：致命错误
        空闲超时通过 asyncio.wait_for 包装队列读取实现。
        """
        loop = asyncio.get_running_loop()
        queue: asyncio.Queue = asyncio.Queue()
        start_time = time.time()

        def _producer() -> None:
            """在线程中消费同步 SSE 迭代器，结果投递到 asyncio 队列"""
            try:
                runtime = util_models.RuntimeOptions()
                runtime.connect_timeout = 30000
                runtime.read_timeout = 300000
                response_iterator = self._client.create_chat_with_sse(
                    request, {}, runtime
                )
                for response in response_iterator:
                    loop.call_soon_threadsafe(
                        queue.put_nowait, ("response", response)
                    )
                loop.call_soon_threadsafe(queue.put_nowait, ("closed", None))
            except Exception as exc:  # noqa: BLE001 - 一律视为连接中断
                loop.call_soon_threadsafe(queue.put_nowait, ("error", exc))

        producer = loop.run_in_executor(None, _producer)

        try:
            while True:
                try:
                    kind, payload = await asyncio.wait_for(
                        queue.get(), timeout=config.idle_timeout
                    )
                except asyncio.TimeoutError:
                    print("连接中断，中断原因：空闲超时，未收到消息")
                    yield (ConnectionOutcome.INTERRUPTED, None)
                    return

                if kind == "closed":
                    # 通道关闭且未收到 stream_done → 连接中断
                    print("连接中断，中断原因：通道关闭且未收到 stream_done")
                    yield (ConnectionOutcome.INTERRUPTED, None)
                    return

                if kind == "error":
                    # 错误为 None/关闭标记 → 忽略，继续循环
                    if payload is None:
                        continue
                    # 非 stream_done 的任何错误都视为连接中断，触发重连
                    print(f"SSE 连接错误: {payload}，准备重连...", file=sys.stderr)
                    print("连接中断，中断原因：SSE连接错误")
                    yield (ConnectionOutcome.INTERRUPTED, None)
                    return

                # kind == "response"
                response = payload
                if not response.body:
                    continue
                body_dict = response.body.to_map()
                raw_json = json.dumps(body_dict, ensure_ascii=False)
                event = ChatEvent.from_response(body_dict, raw_json, 200)

                if is_stream_done_event(event):
                    event.is_done = True
                    yield (ConnectionOutcome.DONE, event)
                    return

                if self._forward_event(event, state):
                    state.retry_count = 0
                    yield (None, event)

                # 模拟断连（转发后延迟触发）
                if getattr(self.config, "simulate_network_error", False):
                    if time.time() - start_time > 5:
                        self.config.simulate_network_error = False
                        print("模拟网络断连，触发重连...", file=sys.stderr)
                        yield (ConnectionOutcome.INTERRUPTED, None)
                        return
        finally:
            # 取消后台生产者，避免线程/连接泄漏
            producer.cancel()

    def _forward_event(self, event: ChatEvent, state: RetryState) -> bool:
        """去重判定普通事件，返回是否应转发该消息

        重连后进入去重窗口，仅当收到比 last_timestamp 更新的消息才转发并退出窗口。
        """
        ts = extract_newest_timestamp(event, state.last_timestamp)

        if state.in_dedupe_window:
            if ts == "":
                return False  # 重复消息，跳过
            state.in_dedupe_window = False  # 收到新消息，退出去重窗口

        if ts != "":
            state.last_timestamp = ts
        return True

    async def _prepare_reconnect(
        self, state: RetryState, config: RetryConfig
    ) -> bool:
        """执行退避并判定是否继续重试；返回 False 表示超过最大重试次数应终止"""
        if state.retry_count >= config.max_retries:
            return False
        state.retry_count += 1
        backoff = calculate_backoff(state.retry_count, config)
        print(
            f"连接中断，{backoff}s 后重试 (第 {state.retry_count}/{config.max_retries} 次)",
            file=sys.stderr,
        )
        await asyncio.sleep(backoff)
        state.in_dedupe_window = True  # 进入去重窗口
        return True


    async def interact(
        self, thread_id: str, user_interactive: str,
        base_variables: Optional[Dict[str, Any]] = None
    ) -> AsyncIterator[ChatEvent]:
        """发送交互响应并恢复 SSE 对话 / Send interactive response and resume SSE chat"""
        try:
            variables = dict(base_variables) if base_variables else {}
            variables["userInteractive"] = user_interactive
            variables.setdefault("workspace", self.config.workspace)
            variables.setdefault("region", self.config.region)
            variables.setdefault("language", "zh")
            variables.setdefault("timeZone", "Asia/Shanghai")
            variables.setdefault("timeStamp", str(int(time.time())))

            request = starops_models.CreateChatRequest(
                action="interact",
                thread_id=thread_id,
                digital_employee_name=self.config.employee_name,
                variables=variables,
            )

            runtime = util_models.RuntimeOptions()
            runtime.connect_timeout = 30000
            runtime.read_timeout = 300000
            response_iterator = self._client.create_chat_with_sse(request, {}, runtime)

            for response in response_iterator:
                if response.body:
                    body_dict = response.body.to_map()
                    raw_json = json.dumps(body_dict, ensure_ascii=False)
                    event = ChatEvent.from_response(body_dict, raw_json, 200)
                    yield event
                    if event.is_done:
                        return

            yield ChatEvent.done()
        except Exception as e:
            yield ChatEvent.from_error(SDKException.chat_failed(e))

    async def chat_with_timeout(
        self, thread_id: str, message: str, timeout: float
    ) -> AsyncIterator[ChatEvent]:
        """带超时的对话 / Chat with timeout"""
        try:
            async for event in asyncio.wait_for(
                self._collect_events(thread_id, message), timeout=timeout
            ):
                yield event
        except asyncio.TimeoutError:
            yield ChatEvent.from_error(SDKException.timeout(f"{timeout}s"))

    async def _collect_events(self, thread_id: str, message: str) -> AsyncIterator[ChatEvent]:
        async for event in self.chat(thread_id, message):
            yield event


    def list_threads(self, page_size: int = 20) -> tuple[List[ThreadInfo], int]:
        """列出会话 / List threads"""
        try:
            if page_size <= 0:
                page_size = 20
            if page_size > 100:
                page_size = 100

            request = starops_models.ListThreadsRequest(max_results=page_size)
            response = self._client.list_threads(self.config.employee_name, request)

            if not response.body:
                raise SDKException(
                    ErrorCode.PARSE_ERROR, "无效响应: 响应体为空"
                ).with_suggestion("请稍后重试")

            threads = []
            if response.body.threads:
                for t in response.body.threads:
                    threads.append(ThreadInfo(
                        thread_id=t.thread_id or "",
                        title=t.title or "",
                        status=t.status or "",
                        create_time=t.create_time or "",
                        update_time=t.update_time or "",
                    ))

            total = response.body.total or 0
            return threads, total
        except SDKException:
            raise
        except Exception as e:
            raise SDKException(
                ErrorCode.API_ERROR, "获取会话列表失败", e
            ).with_suggestion("请检查网络连接和 API 权限")

    def get_thread(self, thread_id: str) -> ThreadInfo:
        """获取会话详情 / Get thread details"""
        self._validate_thread_id(thread_id)

        try:
            response = self._client.get_thread(self.config.employee_name, thread_id)

            if not response.body:
                raise SDKException(
                    ErrorCode.PARSE_ERROR, "无效响应: 响应体为空"
                ).with_context("threadId", thread_id).with_suggestion("请稍后重试")

            return ThreadInfo(
                thread_id=response.body.thread_id or "",
                title=response.body.title or "",
                status=response.body.status or "",
                create_time=response.body.create_time or "",
                update_time=response.body.update_time or "",
            )
        except SDKException:
            raise
        except Exception as e:
            if self._is_thread_not_found_error(e):
                raise SDKException.thread_not_found(thread_id)
            raise SDKException(
                ErrorCode.API_ERROR, f"获取会话详情失败: {thread_id}", e
            ).with_context("threadId", thread_id).with_suggestion("请检查会话 ID 是否正确")

    def delete_thread(self, thread_id: str) -> None:
        """删除会话 / Delete thread"""
        self._validate_thread_id(thread_id)

        try:
            self._client.delete_thread(self.config.employee_name, thread_id)
        except Exception as e:
            if self._is_thread_not_found_error(e):
                raise SDKException.thread_not_found(thread_id)
            raise SDKException(
                ErrorCode.API_ERROR, f"删除会话失败: {thread_id}", e
            ).with_context("threadId", thread_id).with_suggestion("请检查会话 ID 是否正确")


    def get_thread_data(self, thread_id: str, limit: int = 50) -> List[ThreadMessage]:
        """获取会话消息 / Get thread messages"""
        self._validate_thread_id(thread_id)

        try:
            if limit <= 0:
                limit = 50
            if limit > 100:
                limit = 100

            request = starops_models.GetThreadDataRequest(max_results=limit)
            response = self._client.get_thread_data(
                self.config.employee_name, thread_id, request
            )

            if not response.body:
                raise SDKException(
                    ErrorCode.PARSE_ERROR, "无效响应: 响应体为空"
                ).with_context("threadId", thread_id).with_suggestion("请稍后重试")

            # Strategy: prefer system Result over assistant streaming messages
            # 策略：优先使用 system Result，而不是 assistant 流式消息
            message_map: Dict[str, Dict[str, str]] = {}
            message_order: List[str] = []
            
            # Check if any system Result exists
            # 检查是否存在 system Result
            has_system_result = False
            if response.body.data:
                for data in response.body.data:
                    if data.messages:
                        for msg in data.messages:
                            if msg.role == 'system':
                                artifacts = getattr(msg, 'artifacts', None)
                                if artifacts:
                                    for artifact in artifacts:
                                        if isinstance(artifact, dict) and artifact.get('name') == 'Result':
                                            has_system_result = True
                                            break
                            if has_system_result:
                                break
                    if has_system_result:
                        break

            # Process messages
            if response.body.data:
                for data in response.body.data:
                    if data.messages:
                        for msg in data.messages:
                            role = msg.role or ''
                            timestamp = msg.timestamp or ''
                            
                            # Skip assistant streaming messages if system Result exists
                            if role == 'assistant' and has_system_result:
                                continue
                            
                            # Use different key strategy based on role
                            if role == 'user':
                                key = f"user-{timestamp}"
                            elif role == 'system':
                                key = f"system-{timestamp}"
                            else:
                                call_id = getattr(msg, 'call_id', '') or ''
                                key = f"assistant-{call_id}"
                            
                            content = self._extract_message_content(msg)
                            if not content:
                                continue

                            if key in message_map:
                                message_map[key]['content'] += content
                            else:
                                # For system messages, display as assistant role
                                display_role = 'assistant' if role == 'system' else role
                                message_map[key] = {
                                    'role': display_role,
                                    'content': content,
                                    'timestamp': timestamp,
                                }
                                message_order.append(key)

            return [
                ThreadMessage(
                    role=message_map[key]['role'],
                    content=message_map[key]['content'],
                    timestamp=message_map[key]['timestamp'],
                )
                for key in message_order
            ]
        except SDKException:
            raise
        except Exception as e:
            if self._is_thread_not_found_error(e):
                raise SDKException.thread_not_found(thread_id)
            raise SDKException(
                ErrorCode.API_ERROR, f"获取会话消息失败: {thread_id}", e
            ).with_context("threadId", thread_id).with_suggestion("请检查会话 ID 是否正确")

    def _validate_thread_id(self, thread_id: str) -> None:
        if not thread_id:
            raise SDKException(
                ErrorCode.CONFIG_INVALID, "会话 ID 不能为空"
            ).with_context("threadId", thread_id).with_suggestion("请提供有效的会话 ID")

        if any(c in thread_id for c in " \t\n\r"):
            raise SDKException(
                ErrorCode.CONFIG_INVALID, f"会话 ID 包含非法字符: {thread_id}"
            ).with_context("threadId", thread_id).with_suggestion("会话 ID 不能包含空白字符")

    def _is_thread_not_found_error(self, e: Exception) -> bool:
        if not e:
            return False
        err_str = str(e)
        patterns = ["NotFound", "not found", "NOT_FOUND", "ThreadNotFound", 
                   "InvalidThreadId", "does not exist"]
        return any(p in err_str for p in patterns)

    def _extract_message_content(self, msg: Any) -> str:
        # 1. Try to extract from contents (streaming text chunks)
        # 尝试从 contents 提取（流式文本块）
        if msg and hasattr(msg, "contents") and msg.contents:
            result = []
            for content in msg.contents:
                if isinstance(content, dict):
                    if content.get("type") == "text" and content.get("value"):
                        result.append(content["value"])
            if result:
                return "".join(result)

        # 2. Try to extract from artifacts (final result)
        # 尝试从 artifacts 提取（最终结果）
        if msg and hasattr(msg, "artifacts") and msg.artifacts:
            for artifact in msg.artifacts:
                if isinstance(artifact, dict) and artifact.get("name") == "Result":
                    parts = artifact.get("parts", [])
                    text_parts = []
                    for part in parts:
                        if isinstance(part, dict) and part.get("kind") == "text" and part.get("text"):
                            text_parts.append(part["text"])
                    if text_parts:
                        return "".join(text_parts)

        return ""
