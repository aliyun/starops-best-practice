/**
 * Tests for config loading defaults and error paths
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { loadConfigFromEnv } from '../src/client/config.js';
import { SDKException, ErrorCode } from '../src/client/errors.js';

const ENV_KEYS = [
  'STAROPS_WORKSPACE',
  'STAROPS_ENDPOINT',
  'STAROPS_REGION',
  'STAROPS_EMPLOYEE_NAME',
  'ALIBABA_CLOUD_ACCESS_KEY_ID',
  'ALIBABA_CLOUD_ACCESS_KEY_SECRET',
];

describe('loadConfigFromEnv', () => {
  const originalEnv: Record<string, string | undefined> = {};

  beforeEach(() => {
    for (const key of ENV_KEYS) {
      originalEnv[key] = process.env[key];
      // 设为空字符串：既模拟"未设置"（|| 兜底生效），又阻止 dotenv 从 .env 文件覆盖
      process.env[key] = '';
    }
  });

  afterEach(() => {
    for (const key of ENV_KEYS) {
      const val = originalEnv[key];
      if (val === undefined) {
        delete process.env[key];
      } else {
        process.env[key] = val;
      }
    }
  });

  it('should default employeeName to apsara-ops when STAROPS_EMPLOYEE_NAME is unset', async () => {
    process.env.STAROPS_ENDPOINT = 'https://example.com';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = 'test-ak-id';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET = 'test-ak-secret';

    const config = await loadConfigFromEnv();
    expect(config.employeeName).toBe('apsara-ops');
  });

  it('should default region to cn-beijing when STAROPS_REGION is unset', async () => {
    process.env.STAROPS_ENDPOINT = 'https://example.com';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = 'test-ak-id';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET = 'test-ak-secret';

    const config = await loadConfigFromEnv();
    expect(config.region).toBe('cn-beijing');
  });

  it('should respect explicitly set region and employeeName', async () => {
    process.env.STAROPS_ENDPOINT = 'https://example.com';
    process.env.STAROPS_REGION = 'cn-shanghai';
    process.env.STAROPS_EMPLOYEE_NAME = 'my-employee';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = 'test-ak-id';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET = 'test-ak-secret';

    const config = await loadConfigFromEnv();
    expect(config.region).toBe('cn-shanghai');
    expect(config.employeeName).toBe('my-employee');
  });

  it('should throw CONFIG_MISSING when STAROPS_ENDPOINT is missing', async () => {
    process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = 'test-ak-id';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET = 'test-ak-secret';

    try {
      await loadConfigFromEnv();
      expect.fail('expected loadConfigFromEnv to throw');
    } catch (e) {
      expect(e).toBeInstanceOf(SDKException);
      const ex = e as SDKException;
      expect(ex.code).toBe(ErrorCode.CONFIG_MISSING);
      expect(ex.message).toContain('STAROPS_ENDPOINT');
    }
  });
});
