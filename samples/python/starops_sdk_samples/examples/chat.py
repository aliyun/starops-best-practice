#!/usr/bin/env python3
"""
交互式对话示例
Interactive chat example

Usage: python -m starops_sdk_samples.examples.chat
"""

import argparse
import asyncio
import json
import os
import signal
import sys
from typing import Optional

from ..client import (
    AgentClient,
    Config,
    SDKException,
    ErrorCode,
    SimplePrinter,
    InteractiveHandler,
    InteractiveResponse,
)
from ..client import (
    ModelsConfig,
    load_models,
    find_config_path,
    parse_model_flag,
    validate_model,
    build_model_json,
    display_menu,
    list_model_flags,
    get_model_by_index,
)


def _parse_args() -> argparse.Namespace:
    """解析命令行参数 / Parse CLI arguments"""
    parser = argparse.ArgumentParser(
        prog="chat",
        description="STAROps 交互式对话示例",
    )
    parser.add_argument(
        "--simulate-error",
        action="store_true",
        help="模拟网络断连，测试重试逻辑",
    )
    parser.add_argument(
        "--model",
        default="",
        help="指定模型 (格式: provider:modelId)",
    )
    parser.add_argument(
        "--list-models",
        action="store_true",
        help="列出所有可传入 --model 的模型取值后退出",
    )
    return parser.parse_args()


async def main_async(args: argparse.Namespace, models_cfg: Optional[ModelsConfig]):
    print("🚀 STAROps Chat (Python)")
    print("=" * 60)

    simulate_error = args.simulate_error

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
        thread_attrs: Optional[dict] = None
        if args.model:
            provider, model_id = parse_model_flag(args.model)
            if models_cfg is not None and not validate_model(
                models_cfg, provider, model_id
            ):
                raise SDKException(
                    ErrorCode.CONFIG_INVALID,
                    f"未知模型: {provider}:{model_id}，可用 --list-models 查看可选值",
                )
            thread_attrs = {"model": build_model_json(provider, model_id)}
            print(f"🤖 已指定模型: {provider}:{model_id}")
        thread_id = client.create_thread(thread_attrs)
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
            if user_input == "/model":
                if models_cfg is None:
                    print("⚠️  模型配置未加载")
                    continue
                print(display_menu(models_cfg))
                try:
                    choice = input("请输入序号选择模型: ")
                except EOFError:
                    print("\n👋 再见!")
                    break
                try:
                    idx = int(choice.strip())
                except ValueError:
                    print("❌ 无效输入")
                    continue
                try:
                    provider, model_id = get_model_by_index(models_cfg, idx - 1)
                except ValueError as e:
                    print("❌", e)
                    continue
                try:
                    client.update_thread(
                        thread_id, {"model": build_model_json(provider, model_id)}
                    )
                except Exception as e:
                    print(f"❌ 更新失败: {e}")
                    continue
                print("✅ 更新模型成功")
                continue

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
    args = _parse_args()

    # 加载模型配置（不依赖凭据）
    models_cfg: Optional[ModelsConfig] = None
    try:
        models_cfg = load_models(find_config_path())
    except Exception:
        models_cfg = None

    # --list-models 在加载凭据之前短路，不需要凭据
    if args.list_models:
        if models_cfg is None:
            print("⚠️  模型配置未加载")
            return
        print(list_model_flags(models_cfg))
        return

    asyncio.run(main_async(args, models_cfg))


if __name__ == "__main__":
    main()