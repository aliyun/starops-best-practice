#!/usr/bin/env python3
"""
会话管理工具
Thread management tool

Usage:
    python -m starops_sdk_samples.examples.thread_manager list
    python -m starops_sdk_samples.examples.thread_manager get <thread-id>
    python -m starops_sdk_samples.examples.thread_manager messages <thread-id>
    python -m starops_sdk_samples.examples.thread_manager delete <thread-id>
"""

import sys

from ..client import AgentClient, Config, SDKException


def list_threads(client: AgentClient):
    print("📋 会话列表")
    print("-" * 100)

    threads, total = client.list_threads(20)
    print(f"共 {total} 个会话:\n")

    if not threads:
        print("  (无会话)")
        return

    print(f"{'Thread ID':<40} {'标题':<25} {'状态':<10} 创建时间")
    print("-" * 100)

    for t in threads:
        title = t.title
        if title and len(title) > 23:
            title = title[:23] + "..."
        print(f"{t.thread_id:<40} {title or '':<25} {t.status or '':<10} {t.create_time or ''}")


def get_thread(client: AgentClient, thread_id: str):
    print(f"📋 会话详情: {thread_id}")
    print("-" * 60)

    detail = client.get_thread(thread_id)

    print(f"  Thread ID: {detail.thread_id}")
    print(f"  标题: {detail.title}")
    print(f"  状态: {detail.status}")
    print(f"  创建时间: {detail.create_time}")
    print(f"  更新时间: {detail.update_time}")


def list_messages(client: AgentClient, thread_id: str):
    print(f"💬 会话消息: {thread_id}")
    print("-" * 80)

    messages = client.get_thread_data(thread_id, 50)
    print(f"共 {len(messages)} 条消息:\n")

    for i, m in enumerate(messages):
        role_icon = "👤"
        if m.role == "assistant":
            role_icon = "🤖"
        elif m.role == "system":
            role_icon = "⚙️"

        content = m.content
        if content and len(content) > 100:
            content = content[:100] + "..."

        print(f"[{i + 1}] {role_icon} {m.role}")
        if content:
            print(f"    {content}")
        print(f"    时间戳: {m.timestamp}\n")


def delete_thread(client: AgentClient, thread_id: str):
    print(f"🗑️ 删除会话: {thread_id}")

    client.delete_thread(thread_id)
    print("✅ 会话已删除")


def print_usage():
    print("会话管理工具 - STAROps SDK 接口演示 (Python)")
    print("=" * 50)
    print("\n用法:")
    print("  python -m starops_sdk_samples.examples.thread_manager <command> [args]")
    print("\n命令:")
    print("  list                  列出所有会话")
    print("  get <thread-id>       获取会话详情")
    print("  messages <thread-id>  列出会话消息")
    print("  delete <thread-id>    删除会话")


def main():
    if len(sys.argv) < 2:
        print_usage()
        sys.exit(1)

    command = sys.argv[1]

    try:
        # Load configuration
        cfg = Config.load_from_env()

        # Create client
        client = AgentClient(cfg)

        if command == "list":
            list_threads(client)
        elif command == "get":
            if len(sys.argv) < 3:
                print("❌ 请指定 thread-id")
                sys.exit(1)
            get_thread(client, sys.argv[2])
        elif command == "messages":
            if len(sys.argv) < 3:
                print("❌ 请指定 thread-id")
                sys.exit(1)
            list_messages(client, sys.argv[2])
        elif command == "delete":
            if len(sys.argv) < 3:
                print("❌ 请指定 thread-id")
                sys.exit(1)
            delete_thread(client, sys.argv[2])
        else:
            print(f"❌ 未知命令: {command}")
            print_usage()
            sys.exit(1)

    except SDKException as e:
        print(f"❌ 错误: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
