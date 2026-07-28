/**
 * Tests for InteractiveHandler
 * 参考 golang/pkg/interactive/interactive_test.go：
 * 覆盖交互事件识别与 user_ack / user_input 两类事件的 happy path
 */

import { describe, it, expect } from 'vitest';
import { InteractiveHandler, ChatClient, InteractiveResponse } from '../src/client/interactive-handler.js';
import { EventType, InteractionType } from '../src/types/index.js';
import type { ChatEvent } from '../src/types/events.js';

/** 模拟 ChatClient / Mock chat client */
class MockChatClient implements ChatClient {
  interactCalled = false;
  lastThreadId = '';
  lastUserInteractive = '';

  async *interact(
    threadId: string,
    userInteractive: string
  ): AsyncIterable<ChatEvent> {
    this.interactCalled = true;
    this.lastThreadId = threadId;
    this.lastUserInteractive = userInteractive;
    yield { rawJson: '', statusCode: 200, isDone: true };
  }
}

describe('InteractiveHandler.isInteractiveEvent', () => {
  it('should return true for interactive event', () => {
    expect(InteractiveHandler.isInteractiveEvent({ type: EventType.INTERACTIVE })).toBe(true);
  });

  it('should return false for thinking event', () => {
    expect(InteractiveHandler.isInteractiveEvent({ type: EventType.THINKING })).toBe(false);
  });

  it('should return false for null/undefined event', () => {
    expect(InteractiveHandler.isInteractiveEvent(null as unknown as Record<string, unknown>)).toBe(false);
    expect(InteractiveHandler.isInteractiveEvent(undefined as unknown as Record<string, unknown>)).toBe(false);
  });
});

describe('InteractiveHandler.extractInteractiveEvents', () => {
  it('should extract only interactive events from mixed events', () => {
    const message = {
      events: [
        { type: EventType.THINKING },
        { type: EventType.INTERACTIVE },
        { type: EventType.ERROR },
        { type: EventType.INTERACTIVE },
      ],
    };
    const result = InteractiveHandler.extractInteractiveEvents(message);
    expect(result).toHaveLength(2);
    for (const evt of result) {
      expect(evt.type).toBe(EventType.INTERACTIVE);
    }
  });

  it('should return empty array when message has no events', () => {
    expect(InteractiveHandler.extractInteractiveEvents({})).toEqual([]);
  });
});

describe('InteractiveHandler.handleEvent - user_ack happy path', () => {
  it('should confirm with decision yes when user inputs y', async () => {
    const handler = new InteractiveHandler(new MockChatClient());
    handler.setMockInput('y');

    const event = {
      type: EventType.INTERACTIVE,
      payload: {
        type: InteractionType.USER_ACK,
        userAck: {
          message: '是否继续执行？',
          data: { title: '确认操作' },
        },
      },
    };

    const resp = await handler.handleEvent(event, 'call-ack-1');
    expect(resp.callId).toBe('call-ack-1');
    expect(resp.type).toBe(InteractionType.USER_ACK);
    expect(resp.response.confirmed).toBe(true);
    expect(resp.decision).toBe('yes');
  });

  it('should reject with decision no when user inputs n', async () => {
    const handler = new InteractiveHandler(new MockChatClient());
    handler.setMockInput('n');

    const event = {
      type: EventType.INTERACTIVE,
      payload: {
        type: InteractionType.USER_ACK,
        userAck: { message: '是否继续执行？' },
      },
    };

    const resp = await handler.handleEvent(event, 'call-ack-2');
    expect(resp.response.confirmed).toBe(false);
    expect(resp.decision).toBe('no');
  });
});

describe('InteractiveHandler.handleEvent - user_input happy path', () => {
  it('should collect form data and return decision submit', async () => {
    const handler = new InteractiveHandler(new MockChatClient());
    handler.setMockInput('hello-world');

    const event = {
      type: EventType.INTERACTIVE,
      payload: {
        type: InteractionType.USER_INPUT,
        userInput: {
          title: '请填写表单',
          description: '需要补充信息',
          formSpec: {
            ui_schema: {
              elements: [{ field: 'name', label: '名称' }],
            },
          },
        },
      },
    };

    const resp = await handler.handleEvent(event, 'call-input-1');
    expect(resp.callId).toBe('call-input-1');
    expect(resp.type).toBe(InteractionType.USER_INPUT);
    expect(resp.decision).toBe('submit');
    expect(resp.formData).toEqual({ name: 'hello-world' });
    expect(resp.response.value).toEqual({ name: 'hello-world' });
  });
});

describe('InteractiveHandler.handleEvent - error paths', () => {
  it('should throw for non-interactive event type', async () => {
    const handler = new InteractiveHandler(new MockChatClient());
    const event = { type: EventType.THINKING, payload: {} };
    await expect(handler.handleEvent(event, 'call-x')).rejects.toThrow('不支持的事件类型');
  });

  it('should throw for missing payload', async () => {
    const handler = new InteractiveHandler(new MockChatClient());
    const event = { type: EventType.INTERACTIVE };
    await expect(handler.handleEvent(event, 'call-x')).rejects.toThrow('交互负载为空');
  });
});

describe('InteractiveHandler.resumeChat', () => {
  it('should call client.interact with thread id and serialized response', async () => {
    const mock = new MockChatClient();
    const handler = new InteractiveHandler(mock);
    const resp: InteractiveResponse = {
      callId: 'fake-call-id',
      type: InteractionType.USER_ACK,
      response: { confirmed: true },
      source: { app: 'test-app' },
      decision: 'yes',
    };

    const events: ChatEvent[] = [];
    for await (const evt of handler.resumeChat('fake-thread-id', resp)) {
      events.push(evt);
    }

    expect(mock.interactCalled).toBe(true);
    expect(mock.lastThreadId).toBe('fake-thread-id');
    const ui = JSON.parse(mock.lastUserInteractive);
    expect(ui.callId).toBe('fake-call-id');
    expect(ui.decision).toBe('yes');
    expect(events).toHaveLength(1);
    expect(events[0].isDone).toBe(true);
  });

  it('should yield error event for null response', async () => {
    const handler = new InteractiveHandler(new MockChatClient());
    const events: ChatEvent[] = [];
    for await (const evt of handler.resumeChat('fake-thread-id', null as unknown as InteractiveResponse)) {
      events.push(evt);
    }
    expect(events).toHaveLength(1);
    expect(events[0].error).toBeDefined();
  });
});
