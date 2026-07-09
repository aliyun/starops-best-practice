/**
 * Tests for isDoneMessage
 * isDoneMessage 函数测试
 */

import { describe, it, expect } from 'vitest';
import { isDoneMessage } from '../src/client/agent-client.js';

describe('isDoneMessage', () => {
  it('event 字段为 done 时返回 true', () => {
    const body = { event: 'done', messages: [] };
    expect(isDoneMessage(body)).toBe(true);
  });

  it('无 event 字段时 fallback 到 messages[].type', () => {
    const body = { messages: [{ type: 'done' }] };
    expect(isDoneMessage(body)).toBe(true);
  });

  it('非 done 事件返回 false', () => {
    const body = { event: 'text', messages: [{ type: 'text' }] };
    expect(isDoneMessage(body)).toBe(false);
  });

  it('body 为 undefined 时返回 false', () => {
    expect(isDoneMessage(undefined)).toBe(false);
  });
});
