/**
 * Tests for retry utility functions
 * 重试工具函数测试
 */

import { describe, it, expect } from 'vitest';
import {
  isNewerTimestamp,
  extractNewestTimestamp,
  calculateBackoff,
  defaultRetryConfig,
} from '../src/client/retry.js';
import type { ChatEvent } from '../src/client/agent-client.js';

function makeEvent(messages: Array<Record<string, unknown>>): ChatEvent {
  return {
    body: { messages },
    rawJson: '',
    statusCode: 200,
    isDone: false,
  };
}

describe('isNewerTimestamp', () => {
  it('ts 为空时返回 false', () => {
    expect(isNewerTimestamp('', '100')).toBe(false);
  });

  it('base 为空时返回 true', () => {
    expect(isNewerTimestamp('100', '')).toBe(true);
  });

  it('数值比较：更大的时间戳更新', () => {
    expect(isNewerTimestamp('200', '100')).toBe(true);
    expect(isNewerTimestamp('100', '200')).toBe(false);
    expect(isNewerTimestamp('100', '100')).toBe(false);
  });

  it('base 为数值但 ts 无法解析时返回 false', () => {
    expect(isNewerTimestamp('abc', '100')).toBe(false);
  });

  it('非数值时回退到字符串比较', () => {
    expect(isNewerTimestamp('b', 'a')).toBe(true);
    expect(isNewerTimestamp('a', 'b')).toBe(false);
  });
});

describe('extractNewestTimestamp', () => {
  it('无 messages 时返回空字符串', () => {
    const event: ChatEvent = { body: {}, rawJson: '', statusCode: 200, isDone: false };
    expect(extractNewestTimestamp(event, '')).toBe('');
  });

  it('提取比 base 更新的最大时间戳', () => {
    const event = makeEvent([{ timestamp: '100' }, { timestamp: '300' }, { timestamp: '200' }]);
    expect(extractNewestTimestamp(event, '150')).toBe('300');
  });

  it('没有比 base 更新的时间戳时返回空字符串', () => {
    const event = makeEvent([{ timestamp: '100' }, { timestamp: '120' }]);
    expect(extractNewestTimestamp(event, '200')).toBe('');
  });

  it('base 为空时返回最大时间戳', () => {
    const event = makeEvent([{ timestamp: '100' }, { timestamp: '300' }]);
    expect(extractNewestTimestamp(event, '')).toBe('300');
  });
});

describe('calculateBackoff', () => {
  const config = defaultRetryConfig();

  it('第一次重试为初始退避', () => {
    expect(calculateBackoff(1, config)).toBe(1000);
  });

  it('按退避系数指数增长', () => {
    expect(calculateBackoff(2, config)).toBe(2000);
    expect(calculateBackoff(3, config)).toBe(4000);
    expect(calculateBackoff(4, config)).toBe(8000);
  });

  it('不超过最大退避时间', () => {
    expect(calculateBackoff(10, config)).toBe(30000);
    expect(calculateBackoff(100, config)).toBe(30000);
  });
});
