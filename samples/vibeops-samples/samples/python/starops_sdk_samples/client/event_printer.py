"""
事件打印器 - 打印每个 SSE 事件的详细信息
Event printer - prints detailed information for each SSE event
"""

import json
from typing import Any, Dict, List, Optional, Protocol


class ChatEventProtocol(Protocol):
    """Protocol for ChatEvent to avoid circular imports"""
    body: Optional[Dict[str, Any]]
    raw_json: str
    is_done: bool
    error: Optional[Exception]


class EventPrinter:
    """事件打印器 / Event printer"""

    def __init__(self, print_raw_body: bool = False, print_parsed: bool = True):
        self._print_raw_body = print_raw_body
        self._print_parsed = print_parsed
        self._print_separator = True

    def print_event(self, event: Any, event_index: int) -> None:
        """打印事件 / Print event"""
        if event.error is not None:
            print(f"\n❌ 错误: {event.error}")
            return

        if event.is_done and event.body is None:
            print("\n✅ 对话完成")
            return

        if event.body is None:
            return

        if self._print_separator:
            sep = "=" * 30
            print(f"\n{sep} 事件 #{event_index} {sep}")

        if self._print_raw_body and event.raw_json:
            print("\n📦 原始 Body:")
            print(_pretty_json(event.raw_json))

        if self._print_parsed:
            self._print_parsed_event(event.body)

    def _print_parsed_event(self, body: Dict[str, Any]) -> None:
        """打印解析后的事件详情 / Print parsed event details"""
        print("\n📋 解析详情:")

        for msg in body.get("messages", []):
            if not isinstance(msg, dict):
                continue
            print(f"  原始消息: {json.dumps(msg, ensure_ascii=False)}")
            self._print_message_item(msg)

    def _print_message_item(self, item: Dict[str, Any]) -> None:
        """打印消息条目 / Print message item"""
        role = item.get("role", "")
        if role:
            print(f"  📌 角色: {role}")

        call_id = item.get("callId", "")
        if call_id:
            print(f"  🔗 CallID: {call_id}")

        parent_call_id = item.get("parentCallId", "")
        if parent_call_id:
            print(f"  🔗 ParentCallID: {parent_call_id}")

        # 内容
        contents = item.get("contents", [])
        if contents:
            print("  📝 内容:")
            for i, content in enumerate(contents):
                if not isinstance(content, dict):
                    continue
                print(f"    [{i}] 类型: {content.get('type', '')}")
                value = content.get("value", "")
                if value:
                    if len(value) > 200:
                        value = value[:200] + "..."
                    print(f"        值: {value}")
                if content.get("append"):
                    print("        追加: true")
                if content.get("lastChunk"):
                    print("        最后块: true")

        # 工具调用
        tools = item.get("tools", [])
        if tools:
            print("  🔧 工具调用:")
            for i, tool in enumerate(tools):
                if not isinstance(tool, dict):
                    continue
                name = tool.get("name", "")
                status = tool.get("status", "")
                print(f"    [{i}] 名称: {name}, 状态: {status}")
                tool_call_id = tool.get("toolCallId", "")
                if tool_call_id:
                    print(f"        ToolCallID: {tool_call_id}")
                arguments = tool.get("arguments")
                if arguments is not None:
                    args_str = json.dumps(arguments, ensure_ascii=False)
                    if len(args_str) > 200:
                        args_str = args_str[:200] + "..."
                    print(f"        参数: {args_str}")

        # Agent 调用
        agents = item.get("agents", [])
        if agents:
            print("  🤖 Agent调用:")
            for i, agent in enumerate(agents):
                if not isinstance(agent, dict):
                    continue
                name = agent.get("name", "")
                status = agent.get("status", "")
                print(f"    [{i}] 名称: {name}, 状态: {status}")

        # 事件
        events = item.get("events", [])
        if events:
            print("  📢 事件:")
            for i, evt in enumerate(events):
                if not isinstance(evt, dict):
                    continue
                evt_type = evt.get("type", "")
                print(f"    [{i}] 类型: {evt_type}")
                payload = evt.get("payload")
                if payload is not None:
                    self._print_event_payload(evt_type, payload)

    def _print_event_payload(self, evt_type: str, payload: Any) -> None:
        """打印事件负载 / Print event payload"""
        if evt_type == "thinking":
            if isinstance(payload, dict):
                delta = payload.get("reasoningDelta", "")
                if delta:
                    if len(delta) > 100:
                        delta = delta[:100] + "..."
                    print(f"        思考: {delta}")

        elif evt_type == "error":
            if isinstance(payload, dict):
                print(f"        错误码: {payload.get('code', '')}")
                print(f"        消息: {payload.get('message', '')}")

        elif evt_type == "task_finished":
            if isinstance(payload, dict):
                print(f"        成功: {payload.get('success', False)}")
                statistics = payload.get("statistics")
                if isinstance(statistics, dict):
                    duration_ns = statistics.get("duration", 0)
                    print(f"        耗时: {duration_ns // 1000000}ms")

        else:
            payload_str = json.dumps(payload, ensure_ascii=False)
            if len(payload_str) > 200:
                payload_str = payload_str[:200] + "..."
            print(f"        负载: {payload_str}")


def _pretty_json(json_str: str) -> str:
    """格式化 JSON 输出 / Pretty print JSON"""
    try:
        data = json.loads(json_str)
        return json.dumps(data, indent=2, ensure_ascii=False)
    except (json.JSONDecodeError, TypeError):
        return json_str
