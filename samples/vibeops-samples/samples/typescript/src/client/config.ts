/**
 * Configuration for STAROps SDK
 * STAROps SDK 配置
 */

import { config as dotenvConfig } from 'dotenv';
import { SDKException } from './errors.js';
import { loadCredentialsFromChain } from './credentials.js';
import type { RetryConfig } from './retry.js';

/** 应用配置 / Application configuration */
export interface Config {
  workspace: string;
  endpoint: string;
  accessKeyId: string;
  accessKeySecret: string;
  region: string;
  employeeName: string;
  retryConfig?: RetryConfig; // 重试配置，未设置时使用默认配置
  simulateNetworkError?: boolean; // 模拟网络断连，用于测试重试逻辑
}

/** 从环境变量加载配置 / Load configuration from environment variables */
export async function loadConfigFromEnv(): Promise<Config> {
  dotenvConfig();

  const workspace = process.env.VIBEOPS_WORKSPACE || '';
  const endpoint = process.env.VIBEOPS_ENDPOINT || '';
  const region = process.env.VIBEOPS_REGION || 'cn-hangzhou';
  let accessKeyId = process.env.ALIBABA_CLOUD_ACCESS_KEY_ID || '';
  let accessKeySecret = process.env.ALIBABA_CLOUD_ACCESS_KEY_SECRET || '';
  const employeeName = process.env.VIBEOPS_EMPLOYEE_NAME || 'default';

  // 环境变量为空时回退到阿里云默认凭据链
  if (!accessKeyId || !accessKeySecret) {
    try {
      const creds = await loadCredentialsFromChain();
      accessKeyId = creds.accessKeyId;
      accessKeySecret = creds.accessKeySecret;
    } catch (e) {
      console.error(`凭据链加载失败: ${e}，请手动设置环境变量`);
    }
  }

  // Validate required fields
  const missingVars: string[] = [];
  if (!endpoint) missingVars.push('VIBEOPS_ENDPOINT');
  if (!accessKeyId) missingVars.push('ALIBABA_CLOUD_ACCESS_KEY_ID');
  if (!accessKeySecret) missingVars.push('ALIBABA_CLOUD_ACCESS_KEY_SECRET');

  if (missingVars.length > 0) {
    throw SDKException.configMissing(missingVars);
  }

  return {
    workspace,
    endpoint,
    accessKeyId,
    accessKeySecret,
    region: region || 'cn-hangzhou',
    employeeName: employeeName || 'default',
  };
}
