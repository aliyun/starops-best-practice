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
