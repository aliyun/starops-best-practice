/**
 * Tests for SimplePrinter
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { SimplePrinter } from '../src/client/simple-printer.js';
import { ChatEvent } from '../src/client/agent-client.js';

describe('SimplePrinter', () => {
  let printer: SimplePrinter;

  beforeEach(() => {
    printer = new SimplePrinter();
  });

  it('should return empty string for null event', () => {
    const result = printer.processEvent(null);
    expect(result).toBe('');
  });

  it('should return empty string for event with null body', () => {
    const event: ChatEvent = {
      rawJson: '',
      statusCode: 200,
      isDone: false,
    };
    const result = printer.processEvent(event);
    expect(result).toBe('');
  });

  it('should extract text from system artifacts', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Hello World' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('Hello World');
    expect(printer.getFinalText()).toBe('Hello World');
  });

  it('should ignore assistant role', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'assistant',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Should be ignored' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('');
  });

  it('should deduplicate artifacts', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Duplicate' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result1 = printer.processEvent(event);
    const result2 = printer.processEvent(event);

    expect(result1).toBe('Duplicate');
    expect(result2).toBe(''); // Should be empty due to deduplication
    expect(printer.getFinalText()).toBe('Duplicate');
  });

  it('should reset buffer', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Test' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    printer.processEvent(event);
    expect(printer.getFinalText()).toBe('Test');

    printer.reset();
    expect(printer.getFinalText()).toBe('');

    // After reset, same content should be processed again
    const result = printer.processEvent(event);
    expect(result).toBe('Test');
  });

  it('should handle multiple artifacts', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Part1' }] },
              { parts: [{ kind: 'text', text: 'Part2' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('Part1Part2');
  });

  it('should return empty string for non-text/non-task_finished event type', () => {
    const event: ChatEvent = {
      event: 'interaction',
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Should be skipped' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('');
  });

  it('should process event with event field set to text', () => {
    const event: ChatEvent = {
      event: 'text',
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'Text event content' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('Text event content');
  });

  it('should process event with no event field (fallback)', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'text', text: 'No event field' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('No event field');
  });

  it('should ignore non-text kind', () => {
    const event: ChatEvent = {
      body: {
        messages: [
          {
            role: 'system',
            artifacts: [
              { parts: [{ kind: 'image', url: 'http://example.com' }] },
            ],
          },
        ],
      },
      rawJson: '{}',
      statusCode: 200,
      isDone: false,
    };

    const result = printer.processEvent(event);

    expect(result).toBe('');
  });
});
