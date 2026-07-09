/**
 * Enumeration types for STAROps SDK
 * STAROps SDK 枚举类型
 */

/** 事件类型 / Event type */
export enum EventType {
  THREAD_TITLE_UPDATED = 'thread_title_updated',
  ERROR = 'error',
  THINKING = 'thinking',
  INTERACTIVE = 'interactive',
  INTERACTIVE_RESPONSE = 'interactive_response',
  TASK_FINISHED = 'task_finished',
  CANCEL = 'cancel',
}

/** 消息角色 / Message role */
export enum MessageRole {
  USER = 'user',
  ASSISTANT = 'assistant',
  SYSTEM = 'system',
}

/** 内容类型 / Content type */
export enum ContentType {
  TEXT = 'text',
  SPIN_TEXT = 'spin_text',
  IMAGE = 'image',
}

/** 执行状态 / Item status */
export enum ItemStatus {
  INIT = 'init',
  START = 'start',
  PROGRESS = 'progress',
  SUSPENDED = 'suspended',
  SUCCESS = 'success',
  FAIL = 'fail',
}

/** 交互类型 / Interaction type */
export enum InteractionType {
  USER_ACK = 'user_ack',
  USER_SELECT = 'user_select',
  USER_INPUT = 'user_input',
  SLS_QUERY = 'sls_query',
}
