/**
 * Tests for SDK exceptions
 */

import { describe, it, expect } from 'vitest';
import { SDKException, ErrorCode } from '../src/client/errors.js';

describe('SDKException', () => {
  it('should create config missing error', () => {
    const missingVars = ['VAR1', 'VAR2'];
    const ex = SDKException.configMissing(missingVars);

    expect(ex.code).toBe(ErrorCode.CONFIG_MISSING);
    expect(ex.message).toContain('VAR1');
    expect(ex.message).toContain('VAR2');
    expect(ex.suggestion).toBeDefined();
    expect(ex.context.missingVariables).toEqual(missingVars);
  });

  it('should create config invalid error', () => {
    const ex = SDKException.configInvalid('endpoint', 'invalid URL');

    expect(ex.code).toBe(ErrorCode.CONFIG_INVALID);
    expect(ex.message).toContain('endpoint');
    expect(ex.message).toContain('invalid URL');
    expect(ex.context.field).toBe('endpoint');
  });

  it('should create thread not found error', () => {
    const ex = SDKException.threadNotFound('thread-123');

    expect(ex.code).toBe(ErrorCode.THREAD_NOT_FOUND);
    expect(ex.message).toContain('thread-123');
    expect(ex.context.threadId).toBe('thread-123');
  });

  it('should create timeout error', () => {
    const ex = SDKException.timeout('30s');

    expect(ex.code).toBe(ErrorCode.TIMEOUT);
    expect(ex.message).toContain('30s');
  });

  it('should add context', () => {
    const ex = new SDKException(ErrorCode.API_ERROR, 'test error')
      .withContext('key1', 'value1')
      .withContext('key2', 123);

    expect(ex.context.key1).toBe('value1');
    expect(ex.context.key2).toBe(123);
  });

  it('should add suggestion', () => {
    const ex = new SDKException(ErrorCode.API_ERROR, 'test error')
      .withSuggestion('Try again later');

    expect(ex.suggestion).toBe('Try again later');
  });

  it('should format toString correctly', () => {
    const ex = new SDKException(ErrorCode.API_ERROR, 'test error');
    const str = ex.toString();

    expect(str).toContain('API_ERROR');
    expect(str).toContain('test error');
  });

  it('should include cause in toString', () => {
    const cause = new Error('root cause');
    const ex = new SDKException(ErrorCode.NETWORK_ERROR, 'network failed', cause);

    expect(ex.cause).toBe(cause);
    expect(ex.toString()).toContain('root cause');
  });
});
