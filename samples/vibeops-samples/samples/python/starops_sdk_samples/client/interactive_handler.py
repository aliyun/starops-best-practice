"""
Interactive handler for STAROps SDK
STAROps SDK 交互处理器
"""

import json
import sys
import time
from dataclasses import dataclass
from typing import Any, AsyncIterator, Dict, List, Optional, TextIO

from ..types import EventType, InteractionType
from .agent_client import AgentClient, ChatEvent
from .errors import SDKException, ErrorCode


@dataclass
class InteractiveResponse:
    """交互响应 / Interactive response"""
    call_id: str
    type: InteractionType
    response: Dict[str, Any]
    source: Optional[Dict[str, Any]] = None
    modified_data: Optional[Dict[str, Any]] = None
    form_data: Optional[Dict[str, Any]] = None
    decision: str = ""


class InteractiveHandler:
    """交互事件处理器 / Interactive event handler"""

    def __init__(
        self,
        client: AgentClient,
        timeout: Optional[float] = None,
        reader: Optional[TextIO] = None,
        writer: Optional[TextIO] = None,
    ):
        self.client = client
        self.timeout = timeout
        self.reader = reader or sys.stdin
        self.writer = writer or sys.stdout

    def handle_event(self, event: Dict[str, Any], call_id: str) -> InteractiveResponse:
        """处理交互事件 / Handle interactive event"""
        if not event:
            raise SDKException(ErrorCode.PARSE_ERROR, "事件为空")

        event_type = event.get("type")
        if event_type != EventType.INTERACTIVE.value:
            raise SDKException(ErrorCode.PARSE_ERROR, f"不支持的事件类型: {event_type}")

        payload = event.get("payload")
        if not payload:
            raise SDKException(ErrorCode.PARSE_ERROR, "交互负载为空")

        interactive_type = payload.get("type")

        if interactive_type == InteractionType.USER_ACK.value:
            return self.handle_user_ack(payload, call_id)
        elif interactive_type == InteractionType.USER_SELECT.value:
            return self.handle_user_select(payload, call_id)
        elif interactive_type == InteractionType.USER_INPUT.value:
            return self.handle_user_input(payload, call_id)
        else:
            raise SDKException(ErrorCode.PARSE_ERROR, f"不支持的交互类型: {interactive_type}")

    def handle_user_ack(self, payload: Dict[str, Any], call_id: str) -> InteractiveResponse:
        """处理用户确认 / Handle user acknowledgment"""
        title = self._get_title(payload)
        message = self._get_description(payload)
        options = self._get_options(payload)
        modified_data = self._extract_data(payload)

        self._print("\n🔔 确认请求")
        self._print("--------------")
        if title:
            self._print(title)
        if message:
            self._print(f"\n{message}")

        if modified_data:
            self._print("")
            for key, value in modified_data.items():
                if key in ("title", "message"):
                    continue
                self._print(f"{key}: {value}")
        self._print("--------------")

        if options:
            parts = []
            for i, opt in enumerate(options):
                parts.append(f"[{self._get_option_value(opt)}] {self._get_option_label(opt, i)}")
            self._print(f"请输入 {', '.join(parts)}: ", end="")
        else:
            self._print("请输入 [y/yes] 确认，[n/no] 取消: ", end="")

        user_input = self._read_input().strip().lower()
        confirmed = user_input in ("", "y", "yes", "是")
        decision = "yes" if confirmed else "no"

        for opt in options:
            if user_input == self._get_option_value(opt).lower():
                decision = self._get_option_value(opt)
                confirmed = True
                break

        return InteractiveResponse(
            call_id=call_id,
            type=InteractionType.USER_ACK,
            response={"confirmed": confirmed},
            source=self._extract_source(payload),
            modified_data=modified_data,
            decision=decision,
        )

    def handle_user_select(self, payload: Dict[str, Any], call_id: str) -> InteractiveResponse:
        """处理用户选择 / Handle user selection"""
        title = self._get_title(payload)

        self._print("\n📋 请选择")
        if title:
            self._print(f"   标题: {title}")

        options = self._get_options(payload)
        if not options:
            raise SDKException(ErrorCode.PARSE_ERROR, "没有可选项")

        self._print("   选项:")
        for i, opt in enumerate(options):
            label = self._get_option_label(opt, i)
            self._print(f"   [{i + 1}] {label}")
        self._print(f"   请输入选项编号 (1-{len(options)}): ", end="")

        user_input = self._read_input().strip()
        try:
            selected_index = int(user_input)
        except ValueError:
            raise SDKException(
                ErrorCode.PARSE_ERROR,
                f"无效的选择: {user_input}，请输入 1-{len(options)} 之间的数字"
            )

        if selected_index < 1 or selected_index > len(options):
            raise SDKException(
                ErrorCode.PARSE_ERROR,
                f"无效的选择: {user_input}，请输入 1-{len(options)} 之间的数字"
            )

        selected_option = options[selected_index - 1]
        decision = self._get_option_value(selected_option)

        return InteractiveResponse(
            call_id=call_id,
            type=InteractionType.USER_SELECT,
            response={
                "selectedIndex": selected_index - 1,
                "selectedValue": selected_option,
            },
            source=self._extract_source(payload),
            modified_data=self._extract_data(payload),
            decision=decision,
        )

    def handle_user_input(self, payload: Dict[str, Any], call_id: str) -> InteractiveResponse:
        """处理用户输入（表单模式）/ Handle user input (form mode)"""
        title = self._get_title(payload)
        description = self._get_description(payload)
        source = self._extract_source(payload)

        form_spec = self._extract_form_spec(payload)
        elements = self._get_form_elements(form_spec)
        initial_values = self._get_form_initial_values(form_spec)

        self._print(f"\n✏️  {title}")
        if description:
            self._print(f"    {description}")
        self._print(f"    {'-' * 40}")

        form_data = {}
        for elem in elements:
            field = self._get_field_key(elem)
            label = self._get_field_label(elem, field)
            widget = self._get_field_widget(elem)
            placeholder = self._get_field_placeholder(elem)
            default_value = self._get_initial_value(initial_values, field)

            if widget in ("radio", "segmented"):
                enum_opts = self._get_field_enum(form_spec, field)
                if enum_opts:
                    self._print(f"    {label}:")
                    for i, opt in enumerate(enum_opts):
                        marker = "*" if default_value == opt else " "
                        self._print(f"      [{i + 1}]{marker} {opt}")
                    prompt = f"    请选择 (1-{len(enum_opts)})"
                    if default_value:
                        prompt += f" [默认: {default_value}]"
                    prompt += ": "
                    user_input = self._do_prompt(prompt).strip()
                    if not user_input and default_value:
                        form_data[field] = default_value
                    else:
                        try:
                            idx = int(user_input)
                            if 1 <= idx <= len(enum_opts):
                                form_data[field] = enum_opts[idx - 1]
                            else:
                                form_data[field] = default_value
                        except ValueError:
                            form_data[field] = default_value
            else:
                prompt = f"    {label}"
                if placeholder:
                    prompt += f" ({placeholder})"
                if default_value:
                    prompt += f" [默认: {default_value}]"
                prompt += ": "
                user_input = self._do_prompt(prompt).strip()
                if not user_input and default_value:
                    form_data[field] = default_value
                else:
                    form_data[field] = user_input

        self._print(f"    {'-' * 40}")

        return InteractiveResponse(
            call_id=call_id,
            type=InteractionType.USER_INPUT,
            response={"value": form_data},
            source=source,
            form_data=form_data,
            decision="submit",
        )

    async def resume_chat(
        self, thread_id: str, response: InteractiveResponse,
        base_variables: Dict[str, Any] = None
    ) -> AsyncIterator[ChatEvent]:
        """使用交互响应恢复对话 / Resume chat with interactive response"""
        if not self.client:
            yield ChatEvent.from_error(
                SDKException(ErrorCode.CLIENT_CREATE, "客户端未初始化")
            )
            return

        if not response:
            yield ChatEvent.from_error(
                SDKException(ErrorCode.PARSE_ERROR, "交互响应为空")
            )
            return

        user_interactive = {
            "callId": response.call_id,
            "source": response.source,
            "decision": response.decision,
        }
        if response.type == InteractionType.USER_INPUT:
            user_interactive["formData"] = response.form_data
        else:
            user_interactive["modifiedData"] = response.modified_data

        ui_json = json.dumps(user_interactive)
        async for event in self.client.interact(thread_id, ui_json, base_variables):
            yield event

    # =================================================================================
    # 辅助方法
    # =================================================================================

    def _print(self, text: str, end: str = "\n") -> None:
        self.writer.write(text + end)
        self.writer.flush()

    def _do_prompt(self, prompt: str) -> str:
        self.writer.write(prompt)
        self.writer.flush()
        try:
            return self.reader.readline()
        except Exception as e:
            raise SDKException(ErrorCode.PARSE_ERROR, "读取输入失败", e)

    def _read_input(self) -> str:
        try:
            return self.reader.readline()
        except Exception as e:
            raise SDKException(ErrorCode.PARSE_ERROR, "读取输入失败", e)

    def _get_title(self, payload: Dict[str, Any]) -> Optional[str]:
        user_ack = payload.get("userAck", {})
        if user_ack:
            data = user_ack.get("data", {})
            if data.get("title"):
                return data["title"]
        user_input = payload.get("userInput", {})
        if user_input.get("title"):
            return user_input["title"]
        meta = payload.get("meta", {})
        return meta.get("title")

    def _get_description(self, payload: Dict[str, Any]) -> Optional[str]:
        user_ack = payload.get("userAck", {})
        if user_ack.get("message"):
            return user_ack["message"]
        user_input = payload.get("userInput", {})
        if user_input.get("description"):
            return user_input["description"]
        meta = payload.get("meta", {})
        return meta.get("description") or meta.get("desc")

    def _get_options(self, payload: Dict[str, Any]) -> List[Dict[str, Any]]:
        user_ack = payload.get("userAck", {})
        options = user_ack.get("options")
        if options:
            return options
        data = payload.get("data", [])
        if data:
            return data
        meta = payload.get("meta", {})
        return meta.get("options", [])

    def _get_option_label(self, option: Dict[str, Any], index: int) -> str:
        for field in ["label", "name", "title", "value"]:
            value = option.get(field)
            if isinstance(value, str) and value:
                return value
        return f"选项 {index + 1}"

    def _get_option_value(self, option: Dict[str, Any]) -> str:
        value = option.get("value")
        return str(value) if value is not None else ""

    def _extract_source(self, payload: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        user_ack = payload.get("userAck", {})
        if user_ack.get("source"):
            return user_ack["source"]
        user_input = payload.get("userInput", {})
        if user_input.get("source"):
            return user_input["source"]
        meta = payload.get("meta", {})
        return meta.get("source")

    def _extract_data(self, payload: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        user_ack = payload.get("userAck", {})
        if user_ack.get("data"):
            return user_ack["data"]
        meta = payload.get("meta", {})
        return meta.get("data")

    # =================================================================================
    # formSpec 辅助方法 (user_input 表单模式)
    # =================================================================================

    def _extract_form_spec(self, payload: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        user_input = payload.get("userInput", {})
        return user_input.get("formSpec")

    def _get_form_elements(self, form_spec: Optional[Dict[str, Any]]) -> List[Dict[str, Any]]:
        if not form_spec:
            return []
        ui_schema = form_spec.get("ui_schema", {})
        return ui_schema.get("elements", [])

    def _get_form_initial_values(self, form_spec: Optional[Dict[str, Any]]) -> Optional[Dict[str, Any]]:
        if not form_spec:
            return None
        return form_spec.get("initialValues")

    def _get_field_key(self, elem: Dict[str, Any]) -> str:
        return elem.get("field", "")

    def _get_field_label(self, elem: Dict[str, Any], field: str) -> str:
        return elem.get("label", field)

    def _get_field_widget(self, elem: Dict[str, Any]) -> str:
        return elem.get("widget", "input")

    def _get_field_placeholder(self, elem: Dict[str, Any]) -> str:
        return elem.get("placeholder", "")

    def _get_initial_value(self, initial_values: Optional[Dict[str, Any]], field: str) -> str:
        if not initial_values:
            return ""
        val = initial_values.get(field)
        return str(val) if val is not None else ""

    def _get_field_enum(self, form_spec: Optional[Dict[str, Any]], field: str) -> List[str]:
        if not form_spec:
            return []
        schema = form_spec.get("schema", {})
        properties = schema.get("properties", {})
        prop = properties.get(field, {})
        enum_vals = prop.get("enum", [])
        return [str(v) for v in enum_vals]

    @staticmethod
    def is_interactive_event(event: Dict[str, Any]) -> bool:
        """检查事件是否为交互事件 / Check if event is interactive"""
        if not event:
            return False
        return event.get("type") == EventType.INTERACTIVE.value

    @staticmethod
    def extract_interactive_events(message: Dict[str, Any]) -> List[Dict[str, Any]]:
        """从消息中提取交互事件 / Extract interactive events from message"""
        events = []
        if not message:
            return events

        msg_events = message.get("events", [])
        for event in msg_events:
            if InteractiveHandler.is_interactive_event(event):
                events.append(event)
        return events