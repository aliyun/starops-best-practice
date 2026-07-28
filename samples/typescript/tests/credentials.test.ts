/**
 * Tests for credentials chain loading
 * mock @alicloud/credentials，模拟其环境变量凭据提供者行为
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

vi.mock('@alicloud/credentials', () => ({
  default: class MockCredential {
    async getCredential() {
      return {
        accessKeyId: process.env.ALIBABA_CLOUD_ACCESS_KEY_ID,
        accessKeySecret: process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET,
      };
    }
  },
}));

import { loadCredentialsFromChain } from '../src/client/credentials.js';

describe('loadCredentialsFromChain', () => {
  const originalAk = process.env.ALIBABA_CLOUD_ACCESS_KEY_ID;
  const originalSk = process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET;

  beforeEach(() => {
    delete process.env.ALIBABA_CLOUD_ACCESS_KEY_ID;
    delete process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET;
  });

  afterEach(() => {
    if (originalAk === undefined) {
      delete process.env.ALIBABA_CLOUD_ACCESS_KEY_ID;
    } else {
      process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = originalAk;
    }
    if (originalSk === undefined) {
      delete process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET;
    } else {
      process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET = originalSk;
    }
  });

  it('should return credentials injected via process.env', async () => {
    process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = 'test-ak-id';
    process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET = 'test-ak-secret';

    const creds = await loadCredentialsFromChain();
    expect(creds.accessKeyId).toBe('test-ak-id');
    expect(creds.accessKeySecret).toBe('test-ak-secret');
  });

  it('should throw when the chain returns empty credentials', async () => {
    await expect(loadCredentialsFromChain()).rejects.toThrow('凭据链返回的凭证为空');
  });

  it('should throw when only accessKeyId is present', async () => {
    process.env.ALIBABA_CLOUD_ACCESS_KEY_ID = 'test-ak-id';

    await expect(loadCredentialsFromChain()).rejects.toThrow('凭据链返回的凭证为空');
  });
});
