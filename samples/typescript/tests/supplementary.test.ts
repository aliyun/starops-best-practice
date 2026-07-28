/**
 * Supplementary tests for SimplePrinter, EventPrinter, and config
 * 补充测试：简洁打印器、事件打印器、配置
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SimplePrinter } from '../src/client/simple-printer.js';
import { EventPrinter } from '../src/client/event-printer.js';
import type { ChatEvent } from '../src/types/events.js';

// ============================================================
// SimplePrinter 补充测试
// ============================================================

describe('SimplePrinter - getFinalText / reset / dedup', () => {
  function makeTextEvent(text: string): ChatEvent {
    return {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [{ parts: [{ kind: 'text', text }] }],
          },
        ],
      },
      rawJson: '',
      statusCode: 200,
      isDone: false,
    };
  }

  it('getFinalText 累积多次事件文本', () => {
    const printer = new SimplePrinter();
    printer.processEvent(makeTextEvent('Hello '));
    printer.processEvent(makeTextEvent('World'));
    expect(printer.getFinalText()).toBe('Hello World');
  });

  it('reset 清空缓冲区和去重表', () => {
    const printer = new SimplePrinter();
    const event = makeTextEvent('content');
    printer.processEvent(event);
    expect(printer.getFinalText()).toBe('content');

    printer.reset();
    expect(printer.getFinalText()).toBe('');

    // reset 后相同内容应能再次提取
    printer.processEvent(event);
    expect(printer.getFinalText()).toBe('content');
  });

  it('重复内容被去重过滤', () => {
    const printer = new SimplePrinter();
    const event = makeTextEvent('重复');
    expect(printer.processEvent(event)).toBe('重复');
    expect(printer.processEvent(event)).toBe('');
  });

  it('跳过非 system 角色消息', () => {
    const printer = new SimplePrinter();
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'assistant',
            artifacts: [{ parts: [{ kind: 'text', text: 'ignored' }] }],
          },
        ],
      },
      rawJson: '',
      statusCode: 200,
      isDone: false,
    };
    expect(printer.processEvent(event)).toBe('');
  });

  it('跳过非 text/task_finished 事件类型', () => {
    const printer = new SimplePrinter();
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [{ parts: [{ kind: 'text', text: 'skipped' }] }],
          },
        ],
      },
      rawJson: '',
      statusCode: 200,
      isDone: false,
      event: 'thinking',
    };
    expect(printer.processEvent(event)).toBe('');
  });

  it('处理 null/undefined 事件', () => {
    const printer = new SimplePrinter();
    expect(printer.processEvent(null)).toBe('');
    expect(printer.processEvent(undefined)).toBe('');
  });
});

// ============================================================
// EventPrinter 测试
// ============================================================

describe('EventPrinter', () => {
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  it('打印错误事件', () => {
    const printer = new EventPrinter();
    printer.printEvent({ body: null, rawJson: '', statusCode: 500, isDone: false, error: 'boom' }, 1);
    expect(consoleSpy).toHaveBeenCalledWith(expect.stringContaining('boom'));
  });

  it('打印对话完成', () => {
    const printer = new EventPrinter();
    printer.printEvent({ body: null, rawJson: '', statusCode: 200, isDone: true }, 1);
    expect(consoleSpy).toHaveBeenCalledWith(expect.stringContaining('对话完成'));
  });

  it('跳过无 body 且非 done 的事件', () => {
    const printer = new EventPrinter();
    printer.printEvent({ body: null, rawJson: '', statusCode: 200, isDone: false }, 1);
    expect(consoleSpy).not.toHaveBeenCalled();
  });

  it('打印解析详情（角色、内容）', () => {
    const printer = new EventPrinter(false, true);
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'assistant',
            contents: [{ type: 'text', value: 'hello' }],
          },
        ],
      },
      rawJson: '',
      statusCode: 200,
      isDone: false,
    };
    printer.printEvent(event, 1);
    const allOutput = consoleSpy.mock.calls.map((c) => c[0]).join('\n');
    expect(allOutput).toContain('角色: assistant');
    expect(allOutput).toContain('hello');
  });

  it('打印原始 JSON（printRawBody=true）', () => {
    const printer = new EventPrinter(true, false);
    const event: ChatEvent = {
      body: { messages: [] },
      rawJson: '{"test":true}',
      statusCode: 200,
      isDone: false,
    };
    printer.printEvent(event, 1);
    const allOutput = consoleSpy.mock.calls.map((c) => c[0]).join('\n');
    expect(allOutput).toContain('原始 Body');
    expect(allOutput).toContain('"test": true');
  });

  it('打印工具调用信息', () => {
    const printer = new EventPrinter(false, true);
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'assistant',
            tools: [{ name: 'search', status: 'running', toolCallId: 'tc-1', arguments: { q: 'test' } }],
          },
        ],
      },
      rawJson: '',
      statusCode: 200,
      isDone: false,
    };
    printer.printEvent(event, 1);
    const allOutput = consoleSpy.mock.calls.map((c) => c[0]).join('\n');
    expect(allOutput).toContain('search');
    expect(allOutput).toContain('running');
  });
});
