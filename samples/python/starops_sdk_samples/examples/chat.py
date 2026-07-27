#!/usr/bin/env python3
"""
交互式对话示例
Interactive chat example

Usage: python -m starops_sdk_samples.examples.chat
"""

import asyncio
import json
import os
import signal
import sys
from typing import Optional

from ..client import AgentClient, Config, SDKException, SimplePrinter, InteractiveHandler, InteractiveResponse


async def main_async():
    print("🚀 STAROps Chat (Python)")
    print("=" * 60)

    simulate_error = "-simulate-error" in sys.argv

    try:
        # Load configuration
        cfg = Config.load_from_env()
        print(f"📋 Employee: {cfg.employee_name}\n")

        if simulate_error:
            cfg.simulate_network_error = True
            print("⚠️  已启用网络断连模拟，将在收到首个事件后触发重试")

        # Create client
        client = AgentClient(cfg)

        # Create thread
        print("📝 创建会话...")
        thread_id = client.create_thread()
        print(f"✅ ThreadID: {thread_id}\n")

        # 捕获中断信号，发送 stop 请求
        def _stop_handler():
            print("\n⏹️  正在停止对话...")
            try:
                client.stop(thread_id)
            except Exception as e:
                sys.stderr.write(f"stop 请求失败: {e}\n")
            os._exit(0)

        loop = asyncio.get_running_loop()
        for sig in (signal.SIGINT, signal.SIGTERM):
            loop.add_signal_handler(sig, _stop_handler)

        # Create printer
        printer = SimplePrinter()
        interactive_handler = InteractiveHandler(client)

        # Interactive loop
        while True:
            try:
                user_input = input("👤 请输入 (quit 退出): ")
            except EOFError:
                print("\n👋 再见!")
                break

            user_input = user_input.strip()
            if not user_input:
                continue
            if user_input in ("quit", "exit"):
                print("👋 再见!")
                break

            print("-" * 60)

            # Send message
            printer.reset()
            events = client.chat(thread_id, user_input)
            while events is not None:
                try:
                    event = await events.__anext__()
                except StopAsyncIteration:
                    break

                if event.has_error():
                    print(f"❌ 错误: {event.error}")
                    continue

                # 正常输出（先输出）
                text = printer.process_event(event)
                if text:
                    print(text, end="", flush=True)

                # 检测交互事件（在输出之后）
                interactive_resp = _extract_interactive_event(event, interactive_handler)
                if interactive_resp:
                    events = interactive_handler.resume_chat(thread_id, interactive_resp)
                    continue
                    continue

            print()
            print("=" * 60)
            print()

    except SDKException as e:
        print(f"❌ 配置加载失败: {e}")
        print("\n请设置环境变量:")
        print("  STAROPS_ENDPOINT")
        print("  ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET")
        sys.exit(1)
    except Exception as e:
        print(f"❌ 错误: {e}")
        sys.exit(1)


def _extract_interactive_event(event, handler: InteractiveHandler) -> Optional[InteractiveResponse]:
    """从 ChatEvent 中检测交互事件并处理用户响应"""
    if not event.raw_json:
        return None
    try:
        body = json.loads(event.raw_json)
        for msg in body.get("messages", []):
            for evt in msg.get("events", []):
                if InteractiveHandler.is_interactive_event(evt):
                    call_id = msg.get("callId", "")
                    return handler.handle_event(evt, call_id)
    except Exception as e:
        print(f"⚠️ 交互事件解析失败: {e}")
    return None


def main():
    asyncio.run(main_async())


if __name__ == "__main__":
    main()