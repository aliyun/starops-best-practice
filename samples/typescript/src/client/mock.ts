// mock.ts — Mock/Record SDK 层
// 在 openSSEStream 层面拦截，对上层 streamSSE/streamOnce 透明
import * as fs from 'fs';
import * as readline from 'readline';
import type { Config } from './config.js';
import type { ChatEvent } from '../types/events.js';

const STREAM_SEPARATOR = '---';
let mockOffset = 0; // 多流回放的文件偏移量

interface MockSSEEvent {
  id: string;
  event: string;
  data: unknown;
}

// 打开 SSE 流（支持 Mock/Record 模式）
export async function* openSSEStream(
  config: Config,
  request: unknown,
  realStreamFn: () => AsyncGenerator<unknown>
): AsyncGenerator<unknown> {
  if (config.mockMode) {
    yield* mockStream(config);
    return;
  }
  if (config.recordMode) {
    yield* recordStream(config, realStreamFn);
    return;
  }
  yield* realStreamFn();
}

// 从 JSONL 文件读取事件并回放
async function* mockStream(config: Config): AsyncGenerator<unknown> {
  if (!config.mockFile) throw new Error('mockFile is required');

  const fileStream = fs.createReadStream(config.mockFile, { start: mockOffset });
  const rl = readline.createInterface({ input: fileStream, crlfDelay: Infinity });

  let bytesRead = 0;
  for await (const line of rl) {
    bytesRead += Buffer.byteLength(line) + 1;

    if (line.trim() === STREAM_SEPARATOR) {
      mockOffset += bytesRead;
      return;
    }
    if (!line.trim()) continue;

    try {
      const evt: MockSSEEvent = JSON.parse(line);
      yield {
        id: evt.id,
        event: evt.event,
        body: evt.data,
        statusCode: 200,
      };
    } catch {
      // skip invalid lines
    }
  }
}

// 代理真实 SDK 并旁路写入 JSONL 文件
async function* recordStream(
  config: Config,
  realStreamFn: () => AsyncGenerator<unknown>
): AsyncGenerator<unknown> {
  if (!config.mockFile) throw new Error('mockFile is required');

  const ws = fs.createWriteStream(config.mockFile, { flags: 'a' });

  try {
    for await (const resp of realStreamFn()) {
      const r = resp as Record<string, unknown>;
      const evt: MockSSEEvent = {
        id: (r.id as string) || '',
        event: (r.event as string) || 'message',
        data: r.body || r,
      };
      ws.write(JSON.stringify(evt) + '\n');
      yield resp;
    }
    ws.write(STREAM_SEPARATOR + '\n');
  } finally {
    ws.end();
  }
}