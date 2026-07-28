/**
 * Event and message types for STAROps SDK
 * STAROps SDK 事件和消息类型
 */

import { ContentType, EventType, ItemStatus, MessageRole } from './enums.js';

/** 消息内容 / Message content */
export interface ItemContent {
  type: ContentType;
  value: string;
  append?: boolean;
  lastChunk?: boolean;
}

/** 事件定义 / Event definition */
export interface ItemEvent {
  type: EventType;
  payload?: Record<string, unknown>;
}

/** 工具调用详情 / Tool call details */
export interface ItemTool {
  id: string;
  name: string;
  toolCallId: string;
  argumentsDelta?: string;
  arguments?: unknown;
  status: ItemStatus;
  contents?: ItemContent[];
}

/** 消息条目 / Message item */
export interface MessageItem {
  parentCallId: string;
  callId: string;
  role: MessageRole;
  timestamp: string;
  contents?: ItemContent[];
  tools?: ItemTool[];
  events?: ItemEvent[];
  artifacts?: Record<string, unknown>[];
}

/** 聊天事件 / Chat event */
export interface ChatEvent {
  id?: string;
  event?: string;
  body?: Record<string, unknown>;
  rawJson: string;
  statusCode: number;
  isDone: boolean;
  error?: Error;
}

/** 会话信息 / Thread information */
export interface ThreadInfo {
  threadId: string;
  title: string;
  status: string;
  createTime: string;
  updateTime: string;
}

/** 会话消息 / Thread message */
export interface ThreadMessage {
  role: string;
  content: string;
  timestamp: string;
}

/** 判断消息是否为结束消息 / Check if message body indicates done */
export function isDoneMessage(body?: Record<string, unknown>): boolean {
  // 优先使用 response 级别的 event 字段
  if (body?.event === 'done') return true;
  // fallback: 遍历 messages
  if (!body || !body.messages) return false;
  const messages = body.messages as Array<Record<string, unknown>>;
  return messages.some((msg) => msg.type === 'done');
}
