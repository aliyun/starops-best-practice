"""
mock.py — Mock/Record SDK 层
在 open_sse_stream 层面拦截，对上层透明
"""
import json
import os
from typing import AsyncGenerator, Any, Callable, Awaitable

STREAM_SEPARATOR = "---"
_mock_offset = 0  # 多流回放的文件偏移量


async def open_sse_stream(
    config,
    request: Any,
    real_stream_fn: Callable[[], AsyncGenerator[Any, None]],
) -> AsyncGenerator[Any, None]:
    if config.mock_mode:
        async for evt in mock_stream(config):
            yield evt
        return
    if config.record_mode:
        async for evt in record_stream(config, real_stream_fn):
            yield evt
        return
    async for evt in real_stream_fn():
        yield evt


async def mock_stream(config) -> AsyncGenerator[Any, None]:
    global _mock_offset
    if not config.mock_file:
        raise ValueError("mock_file is required")

    with open(config.mock_file, "r") as f:
        if _mock_offset > 0:
            f.seek(_mock_offset)
        bytes_read = 0
        for line in f:
            bytes_read += len(line.encode())
            line = line.strip()
            if line == STREAM_SEPARATOR:
                _mock_offset += bytes_read
                return
            if not line:
                continue
            try:
                evt = json.loads(line)
                yield {
                    "id": evt.get("id", ""),
                    "event": evt.get("event", "message"),
                    "body": evt.get("data"),
                    "statusCode": 200,
                }
            except json.JSONDecodeError:
                continue


async def record_stream(
    config,
    real_stream_fn: Callable[[], AsyncGenerator[Any, None]],
) -> AsyncGenerator[Any, None]:
    if not config.mock_file:
        raise ValueError("mock_file is required")

    with open(config.mock_file, "a") as f:
        async for resp in real_stream_fn():
            evt = {
                "id": getattr(resp, "id", "") or "",
                "event": getattr(resp, "event", "") or "message",
                "data": getattr(resp, "body", None) or resp,
            }
            f.write(json.dumps(evt, ensure_ascii=False) + "\n")
            f.flush()
            yield resp
        f.write(STREAM_SEPARATOR + "\n")
        f.flush()