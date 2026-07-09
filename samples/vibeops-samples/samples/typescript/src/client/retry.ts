/**
 * SSE retry logic for STAROps SDK
 * STAROps SDK SSE 重试逻辑
 *
 * 跨语言对齐 Go 参考实现（samples/golang/internal/client/retry.go）：
 * - 重试触发：所有非 stream_done 的中断均触发重连
 * - 唯一结束：stream_done 是正常结束的唯一标志
 * - 不区分错误：不判断错误类型，一律重试
 * - 去重机制：inDedupeWindow + lastTimestamp，重连后仅转发比上次更新的消息
 */

import * as $Starops20260428 from '@alicloud/starops20260428';
import type { ChatEvent } from './agent-client.js';

// ===================== 一、配置 =====================

/** 重试配置 / Retry config（时间单位均为毫秒） */
export interface RetryConfig {
  maxRetries: number; // 最大重试次数，默认10
  initialBackoff: number; // 初始退避时间(ms)，默认1000
  maxBackoff: number; // 最大退避时间(ms)，默认30000
  backoffFactor: number; // 退避系数，默认2.0
  idleTimeout: number; // 空闲超时(ms)：超过此时长未收到任何消息视为连接中断，默认60000
}

/** 返回默认重试配置 / Return default retry config */
export function defaultRetryConfig(): RetryConfig {
  return {
    maxRetries: 10,
    initialBackoff: 1000,
    maxBackoff: 30000,
    backoffFactor: 2.0,
    idleTimeout: 60000,
  };
}

/** 从环境变量加载重试配置 / Load retry config from env */
export function loadRetryConfigFromEnv(): RetryConfig {
  const cfg = defaultRetryConfig();

  const maxRetries = process.env.VIBEOPS_MAX_RETRIES;
  if (maxRetries) {
    const n = parseInt(maxRetries, 10);
    if (!Number.isNaN(n) && n > 0) {
      cfg.maxRetries = n;
    }
  }

  // 环境变量按秒配置，内部统一转换为毫秒
  const idleTimeout = process.env.VIBEOPS_IDLE_TIMEOUT;
  if (idleTimeout) {
    const n = parseInt(idleTimeout, 10);
    if (!Number.isNaN(n) && n > 0) {
      cfg.idleTimeout = n * 1000;
    }
  }

  return cfg;
}

// ===================== 二、状态定义 =====================

/** 聚合重连过程中的状态 / Retry state */
export interface RetryState {
  lastTimestamp: string; // 最后一条已转发消息的时间戳，用于重连去重
  inDedupeWindow: boolean; // true=重连后去重窗口，仅转发更新的消息
  retryCount: number; // 当前连续重试次数
}

/** 单次连接的结束原因 / Outcome of a single connection */
export enum ConnectionOutcome {
  DONE, // 收到 stream_done，正常结束
  INTERRUPTED, // 连接中断，需重连
  FATAL, // 不可恢复，直接结束
}

// ===================== 三、工具函数 =====================

/**
 * 判断 ts 是否比 base 更新
 * 优先数值比较（Unix 时间戳），无法解析时 fallback 为字符串比较
 */
export function isNewerTimestamp(ts: string, base: string): boolean {
  if (ts === '') {
    return false;
  }
  if (base === '') {
    return true;
  }
  const tsIsNum = /^-?\d+$/.test(ts);
  const baseIsNum = /^-?\d+$/.test(base);
  if (tsIsNum && baseIsNum) {
    return Number(ts) > Number(base);
  }
  if (!tsIsNum && baseIsNum) {
    return false; // 基准是数值但当前 ts 无法解析，视为不更新
  }
  return ts > base;
}

/**
 * 从事件中提取比 base 更新的最大消息 timestamp
 * 返回空字符串表示没有比 base 更新的时间戳
 */
export function extractNewestTimestamp(event: ChatEvent, base: string): string {
  const messages = event?.body?.messages as
    | Array<Record<string, unknown>>
    | undefined;
  if (!messages) {
    return '';
  }

  let newest = base;
  for (const msg of messages) {
    if (!msg) {
      continue;
    }
    const ts = (msg.timestamp as string) || '';
    if (isNewerTimestamp(ts, newest)) {
      newest = ts;
    }
  }
  if (newest === base) {
    return '';
  }
  return newest;
}

/**
 * 计算退避时间(ms)
 * min(initialBackoff * backoffFactor^(retryCount-1), maxBackoff)
 */
export function calculateBackoff(retryCount: number, config: RetryConfig): number {
  const backoff =
    config.initialBackoff * Math.pow(config.backoffFactor, retryCount - 1);
  if (backoff > config.maxBackoff) {
    return config.maxBackoff;
  }
  return backoff;
}

/**
 * 判断事件是否为 stream_done（正常结束标志）
 * 不检查 event.event 字段，只判断 messages[].events[].type == stream_done
 */
export function isStreamDoneEvent(event: ChatEvent): boolean {
  const messages = event?.body?.messages as
    | Array<Record<string, unknown>>
    | undefined;
  if (!messages) {
    return false;
  }
  for (const msg of messages) {
    if (!msg) {
      continue;
    }
    const events = msg.events as Array<Record<string, unknown>> | undefined;
    if (!events) {
      continue;
    }
    for (const evt of events) {
      if (evt && evt.type === 'stream_done') {
        return true;
      }
    }
  }
  return false;
}

/**
 * 构建重连请求
 * action="reconnect"，复制 threadId/digitalEmployeeName/variables，不携带 messages
 */
export function buildReconnectRequest(origReq: any): any {
  const variables: Record<string, unknown> = {};
  if (origReq?.variables) {
    for (const k of Object.keys(origReq.variables)) {
      variables[k] = origReq.variables[k];
    }
  }

  return new $Starops20260428.CreateChatRequest({
    action: 'reconnect',
    threadId: origReq?.threadId,
    digitalEmployeeName: origReq?.digitalEmployeeName,
    variables,
  });
}
